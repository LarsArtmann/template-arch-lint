// Main entry point for the template-arch-lint server
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/container"
)

const (
	// ChannelBufferSize represents the buffer size for error channels
	ChannelBufferSize = 1
	// ExitCodeFailure represents the exit code for application failure
	ExitCodeFailure = 1
	// ErrorConstant represents the repeated string literal for error logging
	ErrorConstant = "error"
	// NewlineConstant represents the newline character constant
	NewlineConstant = "\n"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	diContainer, err := setupContainer()
	if err != nil {
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Failed to register dependencies: %v%s",
			err,
			NewlineConstant,
		)
		cancel() // Call cancel before exit to avoid exitAfterDefer
		os.Exit(ExitCodeFailure)
	}
	defer func() {
		if shutdownErr := shutdownContainer(diContainer); shutdownErr != nil {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Error shutting down container: %v%s",
				shutdownErr,
				NewlineConstant,
			)
		}
	}()

	cfg, logger, router := getDependencies(diContainer)
	logServerStart(logger, cfg)

	server := createHTTPServer(cfg, router)
	serverErrors := startServer(server, logger)
	
	err = runServerWithGracefulShutdown(
		ctx,
		server,
		serverErrors,
		logger,
		cfg,
	)
	if err != nil {
		cancel() // Call cancel before exit to avoid exitAfterDefer
		os.Exit(ExitCodeFailure)
	}
}

// setupContainer creates and registers all dependencies
func setupContainer() (*container.Container, error) {
	diContainer := container.New()
	if err := diContainer.RegisterAll(); err != nil {
		return nil, fmt.Errorf("failed to register dependencies: %w", err)
	}
	return diContainer, nil
}

// shutdownContainer safely shuts down the DI container
func shutdownContainer(diContainer *container.Container) error {
	return diContainer.Shutdown()
}

// getDependencies extracts required dependencies from the container
func getDependencies(
	diContainer *container.Container,
) (*config.Config, *slog.Logger, *gin.Engine) {
	injector := diContainer.GetInjector()
	cfg := do.MustInvoke[*config.Config](injector)
	logger := do.MustInvoke[*slog.Logger](injector)
	router := do.MustInvoke[*gin.Engine](injector)
	return cfg, logger, router
}

// logServerStart logs the server startup information
func logServerStart(logger *slog.Logger, cfg *config.Config) {
	logger.Info("Starting server",
		"name", cfg.App.Name,
		"version", cfg.App.Version,
		"environment", cfg.App.Environment,
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)
}

// createHTTPServer creates and configures the HTTP server
func createHTTPServer(cfg *config.Config, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
}

// startServer starts the HTTP server in a goroutine
func startServer(server *http.Server, logger *slog.Logger) chan error {
	serverErrors := make(chan error, ChannelBufferSize)
	go func() {
		logger.Info("Server listening", "address", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()
	return serverErrors
}

// runServerWithGracefulShutdown handles server lifecycle and graceful shutdown
func runServerWithGracefulShutdown(
	ctx context.Context,
	server *http.Server,
	serverErrors chan error,
	logger *slog.Logger,
	cfg *config.Config,
) error {
	shutdown := make(chan os.Signal, ChannelBufferSize)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to start", ErrorConstant, err)
			return err
		}
	case sig := <-shutdown:
		logger.Info("Graceful shutdown initiated", "signal", sig.String())
		return performGracefulShutdown(ctx, server, logger, cfg)
	}
	return nil
}

// performGracefulShutdown handles the graceful shutdown process
func performGracefulShutdown(
	ctx context.Context,
	server *http.Server,
	logger *slog.Logger,
	cfg *config.Config,
) error {
	shutdownCtx, shutdownCancel := context.WithTimeout(
		ctx,
		cfg.Server.GracefulShutdownTimeout,
	)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown server gracefully", ErrorConstant, err)
		
		if closeErr := server.Close(); closeErr != nil {
			logger.Error("Failed to close server", ErrorConstant, closeErr)
		}
		return err
	}

	logger.Info("Server shutdown completed successfully")
	return nil
}

// init sets up initial configuration before main runs
func init() {
	// Set up basic logging before the DI container is ready
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}
