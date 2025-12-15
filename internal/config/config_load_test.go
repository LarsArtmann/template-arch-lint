package config

import (
	"os"
	"testing"

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
