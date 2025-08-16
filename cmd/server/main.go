// Main entry point for the template-arch-lint server
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	cfg, logger, router, db := getEnhancedDependencies(diContainer)
	logServerStart(logger, cfg)

	server := createHTTPServer(cfg, router)
	serverErrors := startServer(server, logger)

	err = runServerWithGracefulShutdown(
		ctx,
		server,
		serverErrors,
		logger,
		cfg,
		db,
		diContainer,
	)
	
	// Always attempt graceful cleanup
	cleanupErr := performGracefulCleanup(logger, db, diContainer)
	if cleanupErr != nil {
		logger.Error("Cleanup errors during shutdown", ErrorConstant, cleanupErr)
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

// getEnhancedDependencies extracts required dependencies including database from the container.
func getEnhancedDependencies(
	diContainer *container.Container,
) (*config.Config, *slog.Logger, *gin.Engine, *sql.DB) {
	injector := diContainer.GetInjector()
	cfg := do.MustInvoke[*config.Config](injector)
	logger := do.MustInvoke[*slog.Logger](injector)
	router := do.MustInvoke[*gin.Engine](injector)
	db := do.MustInvoke[*sql.DB](injector)
	return cfg, logger, router, db
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

// runServerWithGracefulShutdown handles server lifecycle and graceful shutdown.
func runServerWithGracefulShutdown(
	ctx context.Context,
	server *http.Server,
	serverErrors chan error,
	logger *slog.Logger,
	cfg *config.Config,
	db *sql.DB,
	diContainer *container.Container,
) error {
	shutdown := make(chan os.Signal, ChannelBufferSize)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to start", ErrorConstant, err)
			return err
		}
		logger.Info("Server stopped normally")
		return nil
	case sig := <-shutdown:
		logger.Info("Graceful shutdown initiated", 
			"signal", sig.String(),
			"timeout", cfg.Server.GracefulShutdownTimeout.String())
		return performEnhancedGracefulShutdown(ctx, server, logger, cfg, db, diContainer)
	}
}

// performEnhancedGracefulShutdown handles the comprehensive graceful shutdown process.
func performEnhancedGracefulShutdown(
	ctx context.Context,
	server *http.Server,
	logger *slog.Logger,
	cfg *config.Config,
	db *sql.DB,
	diContainer *container.Container,
) error {
	// Use a shorter timeout for individual operations, with overall max timeout
	shutdownTimeout := cfg.Server.GracefulShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = DefaultShutdownTimeout
	}
	
	// Don't exceed maximum shutdown wait time
	if shutdownTimeout > MaxShutdownWaitTime {
		shutdownTimeout = MaxShutdownWaitTime
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	// Track shutdown progress
	var shutdownErrors []error
	var wg sync.WaitGroup

	logger.Info("Starting graceful shutdown sequence",
		"timeout", shutdownTimeout.String(),
		"steps", "http_server,database,container")

	// Step 1: Shutdown HTTP server (drain connections)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Shutting down HTTP server...")
		
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Failed to shutdown HTTP server gracefully", ErrorConstant, err)
			shutdownErrors = append(shutdownErrors, fmt.Errorf("HTTP server shutdown: %w", err))
			
			// Force close if graceful shutdown fails
			if closeErr := server.Close(); closeErr != nil {
				logger.Error("Failed to force close HTTP server", ErrorConstant, closeErr)
				shutdownErrors = append(shutdownErrors, fmt.Errorf("HTTP server force close: %w", closeErr))
			}
		} else {
			logger.Info("HTTP server shutdown completed")
		}
	}()

	// Step 2: Close database connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Closing database connections...")
		
		dbCtx, dbCancel := context.WithTimeout(shutdownCtx, DatabaseCloseTimeout)
		defer dbCancel()
		
		// Wait a bit for ongoing transactions to complete
		select {
		case <-time.After(2 * time.Second):
		case <-dbCtx.Done():
		}
		
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connections", ErrorConstant, err)
			shutdownErrors = append(shutdownErrors, fmt.Errorf("database close: %w", err))
		} else {
			logger.Info("Database connections closed successfully")
		}
	}()

	// Wait for HTTP server and database to shutdown
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Primary shutdown sequence completed")
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout reached, proceeding with forced cleanup")
		shutdownErrors = append(shutdownErrors, fmt.Errorf("shutdown timeout exceeded: %w", shutdownCtx.Err()))
	}

	// Step 3: Shutdown DI container (always do this last)
	logger.Info("Shutting down dependency injection container...")
	if err := diContainer.Shutdown(); err != nil {
		logger.Error("Failed to shutdown DI container", ErrorConstant, err)
		shutdownErrors = append(shutdownErrors, fmt.Errorf("container shutdown: %w", err))
	} else {
		logger.Info("DI container shutdown completed")
	}

	// Report final shutdown status
	if len(shutdownErrors) > 0 {
		logger.Error("Graceful shutdown completed with errors", 
			"error_count", len(shutdownErrors))
		return fmt.Errorf("shutdown errors: %v", shutdownErrors)
	}

	logger.Info("Graceful shutdown completed successfully")
	return nil
}

// performGracefulCleanup performs final cleanup operations.
func performGracefulCleanup(
	logger *slog.Logger,
	db *sql.DB,
	diContainer *container.Container,
) error {
	var cleanupErrors []error

	// Ensure database is closed
	if db != nil {
		if err := db.Close(); err != nil {
			logger.Warn("Database already closed or error closing", ErrorConstant, err)
			cleanupErrors = append(cleanupErrors, err)
		}
	}

	// Ensure container is shutdown
	if diContainer != nil {
		if err := diContainer.Shutdown(); err != nil {
			logger.Warn("Container already shutdown or error shutting down", ErrorConstant, err)
			cleanupErrors = append(cleanupErrors, err)
		}
	}

	if len(cleanupErrors) > 0 {
		return fmt.Errorf("cleanup errors: %v", cleanupErrors)
	}

	return nil
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

// init sets up initial configuration before main runs.
func init() {
	// Set up basic logging before the DI container is ready
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}
