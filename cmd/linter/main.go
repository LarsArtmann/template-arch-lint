// Main entry point for template-arch-lint - Pure Linting Template
// Demonstrates architecture validation using domain layer for business logic validation.
package main

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/charmbracelet/log"
)

//
// NOTE: Replace hardcoded behavior with cobra CLI for production use.
// NOTE: Load configuration from files/environment variables in production.
// NOTE: Use typed exit codes instead of magic numbers in production.
func main() {
	// Initialize structured logger with enterprise configuration
	logger := log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      "2006-01-02 15:04:05",
		Level:           log.InfoLevel,
	})

	logger.Info("🔥 Template-Arch-Lint - Pure Linting Template")
	logger.Info("✅ This demonstrates enterprise-grade Go architecture enforcement")

	// Demonstrate that domain layer works for linting rule validation
	if err := validateDomainLayer(); err != nil {
		logger.Error("❌ Domain validation failed", "error", err)
		os.Exit(1)
	}

	// Demonstrate config loading works
	if err := validateConfig(); err != nil {
		logger.Error("❌ Config validation failed", "error", err)
		os.Exit(1)
	}

	logger.Info("✅ All validations passed - linting rules are working correctly!")
	logger.Info("📋 Copy .golangci.yml, .go-arch-lint.yml, and justfile to your project")
}

// validateDomainLayer demonstrates that domain value objects work correctly.
func validateDomainLayer() error {
	// Test email validation
	_, err := values.NewEmail("test@example.com")
	if err != nil {
		log.Error("Email validation failed", "error", err)
		return fmt.Errorf("email validation failed: %w", err)
	}

	// Test username validation
	_, err = values.NewUserName("testuser")
	if err != nil {
		log.Error("Username validation failed", "error", err)
		return fmt.Errorf("username validation failed: %w", err)
	}

	// Test user ID validation
	_, err = values.NewUserID("user123")
	if err != nil {
		log.Error("User ID validation failed", "error", err)
		return fmt.Errorf("user ID validation failed: %w", err)
	}

	// NOTE: Add validation for all other value objects (Port, LogLevel, EnvVar).
	// NOTE: Add validation for boundary conditions and error cases.
	return nil
}

// validateConfig demonstrates that configuration loading works.
func validateConfig() error {
	// NOTE: Test actual configuration file path in production.
	_, err := config.LoadConfig("")
	if err != nil {
		log.Error("Config loading failed", "error", err)
		return fmt.Errorf("config loading failed: %w", err)
	}
	// NOTE: Actually validate the loaded configuration values in production.
	// NOTE: Test that all required config fields are present in production.
	return nil
}
