package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// FeatureFlag represents a single feature flag with metadata
type FeatureFlag struct {
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	Description string                 `json:"description"`
	Conditions  []FeatureCondition     `json:"conditions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Version     int                    `json:"version"`
}

// FeatureCondition represents conditions for feature flag evaluation
type FeatureCondition struct {
	Type     string                 `json:"type"`     // "user_id", "percentage", "environment", "time_window"
	Operator string                 `json:"operator"` // "equals", "in", "percentage", "between"
	Values   []interface{}          `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FeatureContext contains context information for feature flag evaluation
type FeatureContext struct {
	UserID      string            `json:"user_id,omitempty"`
	Environment string            `json:"environment"`
	Timestamp   time.Time         `json:"timestamp"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// FeatureManager manages feature flags and runtime configuration
type FeatureManager struct {
	flags          map[string]*FeatureFlag
	config         *FeaturesConfig
	subscribers    []chan FeatureUpdate
	mu             sync.RWMutex
	refreshTicker  *time.Ticker
	ctx            context.Context
	cancel         context.CancelFunc
}

// FeatureUpdate represents a feature flag update
type FeatureUpdate struct {
	Flag      *FeatureFlag `json:"flag"`
	Action    string       `json:"action"` // "created", "updated", "deleted"
	Timestamp time.Time    `json:"timestamp"`
}

// NewFeatureManager creates a new feature manager
func NewFeatureManager(config *FeaturesConfig) *FeatureManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	fm := &FeatureManager{
		flags:       make(map[string]*FeatureFlag),
		config:      config,
		subscribers: make([]chan FeatureUpdate, 0),
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// Initialize built-in feature flags from config
	fm.initializeBuiltInFlags()
	
	// Start refresh ticker if configured
	if config != nil {
		fm.startRefreshTicker()
	}
	
	return fm
}

// initializeBuiltInFlags creates feature flags from the static configuration
func (fm *FeatureManager) initializeBuiltInFlags() {
	if fm.config == nil {
		return
	}
	
	now := time.Now()
	
	// Create flags from the static configuration
	builtInFlags := map[string]*FeatureFlag{
		"debug_endpoints": {
			Name:        "debug_endpoints",
			Enabled:     fm.config.EnableDebugEndpoints,
			Description: "Enable debug endpoints for troubleshooting",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
		"profiling": {
			Name:        "profiling",
			Enabled:     fm.config.EnableProfiling,
			Description: "Enable performance profiling endpoints",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
		"hot_reload": {
			Name:        "hot_reload",
			Enabled:     fm.config.EnableHotReload,
			Description: "Enable hot reloading of configuration and templates",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
		"cors": {
			Name:        "cors",
			Enabled:     fm.config.EnableCORS,
			Description: "Enable Cross-Origin Resource Sharing (CORS)",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
		"request_logging": {
			Name:        "request_logging",
			Enabled:     fm.config.EnableRequestLogging,
			Description: "Enable detailed request logging",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
		"query_logging": {
			Name:        "query_logging",
			Enabled:     fm.config.EnableQueryLogging,
			Description: "Enable database query logging",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
		"metrics_debug": {
			Name:        "metrics_debug",
			Enabled:     fm.config.EnableMetricsDebug,
			Description: "Enable debug-level metrics collection",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
		"beta_features": {
			Name:        "beta_features",
			Enabled:     fm.config.EnableBetaFeatures,
			Description: "Enable access to beta features",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
		"load_testing": {
			Name:        "load_testing",
			Enabled:     fm.config.EnableLoadTesting,
			Description: "Enable load testing utilities and endpoints",
			CreatedAt:   now,
			UpdatedAt:   now,
			Version:     1,
		},
	}
	
	fm.mu.Lock()
	for name, flag := range builtInFlags {
		fm.flags[name] = flag
	}
	fm.mu.Unlock()
}

// startRefreshTicker starts the periodic refresh of feature flags
func (fm *FeatureManager) startRefreshTicker() {
	// In a real implementation, you might refresh flags from an external source
	// For now, we'll just check for configuration changes
	fm.refreshTicker = time.NewTicker(30 * time.Second)
	
	go func() {
		for {
			select {
			case <-fm.refreshTicker.C:
				// In a real implementation, you would fetch updates from a remote source
				// For now, this is a placeholder for runtime configuration updates
			case <-fm.ctx.Done():
				return
			}
		}
	}()
}

// IsEnabled checks if a feature flag is enabled
func (fm *FeatureManager) IsEnabled(flagName string) bool {
	return fm.IsEnabledForContext(flagName, FeatureContext{
		Environment: "default",
		Timestamp:   time.Now(),
	})
}

// IsEnabledForContext checks if a feature flag is enabled for a specific context
func (fm *FeatureManager) IsEnabledForContext(flagName string, ctx FeatureContext) bool {
	fm.mu.RLock()
	flag, exists := fm.flags[flagName]
	fm.mu.RUnlock()
	
	if !exists {
		return false
	}
	
	// If no conditions, return the base enabled state
	if len(flag.Conditions) == 0 {
		return flag.Enabled
	}
	
	// Evaluate conditions
	for _, condition := range flag.Conditions {
		if !fm.evaluateCondition(condition, ctx) {
			return false
		}
	}
	
	return flag.Enabled
}

// evaluateCondition evaluates a single feature flag condition
func (fm *FeatureManager) evaluateCondition(condition FeatureCondition, ctx FeatureContext) bool {
	switch condition.Type {
	case "environment":
		return fm.evaluateEnvironmentCondition(condition, ctx)
	case "percentage":
		return fm.evaluatePercentageCondition(condition, ctx)
	case "user_id":
		return fm.evaluateUserIDCondition(condition, ctx)
	case "time_window":
		return fm.evaluateTimeWindowCondition(condition, ctx)
	default:
		return false
	}
}

// evaluateEnvironmentCondition evaluates environment-based conditions
func (fm *FeatureManager) evaluateEnvironmentCondition(condition FeatureCondition, ctx FeatureContext) bool {
	switch condition.Operator {
	case "equals":
		if len(condition.Values) > 0 {
			if env, ok := condition.Values[0].(string); ok {
				return ctx.Environment == env
			}
		}
	case "in":
		for _, value := range condition.Values {
			if env, ok := value.(string); ok && ctx.Environment == env {
				return true
			}
		}
	}
	return false
}

// evaluatePercentageCondition evaluates percentage-based rollout conditions
func (fm *FeatureManager) evaluatePercentageCondition(condition FeatureCondition, ctx FeatureContext) bool {
	if len(condition.Values) == 0 {
		return false
	}
	
	percentage, ok := condition.Values[0].(float64)
	if !ok {
		return false
	}
	
	// Simple hash-based percentage calculation
	hash := fm.hashString(ctx.UserID + condition.Type)
	return float64(hash%100) < percentage
}

// evaluateUserIDCondition evaluates user ID-based conditions
func (fm *FeatureManager) evaluateUserIDCondition(condition FeatureCondition, ctx FeatureContext) bool {
	switch condition.Operator {
	case "equals":
		if len(condition.Values) > 0 {
			if userID, ok := condition.Values[0].(string); ok {
				return ctx.UserID == userID
			}
		}
	case "in":
		for _, value := range condition.Values {
			if userID, ok := value.(string); ok && ctx.UserID == userID {
				return true
			}
		}
	}
	return false
}

// evaluateTimeWindowCondition evaluates time-based conditions
func (fm *FeatureManager) evaluateTimeWindowCondition(condition FeatureCondition, ctx FeatureContext) bool {
	if len(condition.Values) < 2 {
		return false
	}
	
	startTime, ok1 := condition.Values[0].(string)
	endTime, ok2 := condition.Values[1].(string)
	
	if !ok1 || !ok2 {
		return false
	}
	
	start, err1 := time.Parse(time.RFC3339, startTime)
	end, err2 := time.Parse(time.RFC3339, endTime)
	
	if err1 != nil || err2 != nil {
		return false
	}
	
	return ctx.Timestamp.After(start) && ctx.Timestamp.Before(end)
}

// hashString creates a simple hash of a string for percentage calculations
func (fm *FeatureManager) hashString(s string) int {
	hash := 0
	for _, char := range s {
		hash = hash*31 + int(char)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

// GetFlag returns a feature flag by name
func (fm *FeatureManager) GetFlag(flagName string) (*FeatureFlag, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	flag, exists := fm.flags[flagName]
	if !exists {
		return nil, false
	}
	
	// Return a copy to prevent modification
	flagCopy := *flag
	return &flagCopy, true
}

// ListFlags returns all feature flags
func (fm *FeatureManager) ListFlags() map[string]*FeatureFlag {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	result := make(map[string]*FeatureFlag)
	for name, flag := range fm.flags {
		flagCopy := *flag
		result[name] = &flagCopy
	}
	
	return result
}

// UpdateFlag updates a feature flag
func (fm *FeatureManager) UpdateFlag(flag *FeatureFlag) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	if flag.Name == "" {
		return fmt.Errorf("flag name cannot be empty")
	}
	
	now := time.Now()
	flag.UpdatedAt = now
	flag.Version++
	
	oldFlag, exists := fm.flags[flag.Name]
	fm.flags[flag.Name] = flag
	
	// Notify subscribers
	update := FeatureUpdate{
		Flag:      flag,
		Timestamp: now,
	}
	
	if exists {
		update.Action = "updated"
	} else {
		update.Action = "created"
		flag.CreatedAt = now
	}
	
	go fm.notifySubscribers(update)
	
	// Log the change
	if exists {
		fmt.Printf("Feature flag '%s' updated (v%d -> v%d)\\n", flag.Name, oldFlag.Version, flag.Version)
	} else {
		fmt.Printf("Feature flag '%s' created (v%d)\\n", flag.Name, flag.Version)
	}
	
	return nil
}

// DeleteFlag deletes a feature flag
func (fm *FeatureManager) DeleteFlag(flagName string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	flag, exists := fm.flags[flagName]
	if !exists {
		return fmt.Errorf("flag '%s' not found", flagName)
	}
	
	delete(fm.flags, flagName)
	
	// Notify subscribers
	update := FeatureUpdate{
		Flag:      flag,
		Action:    "deleted",
		Timestamp: time.Now(),
	}
	
	go fm.notifySubscribers(update)
	
	fmt.Printf("Feature flag '%s' deleted\\n", flagName)
	
	return nil
}

// Subscribe subscribes to feature flag updates
func (fm *FeatureManager) Subscribe(ch chan FeatureUpdate) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	fm.subscribers = append(fm.subscribers, ch)
}

// Unsubscribe unsubscribes from feature flag updates
func (fm *FeatureManager) Unsubscribe(ch chan FeatureUpdate) {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	
	for i, subscriber := range fm.subscribers {
		if subscriber == ch {
			fm.subscribers = append(fm.subscribers[:i], fm.subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

// notifySubscribers notifies all subscribers of feature flag updates
func (fm *FeatureManager) notifySubscribers(update FeatureUpdate) {
	fm.mu.RLock()
	subscribers := make([]chan FeatureUpdate, len(fm.subscribers))
	copy(subscribers, fm.subscribers)
	fm.mu.RUnlock()
	
	for _, subscriber := range subscribers {
		select {
		case subscriber <- update:
		default:
			// Non-blocking send - if the channel is full, skip
		}
	}
}

// ExportFlags exports all feature flags to JSON
func (fm *FeatureManager) ExportFlags() ([]byte, error) {
	flags := fm.ListFlags()
	return json.MarshalIndent(flags, "", "  ")
}

// ImportFlags imports feature flags from JSON
func (fm *FeatureManager) ImportFlags(data []byte) error {
	var flags map[string]*FeatureFlag
	if err := json.Unmarshal(data, &flags); err != nil {
		return fmt.Errorf("failed to unmarshal flags: %w", err)
	}
	
	for _, flag := range flags {
		if err := fm.UpdateFlag(flag); err != nil {
			return fmt.Errorf("failed to update flag '%s': %w", flag.Name, err)
		}
	}
	
	return nil
}

// Close closes the feature manager and releases resources
func (fm *FeatureManager) Close() error {
	if fm.cancel != nil {
		fm.cancel()
	}
	
	if fm.refreshTicker != nil {
		fm.refreshTicker.Stop()
	}
	
	// Close all subscriber channels
	fm.mu.Lock()
	for _, subscriber := range fm.subscribers {
		close(subscriber)
	}
	fm.subscribers = nil
	fm.mu.Unlock()
	
	return nil
}

// GetStats returns statistics about the feature manager
func (fm *FeatureManager) GetStats() map[string]interface{} {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	
	enabledCount := 0
	disabledCount := 0
	conditionalCount := 0
	
	for _, flag := range fm.flags {
		if flag.Enabled {
			enabledCount++
		} else {
			disabledCount++
		}
		
		if len(flag.Conditions) > 0 {
			conditionalCount++
		}
	}
	
	return map[string]interface{}{
		"total_flags":       len(fm.flags),
		"enabled_flags":     enabledCount,
		"disabled_flags":    disabledCount,
		"conditional_flags": conditionalCount,
		"subscribers":       len(fm.subscribers),
	}
}