// Package observability provides health check functionality
package observability

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	config          *config.Config
	logger          *slog.Logger
	db              *sql.DB
	businessMetrics *BusinessMetrics
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cfg *config.Config, logger *slog.Logger, db *sql.DB, businessMetrics *BusinessMetrics) (*HealthChecker, error) {
	return &HealthChecker{
		config:          cfg,
		logger:          logger,
		db:              db,
		businessMetrics: businessMetrics,
	}, nil
}

// LivenessHandler provides liveness check
func (hc *HealthChecker) LivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
			"service": hc.config.App.Name,
		})
	}
}

// ReadinessHandler provides readiness check
func (hc *HealthChecker) ReadinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic database ping
		if hc.db != nil {
			if err := hc.db.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "not ready",
					"error": "database unavailable",
				})
				return
			}
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"service": hc.config.App.Name,
		})
	}
}

// HealthHandler provides comprehensive health check
func (hc *HealthChecker) HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "healthy"
		components := make(map[string]interface{})

		// Check database
		if hc.db != nil {
			if err := hc.db.Ping(); err != nil {
				components["database"] = gin.H{"status": "unhealthy", "error": err.Error()}
				status = "degraded"
			} else {
				components["database"] = gin.H{"status": "healthy"}
			}
		}

		// Get business metrics summary
		if hc.businessMetrics != nil {
			components["metrics"] = hc.businessMetrics.GetUserMetricsSummary(c.Request.Context())
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     status,
			"service":    hc.config.App.Name,
			"version":    hc.config.App.Version,
			"components": components,
		})
	}
}

// RegisterHealthRoutes registers health check routes
func (hc *HealthChecker) RegisterHealthRoutes(router *gin.Engine) {
	health := router.Group("/health")
	{
		health.GET("/live", hc.LivenessHandler())
		health.GET("/ready", hc.ReadinessHandler())
		health.GET("", hc.HealthHandler())
		health.GET("/", hc.HealthHandler())
	}

	// Version endpoint
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     hc.config.App.Name,
			"version":     hc.config.App.Version,
			"environment": hc.config.App.Environment,
		})
	})
}