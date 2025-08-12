package config

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ReloadableConfig represents a configuration that can be hot-reloaded
type ReloadableConfig struct {
	mu             sync.RWMutex
	config         *Config
	configPath     string
	watcher        *fsnotify.Watcher
	subscribers    []chan ConfigChange
	ctx            context.Context
	cancel         context.CancelFunc
	lastReload     time.Time
	reloadCount    int
	reloadErrors   int
	featureManager *FeatureManager
}

// ConfigChange represents a configuration change event
type ConfigChange struct {
	Type        string    `json:"type"`        // "config_changed", "file_changed", "reload_success", "reload_error"
	Path        string    `json:"path"`
	OldConfig   *Config   `json:"old_config,omitempty"`
	NewConfig   *Config   `json:"new_config,omitempty"`
	Error       string    `json:"error,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Differences []ConfigDiff `json:"differences,omitempty"`
}

// NewReloadableConfig creates a new reloadable configuration
func NewReloadableConfig(configPath string) (*ReloadableConfig, error) {
	// Load initial configuration
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load initial config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	rc := &ReloadableConfig{
		config:      config,
		configPath:  configPath,
		subscribers: make([]chan ConfigChange, 0),
		ctx:         ctx,
		cancel:      cancel,
		lastReload:  time.Now(),
	}

	// Initialize feature manager
	rc.featureManager = NewFeatureManager(&config.Features)

	// Setup file watcher if hot reload is enabled
	if config.Features.EnableHotReload {
		if err := rc.setupFileWatcher(); err != nil {
			return nil, fmt.Errorf("failed to setup file watcher: %w", err)
		}
	}

	return rc, nil
}

// setupFileWatcher sets up the file system watcher for hot reloading
func (rc *ReloadableConfig) setupFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	rc.watcher = watcher

	// Watch the config file and directory
	configDir := filepath.Dir(rc.configPath)
	if err := watcher.Add(configDir); err != nil {
		return fmt.Errorf("failed to watch config directory: %w", err)
	}

	// Also watch the configs directory for environment-specific files
	if err := watcher.Add("configs"); err != nil {
		// Non-fatal error if configs directory doesn't exist
		fmt.Printf("Warning: Could not watch configs directory: %v\\n", err)
	}

	// Start the watcher goroutine
	go rc.watchFiles()

	return nil
}

// watchFiles monitors file system events and triggers reloads
func (rc *ReloadableConfig) watchFiles() {
	debounceTimer := time.NewTimer(0)
	if !debounceTimer.Stop() {
		<-debounceTimer.C
	}

	for {
		select {
		case event, ok := <-rc.watcher.Events:
			if !ok {
				return
			}

			// Check if this is a config file we care about
			if rc.isConfigFile(event.Name) {
				// Debounce file system events
				debounceTimer.Reset(500 * time.Millisecond)

				// Notify subscribers of file change
				rc.notifySubscribers(ConfigChange{
					Type:      "file_changed",
					Path:      event.Name,
					Timestamp: time.Now(),
				})
			}

		case err, ok := <-rc.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("File watcher error: %v\\n", err)

		case <-debounceTimer.C:
			// Trigger reload after debounce period
			rc.triggerReload()

		case <-rc.ctx.Done():
			return
		}
	}
}

// isConfigFile checks if the file is a configuration file we should watch
func (rc *ReloadableConfig) isConfigFile(filename string) bool {
	ext := filepath.Ext(filename)
	if ext != ".yaml" && ext != ".yml" && ext != ".json" {
		return false
	}

	base := filepath.Base(filename)
	configFiles := []string{
		"config.yaml", "config.yml",
		"development.yaml", "development.yml",
		"staging.yaml", "staging.yml",
		"production.yaml", "production.yml",
		"testing.yaml", "testing.yml",
		"local.yaml", "local.yml",
	}

	for _, configFile := range configFiles {
		if base == configFile {
			return true
		}
	}

	return false
}

// triggerReload attempts to reload the configuration
func (rc *ReloadableConfig) triggerReload() {
	oldConfig := rc.GetConfig()
	
	// Attempt to load new configuration
	newConfig, err := LoadConfig(rc.configPath)
	if err != nil {
		rc.reloadErrors++
		rc.notifySubscribers(ConfigChange{
			Type:      "reload_error",
			Path:      rc.configPath,
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	// Compare configurations
	differences := CompareConfigs(oldConfig, newConfig)

	// Update the configuration
	rc.mu.Lock()
	rc.config = newConfig
	rc.lastReload = time.Now()
	rc.reloadCount++
	rc.mu.Unlock()

	// Update feature manager
	if rc.featureManager != nil {
		rc.updateFeatureManager(newConfig)
	}

	// Notify subscribers of successful reload
	rc.notifySubscribers(ConfigChange{
		Type:        "reload_success",
		Path:        rc.configPath,
		OldConfig:   oldConfig,
		NewConfig:   newConfig,
		Timestamp:   time.Now(),
		Differences: differences,
	})

	fmt.Printf("Configuration reloaded successfully (%d changes detected)\\n", len(differences))
}

// updateFeatureManager updates the feature manager with new configuration
func (rc *ReloadableConfig) updateFeatureManager(newConfig *Config) {
	// Create a new feature manager with the updated configuration
	oldFeatureManager := rc.featureManager
	rc.featureManager = NewFeatureManager(&newConfig.Features)
	
	// Close the old feature manager
	if oldFeatureManager != nil {
		oldFeatureManager.Close()
	}
}

// GetConfig returns the current configuration (thread-safe)
func (rc *ReloadableConfig) GetConfig() *Config {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	// Return a copy to prevent external modifications
	configCopy := *rc.config
	return &configCopy
}

// GetFeatureManager returns the feature manager
func (rc *ReloadableConfig) GetFeatureManager() *FeatureManager {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	return rc.featureManager
}

// Subscribe subscribes to configuration change events
func (rc *ReloadableConfig) Subscribe(ch chan ConfigChange) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	
	rc.subscribers = append(rc.subscribers, ch)
}

// Unsubscribe unsubscribes from configuration change events
func (rc *ReloadableConfig) Unsubscribe(ch chan ConfigChange) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	
	for i, subscriber := range rc.subscribers {
		if subscriber == ch {
			rc.subscribers = append(rc.subscribers[:i], rc.subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

// notifySubscribers notifies all subscribers of configuration changes
func (rc *ReloadableConfig) notifySubscribers(change ConfigChange) {
	rc.mu.RLock()
	subscribers := make([]chan ConfigChange, len(rc.subscribers))
	copy(subscribers, rc.subscribers)
	rc.mu.RUnlock()
	
	for _, subscriber := range subscribers {
		select {
		case subscriber <- change:
		default:
			// Non-blocking send - if channel is full, skip
		}
	}
}

// Reload manually triggers a configuration reload
func (rc *ReloadableConfig) Reload() error {
	rc.triggerReload()
	return nil
}

// GetStats returns statistics about the reloadable configuration
func (rc *ReloadableConfig) GetStats() map[string]interface{} {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	
	stats := map[string]interface{}{
		"config_path":       rc.configPath,
		"last_reload":       rc.lastReload,
		"reload_count":      rc.reloadCount,
		"reload_errors":     rc.reloadErrors,
		"subscribers":       len(rc.subscribers),
		"hot_reload_enabled": rc.config.Features.EnableHotReload,
	}
	
	// Add feature manager stats if available
	if rc.featureManager != nil {
		stats["feature_flags"] = rc.featureManager.GetStats()
	}
	
	return stats
}

// Close closes the reloadable configuration and releases resources
func (rc *ReloadableConfig) Close() error {
	if rc.cancel != nil {
		rc.cancel()
	}
	
	if rc.watcher != nil {
		rc.watcher.Close()
	}
	
	if rc.featureManager != nil {
		rc.featureManager.Close()
	}
	
	// Close all subscriber channels
	rc.mu.Lock()
	for _, subscriber := range rc.subscribers {
		close(subscriber)
	}
	rc.subscribers = nil
	rc.mu.Unlock()
	
	return nil
}

// ValidateReload validates that a configuration reload would be successful
func (rc *ReloadableConfig) ValidateReload() error {
	// Try to load the configuration without actually applying it
	_, err := LoadConfig(rc.configPath)
	return err
}

// SetConfigPath updates the configuration path (useful for testing)
func (rc *ReloadableConfig) SetConfigPath(path string) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	
	// Validate that the new path is valid
	_, err := LoadConfig(path)
	if err != nil {
		return fmt.Errorf("invalid config path: %w", err)
	}
	
	rc.configPath = path
	
	// Update file watcher if enabled
	if rc.watcher != nil {
		// Remove old watcher
		rc.watcher.Close()
		
		// Setup new watcher
		if rc.config.Features.EnableHotReload {
			return rc.setupFileWatcher()
		}
	}
	
	return nil
}

// ReloadFromEnvironment reloads configuration for a specific environment
func (rc *ReloadableConfig) ReloadFromEnvironment(environment string) error {
	configPath := fmt.Sprintf("configs/%s.yaml", environment)
	newConfig, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config for environment %s: %w", environment, err)
	}
	
	oldConfig := rc.GetConfig()
	differences := CompareConfigs(oldConfig, newConfig)
	
	rc.mu.Lock()
	rc.config = newConfig
	rc.lastReload = time.Now()
	rc.reloadCount++
	rc.mu.Unlock()
	
	// Update feature manager
	if rc.featureManager != nil {
		rc.updateFeatureManager(newConfig)
	}
	
	// Notify subscribers
	rc.notifySubscribers(ConfigChange{
		Type:        "config_changed",
		Path:        fmt.Sprintf("environment: %s", environment),
		OldConfig:   oldConfig,
		NewConfig:   newConfig,
		Timestamp:   time.Now(),
		Differences: differences,
	})
	
	fmt.Printf("Configuration reloaded for environment '%s' (%d changes)\\n", environment, len(differences))
	
	return nil
}