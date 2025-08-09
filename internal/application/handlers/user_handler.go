// Application handler - demonstrates violations for testing
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
)

// This file intentionally contains violations to test our linters

// UserHandler handles user-related requests
type UserHandler struct {
	// This would be better with proper dependency injection
}

// CreateUser creates a new user - demonstrates potential issues
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var data interface{} // ❌ This should trigger forbidigo - interface{} violation
	
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		// Error not handled properly - should use structured response
		panic(err) // ❌ This should trigger forbidigo - panic violation
	}
	
	// Type assertion without checking - dangerous
	userData := data.(map[string]interface{}) // ❌ Multiple violations
	
	// Create user
	user, err := entities.NewUser(
		entities.UserID(userData["id"].(string)),
		userData["email"].(string), 
		userData["name"].(string),
	)
	_ = err // ❌ This should trigger errcheck - ignored error
	
	// Return success
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}