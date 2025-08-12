# Configuration Integration Example

This example demonstrates the comprehensive configuration management system with all features enabled:

- Environment-specific configuration loading
- Feature flags with runtime updates
- Secrets management with multiple providers
- Configuration drift detection and alerting
- Hot-reloading capabilities
- HTTP APIs for runtime management

## Running the Example

```bash
# From the project root directory
cd examples/config-integration

# Run with default development configuration
go run main.go

# Run with specific environment
APP_ENVIRONMENT=staging go run main.go

# Run with custom feature flags
APP_FEATURES_ENABLE_DEBUG_ENDPOINTS=true \
APP_FEATURES_ENABLE_BETA_FEATURES=true \
go run main.go
```

## Available Endpoints

Once running, the following endpoints are available:

### Configuration Management
- `GET /api/config/` - Get current configuration
- `POST /api/config/reload` - Reload configuration
- `POST /api/config/validate` - Validate configuration
- `GET /api/config/stats` - Get configuration statistics

### Feature Flags
- `GET /api/config/flags` - List all feature flags
- `GET /api/config/flags/get?name=debug_endpoints` - Get specific flag
- `POST /api/config/flags/toggle?name=debug_endpoints&enabled=true` - Toggle flag

### Secrets Management
- `GET /api/secrets/health` - Check secrets provider health
- `POST /api/secrets/set` - Set a secret (development only)

### Drift Detection
- `GET /api/drift/status` - Get drift detection status
- `GET /api/drift/history` - Get drift history
- `POST /api/drift/baseline/update` - Update configuration baseline

### Health and Demo
- `GET /health` - Application health check
- `GET /demo/feature-gated` - Feature flag protected endpoint
- `GET /demo/stable` - Configuration stability protected endpoint

## Testing Configuration Changes

### Test Hot-Reloading
1. Run the application: `go run main.go`
2. Modify `../../configs/development.yaml`
3. Watch the console for configuration change notifications

### Test Feature Flags
```bash
# Toggle debug endpoints
curl -X POST "http://localhost:8080/api/config/flags/toggle?name=debug_endpoints&enabled=true"

# Check if feature-gated endpoint becomes available
curl -X GET "http://localhost:8080/demo/feature-gated"
```

### Test Drift Detection
1. Start the application
2. Modify the configuration file
3. Check drift detection: `curl -X GET "http://localhost:8080/api/drift/history"`

### Test Secrets
```bash
# Check secrets health
curl -X GET "http://localhost:8080/api/secrets/health"

# Set a secret (if using file provider)
curl -X POST "http://localhost:8080/api/secrets/set" \
  -H "Content-Type: application/json" \
  -d '{"key": "test_secret", "value": "test_value"}'
```

## Configuration Sources

The example loads configuration from multiple sources in this order:
1. Environment-specific YAML file (`configs/{environment}.yaml`)
2. Environment variables (prefixed with `APP_`)
3. Default values defined in the code

## Environment Variables

Key environment variables for testing:

```bash
# Application settings
export APP_APP_ENVIRONMENT=development
export APP_APP_DEBUG=true
export APP_SERVER_PORT=3000

# Feature flags
export APP_FEATURES_ENABLE_DEBUG_ENDPOINTS=true
export APP_FEATURES_ENABLE_PROFILING=true
export APP_FEATURES_ENABLE_BETA_FEATURES=true

# Security settings
export APP_SECURITY_ENABLE_AUTH=false
export APP_SECURITY_ENABLE_RATE_LIMITING=false

# Database (if using PostgreSQL)
export APP_DATABASE_DRIVER=postgres
export APP_DATABASE_DSN="postgres://user:password@localhost:5432/testdb"

# Secrets (for demonstration)
export DATABASE_URL="postgres://demo:password@localhost:5432/demo_db"
export API_KEY="your-api-key-here"
export REDIS_URL="redis://localhost:6379/0"
```

## Monitoring and Observability

The example includes built-in monitoring capabilities:

- **Configuration Changes**: Watch console output for real-time configuration changes
- **Feature Flag Updates**: Monitor feature flag state changes
- **Drift Detection**: Track configuration drift with automatic alerting
- **Health Checks**: Multiple health check endpoints for different components

## Best Practices Demonstrated

1. **Environment Separation**: Different configurations for different environments
2. **Secret Security**: Secrets stored separately from configuration files
3. **Feature Flag Management**: Runtime feature toggles with conditions
4. **Configuration Validation**: Schema validation and business rule checking
5. **Drift Monitoring**: Automatic detection of unauthorized configuration changes
6. **Graceful Shutdown**: Proper cleanup of resources and connections

## Extending the Example

You can extend this example by:

1. **Adding more secrets providers** (Vault, Kubernetes)
2. **Implementing custom alerters** (email, Slack, PagerDuty)
3. **Adding more feature flag conditions** (user-based, time-based)
4. **Creating custom configuration validators**
5. **Adding more drift detection rules**

## Troubleshooting

### Configuration Not Loading
- Check file permissions on config files
- Verify YAML syntax with a validator
- Check environment variable names (must have `APP_` prefix)

### Feature Flags Not Working
- Check feature manager initialization
- Verify flag names match configuration
- Test with `/api/config/flags` endpoint

### Secrets Not Found
- Verify secrets.json file exists and has correct format
- Check environment variables for secret values
- Test with `/api/secrets/health` endpoint

### Drift Detection Not Working
- Check if hot-reload is enabled (`APP_FEATURES_ENABLE_HOT_RELOAD=true`)
- Verify drift detector is running (`/api/drift/status`)
- Check file permissions on configuration directories

For more information, see the main [Configuration Documentation](../../docs/CONFIGURATION.md).