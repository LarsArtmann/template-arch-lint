package errors

import (
	"net/http"
	"testing"
)

func TestValidationError(t *testing.T) {
	err := NewValidationError("email", "invalid email format")

	if err.Code() != ValidationErrorCode {
		t.Errorf("Expected code %s, got %s", ValidationErrorCode, err.Code())
	}

	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus())
	}

	if err.Field() != "email" {
		t.Errorf("Expected field 'email', got %s", err.Field())
	}

	if err.Error() != "invalid email format" {
		t.Errorf("Expected message 'invalid email format', got %s", err.Error())
	}
}

func TestRequiredFieldError(t *testing.T) {
	err := NewRequiredFieldError("name")

	if err.Code() != RequiredFieldCode {
		t.Errorf("Expected code %s, got %s", RequiredFieldCode, err.Code())
	}

	if err.HTTPStatus() != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus())
	}

	expectedMessage := "name cannot be empty"
	if err.Error() != expectedMessage {
		t.Errorf("Expected message '%s', got %s", expectedMessage, err.Error())
	}
}

func TestNotFoundError(t *testing.T) {
	err := NewNotFoundError("user", "123")

	if err.Code() != NotFoundErrorCode {
		t.Errorf("Expected code %s, got %s", NotFoundErrorCode, err.Code())
	}

	if err.HTTPStatus() != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, err.HTTPStatus())
	}

	if err.Resource() != "user" {
		t.Errorf("Expected resource 'user', got %s", err.Resource())
	}

	if err.ID() != "123" {
		t.Errorf("Expected ID '123', got %s", err.ID())
	}

	expectedMessage := "user with id '123' not found"
	if err.Error() != expectedMessage {
		t.Errorf("Expected message '%s', got %s", expectedMessage, err.Error())
	}
}

func TestConflictError(t *testing.T) {
	details := ErrorDetails{
		Field: "email",
		Value: "test@example.com",
	}
	err := NewConflictError("email already exists", details)

	if err.Code() != ConflictErrorCode {
		t.Errorf("Expected code %s, got %s", ConflictErrorCode, err.Code())
	}

	if err.HTTPStatus() != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, err.HTTPStatus())
	}

	expectedMessage := "email already exists"
	if err.Error() != expectedMessage {
		t.Errorf("Expected message '%s', got %s", expectedMessage, err.Error())
	}

	if err.Details().Field != "email" {
		t.Errorf("Expected details field 'email', got %v", err.Details().Field)
	}
}

func TestInternalError(t *testing.T) {
	cause := NewValidationError("test", "test error")
	err := NewInternalError("database connection failed", cause)

	if err.Code() != InternalErrorCode {
		t.Errorf("Expected code %s, got %s", InternalErrorCode, err.Code())
	}

	if err.HTTPStatus() != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, err.HTTPStatus())
	}

	if err.Cause() != cause {
		t.Errorf("Expected cause to be set correctly")
	}

	expectedMessage := "database connection failed: test error"
	if err.Error() != expectedMessage {
		t.Errorf("Expected message '%s', got %s", expectedMessage, err.Error())
	}
}

func TestErrorTypeAssertions(t *testing.T) {
	validationErr := NewValidationError("email", "invalid")
	notFoundErr := NewNotFoundError("user", "123")
	conflictErr := NewConflictError("conflict", ErrorDetails{})
	internalErr := NewInternalError("internal", nil)

	// Test IsDomainError
	if !IsDomainError(validationErr) {
		t.Error("Expected validation error to be a domain error")
	}

	if !IsDomainError(notFoundErr) {
		t.Error("Expected not found error to be a domain error")
	}

	// Test AsValidationError
	if ve, ok := AsValidationError(validationErr); !ok || ve != validationErr {
		t.Error("Expected validation error assertion to succeed")
	}

	if _, ok := AsValidationError(notFoundErr); ok {
		t.Error("Expected validation error assertion to fail for not found error")
	}

	// Test AsNotFoundError
	if nfe, ok := AsNotFoundError(notFoundErr); !ok || nfe != notFoundErr {
		t.Error("Expected not found error assertion to succeed")
	}

	if _, ok := AsNotFoundError(validationErr); ok {
		t.Error("Expected not found error assertion to fail for validation error")
	}

	// Test AsConflictError
	if ce, ok := AsConflictError(conflictErr); !ok || ce != conflictErr {
		t.Error("Expected conflict error assertion to succeed")
	}

	if _, ok := AsConflictError(validationErr); ok {
		t.Error("Expected conflict error assertion to fail for validation error")
	}

	// Test AsInternalError
	if ie, ok := AsInternalError(internalErr); !ok || ie != internalErr {
		t.Error("Expected internal error assertion to succeed")
	}

	if _, ok := AsInternalError(validationErr); ok {
		t.Error("Expected internal error assertion to fail for validation error")
	}
}
