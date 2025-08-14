// Package handlers provides health check endpoints for production readiness.
package handlers

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	httputil "github.com/LarsArtmann/template-arch-lint/internal/application/http"
	"github.com/gin-gonic/gin"
)

const (
	// HealthStatus constants.
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusReady     = "ready"
	StatusNotReady  = "not_ready"
	StatusAlive     = "alive"
)

// HealthHandler provides health check endpoints.
type HealthHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db *sql.DB, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status      string           `json:"status"`
	Timestamp   time.Time        `json:"timestamp"`
	Version     string           `json:"version,omitempty"`
	Environment string           `json:"environment,omitempty"`
	Uptime      string           `json:"uptime,omitempty"`
	Checks      map[string]Check `json:"checks,omitempty"`
}

// Check represents an individual health check.
type Check struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Duration  string    `json:"duration,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// Health returns the overall health status of the application.
func (h *HealthHandler) Health(c *gin.Context) {
	correlationID := httputil.GetCorrelationID(c)
	h.logger.Info("Health check requested", "correlation_id", correlationID)

	response := HealthResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
		Version:   "1.0.0", // This could be injected from build info
		Checks:    make(map[string]Check),
	}

	// Perform database health check
	dbCheck := h.checkDatabase()
	response.Checks["database"] = dbCheck

	// Determine overall status based on individual checks
	overallStatus := StatusHealthy
	for _, check := range response.Checks {
		if check.Status != StatusHealthy {
			overallStatus = StatusUnhealthy
			break
		}
	}
	response.Status = overallStatus

	if overallStatus == StatusHealthy {
		httputil.RespondOK(c, response, "Service is healthy")
	} else {
		h.logger.Warn("Health check failed",
			"status", overallStatus,
			"correlation_id", correlationID)
		httputil.RespondError(c, 503,
			"SERVICE_UNHEALTHY",
			"Service is unhealthy",
			"health_check",
			nil)
	}
}

// Ready returns the readiness status of the application.
func (h *HealthHandler) Ready(c *gin.Context) {
	correlationID := httputil.GetCorrelationID(c)
	h.logger.Info("Readiness check requested", "correlation_id", correlationID)

	response := HealthResponse{
		Status:    StatusReady,
		Timestamp: time.Now(),
		Checks:    make(map[string]Check),
	}

	// Perform readiness checks (similar to health but more strict)
	dbCheck := h.checkDatabase()
	response.Checks["database"] = dbCheck

	// Add more readiness checks here (e.g., external dependencies)

	// Determine overall readiness
	overallStatus := StatusReady
	for _, check := range response.Checks {
		if check.Status != StatusHealthy {
			overallStatus = StatusNotReady
			break
		}
	}
	response.Status = overallStatus

	if overallStatus == StatusReady {
		httputil.RespondOK(c, response, "Service is ready")
	} else {
		h.logger.Warn("Readiness check failed",
			"status", overallStatus,
			"correlation_id", correlationID)
		httputil.RespondError(c, 503,
			"SERVICE_NOT_READY",
			"Service is not ready",
			"readiness_check",
			nil)
	}
}

// Live is a simple liveness check (always returns OK if the service is running).
func (h *HealthHandler) Live(c *gin.Context) {
	correlationID := httputil.GetCorrelationID(c)

	response := HealthResponse{
		Status:    StatusAlive,
		Timestamp: time.Now(),
	}

	httputil.RespondOK(c, response, "Service is alive")

	// Use debug level for liveness to avoid log spam
	h.logger.Debug("Liveness check completed", "correlation_id", correlationID)
}

// checkDatabase performs a database health check.
func (h *HealthHandler) checkDatabase() Check {
	start := time.Now()

	check := Check{
		Status:    StatusHealthy,
		Timestamp: time.Now(),
	}

	if h.db == nil {
		check.Status = StatusUnhealthy
		check.Error = "Database connection is nil"
		check.Duration = time.Since(start).String()
		return check
	}

	// Use a short timeout for health checks
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Simple ping to check database connectivity
	if err := h.db.PingContext(ctx); err != nil {
		check.Status = StatusUnhealthy
		check.Error = err.Error()
		h.logger.Error("Database health check failed", "error", err)
	}

	check.Duration = time.Since(start).String()
	return check
}

// RegisterHealthRoutes registers health check routes.
func RegisterHealthRoutes(router gin.IRouter, handler *HealthHandler) {
	router.GET("/health", handler.Health)
	router.GET("/ready", handler.Ready)
	router.GET("/live", handler.Live)

	// Alternative endpoints for compatibility
	router.GET("/healthz", handler.Health)
	router.GET("/readyz", handler.Ready)
}
