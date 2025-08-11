// Example application demonstrating configuration management
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
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
	fmt.Printf("=== Application Configuration ===\n")
	fmt.Printf("App Name: %s\n", cfg.App.Name)
	fmt.Printf("Version: %s\n", cfg.App.Version)
	fmt.Printf("Environment: %s\n", cfg.App.Environment)
	fmt.Printf("Debug Mode: %t\n", cfg.App.Debug)
	fmt.Printf("\n")
}

// printServerConfig displays server configuration details
func printServerConfig(cfg *config.Config) {
	fmt.Printf("=== Server Configuration ===\n")
	fmt.Printf("Host: %s\n", cfg.Server.Host)
	fmt.Printf("Port: %d\n", cfg.Server.Port)
	fmt.Printf("Read Timeout: %v\n", cfg.Server.ReadTimeout)
	fmt.Printf("Write Timeout: %v\n", cfg.Server.WriteTimeout)
	fmt.Printf("\n")
}

// printDatabaseConfig displays database configuration details
func printDatabaseConfig(cfg *config.Config) {
	fmt.Printf("=== Database Configuration ===\n")
	fmt.Printf("Driver: %s\n", cfg.Database.Driver)
	fmt.Printf("DSN: %s\n", cfg.Database.DSN)
	fmt.Printf("Max Open Connections: %d\n", cfg.Database.MaxOpenConns)
	fmt.Printf("Max Idle Connections: %d\n", cfg.Database.MaxIdleConns)
	fmt.Printf("\n")
}

// printLoggingConfig displays logging configuration details
func printLoggingConfig(cfg *config.Config) {
	fmt.Printf("=== Logging Configuration ===\n")
	fmt.Printf("Level: %s\n", cfg.Logging.Level)
	fmt.Printf("Format: %s\n", cfg.Logging.Format)
	fmt.Printf("Output: %s\n", cfg.Logging.Output)
	fmt.Printf("\n")
}

// demonstrateEnvironmentOverride shows how environment variables can override config
func demonstrateEnvironmentOverride(cfg *config.Config) {
	fmt.Printf("=== Environment Variable Example ===\n")
	fmt.Printf("Set APP_SERVER_PORT=9090 to see port override\n")
	if port := os.Getenv("APP_SERVER_PORT"); port != "" {
		fmt.Printf("Environment override detected: APP_SERVER_PORT=%s\n", port)
	}
	fmt.Printf("Current port from config: %d\n", cfg.Server.Port)
	fmt.Printf("\n")
}

// printUsageInstructions displays usage information for the configuration example
func printUsageInstructions() {
	fmt.Printf("Configuration loaded successfully!\n")
	fmt.Printf("You can override any setting using environment variables with APP_ prefix.\n")
	fmt.Printf("Example: APP_SERVER_PORT=9090 APP_LOGGING_LEVEL=debug go run example/main.go\n")
}
