// Package middleware provides HTTP logging middleware with structured logging
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// StructuredLoggingMiddleware provides structured logging for HTTP requests.
func StructuredLoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log with structured fields instead of returning a formatted string
		logger.Info("HTTP Request",
			"timestamp", param.TimeStamp.Format(time.RFC3339),
			"method", param.Method,
			"path", param.Path,
			"status_code", param.StatusCode,
			"latency", param.Latency.String(),
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
			"response_size", param.BodySize,
			"error_message", param.ErrorMessage,
		)
		return "" // Return empty string since we're doing structured logging
	})
}

// RequestLoggingMiddleware logs detailed request information.
func RequestLoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log level based on status code
		statusCode := c.Writer.Status()
		logLevel := slog.LevelInfo
		if statusCode >= 400 && statusCode < 500 {
			logLevel = slog.LevelWarn
		} else if statusCode >= 500 {
			logLevel = slog.LevelError
		}

		// Build log attributes
		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status_code", statusCode),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
			slog.Int("response_size", c.Writer.Size()),
		}

		// Add query parameters if present
		if c.Request.URL.RawQuery != "" {
			attrs = append(attrs, slog.String("query", c.Request.URL.RawQuery))
		}

		// Add errors if present
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.String()))
		}

		// Log the request with appropriate level
		logger.LogAttrs(c.Request.Context(), logLevel, "HTTP Request Completed", attrs...)
	}
}
