// Package observability provides SLA/SLI tracking and error budget monitoring
package observability

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
)

// SLATier represents different service level agreement tiers
type SLATier string

const (
	SLATierGold   SLATier = "gold"
	SLATierSilver SLATier = "silver"
	SLATierBronze SLATier = "bronze"
)

// SLAConfiguration defines SLA targets for each tier
type SLAConfiguration struct {
	Tier                 SLATier
	AvailabilityTarget   float64 // e.g., 0.995 for 99.5%
	ResponseTimeTarget   float64 // in seconds
	ErrorBudgetPeriod    time.Duration
	AlertThreshold       float64 // error budget threshold for alerting
	CriticalThreshold    float64 // critical error budget threshold
}

// SLAMetrics tracks current SLA performance
type SLAMetrics struct {
	Tier                 SLATier
	CurrentAvailability  float64
	CurrentResponseTime  float64
	ErrorBudgetRemaining float64
	ErrorBudgetBurnRate  float64
	LastUpdate          time.Time
	WindowStart         time.Time
	WindowEnd           time.Time
	TotalRequests       int64
	SuccessfulRequests  int64
	FailedRequests      int64
	ResponseTimes       []float64
}

// SLATracker manages SLA/SLI monitoring and error budget tracking
type SLATracker struct {
	config            *config.Config
	logger            *slog.Logger
	prometheusMetrics *PrometheusMetrics
	configurations    map[SLATier]SLAConfiguration
	metrics          map[SLATier]*SLAMetrics
	mu               sync.RWMutex
	updateInterval    time.Duration
	ticker           *time.Ticker
	stopChan         chan struct{}
}

// NewSLATracker creates a new SLA tracking service
func NewSLATracker(cfg *config.Config, logger *slog.Logger, prometheusMetrics *PrometheusMetrics) *SLATracker {
	tracker := &SLATracker{
		config:            cfg,
		logger:            logger,
		prometheusMetrics: prometheusMetrics,
		configurations:    make(map[SLATier]SLAConfiguration),
		metrics:          make(map[SLATier]*SLAMetrics),
		updateInterval:    30 * time.Second, // Update SLA metrics every 30 seconds
		stopChan:         make(chan struct{}),
	}

	tracker.initializeConfigurations()
	tracker.initializeMetrics()
	
	return tracker
}

// initializeConfigurations sets up default SLA configurations
func (s *SLATracker) initializeConfigurations() {
	s.configurations[SLATierGold] = SLAConfiguration{
		Tier:                 SLATierGold,
		AvailabilityTarget:   0.995,  // 99.5%
		ResponseTimeTarget:   0.2,    // 200ms
		ErrorBudgetPeriod:    30 * 24 * time.Hour, // 30 days
		AlertThreshold:       0.2,    // Alert when 20% of error budget remains
		CriticalThreshold:    0.1,    // Critical when 10% of error budget remains
	}

	s.configurations[SLATierSilver] = SLAConfiguration{
		Tier:                 SLATierSilver,
		AvailabilityTarget:   0.99,   // 99.0%
		ResponseTimeTarget:   1.0,    // 1000ms
		ErrorBudgetPeriod:    30 * 24 * time.Hour, // 30 days
		AlertThreshold:       0.2,
		CriticalThreshold:    0.1,
	}

	s.configurations[SLATierBronze] = SLAConfiguration{
		Tier:                 SLATierBronze,
		AvailabilityTarget:   0.98,   // 98.0%
		ResponseTimeTarget:   2.0,    // 2000ms
		ErrorBudgetPeriod:    30 * 24 * time.Hour, // 30 days
		AlertThreshold:       0.2,
		CriticalThreshold:    0.1,
	}
}

// initializeMetrics creates initial metric tracking structures
func (s *SLATracker) initializeMetrics() {
	now := time.Now()
	
	for tier := range s.configurations {
		s.metrics[tier] = &SLAMetrics{
			Tier:                 tier,
			CurrentAvailability:  1.0, // Start with 100% availability
			CurrentResponseTime:  0.0,
			ErrorBudgetRemaining: 1.0, // Start with full error budget
			ErrorBudgetBurnRate:  0.0,
			LastUpdate:          now,
			WindowStart:         now.Add(-24 * time.Hour), // Look at last 24h initially
			WindowEnd:           now,
			TotalRequests:       0,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			ResponseTimes:       make([]float64, 0),
		}
	}
}

// Start begins SLA tracking and metric updates
func (s *SLATracker) Start(ctx context.Context) error {
	s.logger.Info("Starting SLA tracker", "update_interval", s.updateInterval)
	
	s.ticker = time.NewTicker(s.updateInterval)
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.logger.Info("SLA tracker context cancelled")
				return
			case <-s.stopChan:
				s.logger.Info("SLA tracker stop signal received")
				return
			case <-s.ticker.C:
				s.updateSLAMetrics()
			}
		}
	}()
	
	return nil
}

// Stop stops the SLA tracker
func (s *SLATracker) Stop() error {
	s.logger.Info("Stopping SLA tracker")
	
	if s.ticker != nil {
		s.ticker.Stop()
	}
	
	close(s.stopChan)
	return nil
}

// RecordRequest records a request for SLA tracking
func (s *SLATracker) RecordRequest(responseTime float64, isSuccess bool, endpoint string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Determine SLA tier based on response time
	tier := s.determineSLATier(responseTime)
	
	metrics := s.metrics[tier]
	if metrics == nil {
		s.logger.Warn("No metrics found for SLA tier", "tier", tier)
		return
	}
	
	// Update request counts
	metrics.TotalRequests++
	if isSuccess {
		metrics.SuccessfulRequests++
	} else {
		metrics.FailedRequests++
	}
	
	// Update response times (keep last 1000 for moving average)
	metrics.ResponseTimes = append(metrics.ResponseTimes, responseTime)
	if len(metrics.ResponseTimes) > 1000 {
		metrics.ResponseTimes = metrics.ResponseTimes[1:]
	}
	
	metrics.LastUpdate = time.Now()
	
	s.logger.Debug("Recorded SLA request",
		"tier", tier,
		"response_time", responseTime,
		"success", isSuccess,
		"endpoint", endpoint,
	)
}

// determineSLATier determines which SLA tier a request falls into based on response time
func (s *SLATracker) determineSLATier(responseTime float64) SLATier {
	goldConfig := s.configurations[SLATierGold]
	silverConfig := s.configurations[SLATierSilver]
	
	if responseTime <= goldConfig.ResponseTimeTarget {
		return SLATierGold
	} else if responseTime <= silverConfig.ResponseTimeTarget {
		return SLATierSilver
	}
	
	return SLATierBronze
}

// updateSLAMetrics calculates and updates SLA metrics for all tiers
func (s *SLATracker) updateSLAMetrics() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for tier, metrics := range s.metrics {
		config := s.configurations[tier]
		
		// Calculate current availability
		if metrics.TotalRequests > 0 {
			metrics.CurrentAvailability = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
		}
		
		// Calculate average response time
		if len(metrics.ResponseTimes) > 0 {
			sum := 0.0
			for _, rt := range metrics.ResponseTimes {
				sum += rt
			}
			metrics.CurrentResponseTime = sum / float64(len(metrics.ResponseTimes))
		}
		
		// Calculate error budget remaining
		availabilityGap := config.AvailabilityTarget - metrics.CurrentAvailability
		maxAllowableGap := 1.0 - config.AvailabilityTarget
		
		if maxAllowableGap > 0 {
			metrics.ErrorBudgetRemaining = math.Max(0, 1.0-(availabilityGap/maxAllowableGap))
		} else {
			metrics.ErrorBudgetRemaining = 1.0
		}
		
		// Calculate burn rate (simplified - based on recent trend)
		s.calculateBurnRate(metrics, config)
		
		// Update Prometheus metrics
		s.updatePrometheusMetrics(tier, metrics)
		
		s.logger.Debug("Updated SLA metrics",
			"tier", tier,
			"availability", metrics.CurrentAvailability,
			"response_time", metrics.CurrentResponseTime,
			"error_budget", metrics.ErrorBudgetRemaining,
			"burn_rate", metrics.ErrorBudgetBurnRate,
		)
	}
}

// calculateBurnRate calculates the current error budget burn rate
func (s *SLATracker) calculateBurnRate(metrics *SLAMetrics, config SLAConfiguration) {
	// Simplified burn rate calculation
	// In production, this would be more sophisticated with time windows
	
	if metrics.TotalRequests == 0 {
		metrics.ErrorBudgetBurnRate = 0.0
		return
	}
	
	// Calculate burn rate as: (error_rate / allowable_error_rate) * time_factor
	errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
	allowableErrorRate := 1.0 - config.AvailabilityTarget
	
	if allowableErrorRate > 0 {
		baseBurnRate := errorRate / allowableErrorRate
		
		// Apply time factor (how fast we're burning compared to the budget period)
		// This is a simplified calculation
		metrics.ErrorBudgetBurnRate = baseBurnRate
	} else {
		metrics.ErrorBudgetBurnRate = 0.0
	}
}

// updatePrometheusMetrics updates Prometheus metrics with SLA data
func (s *SLATracker) updatePrometheusMetrics(tier SLATier, metrics *SLAMetrics) {
	serviceName := "template-arch-lint"
	tierStr := string(tier)
	
	s.prometheusMetrics.UpdateSLAMetrics(
		serviceName,
		tierStr,
		metrics.CurrentAvailability,
		metrics.ErrorBudgetRemaining,
		metrics.ErrorBudgetBurnRate,
	)
}

// GetSLAMetrics returns current SLA metrics for a specific tier
func (s *SLATracker) GetSLAMetrics(tier SLATier) (*SLAMetrics, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	metrics, exists := s.metrics[tier]
	if !exists {
		return nil, false
	}
	
	// Return a copy to avoid race conditions
	metricsCopy := *metrics
	metricsCopy.ResponseTimes = make([]float64, len(metrics.ResponseTimes))
	copy(metricsCopy.ResponseTimes, metrics.ResponseTimes)
	
	return &metricsCopy, true
}

// GetAllSLAMetrics returns SLA metrics for all tiers
func (s *SLATracker) GetAllSLAMetrics() map[SLATier]*SLAMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make(map[SLATier]*SLAMetrics)
	
	for tier, metrics := range s.metrics {
		metricsCopy := *metrics
		metricsCopy.ResponseTimes = make([]float64, len(metrics.ResponseTimes))
		copy(metricsCopy.ResponseTimes, metrics.ResponseTimes)
		result[tier] = &metricsCopy
	}
	
	return result
}

// IsErrorBudgetCritical checks if error budget is critically low for any tier
func (s *SLATracker) IsErrorBudgetCritical() map[SLATier]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make(map[SLATier]bool)
	
	for tier, metrics := range s.metrics {
		config := s.configurations[tier]
		result[tier] = metrics.ErrorBudgetRemaining < config.CriticalThreshold
	}
	
	return result
}

// GetSLASummary returns a summary of all SLA performance
func (s *SLATracker) GetSLASummary() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	summary := make(map[string]interface{})
	summary["last_updated"] = time.Now().Format(time.RFC3339)
	summary["tiers"] = make(map[string]interface{})
	
	for tier, metrics := range s.metrics {
		config := s.configurations[tier]
		
		tierSummary := map[string]interface{}{
			"availability": map[string]interface{}{
				"current": metrics.CurrentAvailability,
				"target":  config.AvailabilityTarget,
				"status":  s.getAvailabilityStatus(metrics.CurrentAvailability, config.AvailabilityTarget),
			},
			"response_time": map[string]interface{}{
				"current": metrics.CurrentResponseTime,
				"target":  config.ResponseTimeTarget,
				"status":  s.getResponseTimeStatus(metrics.CurrentResponseTime, config.ResponseTimeTarget),
			},
			"error_budget": map[string]interface{}{
				"remaining": metrics.ErrorBudgetRemaining,
				"burn_rate": metrics.ErrorBudgetBurnRate,
				"status":    s.getErrorBudgetStatus(metrics.ErrorBudgetRemaining, config),
			},
			"requests": map[string]interface{}{
				"total":      metrics.TotalRequests,
				"successful": metrics.SuccessfulRequests,
				"failed":     metrics.FailedRequests,
			},
		}
		
		summary["tiers"].(map[string]interface{})[string(tier)] = tierSummary
	}
	
	return summary
}

// getAvailabilityStatus returns status based on availability vs target
func (s *SLATracker) getAvailabilityStatus(current, target float64) string {
	if current >= target {
		return "healthy"
	} else if current >= target*0.95 { // Within 95% of target
		return "warning"
	}
	return "critical"
}

// getResponseTimeStatus returns status based on response time vs target
func (s *SLATracker) getResponseTimeStatus(current, target float64) string {
	if current <= target {
		return "healthy"
	} else if current <= target*1.5 { // Within 150% of target
		return "warning"
	}
	return "critical"
}

// getErrorBudgetStatus returns status based on error budget remaining
func (s *SLATracker) getErrorBudgetStatus(remaining float64, config SLAConfiguration) string {
	if remaining >= config.AlertThreshold {
		return "healthy"
	} else if remaining >= config.CriticalThreshold {
		return "warning"
	}
	return "critical"
}