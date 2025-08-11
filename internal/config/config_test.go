package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			config, err := LoadConfig(tt.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if config.Server.Port != tt.expectPort {
					t.Errorf("LoadConfig() port = %v, want %v", config.Server.Port, tt.expectPort)
				}
				if config.Logging.Level != tt.expectLevel {
					t.Errorf("LoadConfig() level = %v, want %v", config.Logging.Level, tt.expectLevel)
				}
			}
		})
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
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
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
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 0, // Invalid port
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
			},
			wantErr: true,
		},
		{
			name: "invalid database driver",
			config: Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: 8080,
				},
				Database: DatabaseConfig{
					Driver: "invalid", // Invalid driver
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
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
