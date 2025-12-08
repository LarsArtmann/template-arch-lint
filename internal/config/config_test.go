package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
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

// getLoadConfigTestCases returns test cases for LoadConfig function.
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

// setupTestEnvironment sets up and cleans up environment variables for testing.
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

// runLoadConfigTest executes the LoadConfig test logic.
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

// validateLoadConfigResult validates the loaded config matches expectations.
func validateLoadConfigResult(t *testing.T, config *Config, expectPort int, expectLevel string) {
	t.Helper()

	expectedPort, err := values.NewPort(expectPort)
	if err != nil {
		t.Fatalf("Invalid expected port %d: %v", expectPort, err)
	}
	if config.Server.Port != expectedPort {
		t.Errorf("LoadConfig() port = %v, want %v", config.Server.Port, expectedPort)
	}

	expectedLevel, err := values.NewLogLevel(expectLevel)
	if err != nil {
		t.Fatalf("Invalid expected log level '%s': %v", expectLevel, err)
	}
	if config.Logging.Level != expectedLevel {
		t.Errorf("LoadConfig() level = %v, want %v", config.Logging.Level, expectedLevel)
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

// validateServerDefaults checks server configuration defaults.
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

func TestConfigValidation(t *testing.T) {
	tests := getConfigValidationTestCases()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runConfigValidationTest(t, tt.config, tt.wantErr)
		})
	}
}

// getConfigValidationTestCases returns test cases for config validation.
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

// createValidTestConfig creates a valid configuration for testing.
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
		JWT: JWTConfig{
			SecretKey:          "test-secret-key-that-is-at-least-32-characters-long",
			AccessTokenExpiry:  time.Hour,
			RefreshTokenExpiry: time.Hour * 24 * 7,
			Issuer:             "test-issuer",
			Algorithm:          "HS256",
		},
	}
}

// createConfigWithInvalidPort creates a config with invalid port for testing.
func createConfigWithInvalidPort() Config {
	config := createValidTestConfig()
	config.Server.Port = 0 // Invalid port
	return config
}

// createConfigWithInvalidDriver creates a config with invalid database driver for testing.
func createConfigWithInvalidDriver() Config {
	config := createValidTestConfig()
	config.Database.Driver = "invalid" // Invalid driver
	return config
}

// runConfigValidationTest executes a single config validation test.
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

// testConfigOverrides tests that environment variables override config defaults.
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

// setTestEnvVars sets test environment variables.
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

// unsetTestEnvVars cleans up test environment variables.
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

// validateOverriddenConfig validates that config was properly overridden.
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

// testInvalidEnvironment tests that invalid environment values are rejected.
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

// TestConfigurationParameterValidation tests comprehensive configuration validation at startup.
func TestConfigurationParameterValidation(t *testing.T) {
	t.Run("server configuration validation", func(t *testing.T) {
		testServerConfigValidation(t)
	})

	t.Run("database configuration validation", func(t *testing.T) {
		testDatabaseConfigValidation(t)
	})

	t.Run("logging configuration validation", func(t *testing.T) {
		testLoggingConfigValidation(t)
	})

	t.Run("application configuration validation", func(t *testing.T) {
		testApplicationConfigValidation(t)
	})

	t.Run("cross-configuration dependencies", func(t *testing.T) {
		testCrossConfigurationDependencies(t)
	})
}

// testServerConfigValidation tests server configuration parameter validation.
func testServerConfigValidation(t *testing.T) {
	t.Helper()

	invalidServerConfigs := []struct {
		envVar      string
		value       string
		description string
	}{
		{"APP_SERVER_PORT", "0", "port zero"},
		{"APP_SERVER_PORT", "-1", "negative port"},
		{"APP_SERVER_PORT", "65536", "port too high"},
		{"APP_SERVER_PORT", "99999", "port way too high"},
		{"APP_SERVER_PORT", "abc", "non-numeric port"},
		{"APP_SERVER_PORT", "80.5", "decimal port"},
		{"APP_SERVER_PORT", "", "empty port"},
		{"APP_SERVER_HOST", "", "empty host"},
		{"APP_SERVER_READ_TIMEOUT", "-1s", "negative timeout"},
		{"APP_SERVER_READ_TIMEOUT", "0s", "zero timeout"},
		{"APP_SERVER_READ_TIMEOUT", "invalid", "invalid timeout format"},
		{"APP_SERVER_WRITE_TIMEOUT", "-1s", "negative write timeout"},
		{"APP_SERVER_IDLE_TIMEOUT", "-1s", "negative idle timeout"},
	}

	for _, tc := range invalidServerConfigs {
		t.Run(tc.description, func(t *testing.T) {
			// Set invalid environment variable
			if err := os.Setenv(tc.envVar, tc.value); err != nil {
				t.Fatalf("Failed to set %s: %v", tc.envVar, err)
			}
			defer func() {
				if err := os.Unsetenv(tc.envVar); err != nil {
					t.Errorf("Failed to unset %s: %v", tc.envVar, err)
				}
			}()

			_, err := LoadConfig("")
			if err == nil {
				t.Errorf("Should reject invalid %s: %s", tc.envVar, tc.value)
			}
		})
	}
}

// testDatabaseConfigValidation tests database configuration parameter validation.
func testDatabaseConfigValidation(t *testing.T) {
	t.Helper()

	invalidDatabaseConfigs := []struct {
		envVar      string
		value       string
		description string
	}{
		{"APP_DATABASE_DRIVER", "", "empty driver"},
		{"APP_DATABASE_DRIVER", "mysql", "unsupported driver mysql"},
		{"APP_DATABASE_DRIVER", "postgres", "unsupported driver postgres"},
		{"APP_DATABASE_DRIVER", "invalid", "invalid driver"},
		{"APP_DATABASE_DSN", "", "empty DSN"},
		{"APP_DATABASE_MAX_OPEN_CONNS", "-1", "negative max open connections"},
		{"APP_DATABASE_MAX_IDLE_CONNS", "-1", "negative max idle connections"},
		{"APP_DATABASE_MAX_OPEN_CONNS", "0", "zero max open connections"},
		{"APP_DATABASE_MAX_IDLE_CONNS", "abc", "non-numeric max idle connections"},
		{"APP_DATABASE_CONN_MAX_LIFETIME", "-1s", "negative connection lifetime"},
		{"APP_DATABASE_CONN_MAX_LIFETIME", "invalid", "invalid lifetime format"},
	}

	for _, tc := range invalidDatabaseConfigs {
		t.Run(tc.description, func(t *testing.T) {
			if err := os.Setenv(tc.envVar, tc.value); err != nil {
				t.Fatalf("Failed to set %s: %v", tc.envVar, err)
			}
			defer func() {
				if err := os.Unsetenv(tc.envVar); err != nil {
					t.Errorf("Failed to unset %s: %v", tc.envVar, err)
				}
			}()

			_, err := LoadConfig("")
			if err == nil {
				t.Errorf("Should reject invalid %s: %s", tc.envVar, tc.value)
			}
		})
	}
}

// testLoggingConfigValidation tests logging configuration parameter validation.
func testLoggingConfigValidation(t *testing.T) {
	t.Helper()

	invalidLoggingConfigs := []struct {
		envVar      string
		value       string
		description string
	}{
		{"APP_LOGGING_LEVEL", "", "empty log level"},
		{"APP_LOGGING_LEVEL", "invalid", "invalid log level"},
		{"APP_LOGGING_LEVEL", "INVALID", "invalid log level uppercase"},
		{"APP_LOGGING_LEVEL", "trace", "unsupported trace level"},
		{"APP_LOGGING_LEVEL", "0", "numeric log level"},
		{"APP_LOGGING_FORMAT", "", "empty log format"},
		{"APP_LOGGING_FORMAT", "invalid", "invalid log format"},
		{"APP_LOGGING_FORMAT", "xml", "unsupported xml format"},
		{"APP_LOGGING_FORMAT", "yaml", "unsupported yaml format"},
		{"APP_LOGGING_OUTPUT", "", "empty log output"},
		{"APP_LOGGING_OUTPUT", "invalid", "invalid log output"},
		{"APP_LOGGING_OUTPUT", "/nonexistent/path/log.txt", "nonexistent directory"},
	}

	for _, tc := range invalidLoggingConfigs {
		t.Run(tc.description, func(t *testing.T) {
			if err := os.Setenv(tc.envVar, tc.value); err != nil {
				t.Fatalf("Failed to set %s: %v", tc.envVar, err)
			}
			defer func() {
				if err := os.Unsetenv(tc.envVar); err != nil {
					t.Errorf("Failed to unset %s: %v", tc.envVar, err)
				}
			}()

			_, err := LoadConfig("")
			if err == nil {
				t.Errorf("Should reject invalid %s: %s", tc.envVar, tc.value)
			}
		})
	}
}

// testApplicationConfigValidation tests application configuration parameter validation.
func testApplicationConfigValidation(t *testing.T) {
	t.Helper()

	invalidAppConfigs := []struct {
		envVar      string
		value       string
		description string
	}{
		{"APP_APP_NAME", "", "empty app name"},
		{"APP_APP_NAME", "   ", "whitespace only app name"},
		{"APP_APP_NAME", "app with spaces", "app name with spaces"},
		{"APP_APP_NAME", "app@name", "app name with special chars"},
		{"APP_APP_VERSION", "", "empty version"},
		{"APP_APP_VERSION", "invalid", "invalid version format"},
		{"APP_APP_VERSION", "v1.0.0", "version with v prefix"},
		{"APP_APP_ENVIRONMENT", "", "empty environment"},
		{"APP_APP_ENVIRONMENT", "invalid", "invalid environment"},
		{"APP_APP_ENVIRONMENT", "PRODUCTION", "uppercase environment"},
		{"APP_APP_ENVIRONMENT", "test-env", "environment with hyphen"},
		{"APP_APP_ENVIRONMENT", "development-test", "complex invalid environment"},
	}

	for _, tc := range invalidAppConfigs {
		t.Run(tc.description, func(t *testing.T) {
			if err := os.Setenv(tc.envVar, tc.value); err != nil {
				t.Fatalf("Failed to set %s: %v", tc.envVar, err)
			}
			defer func() {
				if err := os.Unsetenv(tc.envVar); err != nil {
					t.Errorf("Failed to unset %s: %v", tc.envVar, err)
				}
			}()

			_, err := LoadConfig("")
			if err == nil {
				t.Errorf("Should reject invalid %s: %s", tc.envVar, tc.value)
			}
		})
	}
}

// testCrossConfigurationDependencies tests validation of configuration dependencies.
func testCrossConfigurationDependencies(t *testing.T) {
	t.Helper()
	t.Run("database connection limits consistency", testDatabaseConnectionLimits)
	t.Run("production environment security requirements", testProductionSecurityRequirements)
	t.Run("logging output path validation", testLoggingPathValidation)
}

// testDatabaseConnectionLimits validates database connection pool limits.
func testDatabaseConnectionLimits(t *testing.T) {
	envVars := map[string]string{
		"APP_DATABASE_MAX_OPEN_CONNS": "5",
		"APP_DATABASE_MAX_IDLE_CONNS": "10", // Higher than max open
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}
	defer func() {
		for key := range envVars {
			if err := os.Unsetenv(key); err != nil {
				t.Errorf("Failed to unset %s: %v", key, err)
			}
		}
	}()

	_, err := LoadConfig("")
	if err == nil {
		t.Error("Should reject max idle connections > max open connections")
	}
}

// testProductionSecurityRequirements validates production environment security settings.
func testProductionSecurityRequirements(t *testing.T) {
	envVars := map[string]string{
		"APP_APP_ENVIRONMENT": "production",
		"APP_LOGGING_LEVEL":   "debug", // Too verbose for production
	}

	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Failed to set %s: %v", key, err)
		}
	}
	defer func() {
		for key := range envVars {
			if err := os.Unsetenv(key); err != nil {
				t.Errorf("Failed to unset %s: %v", key, err)
			}
		}
	}()

	_, err := LoadConfig("")
	if err == nil {
		t.Error("Should reject debug logging in production environment")
	}
}

// testLoggingPathValidation validates logging output path accessibility.
func testLoggingPathValidation(t *testing.T) {
	if err := os.Setenv("APP_LOGGING_OUTPUT", "/root/app.log"); err != nil {
		t.Fatalf("Failed to set APP_LOGGING_OUTPUT: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("APP_LOGGING_OUTPUT"); err != nil {
			t.Errorf("Failed to unset APP_LOGGING_OUTPUT: %v", err)
		}
	}()

	_, err := LoadConfig("")
	if err == nil {
		t.Error("Should reject inaccessible log file path")
	}
}

// TestConfigValidationErrorMessages tests that validation errors provide clear messages.
func TestConfigValidationErrorMessages(t *testing.T) {
	testCases := []struct {
		envVar      string
		value       string
		expectedMsg string
		description string
	}{
		{"APP_SERVER_PORT", "99999", "port", "port validation error should mention port"},
		{"APP_DATABASE_DRIVER", "mysql", "driver", "driver error should mention driver"},
		{"APP_LOGGING_LEVEL", "invalid", "level", "log level error should mention level"},
		{"APP_APP_ENVIRONMENT", "invalid", "environment", "environment error should mention environment"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if err := os.Setenv(tc.envVar, tc.value); err != nil {
				t.Fatalf("Failed to set %s: %v", tc.envVar, err)
			}
			defer func() {
				if err := os.Unsetenv(tc.envVar); err != nil {
					t.Errorf("Failed to unset %s: %v", tc.envVar, err)
				}
			}()

			_, err := LoadConfig("")
			if err == nil {
				t.Errorf("Should return error for invalid %s", tc.envVar)
				return
			}

			errorMsg := err.Error()
			if !strings.Contains(strings.ToLower(errorMsg), tc.expectedMsg) {
				t.Errorf("Error message should contain '%s', got: %s", tc.expectedMsg, errorMsg)
			}
		})
	}
}

// TestConfigValidationPerformance tests that config validation is efficient.
func TestConfigValidationPerformance(t *testing.T) {
	const iterations = 100

	start := time.Now()

	for range iterations {
		_, err := LoadConfig("")
		if err != nil {
			t.Fatalf("Unexpected error during performance test: %v", err)
		}
	}

	duration := time.Since(start)
	avgDuration := duration / iterations

	// Config loading should be very fast (< 10ms per call on average)
	if avgDuration > 10*time.Millisecond {
		t.Errorf("Config validation too slow: avg %v per call", avgDuration)
	}

	t.Logf("Config validation performance: %d iterations in %v (avg: %v)",
		iterations, duration, avgDuration)
}
