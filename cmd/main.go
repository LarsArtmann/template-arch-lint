package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"charm.land/log/v2"
	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/larsartmann/httputil"
)

const (
	exitCodeSuccess   = 0
	exitCodeFailure   = 1
	defaultServerPort = 8080
)

const (
	defaultServerReadTimeout  = 15 * time.Second
	defaultServerWriteTimeout = 15 * time.Second
	defaultServerIdleTimeout  = 60 * time.Second
	defaultGracefulTimeout    = 30 * time.Second
)

func main() {
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Level:           log.InfoLevel,
	})

	logger.Info("🔥 Template-Arch-Lint - Pure Linting Template")
	logger.Info("✅ This demonstrates enterprise-grade Go architecture enforcement")

	userRepo := repositories.NewInMemoryUserRepository()
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", httputil.HealthHandler())
	userHandler.RegisterRoutes(mux)

	serverCfg := httputil.ServerConfig{
		Addr:         fmt.Sprintf(":%d", defaultServerPort),
		ReadTimeout:  defaultServerReadTimeout,
		WriteTimeout: defaultServerWriteTimeout,
		IdleTimeout:  defaultServerIdleTimeout,
	}

	server, err := httputil.NewServer(serverCfg, mux)
	if err != nil {
		logger.Error("❌ Failed to create HTTP server", "error", err)
		os.Exit(exitCodeFailure)
	}

	logger.Info("🚀 Starting HTTP server", "port", defaultServerPort)

	errChan := server.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("🛑 Shutting down server...")
	case err := <-errChan:
		logger.Error("❌ Server failed", "error", err)
		os.Exit(exitCodeFailure)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultGracefulTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		logger.Error("❌ Server forced to shutdown", "error", err)
		os.Exit(exitCodeFailure)
	}

	logger.Info("✅ Server shutdown complete")
	os.Exit(exitCodeSuccess)
}
