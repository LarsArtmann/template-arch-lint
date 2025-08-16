// Package config provides configuration management for the application.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config represents the application configuration.
type Config struct {
	Server   ServerConfig   `mapstructure:"server" validate:"required"`
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	Logging  LoggingConfig  `mapstructure:"logging" validate:"required"`
	App      AppConfig      `mapstructure:"app" validate:"required"`
	JWT      JWTConfig      `mapstructure:"jwt" validate:"required"`
	Security SecurityConfig `mapstructure:"security"`
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	Host                    string        `mapstructure:"host" validate:"required"`
	Port                    int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}

// DatabaseConfig contains database configuration.
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver" validate:"required,oneof=sqlite3 postgres mysql"`
	DSN             string        `mapstructure:"dsn" validate:"required"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// LoggingConfig contains logging configuration.
type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,valid_log_level"`
	Format string `mapstructure:"format" validate:"required,oneof=json text"`
	Output string `mapstructure:"output" validate:"required"`
}

// AppConfig contains application-specific configuration.
type AppConfig struct {
	Name        string `mapstructure:"name" validate:"required"`
	Version     string `mapstructure:"version" validate:"required"`
	Environment string `mapstructure:"environment" validate:"required,valid_environment"`
	Debug       bool   `mapstructure:"debug"`
}

// JWTConfig contains JWT authentication configuration.
type JWTConfig struct {
	SecretKey          string        `mapstructure:"secret_key" validate:"required,min=32"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
	Issuer             string        `mapstructure:"issuer" validate:"required"`
	Algorithm          string        `mapstructure:"algorithm" validate:"required,oneof=HS256 HS384 HS512"`
}

// SecurityConfig contains security configuration.
type SecurityConfig struct {
	AllowedOrigins    []string      `mapstructure:"allowed_origins"`
	TrustedProxies    []string      `mapstructure:"trusted_proxies"`
	EnableHSTS        bool          `mapstructure:"enable_hsts"`
	EnableCSP         bool          `mapstructure:"enable_csp"`
	CSPReportURI      string        `mapstructure:"csp_report_uri"`
	MaxRequestSize    int64         `mapstructure:"max_request_size"`
	RateLimitEnabled  bool          `mapstructure:"rate_limit_enabled"`
	RateLimitRequests int           `mapstructure:"rate_limit_requests"`
	RateLimitWindow   time.Duration `mapstructure:"rate_limit_window"`
}

// LoadConfig loads configuration from various sources.
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Set defaults
	setDefaults(config)

	// Configure viper
	if err := configureViper(configPath); err != nil {
		return nil, fmt.Errorf("failed to configure viper: %w", err)
	}

	// Unmarshal configuration
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("validation errors: %w", err)
	}

	return config, nil
}

// setDefaults sets default values for the configuration.
func setDefaults(_ *Config) {
	// App defaults
	viper.SetDefault("app.name", "template-arch-lint")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", false)

	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 5*time.Second)
	viper.SetDefault("server.write_timeout", 10*time.Second)
	viper.SetDefault("server.idle_timeout", 120*time.Second)
	viper.SetDefault("server.graceful_shutdown_timeout", 30*time.Second)

	// Database defaults
	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.dsn", "./app.db")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 25)
	viper.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	viper.SetDefault("database.conn_max_idle_time", 5*time.Minute)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// JWT defaults
	viper.SetDefault("jwt.secret_key", "your-super-secret-jwt-key-minimum-32-characters-long-for-security")
	viper.SetDefault("jwt.access_token_expiry", 24*time.Hour)
	viper.SetDefault("jwt.refresh_token_expiry", 7*24*time.Hour)
	viper.SetDefault("jwt.issuer", "template-arch-lint")
	viper.SetDefault("jwt.algorithm", "HS256")

	// Security defaults
	viper.SetDefault("security.allowed_origins", []string{"http://localhost:3000", "http://localhost:8080"})
	viper.SetDefault("security.trusted_proxies", []string{})
	viper.SetDefault("security.enable_hsts", false) // Disabled by default for development
	viper.SetDefault("security.enable_csp", true)
	viper.SetDefault("security.csp_report_uri", "")
	viper.SetDefault("security.max_request_size", 10*1024*1024) // 10MB
	viper.SetDefault("security.rate_limit_enabled", false)
	viper.SetDefault("security.rate_limit_requests", 100)
	viper.SetDefault("security.rate_limit_window", time.Minute)
}

// configureViper sets up viper configuration.
func configureViper(configPath string) error {
	// Environment variable configuration
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// File configuration (optional)
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return nil
}

// validateConfig validates the configuration.
func validateConfig(config *Config) error {
	validate := validator.New()

	// Register custom validators
	if err := validate.RegisterValidation("valid_log_level", validateLogLevel); err != nil {
		return fmt.Errorf("failed to register log level validator: %w", err)
	}

	if err := validate.RegisterValidation("valid_environment", validateEnvironment); err != nil {
		return fmt.Errorf("failed to register environment validator: %w", err)
	}

	return validate.Struct(config)
}

// validateLogLevel validates log level values.
func validateLogLevel(fl validator.FieldLevel) bool {
	level := fl.Field().String()
	switch level {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

// validateEnvironment validates environment values.
func validateEnvironment(fl validator.FieldLevel) bool {
	env := fl.Field().String()
	switch env {
	case "development", "staging", "production", "test":
		return true
	default:
		return false
	}
}
