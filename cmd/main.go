package main

import (
	"context"
	"encoding/json"
	"errors"
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
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status": "healthy",
			"time":   time.Now().UTC(),
		})
	})
	userHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", defaultServerPort),
		Handler:      mux,
		ReadTimeout:  defaultServerReadTimeout,
		WriteTimeout: defaultServerWriteTimeout,
		IdleTimeout:  defaultServerIdleTimeout,
	}

	go func() {
		logger.Info("🚀 Starting HTTP server", "port", defaultServerPort)

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("❌ Server failed to start", "error", err)
			os.Exit(exitCodeFailure)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), defaultGracefulTimeout)

	err := server.Shutdown(ctx)
	if err != nil {
		logger.Error("❌ Server forced to shutdown", "error", err)
		cancel()
		os.Exit(exitCodeFailure)
	}

	logger.Info("✅ Server shutdown complete")
	cancel()
	os.Exit(exitCodeSuccess)
}
