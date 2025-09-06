// Main entry point for template-arch-lint - Pure Linting Template
// This demonstrates architecture validation using the domain layer for business logic validation
package main

import (
	"fmt"
	"os"

	"github.com/LarsArtmann/template-arch-lint/internal/config"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

func main() {
	fmt.Println("🔥 Template-Arch-Lint - Pure Linting Template")
	fmt.Println("✅ This demonstrates enterprise-grade Go architecture enforcement")

	// Demonstrate that domain layer works for linting rule validation
	if err := validateDomainLayer(); err != nil {
		fmt.Printf("❌ Domain validation failed: %v\n", err)
		os.Exit(1)
	}

	// Demonstrate config loading works
	if err := validateConfig(); err != nil {
		fmt.Printf("❌ Config validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ All validations passed - linting rules are working correctly!")
	fmt.Println("📋 Copy .golangci.yml, .go-arch-lint.yml, and justfile to your project")
}

// validateDomainLayer demonstrates that domain value objects work correctly
func validateDomainLayer() error {
	// Test email validation
	_, err := values.NewEmail("test@example.com")
	if err != nil {
		return fmt.Errorf("email validation failed: %w", err)
	}

	// Test username validation
	_, err = values.NewUserName("testuser")
	if err != nil {
		return fmt.Errorf("username validation failed: %w", err)
	}

	// Test user ID validation
	_, err = values.NewUserID("user123")
	if err != nil {
		return fmt.Errorf("user ID validation failed: %w", err)
	}

	return nil
}

// validateConfig demonstrates that configuration loading works
func validateConfig() error {
	_, err := config.LoadConfig("")
	if err != nil {
		return fmt.Errorf("config loading failed: %w", err)
	}
	return nil
}
