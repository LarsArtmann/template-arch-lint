// Package observability provides comprehensive Prometheus metrics collection
package observability

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
)

// PrometheusMetrics provides comprehensive metrics collection
type PrometheusMetrics struct {
	config *config.Config
	logger *slog.Logger
	server *http.Server
	mu     sync.RWMutex

	// HTTP metrics
	httpRequestsTotal        *prometheus.CounterVec
	httpRequestDuration      *prometheus.HistogramVec
	httpRequestSize          *prometheus.HistogramVec
	httpResponseSize         *prometheus.HistogramVec
	httpActiveRequests       prometheus.Gauge
	httpErrorsTotal          *prometheus.CounterVec

	// Application metrics
	appInfo                  *prometheus.GaugeVec
	appUptime                prometheus.Gauge
	appStartTime             prometheus.Gauge

	// Business metrics
	userCreatedTotal         *prometheus.CounterVec
	userValidationTotal      *prometheus.CounterVec
	userValidationDuration   *prometheus.HistogramVec
	featureFlagChecksTotal   *prometheus.CounterVec
	configReloadsTotal       *prometheus.CounterVec

	// Database metrics
	dbConnectionsActive      prometheus.Gauge
	dbConnectionsIdle        prometheus.Gauge
	dbQueriesTotal           *prometheus.CounterVec
	dbQueryDuration          *prometheus.HistogramVec
	dbTransactionsTotal      *prometheus.CounterVec

	// System metrics
	goGoroutines             prometheus.Gauge
	goMemoryUsage            *prometheus.GaugeVec
	goGCDuration             prometheus.Gauge
	goGCRuns                 prometheus.Counter

	// SLA/SLI metrics
	slaResponseTime          *prometheus.HistogramVec
	slaAvailability          *prometheus.GaugeVec
	slaErrorBudget           *prometheus.GaugeVec
	slaErrorBudgetBurn       *prometheus.GaugeVec
}

// NewPrometheusMetrics creates a new Prometheus metrics collector
func NewPrometheusMetrics(cfg *config.Config, logger *slog.Logger) *PrometheusMetrics {
	pm := &PrometheusMetrics{
		config: cfg,
		logger: logger,
	}

	pm.initMetrics()
	pm.registerMetrics()
	
	return pm
}

// initMetrics initializes all Prometheus metrics
func (pm *PrometheusMetrics) initMetrics() {
	// HTTP metrics
	pm.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	pm.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint", "status_code"},
	)

	pm.httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"method", "endpoint"},
	)

	pm.httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 6),
		},
		[]string{"method", "endpoint", "status_code"},
	)

	pm.httpActiveRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_active",
			Help: "Number of active HTTP requests",
		},
	)

	pm.httpErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"method", "endpoint", "error_type", "status_code"},
	)

	// Application metrics
	pm.appInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application information",
		},
		[]string{"version", "environment", "build_time", "git_commit"},
	)

	pm.appUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_uptime_seconds",
			Help: "Application uptime in seconds",
		},
	)

	pm.appStartTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_start_time_seconds",
			Help: "Application start time in Unix timestamp",
		},
	)

	// Business metrics
	pm.userCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_created_total",
			Help: "Total number of users created",
		},
		[]string{"status", "validation_status"},
	)

	pm.userValidationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_validation_total",
			Help: "Total number of user validation attempts",
		},
		[]string{"status", "field"},
	)

	pm.userValidationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "user_validation_duration_seconds",
			Help:    "User validation duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5},
		},
		[]string{"status", "field"},
	)

	pm.featureFlagChecksTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "feature_flag_checks_total",
			Help: "Total number of feature flag checks",
		},
		[]string{"flag_name", "result", "user_segment"},
	)

	pm.configReloadsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "config_reloads_total",
			Help: "Total number of configuration reloads",
		},
		[]string{"status", "source"},
	)

	// Database metrics
	pm.dbConnectionsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	pm.dbConnectionsIdle = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	pm.dbQueriesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	pm.dbQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"operation", "table", "status"},
	)

	pm.dbTransactionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_transactions_total",
			Help: "Total number of database transactions",
		},
		[]string{"status"},
	)

	// System metrics
	pm.goGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_goroutines",
			Help: "Number of goroutines that currently exist",
		},
	)

	pm.goMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "go_memory_usage_bytes",
			Help: "Go memory usage in bytes",
		},
		[]string{"type"},
	)

	pm.goGCDuration = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_gc_duration_seconds",
			Help: "Time spent in garbage collection",
		},
	)

	pm.goGCRuns = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "go_gc_runs_total",
			Help: "Total number of garbage collection runs",
		},
	)

	// SLA/SLI metrics
	pm.slaResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sla_response_time_seconds",
			Help:    "SLA response time tracking",
			Buckets: []float64{0.1, 0.2, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"service", "endpoint", "sla_tier"},
	)

	pm.slaAvailability = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sla_availability_ratio",
			Help: "SLA availability ratio (0-1)",
		},
		[]string{"service", "sla_tier"},
	)

	pm.slaErrorBudget = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sla_error_budget_ratio",
			Help: "SLA error budget remaining (0-1)",
		},
		[]string{"service", "sla_tier", "period"},
	)

	pm.slaErrorBudgetBurn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sla_error_budget_burn_rate",
			Help: "SLA error budget burn rate",
		},
		[]string{"service", "sla_tier", "window"},
	)
}

// registerMetrics registers all metrics with Prometheus
func (pm *PrometheusMetrics) registerMetrics() {
	prometheus.MustRegister(
		// HTTP metrics
		pm.httpRequestsTotal,
		pm.httpRequestDuration,
		pm.httpRequestSize,
		pm.httpResponseSize,
		pm.httpActiveRequests,
		pm.httpErrorsTotal,

		// Application metrics
		pm.appInfo,
		pm.appUptime,
		pm.appStartTime,

		// Business metrics
		pm.userCreatedTotal,
		pm.userValidationTotal,
		pm.userValidationDuration,
		pm.featureFlagChecksTotal,
		pm.configReloadsTotal,

		// Database metrics
		pm.dbConnectionsActive,
		pm.dbConnectionsIdle,
		pm.dbQueriesTotal,
		pm.dbQueryDuration,
		pm.dbTransactionsTotal,

		// System metrics
		pm.goGoroutines,
		pm.goMemoryUsage,
		pm.goGCDuration,
		pm.goGCRuns,

		// SLA/SLI metrics
		pm.slaResponseTime,
		pm.slaAvailability,
		pm.slaErrorBudget,
		pm.slaErrorBudgetBurn,
	)

	// Set initial app info
	pm.appInfo.WithLabelValues(
		pm.config.App.Version,
		pm.config.App.Environment,
		"", // build_time - would be set during build
		"", // git_commit - would be set during build
	).Set(1)

	pm.appStartTime.Set(float64(time.Now().Unix()))
}

// Start starts the Prometheus metrics server
func (pm *PrometheusMetrics) Start(ctx context.Context) error {
	if !pm.config.Observability.Exporters.Prometheus.Enabled {
		pm.logger.Info("Prometheus metrics server disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(pm.config.Observability.Exporters.Prometheus.Path, promhttp.Handler())
	
	// Add health check for metrics server
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	pm.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", pm.config.Observability.Exporters.Prometheus.Port),
		Handler: mux,
	}

	go func() {
		pm.logger.Info("Starting Prometheus metrics server",
			"port", pm.config.Observability.Exporters.Prometheus.Port,
			"path", pm.config.Observability.Exporters.Prometheus.Path,
		)
		
		if err := pm.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			pm.logger.Error("Prometheus metrics server failed", "error", err)
		}
	}()

	// Start metrics collection goroutines
	pm.startMetricsCollection(ctx)

	return nil
}

// Stop stops the Prometheus metrics server
func (pm *PrometheusMetrics) Stop(ctx context.Context) error {
	if pm.server == nil {
		return nil
	}

	pm.logger.Info("Stopping Prometheus metrics server")
	return pm.server.Shutdown(ctx)
}

// startMetricsCollection starts background goroutines for collecting metrics
func (pm *PrometheusMetrics) startMetricsCollection(ctx context.Context) {
	// Update uptime every 30 seconds
	go func() {
		startTime := time.Now()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pm.appUptime.Set(time.Since(startTime).Seconds())
			}
		}
	}()

	// Update Go runtime metrics every 15 seconds
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pm.updateGoMetrics()
			}
		}
	}()
}

// updateGoMetrics updates Go runtime metrics
func (pm *PrometheusMetrics) updateGoMetrics() {
	// This would typically use runtime.ReadMemStats()
	// For now, we'll use placeholder values
	pm.goGoroutines.Set(100) // placeholder
	pm.goMemoryUsage.WithLabelValues("heap").Set(1024 * 1024) // placeholder
	pm.goMemoryUsage.WithLabelValues("stack").Set(512 * 1024) // placeholder
}

// HTTPMiddleware creates a Gin middleware for HTTP metrics collection
func (pm *PrometheusMetrics) HTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !pm.config.Observability.Exporters.Prometheus.Enabled {
			c.Next()
			return
		}

		start := time.Now()
		pm.httpActiveRequests.Inc()

		// Record request size
		if c.Request.ContentLength > 0 {
			pm.httpRequestSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Record metrics after request
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		isSuccess := c.Writer.Status() < 400
		
		pm.httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()

		pm.httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Observe(duration)

		// Record response size
		pm.httpResponseSize.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Observe(float64(c.Writer.Size()))

		// Record SLA response time
		var slaTier string
		switch {
		case duration <= 0.2:
			slaTier = "gold"
		case duration <= 1.0:
			slaTier = "silver"
		default:
			slaTier = "bronze"
		}
		
		pm.slaResponseTime.WithLabelValues(
			"template-arch-lint",
			c.FullPath(),
			slaTier,
		).Observe(duration)

		// Record errors
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			
			pm.httpErrorsTotal.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
				errorType,
				statusCode,
			).Inc()
		}

		pm.httpActiveRequests.Dec()

		// Store SLA data in context for SLA tracker to use
		c.Set("sla_response_time", duration)
		c.Set("sla_success", isSuccess)
		c.Set("sla_endpoint", c.FullPath())
	}
}

// RecordUserCreated records user creation metrics
func (pm *PrometheusMetrics) RecordUserCreated(status, validationStatus string) {
	if !pm.config.Observability.Exporters.Prometheus.Enabled {
		return
	}
	
	pm.userCreatedTotal.WithLabelValues(status, validationStatus).Inc()
}

// RecordUserValidation records user validation metrics
func (pm *PrometheusMetrics) RecordUserValidation(field, status string, duration time.Duration) {
	if !pm.config.Observability.Exporters.Prometheus.Enabled {
		return
	}
	
	pm.userValidationTotal.WithLabelValues(status, field).Inc()
	pm.userValidationDuration.WithLabelValues(status, field).Observe(duration.Seconds())
}

// RecordFeatureFlagCheck records feature flag check metrics
func (pm *PrometheusMetrics) RecordFeatureFlagCheck(flagName, result, userSegment string) {
	if !pm.config.Observability.Exporters.Prometheus.Enabled {
		return
	}
	
	pm.featureFlagChecksTotal.WithLabelValues(flagName, result, userSegment).Inc()
}

// RecordConfigReload records configuration reload metrics
func (pm *PrometheusMetrics) RecordConfigReload(status, source string) {
	if !pm.config.Observability.Exporters.Prometheus.Enabled {
		return
	}
	
	pm.configReloadsTotal.WithLabelValues(status, source).Inc()
}

// RecordDatabaseQuery records database query metrics
func (pm *PrometheusMetrics) RecordDatabaseQuery(operation, table, status string, duration time.Duration) {
	if !pm.config.Observability.Exporters.Prometheus.Enabled {
		return
	}
	
	pm.dbQueriesTotal.WithLabelValues(operation, table, status).Inc()
	pm.dbQueryDuration.WithLabelValues(operation, table, status).Observe(duration.Seconds())
}

// UpdateSLAMetrics updates SLA/SLI metrics
func (pm *PrometheusMetrics) UpdateSLAMetrics(service, slaTier string, availability, errorBudget, burnRate float64) {
	if !pm.config.Observability.Exporters.Prometheus.Enabled {
		return
	}
	
	pm.slaAvailability.WithLabelValues(service, slaTier).Set(availability)
	pm.slaErrorBudget.WithLabelValues(service, slaTier, "30d").Set(errorBudget)
	pm.slaErrorBudgetBurn.WithLabelValues(service, slaTier, "1h").Set(burnRate)
}