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
