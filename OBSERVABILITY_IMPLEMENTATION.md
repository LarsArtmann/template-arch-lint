# OpenTelemetry Observability Implementation

## Overview

This document describes the comprehensive OpenTelemetry observability implementation added to the template-arch-lint project. The implementation provides enterprise-grade monitoring, tracing, and metrics collection with seamless integration into existing functional programming patterns.

## Features Implemented

### 1. OpenTelemetry Core Infrastructure

#### Dependencies Added
- `go.opentelemetry.io/otel` - Core OpenTelemetry SDK
- `go.opentelemetry.io/otel/sdk` - OpenTelemetry SDK
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp` - OTLP trace exporter
- `go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp` - OTLP metrics exporter  
- `go.opentelemetry.io/otel/exporters/prometheus` - Prometheus metrics exporter
- `go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin` - Gin instrumentation

#### Configuration System
Extended the existing configuration system in `internal/config/config.go` with comprehensive observability settings:

```yaml
observability:
  enabled: true
  service_name: "template-arch-lint"
  service_version: "1.0.0"
  environment: "development"
  sampling_rate: 1.0
  
  tracing:
    enabled: true
    endpoint: "http://localhost:4318/v1/traces"
    http_details: true
    db_queries: true
  
  metrics:
    enabled: true
    endpoint: "http://localhost:4318/v1/metrics"
    push_interval: "15s"
    business_metrics: true
  
  exporters:
    prometheus:
      enabled: true
      port: 2112
      path: "/metrics"
    
    otlp:
      enabled: true
      endpoint: "http://localhost:4318"
      insecure: true
    
    jaeger:
      enabled: false
      endpoint: "http://localhost:14268/api/traces"
```

### 2. Observability Package Structure

#### `/internal/observability/otel.go`
- `OTELManager` - Manages OpenTelemetry providers and exporters
- Resource creation with service information
- Tracer and metrics provider setup
- Multiple exporter support (OTLP, Prometheus, Jaeger)
- Graceful shutdown handling

#### `/internal/observability/middleware.go`
- `TracingMiddleware` - HTTP request tracing with OpenTelemetry
- `MetricsMiddleware` - HTTP metrics collection (request count, duration, size)
- `CorrelationIDMiddleware` - Request correlation tracking
- `StructuredLoggingWithTracingMiddleware` - Enhanced logging with trace context
- `FunctionalProgrammingMetricsMiddleware` - Tracks functional programming operations

#### `/internal/observability/business_metrics.go`
- `BusinessMetrics` - Business-specific metrics collection
- User lifecycle metrics (created, updated, deleted, validated)
- Functional programming pattern metrics (Result, Option, Either patterns)
- Lo operation tracking (Map, Filter, Reduce operations)
- Performance monitoring for functional operations

#### `/internal/observability/database.go`
- `DatabaseTracer` - Database operation tracing
- `TracedDB` - Wrapper for sql.DB with automatic tracing
- `TracedTx` - Transaction tracing
- SQLC integration utilities
- Connection pool monitoring

#### `/internal/observability/health.go`
- `HealthChecker` - Comprehensive health checking
- Liveness endpoint (`/health/live`)
- Readiness endpoint (`/health/ready`) 
- Detailed health endpoint (`/health`) with component status
- Version endpoint (`/version`)
- Integration with business metrics

### 3. HTTP Middleware Integration

The middleware stack includes:
1. `gin.Recovery()` - Panic recovery
2. `TracingMiddleware` - Distributed tracing
3. `MetricsMiddleware` - HTTP metrics
4. `CorrelationIDMiddleware` - Request correlation
5. `StructuredLoggingWithTracingMiddleware` - Enhanced logging
6. `FunctionalProgrammingMetricsMiddleware` - FP operation tracking

### 4. Functional Programming Integration

#### Result Pattern Monitoring
- Tracks success/failure rates of Result pattern operations
- Records operation types and outcomes
- Integrates with tracing spans

#### Lo Operations Tracking  
- Monitors Map, Filter, Reduce, and other lo operations
- Tracks item counts and execution duration
- Performance monitoring for functional transformations

#### Business Logic Metrics
- User creation/update/deletion rates
- Validation error tracking
- Domain-specific operation monitoring
- Email domain analysis

### 5. Health Check System

#### Endpoints Provided
- `GET /health/live` - Simple liveness check
- `GET /health/ready` - Readiness check with dependencies
- `GET /health` - Comprehensive health with component status
- `GET /version` - Service version information

#### Health Monitoring
- Database connectivity checks
- OpenTelemetry system status
- System resource monitoring (memory, goroutines)
- Business metrics summary

### 6. Container Integration

Updated dependency injection container (`internal/container/container.go`):
- Registers all observability components
- Proper initialization order
- Graceful shutdown handling
- Error resilience (observability failures don't crash the app)

### 7. Handler Integration

Enhanced user handlers with observability:
- Business operation tracing
- User lifecycle metrics
- Validation tracking
- Functional programming operation monitoring

## Usage Examples

### Starting the Server
```bash
./server
```

The server will:
1. Initialize OpenTelemetry with configured exporters
2. Start Prometheus metrics server on port 2112
3. Begin collecting traces and metrics
4. Serve health check endpoints

### Health Checks
```bash
# Liveness check
curl http://localhost:8080/health/live

# Readiness check  
curl http://localhost:8080/health/ready

# Comprehensive health
curl http://localhost:8080/health

# Version info
curl http://localhost:8080/version
```

### Metrics
```bash
# Prometheus metrics
curl http://localhost:2112/metrics
```

### Testing API Endpoints
```bash
# Create user (triggers business metrics)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"id":"test-123","email":"test@example.com","name":"Test User"}'

# Get user stats (functional programming metrics)
curl http://localhost:8080/api/v1/users/stats
```

## Monitoring Stack Integration

### Prometheus + Grafana
1. Configure Prometheus to scrape `http://localhost:2112/metrics`
2. Import dashboards for HTTP, database, and business metrics
3. Set up alerts for error rates and performance thresholds

### OTLP Collector
1. Configure OTEL Collector to receive on `http://localhost:4318`
2. Export to your observability backend (Jaeger, DataDog, etc.)
3. Set up distributed tracing visualization

### Example Collector Configuration
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317
      http:
        endpoint: 0.0.0.0:4318

exporters:
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [jaeger]
    metrics:
      receivers: [otlp]
      exporters: [prometheus]
```

## Metrics Collected

### HTTP Metrics
- `http_requests_total` - Total HTTP requests by method, route, status
- `http_request_duration_seconds` - Request duration histogram
- `http_request_size_bytes` - Request size histogram
- `http_response_size_bytes` - Response size histogram

### Business Metrics
- `users_created_total` - Total users created
- `users_updated_total` - Total users updated  
- `users_deleted_total` - Total users deleted
- `user_validations_total` - Total validations performed
- `result_pattern_operations_total` - Result pattern usage
- `lo_operations_total` - Lo functional operations
- `business_operations_total` - Business operation counts
- `validation_errors_total` - Validation error counts

### Database Metrics
- `db_queries_total` - Total database queries
- `db_query_duration_seconds` - Query duration histogram
- `db_connections_active` - Active database connections
- `db_errors_total` - Database error counts

### Health Metrics
- `health_checks_total` - Total health checks performed
- `health_check_duration_seconds` - Health check duration
- `system_health_status` - Current system health status
- `component_health_status` - Component health status

## Performance Considerations

### Sampling
- Default sampling rate is 100% for development
- Configure lower rates for production (e.g., 0.1 = 10%)
- Head-based sampling on trace ID

### Resource Usage
- Prometheus metrics server uses minimal resources
- OTLP exporters batch data for efficiency
- Graceful degradation if observability backends are unavailable

### Zero Performance Impact Option
All observability can be disabled via configuration:
```yaml
observability:
  enabled: false
```

## Future Enhancements

### Planned Features
1. Custom span attributes for business context
2. Distributed tracing across service boundaries  
3. Advanced metrics aggregation
4. Alerting rule templates
5. Dashboard templates for common use cases
6. Integration with external APM tools

### SQLC Integration
The database tracing is designed to integrate seamlessly with SQLC-generated code:
- Wrapper functions for generated queries
- Automatic parameter sanitization
- Query performance tracking
- Transaction boundary monitoring

## Troubleshooting

### Common Issues
1. **Port conflicts** - Prometheus port 2112 already in use
   - Solution: Change `observability.exporters.prometheus.port` in config

2. **OTLP connection failed** - Collector not running
   - Solution: Start OTEL Collector or disable OTLP exporter

3. **High memory usage** - Too many metrics/traces
   - Solution: Reduce sampling rate or disable detailed tracing

### Debug Mode
Enable debug logging to see observability operations:
```yaml
logging:
  level: "debug"
```

## Conclusion

This comprehensive OpenTelemetry implementation provides:
- ✅ Production-ready distributed tracing
- ✅ Comprehensive metrics collection  
- ✅ Health monitoring and alerting
- ✅ Functional programming pattern observability
- ✅ Zero-downtime graceful shutdown
- ✅ Configurable and extensible architecture
- ✅ Integration with existing Go patterns

The implementation follows OpenTelemetry best practices and provides a solid foundation for monitoring microservices in production environments.