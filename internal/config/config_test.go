package config

import (
	"testing"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

func TestConfigDefaults(t *testing.T) {
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	validateServerDefaults(t, config.Server)
	validateDatabaseDefaults(t, config.Database)
	validateLoggingDefaults(t, config.Logging)
	validateAppDefaults(t, config.App)
}

// validateServerDefaults checks server configuration defaults.
func validateServerDefaults(t *testing.T, server ServerConfig) {
	t.Helper()

	expectedPort, err := values.NewPort(8080)
	if err != nil {
		t.Fatalf("Invalid expected port 8080: %v", err)
	}
	if server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", server.Host)
	}
	if server.Port != expectedPort {
		t.Errorf("Expected default port 8080, got %v", server.Port)
	}
	if server.ReadTimeout != 5*time.Second {
		t.Errorf("Expected default read timeout 5s, got %v", server.ReadTimeout)
	}
}

// validateDatabaseDefaults checks database configuration defaults.
func validateDatabaseDefaults(t *testing.T, database DatabaseConfig) {
	t.Helper()

	if database.Driver != "sqlite3" {
		t.Errorf("Expected default database driver 'sqlite3', got '%s'", database.Driver)
	}
	if database.DSN != "./app.db" {
		t.Errorf("Expected default DSN './app.db', got '%s'", database.DSN)
	}
}

// validateLoggingDefaults checks logging configuration defaults.
func validateLoggingDefaults(t *testing.T, logging LoggingConfig) {
	t.Helper()

	if logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", logging.Level)
	}
	if logging.Format != "json" {
		t.Errorf("Expected default log format 'json', got '%s'", logging.Format)
	}
}

// validateAppDefaults checks app configuration defaults.
func validateAppDefaults(t *testing.T, app AppConfig) {
	t.Helper()

	if app.Name != "template-arch-lint" {
		t.Errorf("Expected default app name 'template-arch-lint', got '%s'", app.Name)
	}
	if app.Environment != "development" {
		t.Errorf("Expected default environment 'development', got '%s'", app.Environment)
	}
}
