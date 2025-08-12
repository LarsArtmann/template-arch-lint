# Configuration Management System

This document provides comprehensive documentation for the template-arch-lint configuration management system, including environment-specific configurations, feature flags, secrets management, and drift detection.

## Table of Contents

- [Overview](#overview)
- [Configuration Structure](#configuration-structure)
- [Environment-Specific Configurations](#environment-specific-configurations)
- [Feature Flags](#feature-flags)
- [Secrets Management](#secrets-management)
- [Configuration Drift Detection](#configuration-drift-detection)
- [Runtime Configuration Management](#runtime-configuration-management)
- [Best Practices](#best-practices)
- [Examples](#examples)
- [API Reference](#api-reference)
- [Troubleshooting](#troubleshooting)

## Overview

The configuration management system provides:

- **Environment-specific configurations** with validation
- **Feature flags** with conditional logic and runtime updates
- **Secure secrets management** with multiple provider support
- **Configuration drift detection** with alerting
- **Hot-reloading** capabilities for development
- **Runtime configuration management** via HTTP APIs

## Configuration Structure

### Core Configuration Sections

```yaml
# Server configuration
server:
  host: "localhost"
  port: 8080
  read_timeout: "5s"
  write_timeout: "10s"
  idle_timeout: "60s"
  graceful_shutdown_timeout: "30s"

# Database configuration
database:
  driver: "sqlite3"
  dsn: "./app.db"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "5m"
  conn_max_idle_time: "5m"

# Logging configuration
logging:
  level: "info"
  format: "json"
  output: "stdout"

# Application configuration
app:
  name: "template-arch-lint"
  version: "1.0.0"
  environment: "development"
  debug: false

# Feature flags
features:
  enable_debug_endpoints: false
  enable_profiling: false
  enable_hot_reload: true
  enable_cors: true
  enable_request_logging: false
  enable_query_logging: false
  enable_metrics_debug: false
  enable_beta_features: false
  enable_load_testing: false

# Security configuration
security:
  enable_auth: false
  enable_rate_limiting: false
  cors:
    allowed_origins: ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["*"]
    allow_credentials: false
  rate_limit:
    requests_per_minute: 60
    burst: 10
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
    min_version: "1.2"
  api_keys:
    enabled: false
    header_name: "X-API-Key"
    rotate_interval: "24h"

# Health check configuration
health:
  enabled: true
  endpoint: "/health"
  timeout: "5s"
  check_database: true
  check_external_services: false

# Cache configuration
cache:
  enabled: false
  redis_url: ""
  default_ttl: "1h"
  max_memory: "256mb"
  cluster_mode: false

# External services configuration
external:
  api_timeout: "30s"
  retry_attempts: 3
  circuit_breaker:
    enabled: true
    threshold: 5
    timeout: "60s"

# Backup configuration
backup:
  enabled: false
  schedule: "0 2 * * *"
  retention_days: 30
  s3_bucket: ""

# Resource limits
resources:
  max_memory: "512MB"
  max_cpu_cores: 1
  max_connections: 1000
  request_timeout: "30s"
  max_request_size: "10MB"

# Observability configuration
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

## Environment-Specific Configurations

### Available Environments

The system supports the following environments:

- **development** - Local development with debug features
- **staging** - Pre-production testing environment
- **production** - Production environment with security and performance optimizations
- **testing** - Automated testing with minimal dependencies
- **local** - Individual developer customizations

### Environment Configuration Files

Configuration files are located in the `configs/` directory:

```
configs/
├── development.yaml  # Development environment
├── staging.yaml      # Staging environment
├── production.yaml   # Production environment
├── testing.yaml      # Testing environment
└── local.yaml        # Local development overrides
```

### Loading Environment-Specific Configuration

```go
// Load configuration for specific environment
config, err := config.LoadConfigForEnvironment("production")
if err != nil {
    log.Fatal(err)
}

// Load with overrides
config, err := config.LoadConfigWithOverrides("production", "local")
if err != nil {
    log.Fatal(err)
}
```

### Environment Variable Override

All configuration values can be overridden using environment variables with the `APP_` prefix:

```bash
# Override server port
export APP_SERVER_PORT=3000

# Override database DSN
export APP_DATABASE_DSN="postgres://user:pass@localhost/db"

# Override feature flags
export APP_FEATURES_ENABLE_DEBUG_ENDPOINTS=true

# Override security settings
export APP_SECURITY_ENABLE_AUTH=true
export APP_SECURITY_TLS_ENABLED=true
```

## Feature Flags

### Basic Feature Flags

Feature flags are configured in the `features` section:

```yaml
features:
  enable_debug_endpoints: false
  enable_profiling: false
  enable_hot_reload: true
  enable_cors: true
  enable_request_logging: false
  enable_query_logging: false
  enable_metrics_debug: false
  enable_beta_features: false
  enable_load_testing: false
```

### Using Feature Flags in Code

```go
// Initialize feature manager
featureManager := config.NewFeatureManager(&config.Features)

// Check if feature is enabled
if featureManager.IsEnabled("debug_endpoints") {
    // Enable debug endpoints
    mux.HandleFunc("/debug/pprof/", pprof.Index)
}

// Check with context
ctx := config.FeatureContext{
    UserID:      "user123",
    Environment: "production",
    Timestamp:   time.Now(),
}

if featureManager.IsEnabledForContext("beta_features", ctx) {
    // Enable beta features for this user
}
```

### Advanced Feature Flags with Conditions

```go
// Create a feature flag with conditions
flag := &config.FeatureFlag{
    Name:        "new_ui",
    Enabled:     true,
    Description: "Enable new user interface",
    Conditions: []config.FeatureCondition{
        {
            Type:     "percentage",
            Operator: "percentage",
            Values:   []interface{}{50.0}, // 50% rollout
        },
        {
            Type:     "environment",
            Operator: "in",
            Values:   []interface{}{"development", "staging"},
        },
    },
}

// Update the flag
featureManager.UpdateFlag(flag)
```

### Feature Flag HTTP Middleware

```go
// Create feature gate middleware
configMiddleware := config.NewConfigMiddleware(reloadableConfig)

// Protect endpoint with feature flag
mux.Handle("/api/beta", configMiddleware.FeatureGate("beta_features")(betaHandler))

// Conditional feature gate with context
contextFunc := func(r *http.Request) config.FeatureContext {
    return config.FeatureContext{
        UserID:      r.Header.Get("X-User-ID"),
        Environment: "production",
        Timestamp:   time.Now(),
    }
}
mux.Handle("/api/new-ui", configMiddleware.ConditionalFeatureGate("new_ui", contextFunc)(newUIHandler))
```

## Secrets Management

### Supported Providers

The secrets management system supports multiple providers:

- **Environment Variables** - System environment variables
- **File-based** - JSON/YAML/ENV files with optional encryption
- **HashiCorp Vault** - Enterprise secrets management
- **Kubernetes Secrets** - Native Kubernetes secret store

### Secrets Configuration

```yaml
# Secrets management configuration
secrets:
  provider: "env"  # env, file, vault, kubernetes
  cache_ttl: "5m"
  
  # Vault configuration
  vault:
    address: "https://vault.example.com"
    token_file: "/var/secrets/vault-token"
    mount: "secret"
    path: "myapp"
    namespace: "production"
    
  # Kubernetes configuration
  kubernetes:
    namespace: "default"
    secret_name: "app-secrets"
    in_cluster: true
    
  # File configuration
  file:
    path: "/etc/secrets/app.json"
    format: "json"
    encrypted: false
```

### Using Secrets in Configuration

Use the `${}` syntax to reference secrets in configuration files:

```yaml
database:
  dsn: "${DATABASE_URL}"

observability:
  exporters:
    otlp:
      headers:
        Authorization: "Bearer ${OTEL_AUTH_TOKEN}"

security:
  tls:
    cert_file: "${TLS_CERT_FILE}"
    key_file: "${TLS_KEY_FILE}"

cache:
  redis_url: "${REDIS_URL}"
```

### Programmatic Secrets Access

```go
// Initialize secrets manager
secretConfig := config.SecretConfig{
    Provider: "vault",
    VaultConfig: config.VaultConfig{
        Address: "https://vault.example.com",
        Token:   "vault-token",
        Mount:   "secret",
        Path:    "myapp",
    },
}

secretsManager, err := config.NewSecretsManager(secretConfig)
if err != nil {
    log.Fatal(err)
}
defer secretsManager.Close()

// Get a secret
ctx := context.Background()
dbPassword, err := secretsManager.GetSecret(ctx, "db_password")
if err != nil {
    log.Fatal(err)
}

// Set a secret
err = secretsManager.SetSecret(ctx, "api_key", "new-api-key-value")
if err != nil {
    log.Fatal(err)
}

// Expand secrets in configuration
err = config.ExpandSecrets(config, secretsManager)
if err != nil {
    log.Fatal(err)
}
```

### Secrets HTTP API

```bash
# Check if a secret exists (returns metadata only)
curl -X GET "http://localhost:8080/api/secrets/get?key=api_key"

# Set a secret
curl -X POST "http://localhost:8080/api/secrets/set" \
  -H "Content-Type: application/json" \
  -d '{"key": "api_key", "value": "secret-value"}'

# Delete a secret
curl -X DELETE "http://localhost:8080/api/secrets/delete?key=api_key"

# Rotate a secret
curl -X POST "http://localhost:8080/api/secrets/rotate?key=api_key"

# Check secrets health
curl -X GET "http://localhost:8080/api/secrets/health"
```

## Configuration Drift Detection

### Overview

Configuration drift detection monitors changes to your application configuration and alerts when unexpected changes occur.

### Setting Up Drift Detection

```go
// Create drift detector
driftDetector := config.NewDriftDetector(
    "production-app",
    reloadableConfig,
    config.WithCheckInterval(30*time.Second),
    config.WithAlertThreshold(5*time.Minute),
    config.WithAlerter(config.NewLogAlerter()),
    config.WithAlerter(config.NewEmailAlerter("smtp.example.com", "alerts@example.com", []string{"admin@example.com"})),
    config.WithAlerter(config.NewSlackAlerter("https://hooks.slack.com/...", "#alerts")),
)

// Start monitoring
err := driftDetector.Start()
if err != nil {
    log.Fatal(err)
}
defer driftDetector.Stop()
```

### Drift Detection Severity Levels

- **Critical** - Changes to database connection, environment, or security settings
- **High** - Changes to server configuration, logging, or multiple features
- **Medium** - Multiple minor changes or moderate impact changes
- **Low** - Single minor changes with low impact

### Drift Detection HTTP API

```bash
# Get drift detection status
curl -X GET "http://localhost:8080/api/drift/status"

# Get drift history
curl -X GET "http://localhost:8080/api/drift/history?limit=10&severity=high"

# Get current baseline
curl -X GET "http://localhost:8080/api/drift/baseline"

# Update baseline
curl -X POST "http://localhost:8080/api/drift/baseline/update"

# Trigger manual check
curl -X POST "http://localhost:8080/api/drift/check"

# Get drift statistics
curl -X GET "http://localhost:8080/api/drift/stats"

# Get recent alerts
curl -X GET "http://localhost:8080/api/drift/alerts"

# Check system health
curl -X GET "http://localhost:8080/api/drift/health"

# Export baseline
curl -X GET "http://localhost:8080/api/drift/baseline/export" -o baseline.json

# Compare snapshots
curl -X POST "http://localhost:8080/api/drift/compare" \
  -H "Content-Type: application/json" \
  -d '{
    "old_snapshot": {...},
    "new_snapshot": {...}
  }'
```

## Runtime Configuration Management

### Hot Reloading

Enable hot reloading in development:

```yaml
features:
  enable_hot_reload: true
```

```go
// Create reloadable configuration
reloadableConfig, err := config.NewReloadableConfig("configs/development.yaml")
if err != nil {
    log.Fatal(err)
}
defer reloadableConfig.Close()

// Subscribe to configuration changes
configChanges := make(chan config.ConfigChange, 10)
reloadableConfig.Subscribe(configChanges)

go func() {
    for change := range configChanges {
        fmt.Printf("Configuration changed: %s\n", change.Type)
        if len(change.Differences) > 0 {
            fmt.Printf("  %d changes detected\n", len(change.Differences))
        }
    }
}()
```

### Configuration HTTP API

```bash
# Get current configuration
curl -X GET "http://localhost:8080/api/config/"

# Reload configuration
curl -X POST "http://localhost:8080/api/config/reload"

# Validate configuration
curl -X POST "http://localhost:8080/api/config/validate"

# Get configuration statistics
curl -X GET "http://localhost:8080/api/config/stats"

# Get all feature flags
curl -X GET "http://localhost:8080/api/config/flags"

# Get specific feature flag
curl -X GET "http://localhost:8080/api/config/flags/get?name=debug_endpoints"

# Check feature flag for context
curl -X POST "http://localhost:8080/api/config/flags/check" \
  -H "Content-Type: application/json" \
  -d '{
    "flag_name": "beta_features",
    "context": {
      "user_id": "user123",
      "environment": "production",
      "timestamp": "2023-10-01T12:00:00Z"
    }
  }'

# Update feature flag
curl -X PUT "http://localhost:8080/api/config/flags/update" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "new_feature",
    "enabled": true,
    "description": "Enable new feature",
    "conditions": [
      {
        "type": "environment",
        "operator": "equals",
        "values": ["development"]
      }
    ]
  }'

# Toggle feature flag
curl -X POST "http://localhost:8080/api/config/flags/toggle?name=debug_endpoints&enabled=true"
```

## Best Practices

### Configuration Security

1. **Never commit secrets** to version control
2. **Use environment variables** or external secret stores for sensitive data
3. **Enable TLS** in production environments
4. **Rotate secrets regularly** using automated processes
5. **Validate configuration** before deployment
6. **Monitor configuration changes** with drift detection

### Environment Management

1. **Use environment-specific configurations** for different deployment stages
2. **Validate configurations** against schemas
3. **Test configuration changes** in staging before production
4. **Document configuration changes** in commit messages
5. **Use feature flags** for gradual rollouts

### Feature Flags

1. **Use descriptive names** for feature flags
2. **Document feature flags** with clear descriptions
3. **Clean up unused flags** regularly
4. **Use conditions** for complex rollout scenarios
5. **Monitor flag usage** and performance impact

### Operational Excellence

1. **Monitor configuration drift** in production
2. **Set up alerts** for critical configuration changes
3. **Maintain configuration baselines** for rollback scenarios
4. **Test hot-reloading** in development environments
5. **Regular backup** of configuration snapshots

## Examples

### Example 1: Development Setup

```yaml
# configs/development.yaml
server:
  host: "localhost"
  port: 3000

database:
  driver: "sqlite3"
  dsn: "./dev.db"

logging:
  level: "debug"
  format: "text"
  output: "stdout"

app:
  environment: "development"
  debug: true

features:
  enable_debug_endpoints: true
  enable_profiling: true
  enable_hot_reload: true
  enable_cors: true
  enable_request_logging: true
  enable_query_logging: true

observability:
  enabled: true
  sampling_rate: 1.0
```

### Example 2: Production Setup

```yaml
# configs/production.yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "postgres"
  dsn: "${DATABASE_URL}"

logging:
  level: "warn"
  format: "json"
  output: "stdout"

app:
  environment: "production"
  debug: false

features:
  enable_debug_endpoints: false
  enable_profiling: false
  enable_hot_reload: false
  enable_cors: true

security:
  enable_auth: true
  enable_rate_limiting: true
  tls:
    enabled: true
    cert_file: "${TLS_CERT_FILE}"
    key_file: "${TLS_KEY_FILE}"

observability:
  enabled: true
  sampling_rate: 0.1
```

### Example 3: Docker Environment Variables

```dockerfile
# Dockerfile
ENV APP_SERVER_HOST=0.0.0.0
ENV APP_SERVER_PORT=8080
ENV APP_DATABASE_DRIVER=postgres
ENV APP_DATABASE_DSN=${DATABASE_URL}
ENV APP_APP_ENVIRONMENT=production
ENV APP_SECURITY_ENABLE_AUTH=true
ENV APP_SECURITY_TLS_ENABLED=true
ENV APP_OBSERVABILITY_SAMPLING_RATE=0.1
```

### Example 4: Kubernetes ConfigMap

```yaml
# k8s-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  APP_SERVER_HOST: "0.0.0.0"
  APP_SERVER_PORT: "8080"
  APP_DATABASE_DRIVER: "postgres"
  APP_APP_ENVIRONMENT: "production"
  APP_SECURITY_ENABLE_AUTH: "true"
  APP_SECURITY_TLS_ENABLED: "true"
  APP_OBSERVABILITY_SAMPLING_RATE: "0.1"
```

### Example 5: Complete Go Application

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "your-project/internal/config"
)

func main() {
    // Load configuration
    reloadableConfig, err := config.NewReloadableConfig("")
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }
    defer reloadableConfig.Close()

    // Initialize secrets manager
    cfg := reloadableConfig.GetConfig()
    secretConfig := config.SecretConfig{
        Provider: "env",
        CacheTTL: 5 * time.Minute,
    }
    
    secretsManager, err := config.NewSecretsManager(secretConfig)
    if err != nil {
        log.Fatal("Failed to initialize secrets manager:", err)
    }
    defer secretsManager.Close()

    // Expand secrets in configuration
    if err := config.ExpandSecrets(cfg, secretsManager); err != nil {
        log.Fatal("Failed to expand secrets:", err)
    }

    // Initialize drift detection
    driftDetector := config.NewDriftDetector(
        "main-app",
        reloadableConfig,
        config.WithCheckInterval(30*time.Second),
        config.WithAlerter(config.NewLogAlerter()),
    )
    
    if err := driftDetector.Start(); err != nil {
        log.Fatal("Failed to start drift detector:", err)
    }
    defer driftDetector.Stop()

    // Setup HTTP routes
    mux := http.NewServeMux()
    
    // Register configuration management routes
    configHandler := config.NewRuntimeConfigHandler(reloadableConfig)
    configHandler.RegisterRoutes(mux, "/api/config")
    
    // Register secrets management routes
    secretsHandler := config.NewSecretsHandler(secretsManager)
    secretsHandler.RegisterRoutes(mux, "/api/secrets")
    
    // Register drift detection routes
    driftHandler := config.NewDriftHandler(driftDetector)
    driftHandler.RegisterRoutes(mux, "/api/drift")

    // Setup middleware
    configMiddleware := config.NewConfigMiddleware(reloadableConfig)
    
    // Apply configuration headers middleware
    handler := configMiddleware.ConfigHeaders()(mux)

    // Start server
    server := &http.Server{
        Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
        Handler: handler,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
    }

    // Graceful shutdown
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server failed to start:", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Shutdown server
    ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
    
    log.Println("Server exited")
}
```

## API Reference

### Configuration Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/config/` | Get current configuration |
| POST | `/api/config/reload` | Reload configuration |
| POST | `/api/config/validate` | Validate configuration |
| GET | `/api/config/stats` | Get configuration statistics |

### Feature Flag Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/config/flags` | List all feature flags |
| GET | `/api/config/flags/get` | Get specific feature flag |
| POST | `/api/config/flags/check` | Check feature flag for context |
| PUT | `/api/config/flags/update` | Update feature flag |
| POST | `/api/config/flags/toggle` | Toggle feature flag |

### Secrets Management Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/secrets/get` | Check secret existence |
| POST | `/api/secrets/set` | Set secret value |
| DELETE | `/api/secrets/delete` | Delete secret |
| POST | `/api/secrets/rotate` | Rotate secret |
| GET | `/api/secrets/health` | Check secrets health |

### Drift Detection Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/drift/status` | Get drift detection status |
| GET | `/api/drift/history` | Get drift history |
| GET | `/api/drift/baseline` | Get current baseline |
| POST | `/api/drift/baseline/update` | Update baseline |
| GET | `/api/drift/baseline/export` | Export baseline |
| POST | `/api/drift/check` | Trigger manual check |
| GET | `/api/drift/stats` | Get drift statistics |
| GET | `/api/drift/alerts` | Get recent alerts |
| GET | `/api/drift/health` | Check drift system health |
| POST | `/api/drift/compare` | Compare snapshots |

## Troubleshooting

### Common Issues

#### Configuration Not Loading

```bash
# Check if configuration file exists
ls -la configs/

# Check environment variables
env | grep APP_

# Validate configuration syntax
go run cmd/server/main.go --validate-config
```

#### Feature Flags Not Working

```bash
# Check feature manager status
curl -X GET "http://localhost:8080/api/config/stats"

# List all feature flags
curl -X GET "http://localhost:8080/api/config/flags"

# Check specific flag
curl -X GET "http://localhost:8080/api/config/flags/get?name=your_flag"
```

#### Secrets Not Found

```bash
# Check secrets provider health
curl -X GET "http://localhost:8080/api/secrets/health"

# Check environment variables
env | grep -E "(DATABASE_URL|API_KEY|TOKEN)"

# Test secret expansion
echo "Database URL: ${DATABASE_URL:-not-set}"
```

#### Configuration Drift Alerts

```bash
# Check drift detection status
curl -X GET "http://localhost:8080/api/drift/status"

# Check recent drift events
curl -X GET "http://localhost:8080/api/drift/history?limit=5"

# Update baseline if changes are legitimate
curl -X POST "http://localhost:8080/api/drift/baseline/update"
```

### Debugging Tips

1. **Enable debug logging** in development
2. **Use validation endpoints** to check configuration
3. **Monitor configuration changes** with drift detection
4. **Check environment variables** for overrides
5. **Verify secret providers** are accessible
6. **Test feature flags** with different contexts
7. **Export baselines** for comparison and rollback

### Performance Considerations

1. **Cache secrets** appropriately (default 5 minutes)
2. **Limit drift check frequency** (default 30 seconds)
3. **Use lower sampling rates** in production
4. **Minimize hot-reloading** in production
5. **Optimize feature flag conditions** for performance

---

For more information, see the source code in the `internal/config/` directory.