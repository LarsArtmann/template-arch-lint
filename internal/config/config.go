package config

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server" validate:"required"`
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	Logging  LoggingConfig  `mapstructure:"logging" validate:"required"`
	App      AppConfig      `mapstructure:"app" validate:"required"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host                    string        `mapstructure:"host" validate:"required"`
	Port                    int           `mapstructure:"port" validate:"required,min=1,max=65535"`
	ReadTimeout             time.Duration `mapstructure:"read_timeout"`
	WriteTimeout            time.Duration `mapstructure:"write_timeout"`
	IdleTimeout             time.Duration `mapstructure:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `mapstructure:"graceful_shutdown_timeout"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver" validate:"required,oneof=sqlite3 postgres mysql"`
	DSN             string        `mapstructure:"dsn" validate:"required"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Format string `mapstructure:"format" validate:"required,oneof=json text"`
	Output string `mapstructure:"output" validate:"required"`
}

// AppConfig contains application-specific configuration
type AppConfig struct {
	Name        string `mapstructure:"name" validate:"required"`
	Version     string `mapstructure:"version" validate:"required"`
	Environment string `mapstructure:"environment" validate:"required,oneof=development staging production"`
	Debug       bool   `mapstructure:"debug"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Set default values
	setDefaults()

	// Setup configuration paths and sources
	if err := setupConfigPaths(configPath); err != nil {
		return nil, err
	}

	// Setup environment variable bindings
	if err := setupEnvironmentBindings(); err != nil {
		return nil, err
	}

	// Read and parse configuration
	config, err := readAndParseConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

// setupConfigPaths configures viper to read from the specified config file or default locations
func setupConfigPaths(configPath string) error {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Look for config in current directory and /etc/
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/template-arch-lint/")
		viper.AddConfigPath("$HOME/.template-arch-lint/")
	}

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP") // Environment variables will be prefixed with APP_
	return nil
}

// setupEnvironmentBindings configures all environment variable bindings
func setupEnvironmentBindings() error {
	if err := bindServerEnvVars(); err != nil {
		return err
	}
	if err := bindDatabaseEnvVars(); err != nil {
		return err
	}
	if err := bindLoggingEnvVars(); err != nil {
		return err
	}
	return bindAppEnvVars()
}

// bindServerEnvVars binds server-related environment variables
func bindServerEnvVars() error {
	serverBindings := map[string]string{
		"server.host":                        "APP_SERVER_HOST",
		"server.port":                        "APP_SERVER_PORT",
		"server.read_timeout":                "APP_SERVER_READ_TIMEOUT",
		"server.write_timeout":               "APP_SERVER_WRITE_TIMEOUT",
		"server.idle_timeout":                "APP_SERVER_IDLE_TIMEOUT",
		"server.graceful_shutdown_timeout":   "APP_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT",
	}
	return bindEnvVars(serverBindings, "server")
}

// bindDatabaseEnvVars binds database-related environment variables
func bindDatabaseEnvVars() error {
	databaseBindings := map[string]string{
		"database.driver":              "APP_DATABASE_DRIVER",
		"database.dsn":                 "APP_DATABASE_DSN",
		"database.max_open_conns":      "APP_DATABASE_MAX_OPEN_CONNS",
		"database.max_idle_conns":      "APP_DATABASE_MAX_IDLE_CONNS",
		"database.conn_max_lifetime":   "APP_DATABASE_CONN_MAX_LIFETIME",
		"database.conn_max_idle_time":  "APP_DATABASE_CONN_MAX_IDLE_TIME",
	}
	return bindEnvVars(databaseBindings, "database")
}

// bindLoggingEnvVars binds logging-related environment variables
func bindLoggingEnvVars() error {
	loggingBindings := map[string]string{
		"logging.level":  "APP_LOGGING_LEVEL",
		"logging.format": "APP_LOGGING_FORMAT",
		"logging.output": "APP_LOGGING_OUTPUT",
	}
	return bindEnvVars(loggingBindings, "logging")
}

// bindAppEnvVars binds app-related environment variables
func bindAppEnvVars() error {
	appBindings := map[string]string{
		"app.name":        "APP_APP_NAME",
		"app.version":     "APP_APP_VERSION",
		"app.environment": "APP_APP_ENVIRONMENT",
		"app.debug":       "APP_APP_DEBUG",
	}
	return bindEnvVars(appBindings, "app")
}

// bindEnvVars is a helper function that binds a map of configuration keys to environment variables
func bindEnvVars(bindings map[string]string, section string) error {
	for configKey, envVar := range bindings {
		if err := viper.BindEnv(configKey, envVar); err != nil {
			return fmt.Errorf("failed to bind %s env var: %w", configKey, err)
		}
	}
	return nil
}

// readAndParseConfig reads the config file and parses it into a Config struct
func readAndParseConfig() (*Config, error) {
	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}

	// Unmarshal config into struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "5s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.graceful_shutdown_timeout", "30s")

	// Database defaults
	viper.SetDefault("database.driver", "sqlite3")
	viper.SetDefault("database.dsn", "./app.db")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")
	viper.SetDefault("database.conn_max_idle_time", "5m")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// App defaults
	viper.SetDefault("app.name", "template-arch-lint")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.debug", false)
}

// validateConfig validates the configuration using struct tags
func validateConfig(config *Config) error {
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return fmt.Errorf("validation errors: %w", err)
	}
	return nil
}