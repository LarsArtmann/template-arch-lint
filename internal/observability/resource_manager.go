// Resource management and optimization for application performance
package observability

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// ResourceManager manages system resources and provides optimization recommendations
type ResourceManager struct {
	logger             *slog.Logger
	performanceMetrics *PerformanceMetrics
	
	// Resource tracking
	config             *ResourceConfig
	metrics            *ResourceMetrics
	mutex              sync.RWMutex
	
	// Optimization state
	lastOptimization   time.Time
	optimizations      []OptimizationAction
}

// ResourceConfig holds resource management configuration
type ResourceConfig struct {
	// Memory management
	MaxMemoryBytes       int64         `json:"max_memory_bytes"`
	GCTarget             float64       `json:"gc_target_percent"`
	GCInterval           time.Duration `json:"gc_interval"`
	
	// Goroutine management
	MaxGoroutines        int           `json:"max_goroutines"`
	GoroutineThreshold   int           `json:"goroutine_threshold"`
	
	// Monitoring and optimization
	MonitoringInterval   time.Duration `json:"monitoring_interval"`
	OptimizationEnabled  bool          `json:"optimization_enabled"`
	OptimizationCooldown time.Duration `json:"optimization_cooldown"`
}

// ResourceMetrics holds current resource usage metrics
type ResourceMetrics struct {
	Timestamp          time.Time `json:"timestamp"`
	
	// Memory metrics
	MemoryAllocated    uint64    `json:"memory_allocated"`
	MemoryTotal        uint64    `json:"memory_total"`
	MemorySystem       uint64    `json:"memory_system"`
	MemoryHeapObjects  uint64    `json:"memory_heap_objects"`
	GCCycles           uint32    `json:"gc_cycles"`
	GCPauseDuration    time.Duration `json:"gc_pause_duration"`
	
	// CPU and goroutine metrics
	NumCPU             int       `json:"num_cpu"`
	NumGoroutines      int       `json:"num_goroutines"`
	CGOCalls           int64     `json:"cgo_calls"`
	
	// Calculated metrics
	MemoryUtilization  float64   `json:"memory_utilization_percent"`
	GoroutineRatio     float64   `json:"goroutine_per_cpu"`
	
	// System limits
	SystemLimits       SystemLimits `json:"system_limits"`
}

// SystemLimits holds system resource limits
type SystemLimits struct {
	MaxMemoryBytes     int64  `json:"max_memory_bytes"`
	MaxGoroutines      int    `json:"max_goroutines"`
	MaxFileDescriptors int    `json:"max_file_descriptors"`
	MaxConnections     int    `json:"max_connections"`
}

// OptimizationAction represents a resource optimization action
type OptimizationAction struct {
	Type        string                 `json:"type"`
	Timestamp   time.Time             `json:"timestamp"`
	Parameters  map[string]interface{} `json:"parameters"`
	Result      string                `json:"result"`
	Impact      string                `json:"impact"`
}

// NewResourceManager creates a new resource manager
func NewResourceManager(logger *slog.Logger, performanceMetrics *PerformanceMetrics, config *ResourceConfig) *ResourceManager {
	if config == nil {
		config = DefaultResourceConfig()
	}
	
	return &ResourceManager{
		logger:             logger,
		performanceMetrics: performanceMetrics,
		config:            config,
		metrics:           &ResourceMetrics{},
		optimizations:     make([]OptimizationAction, 0),
	}
}

// DefaultResourceConfig returns default resource management configuration
func DefaultResourceConfig() *ResourceConfig {
	return &ResourceConfig{
		MaxMemoryBytes:       512 * 1024 * 1024, // 512MB default
		GCTarget:            100,                  // Default GC target
		GCInterval:          5 * time.Minute,
		MaxGoroutines:       10000,
		GoroutineThreshold:  1000,
		MonitoringInterval:  30 * time.Second,
		OptimizationEnabled: true,
		OptimizationCooldown: 5 * time.Minute,
	}
}

// StartMonitoring starts resource monitoring and optimization
func (rm *ResourceManager) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(rm.config.MonitoringInterval)
	defer ticker.Stop()
	
	rm.logger.Info("Resource monitoring started",
		"interval", rm.config.MonitoringInterval,
		"optimization_enabled", rm.config.OptimizationEnabled,
	)
	
	for {
		select {
		case <-ctx.Done():
			rm.logger.Info("Resource monitoring stopped")
			return
		case <-ticker.C:
			rm.collectMetrics()
			if rm.config.OptimizationEnabled {
				rm.optimizeIfNeeded()
			}
		}
	}
}

// collectMetrics collects current resource usage metrics
func (rm *ResourceManager) collectMetrics() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	rm.metrics = &ResourceMetrics{
		Timestamp:         time.Now(),
		MemoryAllocated:   memStats.Alloc,
		MemoryTotal:       memStats.TotalAlloc,
		MemorySystem:      memStats.Sys,
		MemoryHeapObjects: memStats.HeapObjects,
		GCCycles:          memStats.NumGC,
		GCPauseDuration:   time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256]),
		NumCPU:            runtime.NumCPU(),
		NumGoroutines:     runtime.NumGoroutine(),
		CGOCalls:          runtime.NumCgoCall(),
	}
	
	// Calculate derived metrics
	if rm.config.MaxMemoryBytes > 0 {
		rm.metrics.MemoryUtilization = float64(rm.metrics.MemoryAllocated) / float64(rm.config.MaxMemoryBytes) * 100
	}
	
	rm.metrics.GoroutineRatio = float64(rm.metrics.NumGoroutines) / float64(rm.metrics.NumCPU)
	
	// Set system limits
	rm.metrics.SystemLimits = SystemLimits{
		MaxMemoryBytes:     rm.config.MaxMemoryBytes,
		MaxGoroutines:      rm.config.MaxGoroutines,
		MaxFileDescriptors: 1024, // Default limit
		MaxConnections:     1000, // Default limit
	}
	
	// Record metrics for observability
	rm.recordMetrics()
}

// recordMetrics records resource metrics for observability
func (rm *ResourceManager) recordMetrics() {
	ctx := context.Background()
	
	// Record memory allocation
	rm.performanceMetrics.RecordMemoryAllocation(ctx, int64(rm.metrics.MemoryAllocated))
	
	// Record GC duration if we have a recent GC
	if rm.metrics.GCPauseDuration > 0 {
		rm.performanceMetrics.RecordGCDuration(ctx, rm.metrics.GCPauseDuration)
	}
	
	rm.logger.Debug("Resource metrics collected",
		"memory_allocated_mb", rm.metrics.MemoryAllocated/1024/1024,
		"memory_utilization_percent", rm.metrics.MemoryUtilization,
		"goroutines", rm.metrics.NumGoroutines,
		"gc_cycles", rm.metrics.GCCycles,
		"gc_pause_duration_ms", rm.metrics.GCPauseDuration.Milliseconds(),
	)
}

// optimizeIfNeeded performs automatic optimization if needed and cooldown period has passed
func (rm *ResourceManager) optimizeIfNeeded() {
	now := time.Now()
	if now.Sub(rm.lastOptimization) < rm.config.OptimizationCooldown {
		return
	}
	
	recommendations := rm.getOptimizationRecommendations()
	
	for _, recommendation := range recommendations {
		if recommendation.Impact == "high" {
			if rm.applyOptimization(recommendation) {
				rm.lastOptimization = now
				break // Only apply one optimization at a time
			}
		}
	}
}

// getOptimizationRecommendations analyzes metrics and provides optimization recommendations
func (rm *ResourceManager) getOptimizationRecommendations() []OptimizationRecommendation {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	var recommendations []OptimizationRecommendation
	
	// High memory usage
	if rm.metrics.MemoryUtilization > 85 {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:        "force_gc",
			Description: fmt.Sprintf("High memory utilization (%.2f%%)", rm.metrics.MemoryUtilization),
			Impact:      "high",
			Parameters: map[string]interface{}{
				"current_memory_mb": rm.metrics.MemoryAllocated / 1024 / 1024,
				"utilization":       rm.metrics.MemoryUtilization,
			},
		})
	}
	
	// Excessive goroutines
	if rm.metrics.NumGoroutines > rm.config.GoroutineThreshold {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:        "reduce_goroutines",
			Description: fmt.Sprintf("High goroutine count (%d)", rm.metrics.NumGoroutines),
			Impact:      "medium",
			Parameters: map[string]interface{}{
				"current_goroutines": rm.metrics.NumGoroutines,
				"threshold":          rm.config.GoroutineThreshold,
			},
		})
	}
	
	// Long GC pause times
	if rm.metrics.GCPauseDuration > 50*time.Millisecond {
		recommendations = append(recommendations, OptimizationRecommendation{
			Type:        "tune_gc",
			Description: fmt.Sprintf("Long GC pause time (%v)", rm.metrics.GCPauseDuration),
			Impact:      "medium",
			Parameters: map[string]interface{}{
				"current_pause_ms": rm.metrics.GCPauseDuration.Milliseconds(),
				"target_percent":   rm.config.GCTarget,
			},
		})
	}
	
	return recommendations
}

// OptimizationRecommendation represents a resource optimization recommendation
type OptimizationRecommendation struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Impact      string                 `json:"impact"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// applyOptimization applies a specific optimization
func (rm *ResourceManager) applyOptimization(recommendation OptimizationRecommendation) bool {
	action := OptimizationAction{
		Type:       recommendation.Type,
		Timestamp:  time.Now(),
		Parameters: recommendation.Parameters,
		Impact:     recommendation.Impact,
	}
	
	switch recommendation.Type {
	case "force_gc":
		return rm.forceGC(&action)
	case "reduce_goroutines":
		return rm.optimizeGoroutines(&action)
	case "tune_gc":
		return rm.tuneGC(&action)
	default:
		action.Result = fmt.Sprintf("Unknown optimization type: %s", recommendation.Type)
		rm.addOptimizationAction(action)
		return false
	}
}

// forceGC forces garbage collection
func (rm *ResourceManager) forceGC(action *OptimizationAction) bool {
	memBefore := rm.metrics.MemoryAllocated
	
	// Force garbage collection
	runtime.GC()
	
	// Collect metrics after GC
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	memAfter := memStats.Alloc
	
	memoryReclaimed := memBefore - memAfter
	
	action.Result = fmt.Sprintf("GC completed, reclaimed %d MB", memoryReclaimed/1024/1024)
	action.Parameters["memory_reclaimed_mb"] = memoryReclaimed / 1024 / 1024
	
	rm.logger.Info("Forced garbage collection",
		"memory_before_mb", memBefore/1024/1024,
		"memory_after_mb", memAfter/1024/1024,
		"memory_reclaimed_mb", memoryReclaimed/1024/1024,
	)
	
	rm.addOptimizationAction(*action)
	return true
}

// optimizeGoroutines attempts to optimize goroutine usage
func (rm *ResourceManager) optimizeGoroutines(action *OptimizationAction) bool {
	// This is a placeholder for goroutine optimization
	// In practice, this would involve application-specific logic
	// to reduce goroutine creation or improve pooling
	
	action.Result = "Goroutine optimization recommendation logged for manual action"
	
	rm.logger.Warn("High goroutine count detected - consider reviewing goroutine usage",
		"current_count", rm.metrics.NumGoroutines,
		"threshold", rm.config.GoroutineThreshold,
		"ratio_per_cpu", rm.metrics.GoroutineRatio,
	)
	
	rm.addOptimizationAction(*action)
	return true
}

// tuneGC tunes garbage collection parameters
func (rm *ResourceManager) tuneGC(action *OptimizationAction) bool {
	currentTarget := debug.SetGCPercent(-1) // Get current GC target
	debug.SetGCPercent(currentTarget)        // Restore it
	
	// Adjust GC target based on memory pressure
	newTarget := int(rm.config.GCTarget)
	if rm.metrics.MemoryUtilization > 80 {
		newTarget = int(rm.config.GCTarget * 0.8) // More aggressive GC
	}
	
	debug.SetGCPercent(newTarget)
	
	action.Result = fmt.Sprintf("GC target adjusted from %d%% to %d%%", currentTarget, newTarget)
	action.Parameters["old_target"] = currentTarget
	action.Parameters["new_target"] = newTarget
	
	rm.logger.Info("GC tuning applied",
		"old_target_percent", currentTarget,
		"new_target_percent", newTarget,
		"memory_utilization", rm.metrics.MemoryUtilization,
	)
	
	rm.addOptimizationAction(*action)
	return true
}

// addOptimizationAction adds an optimization action to the history
func (rm *ResourceManager) addOptimizationAction(action OptimizationAction) {
	rm.optimizations = append(rm.optimizations, action)
	
	// Keep only the last 100 optimization actions
	if len(rm.optimizations) > 100 {
		rm.optimizations = rm.optimizations[1:]
	}
}

// GetCurrentMetrics returns current resource metrics
func (rm *ResourceManager) GetCurrentMetrics() *ResourceMetrics {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := *rm.metrics
	return &metrics
}

// GetOptimizationHistory returns the history of optimization actions
func (rm *ResourceManager) GetOptimizationHistory() []OptimizationAction {
	return append([]OptimizationAction(nil), rm.optimizations...)
}

// GetHealthStatus returns resource health status
func (rm *ResourceManager) GetHealthStatus() map[string]interface{} {
	metrics := rm.GetCurrentMetrics()
	
	healthStatus := map[string]interface{}{
		"healthy": true,
		"metrics": metrics,
		"issues":  []string{},
	}
	
	var issues []string
	
	// Check memory utilization
	if metrics.MemoryUtilization > 90 {
		issues = append(issues, fmt.Sprintf("Critical memory utilization: %.2f%%", metrics.MemoryUtilization))
		healthStatus["healthy"] = false
	} else if metrics.MemoryUtilization > 80 {
		issues = append(issues, fmt.Sprintf("High memory utilization: %.2f%%", metrics.MemoryUtilization))
	}
	
	// Check goroutine count
	if metrics.NumGoroutines > rm.config.MaxGoroutines {
		issues = append(issues, fmt.Sprintf("Excessive goroutines: %d", metrics.NumGoroutines))
		healthStatus["healthy"] = false
	} else if metrics.NumGoroutines > rm.config.GoroutineThreshold {
		issues = append(issues, fmt.Sprintf("High goroutine count: %d", metrics.NumGoroutines))
	}
	
	// Check GC pause times
	if metrics.GCPauseDuration > 100*time.Millisecond {
		issues = append(issues, fmt.Sprintf("Long GC pause: %v", metrics.GCPauseDuration))
		healthStatus["healthy"] = false
	}
	
	healthStatus["issues"] = issues
	return healthStatus
}

// OptimizeForEnvironment optimizes resources for specific environment
func (rm *ResourceManager) OptimizeForEnvironment(environment string) error {
	switch environment {
	case "production":
		return rm.optimizeForProduction()
	case "staging":
		return rm.optimizeForStaging()
	case "development":
		return rm.optimizeForDevelopment()
	case "testing":
		return rm.optimizeForTesting()
	default:
		return fmt.Errorf("unknown environment: %s", environment)
	}
}

// optimizeForProduction optimizes for production environment
func (rm *ResourceManager) optimizeForProduction() error {
	rm.config.GCTarget = 100      // Standard GC target
	rm.config.GCInterval = 5 * time.Minute
	debug.SetGCPercent(int(rm.config.GCTarget))
	
	rm.logger.Info("Resource optimization applied for production environment")
	return nil
}

// optimizeForStaging optimizes for staging environment
func (rm *ResourceManager) optimizeForStaging() error {
	rm.config.GCTarget = 120      // Slightly less aggressive GC
	rm.config.GCInterval = 3 * time.Minute
	debug.SetGCPercent(int(rm.config.GCTarget))
	
	rm.logger.Info("Resource optimization applied for staging environment")
	return nil
}

// optimizeForDevelopment optimizes for development environment
func (rm *ResourceManager) optimizeForDevelopment() error {
	rm.config.GCTarget = 200      // Less aggressive GC for development
	rm.config.GCInterval = 2 * time.Minute
	debug.SetGCPercent(int(rm.config.GCTarget))
	
	rm.logger.Info("Resource optimization applied for development environment")
	return nil
}

// optimizeForTesting optimizes for testing environment
func (rm *ResourceManager) optimizeForTesting() error {
	rm.config.GCTarget = 50       // More aggressive GC for testing
	rm.config.GCInterval = 1 * time.Minute
	debug.SetGCPercent(int(rm.config.GCTarget))
	
	rm.logger.Info("Resource optimization applied for testing environment")
	return nil
}