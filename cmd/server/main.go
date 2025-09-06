// Main entry point for the template-arch-lint server
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/container"
)

const (
	// ChannelBufferSize represents the buffer size for error channels.
	ChannelBufferSize = 1
	// ExitCodeFailure represents the exit code for application failure.
	ExitCodeFailure = 1
	// ErrorConstant represents the repeated string literal for error logging.
	ErrorConstant = "error"
	// NewlineConstant represents the newline character constant.
	NewlineConstant = "\n"
	// ErrorShuttingDownContainer represents the repeated error message for
	// container shutdown.
	ErrorShuttingDownContainer = "Error shutting down container"
	// HealthCheckFlag represents the health check command line flag.
	HealthCheckFlag = "health-check"
	// HealthCheckFlagDescription represents the health check flag description.
	HealthCheckFlagDescription = "Perform health check and exit"
	// VersionKey represents the version key for logging.
	VersionKey = "version"
	// ServiceKey represents the service key for logging.
	ServiceKey = "service"
	// HealthCheckWarningMsg represents health check warning message.
	HealthCheckWarningMsg = "Health check warning: unable to shutdown " +
		"container cleanly"
	// HealthCheckPassedMsg represents health check success message.
	HealthCheckPassedMsg = "Health check passed"
	// DefaultShutdownTimeout represents the default shutdown timeout.
	DefaultShutdownTimeout = 30 * time.Second
	// DatabaseCloseTimeout represents the database connection close timeout.
	DatabaseCloseTimeout = 5 * time.Second
	// MaxShutdownWaitTime represents the maximum time to wait for graceful shutdown.
	MaxShutdownWaitTime = 60 * time.Second
)

func main() {
	// Parse command line flags
	healthCheck := flag.Bool(HealthCheckFlag, false, HealthCheckFlagDescription)
	flag.Parse()

	// Handle health check flag
	if *healthCheck {
		if err := performHealthCheck(); err != nil {
			slog.Error("Health check failed", ErrorConstant, err)
			os.Exit(ExitCodeFailure)
		}
		return
	}

	// Run the main server
	if err := runServer(); err != nil {
		os.Exit(ExitCodeFailure)
	}
}

// runServer initializes and runs the main server.
func runServer() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	diContainer, err := setupContainer()
	if err != nil {
		slog.Error("Failed to register dependencies", ErrorConstant, err)
		return err
	}

	cfg, logger, router := getEnhancedDependencies(diContainer)
	logServerStart(logger, cfg)

	server := createHTTPServer(cfg, router)
	serverErrors := startServer(server, logger)

	// Simple server lifecycle for linting template
	select {
	case err := <-serverErrors:
		logger.Error("Server startup failed", ErrorConstant, err)
		return err
	case <-ctx.Done():
		logger.Info("Received shutdown signal, stopping server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server shutdown failed", ErrorConstant, err)
			return err
		}

		logger.Info("Server stopped gracefully")
	}

	return err
}

// setupContainer creates and registers all dependencies.
func setupContainer() (*container.Container, error) {
	diContainer := container.New()
	if err := diContainer.RegisterAll(); err != nil {
		return nil, fmt.Errorf("failed to register dependencies: %w", err)
	}
	return diContainer, nil
}

// shutdownContainer safely shuts down the DI container.
func shutdownContainer(diContainer *container.Container) error {
	if err := diContainer.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown DI container: %w", err)
	}
	return nil
}

// getDependencies extracts required dependencies from the container.
func getDependencies(
	diContainer *container.Container,
) (*config.Config, *slog.Logger, *gin.Engine) {
	injector := diContainer.GetInjector()
	cfg := do.MustInvoke[*config.Config](injector)
	logger := do.MustInvoke[*slog.Logger](injector)
	router := do.MustInvoke[*gin.Engine](injector)
	return cfg, logger, router
}

// getEnhancedDependencies extracts required dependencies from the container.
func getEnhancedDependencies(
	diContainer *container.Container,
) (*config.Config, *slog.Logger, *gin.Engine) {
	injector := diContainer.GetInjector()
	cfg := do.MustInvoke[*config.Config](injector)
	logger := do.MustInvoke[*slog.Logger](injector)
	router := do.MustInvoke[*gin.Engine](injector)
	return cfg, logger, router
}

// logServerStart logs the server startup information.
func logServerStart(logger *slog.Logger, cfg *config.Config) {
	logger.Info("Starting server",
		"name", cfg.App.Name,
		"version", cfg.App.Version,
		"environment", cfg.App.Environment,
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)
}

// createHTTPServer creates and configures the HTTP server.
func createHTTPServer(cfg *config.Config, router *gin.Engine) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
}

// startServer starts the HTTP server in a goroutine.
func startServer(server *http.Server, logger *slog.Logger) chan error {
	serverErrors := make(chan error, ChannelBufferSize)
	go func() {
		logger.Info("Server listening", "address", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()
	return serverErrors
}

// performHealthCheck performs a simple health check for Docker health checks.
func performHealthCheck() error {
	// For Docker health checks, we verify basic application health
	// This includes config loading and basic dependency validation

	// Try to load config
	cfg, err := config.LoadConfig("")
	if err != nil {
		return fmt.Errorf("unable to load config: %w", err)
	}

	// Validate that we can create a basic container (dependency injection)
	diContainer := container.New()
	if err := diContainer.RegisterAll(); err != nil {
		return fmt.Errorf("unable to register dependencies: %w", err)
	}

	// Clean up
	if err := diContainer.Shutdown(); err != nil {
		slog.Error(HealthCheckWarningMsg, ErrorConstant, err)
	}

	slog.Info(HealthCheckPassedMsg,
		ServiceKey, cfg.App.Name,
		VersionKey, cfg.App.Version)
	return nil
}
