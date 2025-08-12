// Connection pool optimization and monitoring for database performance
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/observability"
)

// ConnectionPoolOptimizer monitors and optimizes database connection pool settings
type ConnectionPoolOptimizer struct {
	db                 *sql.DB
	logger             *slog.Logger
	performanceMetrics *observability.PerformanceMetrics
	
	// Monitoring state
	metrics            *PoolMetrics
	mutex              sync.RWMutex
	
	// Configuration
	config             *PoolConfig
}

// PoolConfig holds connection pool configuration
type PoolConfig struct {
	MaxOpenConns        int           `json:"max_open_conns"`
	MaxIdleConns        int           `json:"max_idle_conns"`
	ConnMaxLifetime     time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime     time.Duration `json:"conn_max_idle_time"`
	MonitoringInterval  time.Duration `json:"monitoring_interval"`
	OptimizationEnabled bool          `json:"optimization_enabled"`
}

// PoolMetrics holds connection pool performance metrics
type PoolMetrics struct {
	Timestamp           time.Time `json:"timestamp"`
	OpenConnections     int       `json:"open_connections"`
	InUse              int       `json:"in_use"`
	Idle               int       `json:"idle"`
	MaxOpenConnections int       `json:"max_open_connections"`
	MaxIdleConnections int       `json:"max_idle_connections"`
	
	// Performance metrics
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
	
	// Calculated metrics
	ConnectionUtilization float64 `json:"connection_utilization"`
	AverageWaitTime      float64 `json:"average_wait_time_ms"`
}

// PoolOptimizationRecommendation provides optimization recommendations
type PoolOptimizationRecommendation struct {
	Parameter     string      `json:"parameter"`
	CurrentValue  interface{} `json:"current_value"`
	RecommendedValue interface{} `json:"recommended_value"`
	Reason       string      `json:"reason"`
	ImpactLevel  string      `json:"impact_level"` // "low", "medium", "high"
}

// NewConnectionPoolOptimizer creates a new connection pool optimizer
func NewConnectionPoolOptimizer(
	db *sql.DB,
	logger *slog.Logger,
	performanceMetrics *observability.PerformanceMetrics,
	config *PoolConfig,
) *ConnectionPoolOptimizer {
	if config == nil {
		config = DefaultPoolConfig()
	}
	
	return &ConnectionPoolOptimizer{
		db:                 db,
		logger:             logger,
		performanceMetrics: performanceMetrics,
		config:            config,
		metrics:           &PoolMetrics{},
	}
}

// DefaultPoolConfig returns default connection pool configuration
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:        runtime.NumCPU() * 4, // 4 connections per CPU core
		MaxIdleConns:        runtime.NumCPU() * 2, // 2 idle connections per CPU core
		ConnMaxLifetime:     30 * time.Minute,
		ConnMaxIdleTime:     15 * time.Minute,
		MonitoringInterval:  30 * time.Second,
		OptimizationEnabled: true,
	}
}

// StartMonitoring starts connection pool monitoring
func (cpo *ConnectionPoolOptimizer) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(cpo.config.MonitoringInterval)
	defer ticker.Stop()
	
	cpo.logger.Info("Connection pool monitoring started",
		"interval", cpo.config.MonitoringInterval,
		"optimization_enabled", cpo.config.OptimizationEnabled,
	)
	
	for {
		select {
		case <-ctx.Done():
			cpo.logger.Info("Connection pool monitoring stopped")
			return
		case <-ticker.C:
			cpo.collectMetrics()
			if cpo.config.OptimizationEnabled {
				cpo.optimizeIfNeeded(ctx)
			}
		}
	}
}

// collectMetrics collects current connection pool metrics
func (cpo *ConnectionPoolOptimizer) collectMetrics() {
	cpo.mutex.Lock()
	defer cpo.mutex.Unlock()
	
	stats := cpo.db.Stats()
	
	cpo.metrics = &PoolMetrics{
		Timestamp:           time.Now(),
		OpenConnections:     stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		MaxOpenConnections: stats.MaxOpenConnections,
		MaxIdleConnections: cpo.config.MaxIdleConns,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxIdleTimeClosed:  stats.MaxIdleTimeClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
	
	// Calculate derived metrics
	if cpo.metrics.MaxOpenConnections > 0 {
		cpo.metrics.ConnectionUtilization = float64(cpo.metrics.InUse) / float64(cpo.metrics.MaxOpenConnections) * 100
	}
	
	if cpo.metrics.WaitCount > 0 {
		cpo.metrics.AverageWaitTime = float64(cpo.metrics.WaitDuration.Nanoseconds()) / float64(cpo.metrics.WaitCount) / 1e6 // Convert to milliseconds
	}
	
	// Record metrics for observability
	cpo.recordMetrics()
}

// recordMetrics records pool metrics for observability
func (cpo *ConnectionPoolOptimizer) recordMetrics() {
	ctx := context.Background()
	
	// Register connection pool metrics callback
	cpo.performanceMetrics.StartRuntimeMetricsCollection(ctx)
	
	cpo.logger.Debug("Connection pool metrics collected",
		"open_connections", cpo.metrics.OpenConnections,
		"in_use", cpo.metrics.InUse,
		"idle", cpo.metrics.Idle,
		"utilization_percent", cpo.metrics.ConnectionUtilization,
		"wait_count", cpo.metrics.WaitCount,
		"avg_wait_time_ms", cpo.metrics.AverageWaitTime,
	)
}

// optimizeIfNeeded performs automatic optimization if needed
func (cpo *ConnectionPoolOptimizer) optimizeIfNeeded(ctx context.Context) {
	recommendations := cpo.GetOptimizationRecommendations()
	
	for _, recommendation := range recommendations {
		if recommendation.ImpactLevel == "high" {
			cpo.applyOptimization(recommendation)
		}
	}
}

// GetOptimizationRecommendations analyzes metrics and provides optimization recommendations
func (cpo *ConnectionPoolOptimizer) GetOptimizationRecommendations() []PoolOptimizationRecommendation {
	cpo.mutex.RLock()
	defer cpo.mutex.RUnlock()
	
	var recommendations []PoolOptimizationRecommendation
	
	// High connection utilization
	if cpo.metrics.ConnectionUtilization > 90 {
		recommendations = append(recommendations, PoolOptimizationRecommendation{
			Parameter:        "max_open_conns",
			CurrentValue:     cpo.metrics.MaxOpenConnections,
			RecommendedValue: cpo.metrics.MaxOpenConnections * 2,
			Reason:          "High connection utilization (>90%) may cause bottlenecks",
			ImpactLevel:     "high",
		})
	}
	
	// Excessive wait times
	if cpo.metrics.AverageWaitTime > 50 { // 50ms threshold
		recommendations = append(recommendations, PoolOptimizationRecommendation{
			Parameter:        "max_open_conns",
			CurrentValue:     cpo.metrics.MaxOpenConnections,
			RecommendedValue: cpo.metrics.MaxOpenConnections + runtime.NumCPU(),
			Reason:          fmt.Sprintf("High average wait time (%.2fms)", cpo.metrics.AverageWaitTime),
			ImpactLevel:     "high",
		})
	}
	
	// Low idle connection usage
	idleUtilization := float64(cpo.metrics.Idle) / float64(cpo.metrics.MaxIdleConnections) * 100
	if idleUtilization < 20 && cpo.metrics.MaxIdleConnections > 5 {
		recommendations = append(recommendations, PoolOptimizationRecommendation{
			Parameter:        "max_idle_conns",
			CurrentValue:     cpo.metrics.MaxIdleConnections,
			RecommendedValue: cpo.metrics.MaxIdleConnections / 2,
			Reason:          "Low idle connection utilization, reducing memory usage",
			ImpactLevel:     "medium",
		})
	}
	
	// Excessive connection closures
	totalClosures := cpo.metrics.MaxIdleClosed + cpo.metrics.MaxIdleTimeClosed + cpo.metrics.MaxLifetimeClosed
	if totalClosures > 100 { // Arbitrary threshold
		if cpo.metrics.MaxIdleTimeClosed > totalClosures/2 {
			recommendations = append(recommendations, PoolOptimizationRecommendation{
				Parameter:        "conn_max_idle_time",
				CurrentValue:     cpo.config.ConnMaxIdleTime,
				RecommendedValue: cpo.config.ConnMaxIdleTime * 2,
				Reason:          "High idle timeout closures, increasing idle time",
				ImpactLevel:     "medium",
			})
		}
		
		if cpo.metrics.MaxLifetimeClosed > totalClosures/2 {
			recommendations = append(recommendations, PoolOptimizationRecommendation{
				Parameter:        "conn_max_lifetime",
				CurrentValue:     cpo.config.ConnMaxLifetime,
				RecommendedValue: cpo.config.ConnMaxLifetime * 2,
				Reason:          "High lifetime closures, increasing connection lifetime",
				ImpactLevel:     "medium",
			})
		}
	}
	
	return recommendations
}

// applyOptimization applies a specific optimization recommendation
func (cpo *ConnectionPoolOptimizer) applyOptimization(recommendation PoolOptimizationRecommendation) {
	cpo.logger.Info("Applying connection pool optimization",
		"parameter", recommendation.Parameter,
		"current_value", recommendation.CurrentValue,
		"recommended_value", recommendation.RecommendedValue,
		"reason", recommendation.Reason,
	)
	
	switch recommendation.Parameter {
	case "max_open_conns":
		if value, ok := recommendation.RecommendedValue.(int); ok {
			cpo.db.SetMaxOpenConns(value)
			cpo.config.MaxOpenConns = value
		}
	case "max_idle_conns":
		if value, ok := recommendation.RecommendedValue.(int); ok {
			cpo.db.SetMaxIdleConns(value)
			cpo.config.MaxIdleConns = value
		}
	case "conn_max_lifetime":
		if value, ok := recommendation.RecommendedValue.(time.Duration); ok {
			cpo.db.SetConnMaxLifetime(value)
			cpo.config.ConnMaxLifetime = value
		}
	case "conn_max_idle_time":
		if value, ok := recommendation.RecommendedValue.(time.Duration); ok {
			cpo.db.SetConnMaxIdleTime(value)
			cpo.config.ConnMaxIdleTime = value
		}
	}
}

// GetCurrentMetrics returns current connection pool metrics
func (cpo *ConnectionPoolOptimizer) GetCurrentMetrics() *PoolMetrics {
	cpo.mutex.RLock()
	defer cpo.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := *cpo.metrics
	return &metrics
}

// GetHealthStatus returns health status based on pool metrics
func (cpo *ConnectionPoolOptimizer) GetHealthStatus() map[string]interface{} {
	metrics := cpo.GetCurrentMetrics()
	
	healthStatus := map[string]interface{}{
		"healthy": true,
		"metrics": metrics,
		"issues": []string{},
	}
	
	var issues []string
	
	// Check for high utilization
	if metrics.ConnectionUtilization > 95 {
		issues = append(issues, "Very high connection utilization")
		healthStatus["healthy"] = false
	}
	
	// Check for excessive wait times
	if metrics.AverageWaitTime > 100 { // 100ms threshold
		issues = append(issues, fmt.Sprintf("High average wait time: %.2fms", metrics.AverageWaitTime))
		healthStatus["healthy"] = false
	}
	
	// Check for connection pool exhaustion
	if metrics.WaitCount > 0 && metrics.OpenConnections == metrics.MaxOpenConnections {
		issues = append(issues, "Connection pool exhaustion detected")
		healthStatus["healthy"] = false
	}
	
	healthStatus["issues"] = issues
	return healthStatus
}

// OptimizeForWorkload optimizes connection pool for specific workload patterns
func (cpo *ConnectionPoolOptimizer) OptimizeForWorkload(workloadType string) error {
	switch workloadType {
	case "high_throughput":
		return cpo.optimizeForHighThroughput()
	case "low_latency":
		return cpo.optimizeForLowLatency()
	case "batch_processing":
		return cpo.optimizeForBatchProcessing()
	case "development":
		return cpo.optimizeForDevelopment()
	default:
		return fmt.Errorf("unknown workload type: %s", workloadType)
	}
}

// optimizeForHighThroughput configures pool for high throughput workloads
func (cpo *ConnectionPoolOptimizer) optimizeForHighThroughput() error {
	cpo.db.SetMaxOpenConns(runtime.NumCPU() * 8)
	cpo.db.SetMaxIdleConns(runtime.NumCPU() * 4)
	cpo.db.SetConnMaxLifetime(1 * time.Hour)
	cpo.db.SetConnMaxIdleTime(30 * time.Minute)
	
	cpo.logger.Info("Connection pool optimized for high throughput")
	return nil
}

// optimizeForLowLatency configures pool for low latency workloads
func (cpo *ConnectionPoolOptimizer) optimizeForLowLatency() error {
	cpo.db.SetMaxOpenConns(runtime.NumCPU() * 2)
	cpo.db.SetMaxIdleConns(runtime.NumCPU() * 2)
	cpo.db.SetConnMaxLifetime(2 * time.Hour)
	cpo.db.SetConnMaxIdleTime(1 * time.Hour)
	
	cpo.logger.Info("Connection pool optimized for low latency")
	return nil
}

// optimizeForBatchProcessing configures pool for batch processing workloads
func (cpo *ConnectionPoolOptimizer) optimizeForBatchProcessing() error {
	cpo.db.SetMaxOpenConns(runtime.NumCPU())
	cpo.db.SetMaxIdleConns(2)
	cpo.db.SetConnMaxLifetime(4 * time.Hour)
	cpo.db.SetConnMaxIdleTime(1 * time.Hour)
	
	cpo.logger.Info("Connection pool optimized for batch processing")
	return nil
}

// optimizeForDevelopment configures pool for development environment
func (cpo *ConnectionPoolOptimizer) optimizeForDevelopment() error {
	cpo.db.SetMaxOpenConns(5)
	cpo.db.SetMaxIdleConns(2)
	cpo.db.SetConnMaxLifetime(30 * time.Minute)
	cpo.db.SetConnMaxIdleTime(5 * time.Minute)
	
	cpo.logger.Info("Connection pool optimized for development")
	return nil
}