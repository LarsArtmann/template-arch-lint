package config

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
}) {
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

	// Test server defaults
	if config.Server.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}
	if config.Server.ReadTimeout != 5*time.Second {
		t.Errorf("Expected default read timeout 5s, got %v", config.Server.ReadTimeout)
	}

	// Test database defaults
	if config.Database.Driver != "sqlite3" {
		t.Errorf("Expected default database driver 'sqlite3', got '%s'", config.Database.Driver)
	}
	if config.Database.DSN != "./app.db" {
		t.Errorf("Expected default DSN './app.db', got '%s'", config.Database.DSN)
	}

	// Test logging defaults
	if config.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got '%s'", config.Logging.Level)
	}
	if config.Logging.Format != "json" {
		t.Errorf("Expected default log format 'json', got '%s'", config.Logging.Format)
	}

	// Test app defaults
	if config.App.Name != "template-arch-lint" {
		t.Errorf("Expected default app name 'template-arch-lint', got '%s'", config.App.Name)
	}
	if config.App.Environment != "development" {
		t.Errorf("Expected default environment 'development', got '%s'", config.App.Environment)
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
		Observability: ObservabilityConfig{
			ServiceName:    "test-app",
			ServiceVersion: "1.0.0",
			Environment:    "development",
			Exporters: ExportersConfig{
				Prometheus: PrometheusConfig{
					Port: 2112,
				},
			},
		},
		Security: SecurityConfig{
			CORS: CORSConfig{
				AllowedOrigins: []string{"http://localhost:3000"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
				AllowedHeaders: []string{"Content-Type", "Authorization"},
			},
			RateLimit: RateLimitConfig{
				RequestsPerMinute: 100,
				Burst:            20,
			},
		},
		External: ExternalConfig{
			CircuitBreaker: CircuitBreakerConfig{
				Threshold: 5,
			},
		},
		Backup: BackupConfig{
			RetentionDays: 30,
		},
		Resources: ResourcesConfig{
			MaxCPUCores:    4,
			MaxConnections: 100,
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
		os.Setenv("APP_APP_ENVIRONMENT", "staging")
		os.Setenv("APP_SERVER_PORT", "9090")
		os.Setenv("APP_LOGGING_LEVEL", "warn")
		defer func() {
			os.Unsetenv("APP_APP_ENVIRONMENT")
			os.Unsetenv("APP_SERVER_PORT") 
			os.Unsetenv("APP_LOGGING_LEVEL")
		}()

		config, err := LoadConfig("")
		require.NoError(t, err)
		
		assert.Equal(t, "staging", config.App.Environment)
		assert.Equal(t, 9090, config.Server.Port)
		assert.Equal(t, "warn", config.Logging.Level)
	})
	
	t.Run("invalid environment validation", func(t *testing.T) {
		os.Setenv("APP_APP_ENVIRONMENT", "invalid")
		defer os.Unsetenv("APP_APP_ENVIRONMENT")
		
		_, err := LoadConfig("")
		assert.Error(t, err, "Should reject invalid environment")
	})
}

// Test feature flags functionality
func TestFeatureFlags(t *testing.T) {
	featureConfig := &FeaturesConfig{
		EnableDebugEndpoints: true,
		EnableProfiling:      false,
		EnableBetaFeatures:   true,
	}
	
	manager := NewFeatureManager(featureConfig)
	defer manager.Close()
	
	t.Run("basic feature flag check", func(t *testing.T) {
		assert.True(t, manager.IsEnabled("debug_endpoints"))
		assert.False(t, manager.IsEnabled("profiling"))
		assert.True(t, manager.IsEnabled("beta_features"))
		assert.False(t, manager.IsEnabled("non_existent"))
	})
	
	t.Run("feature flag with context", func(t *testing.T) {
		ctx := FeatureContext{
			UserID:      "test_user",
			Environment: "development",
			Timestamp:   time.Now(),
		}
		
		assert.True(t, manager.IsEnabledForContext("debug_endpoints", ctx))
	})
	
	t.Run("dynamic feature flag creation", func(t *testing.T) {
		flag := &FeatureFlag{
			Name:        "dynamic_test",
			Enabled:     true,
			Description: "Test dynamic flag",
		}
		
		err := manager.UpdateFlag(flag)
		assert.NoError(t, err)
		
		assert.True(t, manager.IsEnabled("dynamic_test"))
		
		// Get flag details
		retrievedFlag, exists := manager.GetFlag("dynamic_test")
		assert.True(t, exists)
		assert.Equal(t, "Test dynamic flag", retrievedFlag.Description)
	})
}

// Test secrets management functionality
func TestSecretsManager(t *testing.T) {
	t.Run("environment provider", func(t *testing.T) {
		provider := NewEnvProvider()
		ctx := context.Background()
		
		// Test getting existing environment variable
		os.Setenv("TEST_SECRET", "test_value")
		defer os.Unsetenv("TEST_SECRET")
		
		value, err := provider.GetSecret(ctx, "TEST_SECRET")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)
		
		// Test getting non-existent secret
		_, err = provider.GetSecret(ctx, "NON_EXISTENT")
		assert.Error(t, err)
	})
	
	t.Run("file provider", func(t *testing.T) {
		config := FileConfig{
			Path:   "/tmp/test_secrets.json",
			Format: "json",
		}
		
		provider, err := NewFileProvider(config)
		require.NoError(t, err)
		defer provider.Close()
		
		ctx := context.Background()
		
		// Test setting and getting secret
		err = provider.SetSecret(ctx, "test_key", "test_value")
		assert.NoError(t, err)
		
		value, err := provider.GetSecret(ctx, "test_key")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)
		
		// Test deleting secret
		err = provider.DeleteSecret(ctx, "test_key")
		assert.NoError(t, err)
		
		_, err = provider.GetSecret(ctx, "test_key")
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkConfigLoading(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := LoadConfig("")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	config, err := LoadConfig("")
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := validateConfig(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFeatureFlagCheck(b *testing.B) {
	featureConfig := &FeaturesConfig{
		EnableDebugEndpoints: true,
		EnableProfiling:      false,
		EnableBetaFeatures:   true,
	}
	
	manager := NewFeatureManager(featureConfig)
	defer manager.Close()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.IsEnabled("debug_endpoints")
	}
}
