package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := getLoadConfigTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupTestEnvironment(t, tt.envVars)
			runLoadConfigTest(t, tt)
		})
	}
}

// getLoadConfigTestCases returns test cases for LoadConfig function
func getLoadConfigTestCases() []struct {
	name        string
	configPath  string
	envVars     map[string]string
	wantErr     bool
	expectPort  int
	expectLevel string
} {
	return []struct {
		name        string
		configPath  string
		envVars     map[string]string
		wantErr     bool
		expectPort  int
		expectLevel string
	}{
		{
			name:        "load with defaults",
			configPath:  "",
			envVars:     map[string]string{},
			wantErr:     false,
			expectPort:  8080,
			expectLevel: "info",
		},
		{
			name:       "override with environment variables",
			configPath: "",
			envVars: map[string]string{
				"APP_SERVER_PORT":   "9090",
				"APP_LOGGING_LEVEL": "debug",
			},
			wantErr:     false,
			expectPort:  9090,
			expectLevel: "debug",
		},
		{
			name:       "invalid port value",
			configPath: "",
			envVars: map[string]string{
				"APP_SERVER_PORT": "99999",
			},
			wantErr: true,
		},
		{
			name:       "invalid log level",
			configPath: "",
			envVars: map[string]string{
				"APP_LOGGING_LEVEL": "invalid",
			},
			wantErr: true,
		},
	}
}

// setupTestEnvironment sets up and cleans up environment variables for testing
func setupTestEnvironment(t *testing.T, envVars map[string]string) {
	t.Helper()
	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
		t.Cleanup(func() {
			if err := os.Unsetenv(key); err != nil {
				t.Errorf("Failed to unset environment variable %s: %v", key, err)
			}
		})
	}
}

// runLoadConfigTest executes the LoadConfig test logic
func runLoadConfigTest(t *testing.T, tt struct {
	name        string
	configPath  string
	envVars     map[string]string
	wantErr     bool
	expectPort  int
	expectLevel string
},
) {
	t.Helper()

	config, err := LoadConfig(tt.configPath)
	if (err != nil) != tt.wantErr {
		t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
		return
	}

	if !tt.wantErr {
		validateLoadConfigResult(t, config, tt.expectPort, tt.expectLevel)
	}
}

// validateLoadConfigResult validates the loaded config matches expectations
func validateLoadConfigResult(t *testing.T, config *Config, expectPort int, expectLevel string) {
	t.Helper()

	if config.Server.Port != expectPort {
		t.Errorf("LoadConfig() port = %v, want %v", config.Server.Port, expectPort)
	}
	if config.Logging.Level != expectLevel {
		t.Errorf("LoadConfig() level = %v, want %v", config.Logging.Level, expectLevel)
	}
}

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

// validateServerDefaults checks server configuration defaults
func validateServerDefaults(t *testing.T, server ServerConfig) {
	t.Helper()

	if server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", server.Host)
	}
	if server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", server.Port)
	}
	if server.ReadTimeout != 5*time.Second {
		t.Errorf("Expected default read timeout 5s, got %v", server.ReadTimeout)
	}
}

// validateDatabaseDefaults checks database configuration defaults
func validateDatabaseDefaults(t *testing.T, database DatabaseConfig) {
	t.Helper()

	if database.Driver != "sqlite3" {
		t.Errorf("Expected default database driver 'sqlite3', got '%s'", database.Driver)
	}
	if database.DSN != "./app.db" {
		t.Errorf("Expected default DSN './app.db', got '%s'", database.DSN)
	}
}

// validateLoggingDefaults checks logging configuration defaults
func validateLoggingDefaults(t *testing.T, logging LoggingConfig) {
	t.Helper()

	if logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", logging.Level)
	}
	if logging.Format != "json" {
		t.Errorf("Expected default log format 'json', got '%s'", logging.Format)
	}
}

// validateAppDefaults checks app configuration defaults
func validateAppDefaults(t *testing.T, app AppConfig) {
	t.Helper()

	if app.Name != "template-arch-lint" {
		t.Errorf("Expected default app name 'template-arch-lint', got '%s'", app.Name)
	}
	if app.Environment != "development" {
		t.Errorf("Expected default environment 'development', got '%s'", app.Environment)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := getConfigValidationTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runConfigValidationTest(t, tt.config, tt.wantErr)
		})
	}
}

// getConfigValidationTestCases returns test cases for config validation
func getConfigValidationTestCases() []struct {
	name    string
	config  Config
	wantErr bool
} {
	return []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  createValidTestConfig(),
			wantErr: false,
		},
		{
			name:    "invalid port",
			config:  createConfigWithInvalidPort(),
			wantErr: true,
		},
		{
			name:    "invalid database driver",
			config:  createConfigWithInvalidDriver(),
			wantErr: true,
		},
	}
}

// createValidTestConfig creates a valid configuration for testing
func createValidTestConfig() Config {
	return Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
		Database: DatabaseConfig{
			Driver: "sqlite3",
			DSN:    "./test.db",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		App: AppConfig{
			Name:        "test-app",
			Version:     "1.0.0",
			Environment: "development",
		},
	}
}

// createConfigWithInvalidPort creates a config with invalid port for testing
func createConfigWithInvalidPort() Config {
	config := createValidTestConfig()
	config.Server.Port = 0 // Invalid port
	return config
}

// createConfigWithInvalidDriver creates a config with invalid database driver for testing
func createConfigWithInvalidDriver() Config {
	config := createValidTestConfig()
	config.Database.Driver = "invalid" // Invalid driver
	return config
}

// runConfigValidationTest executes a single config validation test
func runConfigValidationTest(t *testing.T, config Config, wantErr bool) {
	t.Helper()

	err := validateConfig(&config)
	if (err != nil) != wantErr {
		t.Errorf("validateConfig() error = %v, wantErr %v", err, wantErr)
	}
}

// Additional comprehensive tests for all configuration features

func TestConfigWithEnvironmentOverrides(t *testing.T) {
	t.Run("environment variable overrides", func(t *testing.T) {
		testConfigOverrides(t)
	})

	t.Run("invalid environment validation", func(t *testing.T) {
		testInvalidEnvironment(t)
	})
}

// testConfigOverrides tests that environment variables override config defaults
func testConfigOverrides(t *testing.T) {
	t.Helper()

	setTestEnvVars(t)
	defer unsetTestEnvVars(t)

	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	validateOverriddenConfig(t, config)
}

// setTestEnvVars sets test environment variables
func setTestEnvVars(t *testing.T) {
	t.Helper()

	envVars := map[string]string{
		"APP_APP_ENVIRONMENT": "staging",
		"APP_SERVER_PORT":     "9090",
		"APP_LOGGING_LEVEL":   "warn",
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}
}

// unsetTestEnvVars cleans up test environment variables
func unsetTestEnvVars(t *testing.T) {
	t.Helper()

	envVars := []string{
		"APP_APP_ENVIRONMENT",
		"APP_SERVER_PORT",
		"APP_LOGGING_LEVEL",
	}

	for _, key := range envVars {
		if err := os.Unsetenv(key); err != nil {
			t.Errorf("Failed to unset %s: %v", key, err)
		}
	}
}

// validateOverriddenConfig validates that config was properly overridden
func validateOverriddenConfig(t *testing.T, config *Config) {
	t.Helper()

	if config.App.Environment != "staging" {
		t.Errorf("Expected environment 'staging', got '%s'", config.App.Environment)
	}
	if config.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Server.Port)
	}
	if config.Logging.Level != "warn" {
		t.Errorf("Expected log level 'warn', got '%s'", config.Logging.Level)
	}
}

// testInvalidEnvironment tests that invalid environment values are rejected
func testInvalidEnvironment(t *testing.T) {
	t.Helper()

	if err := os.Setenv("APP_APP_ENVIRONMENT", "invalid"); err != nil {
		t.Fatalf("Failed to set APP_APP_ENVIRONMENT: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("APP_APP_ENVIRONMENT"); err != nil {
			t.Errorf("Failed to unset APP_APP_ENVIRONMENT: %v", err)
		}
	}()

	_, err := LoadConfig("")
	if err == nil {
		t.Error("Should reject invalid environment")
	}
}
