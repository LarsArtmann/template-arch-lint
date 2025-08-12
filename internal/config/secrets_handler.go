package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SecretsHandler provides HTTP endpoints for secrets management
type SecretsHandler struct {
	secretsManager *SecretsManager
}

// NewSecretsHandler creates a new secrets handler
func NewSecretsHandler(secretsManager *SecretsManager) *SecretsHandler {
	return &SecretsHandler{
		secretsManager: secretsManager,
	}
}

// SecretRequest represents a request to get/set a secret
type SecretRequest struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

// SecretResponse represents a response containing secret information
type SecretResponse struct {
	Key       string    `json:"key"`
	Exists    bool      `json:"exists"`
	Timestamp time.Time `json:"timestamp"`
	Provider  string    `json:"provider,omitempty"`
}

// SecretsListResponse represents a response containing list of secret keys
type SecretsListResponse struct {
	Keys      []string  `json:"keys"`
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

// HandleGetSecret retrieves a secret (returns only metadata, not the actual value for security)
func (sh *SecretsHandler) HandleGetSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	_, err := sh.secretsManager.GetSecret(ctx, key)
	
	response := SecretResponse{
		Key:       key,
		Exists:    err == nil,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

// HandleSetSecret stores a secret
func (sh *SecretsHandler) HandleSetSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request SecretRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if request.Key == "" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}

	if request.Value == "" {
		http.Error(w, "Value is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := sh.secretsManager.SetSecret(ctx, request.Key, request.Value); err != nil {
		http.Error(w, fmt.Sprintf("Failed to set secret: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "success",
		"message":   fmt.Sprintf("Secret '%s' set successfully", request.Key),
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleDeleteSecret deletes a secret
func (sh *SecretsHandler) HandleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := sh.secretsManager.DeleteSecret(ctx, key); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete secret: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "success",
		"message":   fmt.Sprintf("Secret '%s' deleted successfully", key),
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// HandleListSecrets lists all available secret keys (not values)
func (sh *SecretsHandler) HandleListSecrets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For security reasons, we don't actually list all secrets
	// This endpoint would be restricted to admin users in a real implementation
	
	w.Header().Set("Content-Type", "application/json")
	response := SecretsListResponse{
		Keys:      []string{}, // Empty for security
		Count:     0,
		Timestamp: time.Now(),
	}
	
	// In a real implementation, you might want to:
	// 1. Check if user has admin privileges
	// 2. Return only certain types of secrets
	// 3. Return masked/filtered results
	
	json.NewEncoder(w).Encode(response)
}

// HandleRotateSecret rotates a secret (generates a new value)
func (sh *SecretsHandler) HandleRotateSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	// Generate a new secret value (this is a simplified example)
	newValue := sh.generateSecretValue()

	ctx := r.Context()
	if err := sh.secretsManager.SetSecret(ctx, key, newValue); err != nil {
		http.Error(w, fmt.Sprintf("Failed to rotate secret: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "success",
		"message":   fmt.Sprintf("Secret '%s' rotated successfully", key),
		"timestamp": time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// generateSecretValue generates a new secret value (simplified implementation)
func (sh *SecretsHandler) generateSecretValue() string {
	// In a real implementation, you would use a cryptographically secure
	// random generator and generate appropriate length/complexity
	return fmt.Sprintf("secret_%d", time.Now().UnixNano())
}

// HandleHealthCheck checks the health of secrets providers
func (sh *SecretsHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check availability of all providers
	providerStatuses := make(map[string]bool)
	
	// This is a simplified check - in reality you would iterate through
	// actual providers and check their health
	providerStatuses["environment"] = true
	providerStatuses["file"] = true
	providerStatuses["vault"] = false // Mock: assume Vault is not available
	providerStatuses["kubernetes"] = false // Mock: assume K8s is not available

	allHealthy := true
	for _, healthy := range providerStatuses {
		if !healthy {
			allHealthy = false
			break
		}
	}

	status := "healthy"
	statusCode := http.StatusOK
	
	if !allHealthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"status":            status,
		"providers":         providerStatuses,
		"timestamp":         time.Now(),
		"cache_entries":     len(sh.secretsManager.cache), // This would need to be exposed properly
	}
	
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers all secrets management routes
func (sh *SecretsHandler) RegisterRoutes(mux *http.ServeMux, pathPrefix string) {
	if pathPrefix == "" {
		pathPrefix = "/api/secrets"
	}

	mux.HandleFunc(pathPrefix+"/get", sh.HandleGetSecret)
	mux.HandleFunc(pathPrefix+"/set", sh.HandleSetSecret)
	mux.HandleFunc(pathPrefix+"/delete", sh.HandleDeleteSecret)
	mux.HandleFunc(pathPrefix+"/list", sh.HandleListSecrets)
	mux.HandleFunc(pathPrefix+"/rotate", sh.HandleRotateSecret)
	mux.HandleFunc(pathPrefix+"/health", sh.HandleHealthCheck)
}

// SecretsMiddleware provides secrets-aware middleware
type SecretsMiddleware struct {
	secretsManager *SecretsManager
}

// NewSecretsMiddleware creates a new secrets middleware
func NewSecretsMiddleware(secretsManager *SecretsManager) *SecretsMiddleware {
	return &SecretsMiddleware{
		secretsManager: secretsManager,
	}
}

// RequireSecret creates middleware that requires a specific secret to be present
func (sm *SecretsMiddleware) RequireSecret(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			_, err := sm.secretsManager.GetSecret(ctx, secretKey)
			if err != nil {
				http.Error(w, "Required secret not available", http.StatusServiceUnavailable)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// InjectSecret creates middleware that injects a secret into the request context
func (sm *SecretsMiddleware) InjectSecret(secretKey, contextKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			_, err := sm.secretsManager.GetSecret(ctx, secretKey)
			if err != nil {
				// Log error but continue - the handler can decide how to handle missing secrets
				fmt.Printf("Warning: Could not retrieve secret '%s': %v\\n", secretKey, err)
			} else {
				// In a real implementation, you would use a proper context key type
				ctx = r.Context() // This is simplified - you'd inject the secret properly
			}
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}