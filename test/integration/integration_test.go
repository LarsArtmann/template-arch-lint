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

const (
	// HTTP status codes
	StatusOK                = 200
	StatusCreated           = 201
	StatusBadRequest        = 400
	StatusNotFound          = 404
	// HTTP methods
	MethodGET               = "GET"
	MethodPOST              = "POST"
	MethodPUT               = "PUT"
	MethodDELETE            = "DELETE"
	// Common strings
	ContentTypeHeader       = "Content-Type"
	ApplicationJSON         = "application/json"
	ResponseBodyMsg         = "Response body: %s"
	FailedToCreateReqMsg    = "Failed to create request: %v"
	// API endpoints
	UsersEndpoint           = "/api/v1/users"
	UserEndpoint            = "/api/v1/users/testuser123"
)

// Test response types
type HealthResponse struct {
	Status string `json:"status"`
}

// TestHTTPIntegration tests the full HTTP server integration
func TestHTTPIntegration(t *testing.T) {
	router := setupTestRouter(t)

	t.Run("Health Check", func(t *testing.T) {
		testHealthEndpoint(t, router)
	})

	t.Run("User CRUD Operations", func(t *testing.T) {
		testUserCRUDOperations(t, router)
	})

	t.Run("Error Handling", func(t *testing.T) {
		testErrorHandling(t, router)
	})
}

// setupTestRouter creates and configures the test router with dependencies
func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Create and initialize DI container
	diContainer := container.New()
	t.Cleanup(func() {
		if err := diContainer.Shutdown(); err != nil {
			t.Errorf("Error shutting down container: %v", err)
		}
	})

	// Register all dependencies
	if err := diContainer.RegisterAll(); err != nil {
		t.Fatalf("Failed to register dependencies: %v", err)
	}

	// Get router from container
	injector := diContainer.GetInjector()
	return do.MustInvoke[*gin.Engine](injector)
}

// testHealthEndpoint tests the health check endpoint
func testHealthEndpoint(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	req, err := http.NewRequest("GET", "/health", http.NoBody)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got %v", response.Status)
	}

	_, _ = fmt.Printf("✅ Health check passed: %+v\n", response)
}

// testUserCRUDOperations tests all user CRUD operations
func testUserCRUDOperations(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	testUserCreation(t, router)
	testUserRetrieval(t, router)
	testUserListing(t, router)
	testUserUpdate(t, router)
	testUserDeletion(t, router)
}

// testUserCreation tests user creation endpoint
func testUserCreation(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	userPayload := map[string]string{
		"id":    "testuser123",
		"email": "test@example.com",
		"name":  "Test User",
	}
	
	jsonPayload, err := json.Marshal(userPayload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	
	req, err := http.NewRequest(
		MethodPOST,
		UsersEndpoint,
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	req.Header.Set(ContentTypeHeader, ApplicationJSON)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusCreated {
		t.Errorf("Expected status 201 for user creation, got %d", w.Code)
		t.Errorf(ResponseBodyMsg, w.Body.String())
	} else {
		_, _ = fmt.Printf("✅ User creation passed: %s\n", w.Body.String())
	}
}

// testUserRetrieval tests user retrieval by ID
func testUserRetrieval(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	req, err := http.NewRequest(
		MethodGET,
		UserEndpoint,
		http.NoBody,
	)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusOK {
		t.Errorf("Expected status 200 for user retrieval, got %d", w.Code)
		t.Errorf(ResponseBodyMsg, w.Body.String())
	} else {
		_, _ = fmt.Printf("✅ User retrieval passed: %s\n", w.Body.String())
	}
}

// testUserListing tests user listing endpoint
func testUserListing(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	req, err := http.NewRequest(MethodGET, UsersEndpoint, http.NoBody)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusOK {
		t.Errorf("Expected status 200 for user list, got %d", w.Code)
	} else {
		_, _ = fmt.Printf("✅ User list passed: %s\n", w.Body.String())
	}
}

// testUserUpdate tests user update endpoint
func testUserUpdate(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	updatePayload := map[string]string{
		"email": "updated@example.com",
		"name":  "Updated User",
	}
	
	jsonPayload, err := json.Marshal(updatePayload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	
	req, err := http.NewRequest(
		MethodPUT,
		UserEndpoint,
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	req.Header.Set(ContentTypeHeader, ApplicationJSON)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusOK {
		t.Errorf("Expected status 200 for user update, got %d", w.Code)
		t.Errorf(ResponseBodyMsg, w.Body.String())
	} else {
		_, _ = fmt.Printf("✅ User update passed: %s\n", w.Body.String())
	}
}

// testUserDeletion tests user deletion endpoint
func testUserDeletion(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	req, err := http.NewRequest(
		MethodDELETE,
		UserEndpoint,
		http.NoBody,
	)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusOK {
		t.Errorf("Expected status 200 for user deletion, got %d", w.Code)
	} else {
		_, _ = fmt.Printf("✅ User deletion passed: %s\n", w.Body.String())
	}
}

// testErrorHandling tests various error scenarios
func testErrorHandling(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	testInvalidUserIDError(t, router)
	testUserNotFoundError(t, router)
	testInvalidJSONError(t, router)
}

// testInvalidUserIDError tests error handling for invalid user IDs
func testInvalidUserIDError(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	req, err := http.NewRequest(
		MethodGET,
		"/api/v1/users/invalid@user",
		http.NoBody,
	)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusBadRequest {
		t.Errorf("Expected status 400 for invalid user ID, got %d", w.Code)
	} else {
		_, _ = fmt.Printf(
			"✅ Invalid user ID error handling passed: %s\n",
			w.Body.String(),
		)
	}
}

// testUserNotFoundError tests error handling for non-existent users
func testUserNotFoundError(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	req, err := http.NewRequest(
		MethodGET,
		"/api/v1/users/nonexistent-user",
		http.NoBody,
	)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusNotFound {
		t.Errorf("Expected status 404 for nonexistent user, got %d", w.Code)
	} else {
		_, _ = fmt.Printf(
			"✅ User not found error handling passed: %s\n",
			w.Body.String(),
		)
	}
}

// testInvalidJSONError tests error handling for invalid JSON payloads
func testInvalidJSONError(t *testing.T, router *gin.Engine) {
	t.Helper()
	
	req, err := http.NewRequest(
		MethodPOST,
		UsersEndpoint,
		bytes.NewBuffer([]byte("invalid json")),
	)
	if err != nil {
		t.Fatalf(FailedToCreateReqMsg, err)
	}
	req.Header.Set(ContentTypeHeader, ApplicationJSON)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != StatusBadRequest {
		t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
	} else {
		_, _ = fmt.Printf(
			"✅ Invalid JSON error handling passed: %s\n",
			w.Body.String(),
		)
	}
}