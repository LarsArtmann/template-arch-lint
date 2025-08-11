package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/LarsArtmann/template-arch-lint/internal/container"
)

// TestHTTPIntegration tests the full HTTP server integration
func TestHTTPIntegration(t *testing.T) {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create and initialize DI container
	diContainer := container.New()
	defer func() {
		if err := diContainer.Shutdown(); err != nil {
			t.Errorf("Error shutting down container: %v", err)
		}
	}()

	// Register all dependencies
	if err := diContainer.RegisterAll(); err != nil {
		t.Fatalf("Failed to register dependencies: %v", err)
	}

	// Get router from container
	injector := diContainer.GetInjector()
	router := do.MustInvoke[*gin.Engine](injector)

	// Test health endpoint
	t.Run("Health Check", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/health", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}

		if response["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %v", response["status"])
		}

		fmt.Printf("✅ Health check passed: %+v\n", response)
	})

	// Test user CRUD operations
	t.Run("User CRUD Operations", func(t *testing.T) {
		// Create a user
		userPayload := map[string]string{
			"id":    "testuser123",
			"email": "test@example.com",
			"name":  "Test User",
		}
		
		jsonPayload, err := json.Marshal(userPayload)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		
		// Create user
		req, err := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonPayload))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 201 {
			t.Errorf("Expected status 201 for user creation, got %d", w.Code)
			t.Errorf("Response body: %s", w.Body.String())
		} else {
			fmt.Printf("✅ User creation passed: %s\n", w.Body.String())
		}

		// Get user by ID
		req, err = http.NewRequest("GET", "/api/v1/users/testuser123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for user retrieval, got %d", w.Code)
			t.Errorf("Response body: %s", w.Body.String())
		} else {
			fmt.Printf("✅ User retrieval passed: %s\n", w.Body.String())
		}

		// List users
		req, err = http.NewRequest("GET", "/api/v1/users", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for user list, got %d", w.Code)
		} else {
			fmt.Printf("✅ User list passed: %s\n", w.Body.String())
		}

		// Update user
		updatePayload := map[string]string{
			"email": "updated@example.com",
			"name":  "Updated User",
		}
		
		jsonPayload, err = json.Marshal(updatePayload)
		if err != nil {
			t.Fatalf("Failed to marshal JSON: %v", err)
		}
		
		req, err = http.NewRequest("PUT", "/api/v1/users/testuser123", bytes.NewBuffer(jsonPayload))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for user update, got %d", w.Code)
			t.Errorf("Response body: %s", w.Body.String())
		} else {
			fmt.Printf("✅ User update passed: %s\n", w.Body.String())
		}

		// Delete user
		req, err = http.NewRequest("DELETE", "/api/v1/users/testuser123", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Expected status 200 for user deletion, got %d", w.Code)
		} else {
			fmt.Printf("✅ User deletion passed: %s\n", w.Body.String())
		}
	})

	// Test error handling
	t.Run("Error Handling", func(t *testing.T) {
		// Test invalid user ID
		req, err := http.NewRequest("GET", "/api/v1/users/invalid@user", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("Expected status 400 for invalid user ID, got %d", w.Code)
		} else {
			fmt.Printf("✅ Invalid user ID error handling passed: %s\n", w.Body.String())
		}

		// Test user not found
		req, err = http.NewRequest("GET", "/api/v1/users/nonexistent-user", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 404 {
			t.Errorf("Expected status 404 for nonexistent user, got %d", w.Code)
		} else {
			fmt.Printf("✅ User not found error handling passed: %s\n", w.Body.String())
		}

		// Test invalid JSON
		req, err = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer([]byte("invalid json")))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		} else {
			fmt.Printf("✅ Invalid JSON error handling passed: %s\n", w.Body.String())
		}
	})
}