// Package middleware provides structured request logging with correlation IDs.
package middleware

import (
	"log/slog"
	"time"

	httputil "github.com/LarsArtmann/template-arch-lint/internal/application/http"
	"github.com/gin-gonic/gin"
)

// RequestLoggingConfig contains configuration for request logging middleware
type RequestLoggingConfig struct {
	// Logger is the structured logger to use
	Logger *slog.Logger
	// SkipPaths contains paths to skip logging (e.g., health checks)
	SkipPaths []string
	// LogRequestBody whether to log request body (be careful with sensitive data)
	LogRequestBody bool
	// LogResponseBody whether to log response body (be careful with sensitive data)
	LogResponseBody bool
}

// DefaultRequestLoggingConfig returns a default configuration
func DefaultRequestLoggingConfig() RequestLoggingConfig {
	return RequestLoggingConfig{
		Logger: slog.Default(),
		SkipPaths: []string{
			"/health",
			"/ready",
			"/metrics",
		},
		LogRequestBody:  false, // Disabled by default for security
		LogResponseBody: false, // Disabled by default for security
	}
}

// StructuredRequestLogging returns middleware for structured request logging
func StructuredRequestLogging(config ...RequestLoggingConfig) gin.HandlerFunc {
	cfg := DefaultRequestLoggingConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// Skip logging for specified paths
		for _, skipPath := range cfg.SkipPaths {
			if c.Request.URL.Path == skipPath {
				c.Next()
				return
			}
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Get correlation ID for structured logging
		correlationID := httputil.GetCorrelationID(c)

		// Log request start
		logFields := []any{
			"method", c.Request.Method,
			"path", path,
			"correlation_id", correlationID,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		}

		if raw != "" {
			logFields = append(logFields, "query", raw)
		}

		cfg.Logger.Info("Request started", logFields...)

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		// Prepare response log fields
		responseFields := []any{
			"method", c.Request.Method,
			"path", path,
			"status_code", statusCode,
			"correlation_id", correlationID,
			"client_ip", c.ClientIP(),
			"duration_ms", duration.Milliseconds(),
			"response_size", c.Writer.Size(),
		}

		if raw != "" {
			responseFields = append(responseFields, "query", raw)
		}

		// Log based on status code
		if statusCode >= 500 {
			cfg.Logger.Error("Request completed with server error", responseFields...)
		} else if statusCode >= 400 {
			cfg.Logger.Warn("Request completed with client error", responseFields...)
		} else {
			cfg.Logger.Info("Request completed successfully", responseFields...)
		}
	}
}

// WithStructuredLogging is a convenience function to create structured logging middleware
func WithStructuredLogging(logger *slog.Logger, skipPaths ...string) gin.HandlerFunc {
	config := DefaultRequestLoggingConfig()
	config.Logger = logger
	if len(skipPaths) > 0 {
		config.SkipPaths = skipPaths
	}
	return StructuredRequestLogging(config)
}
