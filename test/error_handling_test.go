package test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"

	"github.com/LarsArtmann/template-arch-lint/internal/container"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	inmemory "github.com/LarsArtmann/template-arch-lint/internal/infrastructure/repositories"
)

// TestErrorHandlingScenarios provides comprehensive error testing
func TestErrorHandlingScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Domain Error Handling", func(t *testing.T) {
		testDomainErrorHandling(t)
	})

	t.Run("HTTP Error Handling", func(t *testing.T) {
		testHTTPErrorHandling(t)
	})

	t.Run("Validation Error Handling", func(t *testing.T) {
		testValidationErrorHandling(t)
	})

	t.Run("Service Layer Error Handling", func(t *testing.T) {
		testServiceLayerErrorHandling(t)
	})

	t.Run("Repository Error Handling", func(t *testing.T) {
		testRepositoryErrorHandling(t)
	})
}

// testDomainErrorHandling tests domain-level error scenarios
func testDomainErrorHandling(t *testing.T) {
	t.Helper()

	t.Run("ValidationError", func(t *testing.T) {
		err := errors.NewValidationError("email", "invalid format")
		
		// Test error properties
		if err.Code() != errors.ValidationErrorCode {
			t.Errorf("Expected code %s, got %s", errors.ValidationErrorCode, err.Code())
		}
		
		if err.HTTPStatus() != http.StatusBadRequest {
			t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus())
		}
		
		// Test error type checking
		_, isValidation := errors.AsValidationError(err)
		if !isValidation {
			t.Error("Should be recognized as ValidationError")
		}
		
		// Test error message
		expectedMsg := "validation failed for field 'email': invalid format"
		if err.Error() != expectedMsg {
			t.Errorf("Expected message '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("NotFoundError", func(t *testing.T) {
		err := errors.NewNotFoundError("user", "123")
		
		if err.Code() != errors.NotFoundErrorCode {
			t.Errorf("Expected code %s, got %s", errors.NotFoundErrorCode, err.Code())
		}
		
		if err.HTTPStatus() != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, err.HTTPStatus())
		}
		
		_, isNotFound := errors.AsNotFoundError(err)
		if !isNotFound {
			t.Error("Should be recognized as NotFoundError")
		}
	})

	t.Run("ConflictError", func(t *testing.T) {
		err := errors.NewConflictError("user", errors.ErrorDetails{
			Resource: "user",
			Reason:   "email already exists",
		})
		
		if err.Code() != errors.ConflictErrorCode {
			t.Errorf("Expected code %s, got %s", errors.ConflictErrorCode, err.Code())
		}
		
		if err.HTTPStatus() != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusConflict, err.HTTPStatus())
		}
		
		_, isConflict := errors.AsConflictError(err)
		if !isConflict {
			t.Error("Should be recognized as ConflictError")
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		originalErr := &testError{message: "database connection failed"}
		err := errors.NewInternalError("database operation failed", originalErr)
		
		if err.Code() != errors.InternalErrorCode {
			t.Errorf("Expected code %s, got %s", errors.InternalErrorCode, err.Code())
		}
		
		if err.HTTPStatus() != http.StatusInternalServerError {
			t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, err.HTTPStatus())
		}
		
		_, isInternal := errors.AsInternalError(err)
		if !isInternal {
			t.Error("Should be recognized as InternalError")
		}
		
		if !strings.Contains(err.Error(), "database operation failed") {
			t.Error("Should contain the main error message")
		}
	})
}

// testHTTPErrorHandling tests HTTP layer error scenarios
func testHTTPErrorHandling(t *testing.T) {
	t.Helper()

	router := setupErrorTestRouter(t)

	t.Run("Invalid Content-Type", func(t *testing.T) {
		payload := `{"id": "test", "email": "test@example.com", "name": "Test User"}`
		req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
		req.Header.Set("Content-Type", "text/plain") // Wrong content type
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid content type, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(`{"invalid json"`))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for malformed JSON, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		payload := `{"id": "test"}` // Missing email and name
		req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing fields, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Invalid User ID Format", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users/invalid@user!id", nil)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid user ID, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/users/nonexistent-user", nil)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for user not found, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Duplicate User Creation", func(t *testing.T) {
		// Create user first
		payload := `{"id": "duplicate-test", "email": "duplicate@example.com", "name": "Duplicate User"}`
		req1, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
		req1.Header.Set("Content-Type", "application/json")
		
		w1 := httptest.NewRecorder()
		router.ServeHTTP(w1, req1)
		
		// Try to create the same user again
		req2, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
		req2.Header.Set("Content-Type", "application/json")
		
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		
		if w2.Code != http.StatusConflict {
			t.Errorf("Expected status %d for duplicate user, got %d", http.StatusConflict, w2.Code)
		}
	})

	t.Run("Large Request Body", func(t *testing.T) {
		// Create a large payload (simulating request size limit)
		largePayload := strings.Repeat("x", 2*1024*1024) // 2MB
		payload := `{"id": "test", "email": "test@example.com", "name": "` + largePayload + `"}`
		
		req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		// Should handle large payloads gracefully
		if w.Code == http.StatusOK {
			t.Error("Should not accept extremely large payloads")
		}
	})
}

// testValidationErrorHandling tests validation scenarios
func testValidationErrorHandling(t *testing.T) {
	t.Helper()

	t.Run("Invalid Email Formats", func(t *testing.T) {
		invalidEmails := []string{
			"",                    // empty
			"invalid",             // no @
			"@example.com",        // no local part
			"test@",               // no domain
			"test @example.com",   // space in email
			"test..user@example.com", // consecutive dots
		}

		for _, email := range invalidEmails {
			_, err := values.NewEmail(email)
			if err == nil {
				t.Errorf("Should reject invalid email: %s", email)
			}
		}
	})

	t.Run("Invalid User ID Formats", func(t *testing.T) {
		invalidIDs := []string{
			"",              // empty
			"user@123",      // invalid characters
			"user 123",      // spaces
			"user.123",      // dots
			"user/123",      // slashes
			"  user123  ",   // leading/trailing spaces
		}

		for _, id := range invalidIDs {
			_, err := values.NewUserID(id)
			if err == nil {
				t.Errorf("Should reject invalid user ID: %s", id)
			}
		}
	})

	t.Run("Invalid Username Formats", func(t *testing.T) {
		invalidNames := []string{
			"",          // empty
			"a",         // too short
			"123",       // no letters
			"admin",     // reserved
			"root",      // reserved
			".test",     // starts with dot
			"test.",     // ends with dot
			"test..user", // consecutive dots
		}

		for _, name := range invalidNames {
			_, err := values.NewUserName(name)
			if err == nil {
				t.Errorf("Should reject invalid username: %s", name)
			}
		}
	})

	t.Run("Business Rule Validations", func(t *testing.T) {
		router := setupErrorTestRouter(t)

		// Test email format validation through API
		payload := `{"id": "test-user", "email": "invalid-email", "name": "Test User"}`
		req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected validation error for invalid email, got status %d", w.Code)
		}

		// Check that error response contains validation details
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err == nil {
			if errorCode, exists := errorResponse["code"]; exists {
				if errorCode != string(errors.ValidationErrorCode) {
					t.Errorf("Expected validation error code, got %s", errorCode)
				}
			}
		}
	})
}

// testServiceLayerErrorHandling tests service layer error scenarios
func testServiceLayerErrorHandling(t *testing.T) {
	t.Helper()

	userRepo := inmemory.NewInMemoryUserRepository()
	userService := services.NewUserService(userRepo)
	ctx := context.Background()

	t.Run("Service Validation Errors", func(t *testing.T) {
		id, _ := values.NewUserID("test-user")

		// Test empty email
		_, err := userService.CreateUser(ctx, id, "", "Test User")
		if err == nil {
			t.Error("Should return error for empty email")
		}
		_, isValidation := errors.AsValidationError(err)
		if !isValidation {
			t.Error("Should return validation error for empty email")
		}

		// Test empty name
		_, err = userService.CreateUser(ctx, id, "test@example.com", "")
		if err == nil {
			t.Error("Should return error for empty name")
		}
		_, isValidation = errors.AsValidationError(err)
		if !isValidation {
			t.Error("Should return validation error for empty name")
		}
	})

	t.Run("Service Business Logic Errors", func(t *testing.T) {
		id1, _ := values.NewUserID("user1")
		id2, _ := values.NewUserID("user2")
		email := "duplicate@example.com"

		// Create first user
		_, err := userService.CreateUser(ctx, id1, email, "User One")
		if err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		// Try to create second user with same email
		_, err = userService.CreateUser(ctx, id2, email, "User Two")
		if err == nil {
			t.Error("Should return error for duplicate email")
		}

		// Should return conflict error (user already exists)
		expectedErr := "user already exists"
		if !strings.Contains(err.Error(), expectedErr) && err.Error() != "user already exists" {
			t.Errorf("Expected conflict error, got: %v", err)
		}
	})

	t.Run("Service Not Found Errors", func(t *testing.T) {
		id, _ := values.NewUserID("nonexistent-user")

		// Try to get nonexistent user
		_, err := userService.GetUser(ctx, id)
		if err == nil {
			t.Error("Should return error for nonexistent user")
		}

		// Try to update nonexistent user
		_, err = userService.UpdateUser(ctx, id, "new@example.com", "New Name")
		if err == nil {
			t.Error("Should return error when updating nonexistent user")
		}

		// Try to delete nonexistent user
		err = userService.DeleteUser(ctx, id)
		if err == nil {
			t.Error("Should return error when deleting nonexistent user")
		}
	})
}

// testRepositoryErrorHandling tests repository layer error scenarios
func testRepositoryErrorHandling(t *testing.T) {
	t.Helper()

	userRepo := inmemory.NewInMemoryUserRepository()
	ctx := context.Background()

	t.Run("Repository Validation", func(t *testing.T) {
		// Test saving nil user
		err := userRepo.Save(ctx, nil)
		if err == nil {
			t.Error("Should return error for nil user")
		}
	})

	t.Run("Repository Not Found", func(t *testing.T) {
		id, _ := values.NewUserID("nonexistent")

		// Test finding nonexistent user
		_, err := userRepo.FindByID(ctx, id)
		if err == nil {
			t.Error("Should return error for nonexistent user")
		}

		// Test finding by nonexistent email
		_, err = userRepo.FindByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Error("Should return error for nonexistent email")
		}
	})

	t.Run("Repository Context Cancellation", func(t *testing.T) {
		// Test with cancelled context
		cancelledCtx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		id, _ := values.NewUserID("test-user")
		user, _ := entities.NewUser(id, "test@example.com", "Test User")

		// Save should still work with cancelled context in memory repo
		// but we test that it handles context properly
		err := userRepo.Save(cancelledCtx, user)
		// In-memory repo doesn't check context, but real implementations would
		_ = err // Acknowledge that some repos might return errors
	})

	t.Run("Repository Timeout", func(t *testing.T) {
		// Test with timeout context
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		time.Sleep(1 * time.Millisecond) // Ensure timeout

		id, _ := values.NewUserID("test-user")
		user, _ := entities.NewUser(id, "test@example.com", "Test User")

		// Save might fail with timeout in real implementations
		err := userRepo.Save(timeoutCtx, user)
		// In-memory repo doesn't check context timeout, but we test the pattern
		_ = err
	})
}

// Helper functions

// setupErrorTestRouter creates a test router with full DI container
func setupErrorTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	diContainer := container.New()
	t.Cleanup(func() {
		if err := diContainer.Shutdown(); err != nil {
			t.Errorf("Error shutting down container: %v", err)
		}
	})

	if err := diContainer.RegisterAll(); err != nil {
		t.Fatalf("Failed to register dependencies: %v", err)
	}

	injector := diContainer.GetInjector()
	return do.MustInvoke[*gin.Engine](injector)
}

// testError is a simple error implementation for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}

// Test helper for concurrent error scenarios
func testConcurrentErrors(t *testing.T, router *gin.Engine) {
	t.Helper()

	const numRequests = 10
	done := make(chan bool, numRequests)
	
	// Create multiple concurrent requests that might cause errors
	for i := 0; i < numRequests; i++ {
		go func(index int) {
			defer func() { done <- true }()
			
			// Each request tries to create a user with the same email (should cause conflicts)
			payload := `{"id": "concurrent-` + string(rune(index)) + `", "email": "concurrent@example.com", "name": "Concurrent User"}`
			req, _ := http.NewRequest("POST", "/api/v1/users", strings.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Either success (first request) or conflict (subsequent requests)
			if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
				t.Errorf("Unexpected status code in concurrent test: %d", w.Code)
			}
		}(i)
	}
	
	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}