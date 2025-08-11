// Example application demonstrating configuration management
package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
)

const (
	// NewlineChar represents the newline character constant
	NewlineChar = "\n"
)

func main() {
	// Set up structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Display all configuration sections
	printAppConfig(cfg, logger)
	printServerConfig(cfg, logger)
	printDatabaseConfig(cfg, logger)
	printLoggingConfig(cfg, logger)
	demonstrateEnvironmentOverride(cfg, logger)
	printUsageInstructions(logger)
}

// printAppConfig displays application configuration details
func printAppConfig(cfg *config.Config, logger *slog.Logger) {
	logger.Info("Application Configuration",
		"name", cfg.App.Name,
		"version", cfg.App.Version,
		"environment", cfg.App.Environment,
		"debug", cfg.App.Debug,
	)
}

// printServerConfig displays server configuration details
func printServerConfig(cfg *config.Config, logger *slog.Logger) {
	logger.Info("Server Configuration",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"read_timeout", cfg.Server.ReadTimeout,
		"write_timeout", cfg.Server.WriteTimeout,
	)
}

// printDatabaseConfig displays database configuration details
func printDatabaseConfig(cfg *config.Config, logger *slog.Logger) {
	logger.Info("Database Configuration",
		"driver", cfg.Database.Driver,
		"dsn", cfg.Database.DSN,
		"max_open_conns", cfg.Database.MaxOpenConns,
		"max_idle_conns", cfg.Database.MaxIdleConns,
	)
}

// printLoggingConfig displays logging configuration details
func printLoggingConfig(cfg *config.Config, logger *slog.Logger) {
	logger.Info("Logging Configuration",
		"level", cfg.Logging.Level,
		"format", cfg.Logging.Format,
		"output", cfg.Logging.Output,
	)
}

// demonstrateEnvironmentOverride shows how environment variables can
// override config
func demonstrateEnvironmentOverride(cfg *config.Config, logger *slog.Logger) {
	logger.Info("Environment Variable Example",
		"instruction", "Set APP_SERVER_PORT=9090 to see port override",
		"current_port", cfg.Server.Port,
	)
	if port := os.Getenv("APP_SERVER_PORT"); port != "" {
		logger.Info("Environment override detected",
			"variable", "APP_SERVER_PORT",
			"value", port,
		)
	}
}

// printUsageInstructions displays usage information for the
// configuration example
func printUsageInstructions(logger *slog.Logger) {
	logger.Info("Configuration loaded successfully")
	logger.Info("Usage Instructions",
		"note", "You can override any setting using environment variables with APP_ prefix",
		"example", "APP_SERVER_PORT=9090 APP_LOGGING_LEVEL=debug go run example/main.go",
	)
}
