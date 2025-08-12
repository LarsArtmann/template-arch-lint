// Performance monitoring and metrics collection for application optimization
package observability

import (
	"context"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	meterName = "template-arch-lint/performance"
)

// PerformanceMetrics collects and tracks application performance metrics
type PerformanceMetrics struct {
	logger *slog.Logger
	meter  metric.Meter

	// Memory metrics
	memoryUsage       metric.Int64ObservableGauge
	memoryAllocations metric.Int64Counter
	gcDuration        metric.Float64Histogram

	// CPU metrics
	cpuUsage    metric.Float64ObservableGauge
	goroutines  metric.Int64ObservableGauge
	cgoCalls    metric.Int64ObservableGauge

	// HTTP metrics
	requestDuration  metric.Float64Histogram
	requestSize      metric.Int64Histogram
	responseSize     metric.Int64Histogram
	activeConnections metric.Int64ObservableGauge

	// Database metrics
	dbConnections     metric.Int64ObservableGauge
	dbQueryDuration   metric.Float64Histogram
	dbQueryCount      metric.Int64Counter
	dbConnectionPool  metric.Int64ObservableGauge

	// Cache metrics
	cacheHits        metric.Int64Counter
	cacheMisses      metric.Int64Counter
	cacheEvictions   metric.Int64Counter
	cacheSize        metric.Int64ObservableGauge

	// Custom business metrics
	userOperations metric.Int64Counter
	errorRate      metric.Float64Histogram
}

// NewPerformanceMetrics creates a new performance metrics collector
func NewPerformanceMetrics(logger *slog.Logger) (*PerformanceMetrics, error) {
	meter := otel.Meter(meterName)

	pm := &PerformanceMetrics{
		logger: logger,
		meter:  meter,
	}

	if err := pm.initializeMetrics(); err != nil {
		return nil, err
	}

	return pm, nil
}

// initializeMetrics initializes all performance metrics
func (pm *PerformanceMetrics) initializeMetrics() error {
	var err error

	// Memory metrics
	pm.memoryUsage, err = pm.meter.Int64ObservableGauge(
		"memory_usage_bytes",
		metric.WithDescription("Current memory usage in bytes"),
	)
	if err != nil {
		return err
	}

	pm.memoryAllocations, err = pm.meter.Int64Counter(
		"memory_allocations_total",
		metric.WithDescription("Total memory allocations"),
	)
	if err != nil {
		return err
	}

	pm.gcDuration, err = pm.meter.Float64Histogram(
		"gc_duration_seconds",
		metric.WithDescription("Garbage collection duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	// CPU metrics
	pm.cpuUsage, err = pm.meter.Float64ObservableGauge(
		"cpu_usage_percent",
		metric.WithDescription("CPU usage percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return err
	}

	pm.goroutines, err = pm.meter.Int64ObservableGauge(
		"goroutines_count",
		metric.WithDescription("Number of active goroutines"),
	)
	if err != nil {
		return err
	}

	pm.cgoCalls, err = pm.meter.Int64ObservableGauge(
		"cgo_calls_count",
		metric.WithDescription("Number of CGO calls"),
	)
	if err != nil {
		return err
	}

	// HTTP metrics
	pm.requestDuration, err = pm.meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	pm.requestSize, err = pm.meter.Int64Histogram(
		"http_request_size_bytes",
		metric.WithDescription("HTTP request size in bytes"),
	)
	if err != nil {
		return err
	}

	pm.responseSize, err = pm.meter.Int64Histogram(
		"http_response_size_bytes",
		metric.WithDescription("HTTP response size in bytes"),
	)
	if err != nil {
		return err
	}

	pm.activeConnections, err = pm.meter.Int64ObservableGauge(
		"http_active_connections",
		metric.WithDescription("Number of active HTTP connections"),
	)
	if err != nil {
		return err
	}

	// Database metrics
	pm.dbConnections, err = pm.meter.Int64ObservableGauge(
		"database_connections_active",
		metric.WithDescription("Active database connections"),
	)
	if err != nil {
		return err
	}

	pm.dbQueryDuration, err = pm.meter.Float64Histogram(
		"database_query_duration_seconds",
		metric.WithDescription("Database query duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	pm.dbQueryCount, err = pm.meter.Int64Counter(
		"database_queries_total",
		metric.WithDescription("Total database queries executed"),
	)
	if err != nil {
		return err
	}

	pm.dbConnectionPool, err = pm.meter.Int64ObservableGauge(
		"database_connection_pool_size",
		metric.WithDescription("Database connection pool size"),
	)
	if err != nil {
		return err
	}

	// Cache metrics
	pm.cacheHits, err = pm.meter.Int64Counter(
		"cache_hits_total",
		metric.WithDescription("Total cache hits"),
	)
	if err != nil {
		return err
	}

	pm.cacheMisses, err = pm.meter.Int64Counter(
		"cache_misses_total",
		metric.WithDescription("Total cache misses"),
	)
	if err != nil {
		return err
	}

	pm.cacheEvictions, err = pm.meter.Int64Counter(
		"cache_evictions_total",
		metric.WithDescription("Total cache evictions"),
	)
	if err != nil {
		return err
	}

	pm.cacheSize, err = pm.meter.Int64ObservableGauge(
		"cache_size_bytes",
		metric.WithDescription("Current cache size in bytes"),
	)
	if err != nil {
		return err
	}

	// Business metrics
	pm.userOperations, err = pm.meter.Int64Counter(
		"user_operations_total",
		metric.WithDescription("Total user operations"),
	)
	if err != nil {
		return err
	}

	pm.errorRate, err = pm.meter.Float64Histogram(
		"error_rate_percent",
		metric.WithDescription("Application error rate"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return err
	}

	return nil
}

// StartRuntimeMetricsCollection starts collecting runtime metrics
func (pm *PerformanceMetrics) StartRuntimeMetricsCollection(ctx context.Context) {
	// Register runtime metrics callbacks
	_, err := pm.meter.RegisterCallback(
		pm.collectRuntimeMetrics,
		pm.memoryUsage,
		pm.cpuUsage,
		pm.goroutines,
		pm.cgoCalls,
	)
	if err != nil {
		pm.logger.Error("Failed to register runtime metrics callback", "error", err)
		return
	}

	pm.logger.Info("Runtime metrics collection started")
}

// collectRuntimeMetrics collects Go runtime metrics
func (pm *PerformanceMetrics) collectRuntimeMetrics(ctx context.Context, observer metric.Observer) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Memory metrics
	observer.ObserveInt64(pm.memoryUsage, int64(memStats.Alloc))

	// CPU and goroutine metrics
	observer.ObserveInt64(pm.goroutines, int64(runtime.NumGoroutine()))
	observer.ObserveInt64(pm.cgoCalls, int64(runtime.NumCgoCall()))

	return nil
}

// RecordHTTPRequest records HTTP request metrics
func (pm *PerformanceMetrics) RecordHTTPRequest(ctx context.Context, method, endpoint string, duration time.Duration, requestSize, responseSize int64, statusCode int) {
	attributes := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
		attribute.Int("status_code", statusCode),
	)

	pm.requestDuration.Record(ctx, duration.Seconds(), attributes)
	pm.requestSize.Record(ctx, requestSize, attributes)
	pm.responseSize.Record(ctx, responseSize, attributes)
}

// RecordDatabaseQuery records database query metrics
func (pm *PerformanceMetrics) RecordDatabaseQuery(ctx context.Context, query string, duration time.Duration, success bool) {
	attributes := metric.WithAttributes(
		attribute.String("query_type", extractQueryType(query)),
		attribute.Bool("success", success),
	)

	pm.dbQueryDuration.Record(ctx, duration.Seconds(), attributes)
	pm.dbQueryCount.Add(ctx, 1, attributes)
}

// RecordCacheOperation records cache operation metrics
func (pm *PerformanceMetrics) RecordCacheOperation(ctx context.Context, operation string, hit bool) {
	attributes := metric.WithAttributes(
		attribute.String("operation", operation),
	)

	if hit {
		pm.cacheHits.Add(ctx, 1, attributes)
	} else {
		pm.cacheMisses.Add(ctx, 1, attributes)
	}
}

// RecordCacheEviction records cache eviction metrics
func (pm *PerformanceMetrics) RecordCacheEviction(ctx context.Context, reason string) {
	attributes := metric.WithAttributes(
		attribute.String("reason", reason),
	)
	pm.cacheEvictions.Add(ctx, 1, attributes)
}

// RecordUserOperation records business operation metrics
func (pm *PerformanceMetrics) RecordUserOperation(ctx context.Context, operation string, success bool) {
	attributes := metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.Bool("success", success),
	)
	pm.userOperations.Add(ctx, 1, attributes)
}

// RecordGCDuration records garbage collection duration
func (pm *PerformanceMetrics) RecordGCDuration(ctx context.Context, duration time.Duration) {
	pm.gcDuration.Record(ctx, duration.Seconds())
}

// RecordMemoryAllocation records memory allocation
func (pm *PerformanceMetrics) RecordMemoryAllocation(ctx context.Context, bytes int64) {
	pm.memoryAllocations.Add(ctx, bytes)
}

// RecordErrorRate records application error rate
func (pm *PerformanceMetrics) RecordErrorRate(ctx context.Context, errorRate float64) {
	pm.errorRate.Record(ctx, errorRate)
}

// extractQueryType extracts the type of SQL query for metrics
func extractQueryType(query string) string {
	query = strings.ToUpper(strings.TrimSpace(query))
	
	if strings.HasPrefix(query, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(query, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(query, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(query, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(query, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(query, "DROP") {
		return "DROP"
	} else if strings.HasPrefix(query, "ALTER") {
		return "ALTER"
	}
	
	return "OTHER"
}

// GetMetricsSummary returns a summary of current performance metrics
func (pm *PerformanceMetrics) GetMetricsSummary() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"memory": map[string]interface{}{
			"alloc_bytes":      memStats.Alloc,
			"total_alloc":      memStats.TotalAlloc,
			"sys_bytes":        memStats.Sys,
			"heap_objects":     memStats.HeapObjects,
			"gc_cycles":        memStats.NumGC,
			"next_gc_bytes":    memStats.NextGC,
		},
		"runtime": map[string]interface{}{
			"goroutines":     runtime.NumGoroutine(),
			"cgo_calls":      runtime.NumCgoCall(),
			"cpu_count":      runtime.NumCPU(),
		},
		"timestamp": time.Now().Unix(),
	}
}