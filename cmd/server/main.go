// Main entry point for the template-arch-lint server
package main

import (
	"context"
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

func main() {
	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create and initialize DI container
	diContainer := container.New()
	defer func() {
		if err := diContainer.Shutdown(); err != nil {
			fmt.Printf("Error shutting down container: %v\n", err)
		}
	}()

	// Register all dependencies
	if err := diContainer.RegisterAll(); err != nil {
		fmt.Printf("Failed to register dependencies: %v\n", err)
		os.Exit(1)
	}

	// Get dependencies from container
	injector := diContainer.GetInjector()
	cfg := do.MustInvoke[*config.Config](injector)
	logger := do.MustInvoke[*slog.Logger](injector)
	router := do.MustInvoke[*gin.Engine](injector)

	logger.Info("Starting server",
		"name", cfg.App.Name,
		"version", cfg.App.Version,
		"environment", cfg.App.Environment,
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("Server listening", "address", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	// Setup signal handling for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	case sig := <-shutdown:
		logger.Info("Graceful shutdown initiated", "signal", sig.String())

		// Create context with timeout for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, cfg.Server.GracefulShutdownTimeout)
		defer shutdownCancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown server gracefully", "error", err)

			// Force close after timeout
			if err := server.Close(); err != nil {
				logger.Error("Failed to close server", "error", err)
			}
			os.Exit(1)
		}

		logger.Info("Server shutdown completed successfully")
	}
}

// init sets up initial configuration before main runs
func init() {
	// Set up basic logging before the DI container is ready
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}
