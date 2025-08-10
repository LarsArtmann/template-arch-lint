# Configuration Management

This project uses [Viper](https://github.com/spf13/viper) for configuration management, providing a flexible and production-ready configuration system.

## Configuration Sources

The configuration system loads settings from multiple sources in order of priority:

1. **Environment Variables** (highest priority)
2. **Configuration Files**
3. **Default Values** (lowest priority)

## Configuration Structure

The configuration is organized into four main sections:

### Server Configuration
- `server.host` - Server bind address (default: "localhost")
- `server.port` - Server port (default: 8080)
- `server.read_timeout` - HTTP read timeout (default: "5s")
- `server.write_timeout` - HTTP write timeout (default: "10s")
- `server.idle_timeout` - HTTP idle timeout (default: "60s")
- `server.graceful_shutdown_timeout` - Graceful shutdown timeout (default: "30s")

### Database Configuration
- `database.driver` - Database driver (sqlite, postgres, mysql)
- `database.dsn` - Database connection string
- `database.max_open_conns` - Maximum open connections (default: 25)
- `database.max_idle_conns` - Maximum idle connections (default: 5)
- `database.conn_max_lifetime` - Connection maximum lifetime (default: "5m")
- `database.conn_max_idle_time` - Connection maximum idle time (default: "5m")

### Logging Configuration
- `logging.level` - Log level (debug, info, warn, error)
- `logging.format` - Log format (json, text)
- `logging.output` - Log output destination (stdout, stderr, or file path)

### Application Configuration
- `app.name` - Application name
- `app.version` - Application version
- `app.environment` - Environment (development, staging, production)
- `app.debug` - Debug mode flag

## Configuration Files

### Default Configuration
The system looks for configuration files in the following locations:
- Current directory: `./config.yaml`
- System directory: `/etc/template-arch-lint/config.yaml`
- User directory: `$HOME/.template-arch-lint/config.yaml`

### Example Files
- `config.yaml` - Development configuration example
- `config.production.yaml` - Production configuration example
- `.env.example` - Environment variables example

## Environment Variables

All configuration values can be overridden using environment variables with the `APP_` prefix:

```bash
# Server configuration
APP_SERVER_HOST=0.0.0.0
APP_SERVER_PORT=8080

# Database configuration
APP_DATABASE_DRIVER=postgres
APP_DATABASE_DSN="postgres://user:pass@localhost/db"

# Logging configuration
APP_LOGGING_LEVEL=info
APP_LOGGING_FORMAT=json

# Application configuration
APP_APP_ENVIRONMENT=production
APP_APP_DEBUG=false
```

## Usage

### Loading Configuration

```go
import "github.com/LarsArtmann/template-arch-lint/internal/config"

func main() {
    // Load configuration with defaults and environment variables
    cfg, err := config.LoadConfig("")
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Use configuration
    fmt.Printf("Server will run on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
}
```

### Loading from Specific File

```go
// Load configuration from specific file
cfg, err := config.LoadConfig("/path/to/config.yaml")
```

## Validation

The configuration system includes comprehensive validation:

- **Type Safety**: All configuration values are type-safe
- **Required Fields**: Essential configuration fields are validated
- **Value Constraints**: Ports, log levels, and other values are validated
- **Clear Error Messages**: Validation errors provide clear guidance

## Production Recommendations

### Security
- Use environment variables for sensitive data (database passwords, API keys)
- Never commit sensitive configuration files to version control
- Use restricted file permissions for configuration files

### Performance  
- Use connection pooling settings appropriate for your load
- Configure appropriate timeouts for your use case
- Monitor and tune configuration based on metrics

### Reliability
- Always validate configuration on startup
- Use structured logging in production
- Configure graceful shutdown timeouts
- Set up health checks and monitoring

## Example Usage

Run the example application to see configuration in action:

```bash
# With default configuration
go run example/main.go

# With environment variable overrides
APP_SERVER_PORT=9090 APP_LOGGING_LEVEL=debug go run example/main.go

# With configuration file
go run example/main.go -config config.production.yaml
```