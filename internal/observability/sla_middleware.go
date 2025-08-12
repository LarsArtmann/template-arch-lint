// Package observability provides SLA tracking middleware
package observability

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// SLAMiddleware creates a Gin middleware for SLA tracking
func SLAMiddleware(slaTracker *SLATracker, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request first
		c.Next()

		// Extract SLA data from context (set by Prometheus middleware)
		responseTime, exists := c.Get("sla_response_time")
		if !exists {
			logger.Debug("No SLA response time found in context")
			return
		}

		isSuccess, exists := c.Get("sla_success")
		if !exists {
			logger.Debug("No SLA success status found in context")
			return
		}

		endpoint, exists := c.Get("sla_endpoint")
		if !exists {
			logger.Debug("No SLA endpoint found in context")
			return
		}

		// Record request in SLA tracker
		if rt, ok := responseTime.(float64); ok {
			if success, ok := isSuccess.(bool); ok {
				if ep, ok := endpoint.(string); ok {
					slaTracker.RecordRequest(rt, success, ep)
				}
			}
		}
	}
}