package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// RuntimeConfigHandler provides HTTP endpoints for runtime configuration management
type RuntimeConfigHandler struct {
	reloadableConfig *ReloadableConfig
}

// NewRuntimeConfigHandler creates a new runtime configuration handler
func NewRuntimeConfigHandler(reloadableConfig *ReloadableConfig) *RuntimeConfigHandler {
	return &RuntimeConfigHandler{
		reloadableConfig: reloadableConfig,
	}
}

// HandleGetConfig returns the current configuration
func (rch *RuntimeConfigHandler) HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := rch.reloadableConfig.GetConfig()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode config: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleReloadConfig triggers a configuration reload
func (rch *RuntimeConfigHandler) HandleReloadConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := rch.reloadableConfig.Reload(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reload config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "success",
		"message":   "Configuration reloaded successfully",
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleValidateConfig validates the current configuration
func (rch *RuntimeConfigHandler) HandleValidateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := rch.reloadableConfig.ValidateReload(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":  "success",
		"message": "Configuration is valid",
	}
	json.NewEncoder(w).Encode(response)
}

// HandleGetStats returns configuration statistics
func (rch *RuntimeConfigHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := rch.reloadableConfig.GetStats()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode stats: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleGetFeatureFlags returns all feature flags
func (rch *RuntimeConfigHandler) HandleGetFeatureFlags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	featureManager := rch.reloadableConfig.GetFeatureManager()
	if featureManager == nil {
		http.Error(w, "Feature manager not available", http.StatusServiceUnavailable)
		return
	}

	flags := featureManager.ListFlags()
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(flags); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode flags: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleGetFeatureFlag returns a specific feature flag
func (rch *RuntimeConfigHandler) HandleGetFeatureFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flagName := r.URL.Query().Get("name")
	if flagName == "" {
		http.Error(w, "Flag name is required", http.StatusBadRequest)
		return
	}

	featureManager := rch.reloadableConfig.GetFeatureManager()
	if featureManager == nil {
		http.Error(w, "Feature manager not available", http.StatusServiceUnavailable)
		return
	}

	flag, exists := featureManager.GetFlag(flagName)
	if !exists {
		http.Error(w, "Flag not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(flag); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode flag: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleCheckFeatureFlag checks if a feature flag is enabled for a context
func (rch *RuntimeConfigHandler) HandleCheckFeatureFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		FlagName string         `json:"flag_name"`
		Context  FeatureContext `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	featureManager := rch.reloadableConfig.GetFeatureManager()
	if featureManager == nil {
		http.Error(w, "Feature manager not available", http.StatusServiceUnavailable)
		return
	}

	enabled := featureManager.IsEnabledForContext(request.FlagName, request.Context)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"flag_name": request.FlagName,
		"enabled":   enabled,
		"context":   request.Context,
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleUpdateFeatureFlag updates a feature flag
func (rch *RuntimeConfigHandler) HandleUpdateFeatureFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var flag FeatureFlag
	if err := json.NewDecoder(r.Body).Decode(&flag); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	featureManager := rch.reloadableConfig.GetFeatureManager()
	if featureManager == nil {
		http.Error(w, "Feature manager not available", http.StatusServiceUnavailable)
		return
	}

	if err := featureManager.UpdateFlag(&flag); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update flag: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "success",
		"message":   fmt.Sprintf("Feature flag '%s' updated successfully", flag.Name),
		"flag":      flag,
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleToggleFeatureFlag toggles a feature flag on/off
func (rch *RuntimeConfigHandler) HandleToggleFeatureFlag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flagName := r.URL.Query().Get("name")
	if flagName == "" {
		http.Error(w, "Flag name is required", http.StatusBadRequest)
		return
	}

	enabledStr := r.URL.Query().Get("enabled")
	enabled, err := strconv.ParseBool(enabledStr)
	if err != nil {
		http.Error(w, "Invalid enabled value (must be true or false)", http.StatusBadRequest)
		return
	}

	featureManager := rch.reloadableConfig.GetFeatureManager()
	if featureManager == nil {
		http.Error(w, "Feature manager not available", http.StatusServiceUnavailable)
		return
	}

	flag, exists := featureManager.GetFlag(flagName)
	if !exists {
		http.Error(w, "Flag not found", http.StatusNotFound)
		return
	}

	flag.Enabled = enabled
	if err := featureManager.UpdateFlag(flag); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update flag: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "success",
		"message":   fmt.Sprintf("Feature flag '%s' %s", flag.Name, map[bool]string{true: "enabled", false: "disabled"}[enabled]),
		"flag":      flag,
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers all runtime configuration routes
func (rch *RuntimeConfigHandler) RegisterRoutes(mux *http.ServeMux, pathPrefix string) {
	if pathPrefix == "" {
		pathPrefix = "/api/config"
	}

	// Configuration endpoints
	mux.HandleFunc(pathPrefix+"/", rch.HandleGetConfig)
	mux.HandleFunc(pathPrefix+"/reload", rch.HandleReloadConfig)
	mux.HandleFunc(pathPrefix+"/validate", rch.HandleValidateConfig)
	mux.HandleFunc(pathPrefix+"/stats", rch.HandleGetStats)

	// Feature flag endpoints
	mux.HandleFunc(pathPrefix+"/flags", rch.HandleGetFeatureFlags)
	mux.HandleFunc(pathPrefix+"/flags/get", rch.HandleGetFeatureFlag)
	mux.HandleFunc(pathPrefix+"/flags/check", rch.HandleCheckFeatureFlag)
	mux.HandleFunc(pathPrefix+"/flags/update", rch.HandleUpdateFeatureFlag)
	mux.HandleFunc(pathPrefix+"/flags/toggle", rch.HandleToggleFeatureFlag)
}

// ConfigMiddleware provides configuration-aware middleware
type ConfigMiddleware struct {
	reloadableConfig *ReloadableConfig
}

// NewConfigMiddleware creates a new configuration middleware
func NewConfigMiddleware(reloadableConfig *ReloadableConfig) *ConfigMiddleware {
	return &ConfigMiddleware{
		reloadableConfig: reloadableConfig,
	}
}

// FeatureGate creates middleware that checks feature flags
func (cm *ConfigMiddleware) FeatureGate(flagName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureManager := cm.reloadableConfig.GetFeatureManager()
			if featureManager == nil || !featureManager.IsEnabled(flagName) {
				http.Error(w, "Feature not available", http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ConditionalFeatureGate creates middleware that checks feature flags with context
func (cm *ConfigMiddleware) ConditionalFeatureGate(flagName string, contextFunc func(*http.Request) FeatureContext) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			featureManager := cm.reloadableConfig.GetFeatureManager()
			if featureManager == nil {
				http.Error(w, "Feature manager not available", http.StatusServiceUnavailable)
				return
			}

			ctx := contextFunc(r)
			if !featureManager.IsEnabledForContext(flagName, ctx) {
				http.Error(w, "Feature not available", http.StatusNotFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ConfigHeaders adds configuration information to response headers
func (cm *ConfigMiddleware) ConfigHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			config := cm.reloadableConfig.GetConfig()
			
			w.Header().Set("X-App-Name", config.App.Name)
			w.Header().Set("X-App-Version", config.App.Version)
			w.Header().Set("X-App-Environment", config.App.Environment)
			
			next.ServeHTTP(w, r)
		})
	}
}