// Package handlers provides monitoring and SLA endpoints
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"log/slog"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/observability"
)

// MonitoringHandler handles monitoring and SLA related endpoints
type MonitoringHandler struct {
	config            *config.Config
	logger            *slog.Logger
	prometheusMetrics *observability.PrometheusMetrics
	slaTracker        *observability.SLATracker
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(
	cfg *config.Config,
	logger *slog.Logger,
	prometheusMetrics *observability.PrometheusMetrics,
	slaTracker *observability.SLATracker,
) *MonitoringHandler {
	return &MonitoringHandler{
		config:            cfg,
		logger:            logger,
		prometheusMetrics: prometheusMetrics,
		slaTracker:        slaTracker,
	}
}

// GetSLAStatus returns current SLA status for all tiers
func (h *MonitoringHandler) GetSLAStatus(c *gin.Context) {
	slaMetrics := h.slaTracker.GetAllSLAMetrics()
	slaSummary := h.slaTracker.GetSLASummary()
	
	response := map[string]interface{}{
		"status":     "success",
		"timestamp":  time.Now().Format(time.RFC3339),
		"sla_tiers":  slaMetrics,
		"summary":    slaSummary,
		"critical":   h.slaTracker.IsErrorBudgetCritical(),
	}
	
	c.JSON(http.StatusOK, response)
}

// GetSLAStatusForTier returns SLA status for a specific tier
func (h *MonitoringHandler) GetSLAStatusForTier(c *gin.Context) {
	tierParam := c.Param("tier")
	
	var tier observability.SLATier
	switch tierParam {
	case "gold":
		tier = observability.SLATierGold
	case "silver":
		tier = observability.SLATierSilver
	case "bronze":
		tier = observability.SLATierBronze
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid SLA tier. Must be one of: gold, silver, bronze",
		})
		return
	}
	
	metrics, exists := h.slaTracker.GetSLAMetrics(tier)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "SLA tier not found",
		})
		return
	}
	
	response := map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now().Format(time.RFC3339),
		"tier":      string(tier),
		"metrics":   metrics,
	}
	
	c.JSON(http.StatusOK, response)
}

// SimulateSlowResponse creates an artificial slow response for testing
func (h *MonitoringHandler) SimulateSlowResponse(c *gin.Context) {
	// Get delay from query parameter (default 2 seconds)
	delayParam := c.DefaultQuery("delay", "2000")
	
	var delayMs int
	if _, err := fmt.Sscanf(delayParam, "%d", &delayMs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid delay parameter. Must be a number in milliseconds.",
		})
		return
	}
	
	// Limit delay to reasonable bounds (max 10 seconds)
	if delayMs < 0 || delayMs > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Delay must be between 0 and 10000 milliseconds.",
		})
		return
	}
	
	h.logger.Info("Simulating slow response", "delay_ms", delayMs, "client_ip", c.ClientIP())
	
	// Sleep for the specified duration
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	
	c.JSON(http.StatusOK, gin.H{
		"message":   "Slow response simulation completed",
		"delay_ms":  delayMs,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// SimulateError creates an artificial error for testing error rate alerts
func (h *MonitoringHandler) SimulateError(c *gin.Context) {
	// Get error type from query parameter
	errorType := c.DefaultQuery("type", "500")
	
	h.logger.Info("Simulating error response", "error_type", errorType, "client_ip", c.ClientIP())
	
	switch errorType {
	case "400":
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Simulated bad request error",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case "401":
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":     "Simulated unauthorized error",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case "403":
		c.JSON(http.StatusForbidden, gin.H{
			"error":     "Simulated forbidden error",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case "404":
		c.JSON(http.StatusNotFound, gin.H{
			"error":     "Simulated not found error",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case "500":
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Simulated internal server error",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case "502":
		c.JSON(http.StatusBadGateway, gin.H{
			"error":     "Simulated bad gateway error",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	case "503":
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":     "Simulated service unavailable error",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid error type. Supported: 400, 401, 403, 404, 500, 502, 503",
		})
	}
}

// TriggerAlert manually triggers specific alert conditions for testing
func (h *MonitoringHandler) TriggerAlert(c *gin.Context) {
	alertType := c.Param("type")
	
	h.logger.Info("Manual alert trigger requested", "alert_type", alertType, "client_ip", c.ClientIP())
	
	switch alertType {
	case "high-error-rate":
		// Simulate multiple errors in quick succession
		for i := 0; i < 10; i++ {
			h.prometheusMetrics.RecordUserCreated("failed", "test_trigger")
			time.Sleep(100 * time.Millisecond)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "High error rate simulation triggered",
			"details": "Generated 10 failed user creation events",
		})
		
	case "user-validation-failure":
		// Simulate validation failures
		for i := 0; i < 5; i++ {
			h.prometheusMetrics.RecordUserValidation("email", "failed", 50*time.Millisecond)
			h.prometheusMetrics.RecordUserValidation("name", "failed", 30*time.Millisecond)
			time.Sleep(200 * time.Millisecond)
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "User validation failure simulation triggered",
			"details": "Generated multiple validation failures",
		})
		
	case "feature-flag-checks":
		// Simulate feature flag usage
		flags := []string{"new_ui", "beta_features", "advanced_search", "premium_features"}
		results := []string{"enabled", "disabled"}
		segments := []string{"premium", "standard", "trial"}
		
		for i := 0; i < 20; i++ {
			flag := flags[i%len(flags)]
			result := results[i%len(results)]
			segment := segments[i%len(segments)]
			
			h.prometheusMetrics.RecordFeatureFlagCheck(flag, result, segment)
			time.Sleep(50 * time.Millisecond)
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Feature flag usage simulation triggered",
			"details": "Generated feature flag check events",
		})
		
	case "config-reload":
		// Simulate configuration reloads
		sources := []string{"file", "api", "webhook"}
		statuses := []string{"success", "failed"}
		
		for i := 0; i < 3; i++ {
			source := sources[i%len(sources)]
			status := statuses[i%len(statuses)]
			
			h.prometheusMetrics.RecordConfigReload(status, source)
			time.Sleep(500 * time.Millisecond)
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Configuration reload simulation triggered",
			"details": "Generated config reload events",
		})
		
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid alert type",
			"valid_types": []string{
				"high-error-rate",
				"user-validation-failure", 
				"feature-flag-checks",
				"config-reload",
			},
		})
	}
}

// GetHealthDetails returns detailed health information including SLA status
func (h *MonitoringHandler) GetHealthDetails(c *gin.Context) {
	slaMetrics := h.slaTracker.GetAllSLAMetrics()
	slaSummary := h.slaTracker.GetSLASummary()
	
	// Determine overall health status
	overallHealth := "healthy"
	criticalBudgets := h.slaTracker.IsErrorBudgetCritical()
	
	for tier, isCritical := range criticalBudgets {
		if isCritical {
			h.logger.Warn("Critical error budget detected", "tier", tier)
			overallHealth = "degraded"
			break
		}
	}
	
	response := map[string]interface{}{
		"status":    overallHealth,
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   h.config.App.Version,
		"environment": h.config.App.Environment,
		"sla": map[string]interface{}{
			"metrics": slaMetrics,
			"summary": slaSummary,
			"critical_budgets": criticalBudgets,
		},
		"features": map[string]interface{}{
			"prometheus_enabled": h.config.Observability.Exporters.Prometheus.Enabled,
			"tracing_enabled":   h.config.Observability.Tracing.Enabled,
			"metrics_enabled":   h.config.Observability.Metrics.Enabled,
		},
		"endpoints": map[string]string{
			"metrics":    fmt.Sprintf(":%d%s", h.config.Observability.Exporters.Prometheus.Port, h.config.Observability.Exporters.Prometheus.Path),
			"health":     "/health",
			"sla_status": "/monitoring/sla",
		},
	}
	
	var statusCode int
	if overallHealth == "healthy" {
		statusCode = http.StatusOK
	} else {
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, response)
}