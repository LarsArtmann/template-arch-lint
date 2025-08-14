// Package middleware provides HTTP middleware for cross-cutting concerns.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"

	httputil "github.com/LarsArtmann/template-arch-lint/internal/application/http"
	"github.com/gin-gonic/gin"
)

// CorrelationIDConfig contains configuration for correlation ID middleware.
type CorrelationIDConfig struct {
	// HeaderName is the name of the HTTP header for correlation ID
	HeaderName string
	// Generator is a custom function to generate correlation IDs
	Generator func() string
	// Logger is used for middleware logging
	Logger *slog.Logger
}

// DefaultCorrelationIDConfig returns a default configuration.
func DefaultCorrelationIDConfig() CorrelationIDConfig {
	return CorrelationIDConfig{
		HeaderName: httputil.CorrelationIDKey,
		Generator:  generateCorrelationID,
		Logger:     slog.Default(),
	}
}

// CorrelationID returns a middleware that adds correlation IDs to requests.
func CorrelationID(config ...CorrelationIDConfig) gin.HandlerFunc {
	cfg := DefaultCorrelationIDConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// Check if correlation ID already exists in request headers
		correlationID := c.GetHeader(cfg.HeaderName)

		if correlationID == "" {
			// Generate new correlation ID if not provided
			correlationID = cfg.Generator()
			cfg.Logger.Debug("Generated new correlation ID",
				"correlation_id", correlationID,
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
		} else {
			cfg.Logger.Debug("Using existing correlation ID",
				"correlation_id", correlationID,
				"path", c.Request.URL.Path,
				"method", c.Request.Method)
		}

		// Store correlation ID in gin context for handlers to access
		c.Set(cfg.HeaderName, correlationID)

		// Set correlation ID in response header
		c.Header(cfg.HeaderName, correlationID)

		// Continue processing request
		c.Next()
	}
}

// generateCorrelationID generates a random hex correlation ID.
func generateCorrelationID() string {
	bytes := make([]byte, 16) // 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return generateTimestampID()
	}
	return hex.EncodeToString(bytes)
}

// generateTimestampID generates a timestamp-based correlation ID as fallback.
func generateTimestampID() string {
	// Simple fallback using nanosecond timestamp
	// In production, you might want to use a more sophisticated approach
	return hex.EncodeToString([]byte("fallback-" + hex.EncodeToString([]byte("timestamp"))))
}

// WithCorrelationID is a convenience function to create correlation ID middleware.
func WithCorrelationID(logger *slog.Logger) gin.HandlerFunc {
	return CorrelationID(CorrelationIDConfig{
		Logger: logger,
	})
}

// GetCorrelationIDFromContext retrieves correlation ID from gin context
// This is a convenience function that wraps the httputil function.
func GetCorrelationIDFromContext(c *gin.Context) string {
	return httputil.GetCorrelationID(c)
}
