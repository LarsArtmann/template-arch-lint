package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
)

const (
	exitCodeSuccess    = 0
	exitCodeFailure    = 1
	defaultServerPort = 8080
)

func main() {
	// Initialize structured logger with enterprise configuration
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Level:           log.InfoLevel,
	})

	logger.Info("🔥 Template-Arch-Lint - Pure Linting Template")
	logger.Info("✅ This demonstrates enterprise-grade Go architecture enforcement")

	// Setup dependency injection
	userRepo := repositories.NewInMemoryUserRepository()
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// Setup HTTP server
	router := gin.Default()
	setupRoutes(router, userHandler)

	// Start server with graceful shutdown
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", defaultServerPort),
		Handler:  router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("🚀 Starting HTTP server", "port", defaultServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("❌ Server failed to start", "error", err)
			os.Exit(exitCodeFailure)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("❌ Server forced to shutdown", "error", err)
		os.Exit(exitCodeFailure)
	}

	logger.Info("✅ Server shutdown complete")
	os.Exit(exitCodeSuccess)
}

// setupRoutes configures HTTP routes.
func setupRoutes(router *gin.Engine, userHandler *handlers.UserHandler) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().UTC(),
		})
	})

	// User management endpoints
	userGroup := router.Group("/api/v1/users")
	{
		userGroup.POST("", userHandler.CreateUser)
		userGroup.GET("/:id", userHandler.GetUser)
		userGroup.PUT("/:id", userHandler.UpdateUser)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}
}