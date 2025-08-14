// Package main demonstrates Clean Architecture with linting enforcement.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/LarsArtmann/template-arch-lint-example/internal/application/handlers"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint-example/internal/infrastructure/persistence"
)

const (
	defaultPort       = "8090"
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 30 * time.Second
)

func main() {
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Setup dependencies (simple manual DI for example)
	productRepo := persistence.NewMemoryProductRepository()
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// Setup HTTP server
	router := setupRouter(productHandler)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// Start server in background
	go func() {
		fmt.Printf("🚀 Example API server starting on port %s\n", port)
		fmt.Printf("📋 Available endpoints:\n")
		fmt.Printf("  GET  /products     - List all products\n")
		fmt.Printf("  POST /products     - Create new product\n")
		fmt.Printf("  GET  /health       - Health check\n")
		fmt.Printf("  GET  /             - Welcome message\n")
		fmt.Printf("\n🔗 Visit: http://localhost:%s\n\n", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("\n🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("✅ Server shutdown completed")
}

func setupRouter(productHandler *handlers.ProductHandler) *gin.Engine {
	// Set gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Routes
	router.GET("/", welcomeHandler)
	router.GET("/health", healthHandler)

	// Product routes
	products := router.Group("/products")
	{
		products.GET("", productHandler.ListProducts)
		products.POST("", productHandler.CreateProduct)
	}

	return router
}

func welcomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "🎯 Template Architecture Lint - Example API",
		"status":  "running",
		"version": "1.0.0",
		"docs":    "See README.md for usage examples",
		"endpoints": gin.H{
			"GET /products":  "List all products",
			"POST /products": "Create new product (JSON: {\"name\": \"...\", \"price\": 1000})",
			"GET /health":    "Health check",
		},
	})
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    "running",
	})
}
