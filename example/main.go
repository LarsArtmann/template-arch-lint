// Example application demonstrating configuration management
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
)

const (
	// NewlineChar represents the newline character constant
	NewlineChar = "\n"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Display all configuration sections
	printAppConfig(cfg)
	printServerConfig(cfg)
	printDatabaseConfig(cfg)
	printLoggingConfig(cfg)
	demonstrateEnvironmentOverride(cfg)
	printUsageInstructions()
}

// printAppConfig displays application configuration details
func printAppConfig(cfg *config.Config) {
	_, _ = fmt.Printf("=== Application Configuration ===\n")
	_, _ = fmt.Printf("App Name: %s\n", cfg.App.Name)
	_, _ = fmt.Printf("Version: %s\n", cfg.App.Version)
	_, _ = fmt.Printf("Environment: %s\n", cfg.App.Environment)
	_, _ = fmt.Printf("Debug Mode: %t\n", cfg.App.Debug)
	_, _ = fmt.Printf(NewlineChar)
}

// printServerConfig displays server configuration details
func printServerConfig(cfg *config.Config) {
	_, _ = fmt.Printf("=== Server Configuration ===\n")
	_, _ = fmt.Printf("Host: %s\n", cfg.Server.Host)
	_, _ = fmt.Printf("Port: %d\n", cfg.Server.Port)
	_, _ = fmt.Printf("Read Timeout: %v\n", cfg.Server.ReadTimeout)
	_, _ = fmt.Printf("Write Timeout: %v\n", cfg.Server.WriteTimeout)
	_, _ = fmt.Printf(NewlineChar)
}

// printDatabaseConfig displays database configuration details
func printDatabaseConfig(cfg *config.Config) {
	_, _ = fmt.Printf("=== Database Configuration ===\n")
	_, _ = fmt.Printf("Driver: %s\n", cfg.Database.Driver)
	_, _ = fmt.Printf("DSN: %s\n", cfg.Database.DSN)
	_, _ = fmt.Printf("Max Open Connections: %d\n", cfg.Database.MaxOpenConns)
	_, _ = fmt.Printf("Max Idle Connections: %d\n", cfg.Database.MaxIdleConns)
	_, _ = fmt.Printf(NewlineChar)
}

// printLoggingConfig displays logging configuration details
func printLoggingConfig(cfg *config.Config) {
	_, _ = fmt.Printf("=== Logging Configuration ===\n")
	_, _ = fmt.Printf("Level: %s\n", cfg.Logging.Level)
	_, _ = fmt.Printf("Format: %s\n", cfg.Logging.Format)
	_, _ = fmt.Printf("Output: %s\n", cfg.Logging.Output)
	_, _ = fmt.Printf(NewlineChar)
}

// demonstrateEnvironmentOverride shows how environment variables can
// override config
func demonstrateEnvironmentOverride(cfg *config.Config) {
	_, _ = fmt.Printf("=== Environment Variable Example ===\n")
	_, _ = fmt.Printf("Set APP_SERVER_PORT=9090 to see port override\n")
	if port := os.Getenv("APP_SERVER_PORT"); port != "" {
		_, _ = fmt.Printf(
			"Environment override detected: APP_SERVER_PORT=%s\n",
			port,
		)
	}
	_, _ = fmt.Printf("Current port from config: %d\n", cfg.Server.Port)
	_, _ = fmt.Printf(NewlineChar)
}

// printUsageInstructions displays usage information for the
// configuration example
func printUsageInstructions() {
	_, _ = fmt.Printf("Configuration loaded successfully!\n")
	_, _ = fmt.Printf(
		"You can override any setting using environment variables with " +
			"APP_ prefix.\n",
	)
	_, _ = fmt.Printf(
		"Example: APP_SERVER_PORT=9090 APP_LOGGING_LEVEL=debug " +
			"go run example/main.go\n",
	)
}
