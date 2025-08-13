package errors

import (
	"errors"
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

	if !errors.Is(err.Cause(), cause) {
		t.Errorf("Expected cause to be set correctly")
	}

	expectedMessage := "database connection failed: test error"
	if err.Error() != expectedMessage {
		t.Errorf("Expected message '%s', got %s", expectedMessage, err.Error())
	}
}

func TestErrorTypeAssertions(t *testing.T) {
	testErrors := createTestErrors()
	
	testDomainErrorIdentification(t, testErrors)
	testValidationErrorAssertions(t, testErrors)
	testNotFoundErrorAssertions(t, testErrors)
	testConflictErrorAssertions(t, testErrors)
	testInternalErrorAssertions(t, testErrors)
}

// testErrorTypes holds all test error instances
type testErrorTypes struct {
	validation *ValidationError
	notFound   *NotFoundError
	conflict   *ConflictError
	internal   *InternalError
}

// createTestErrors creates test error instances
func createTestErrors() testErrorTypes {
	return testErrorTypes{
		validation: NewValidationError("email", "invalid"),
		notFound:   NewNotFoundError("user", "123"),
		conflict:   NewConflictError("conflict", ErrorDetails{}),
		internal:   NewInternalError("internal", nil),
	}
}

// testDomainErrorIdentification tests IsDomainError function
func testDomainErrorIdentification(t *testing.T, errors testErrorTypes) {
	t.Helper()
	
	if !IsDomainError(errors.validation) {
		t.Error("Expected validation error to be a domain error")
	}
	if !IsDomainError(errors.notFound) {
		t.Error("Expected not found error to be a domain error")
	}
}

// testValidationErrorAssertions tests AsValidationError function
func testValidationErrorAssertions(t *testing.T, errors testErrorTypes) {
	t.Helper()
	
	if ve, ok := AsValidationError(errors.validation); !ok || ve != errors.validation {
		t.Error("Expected validation error assertion to succeed")
	}
	if _, ok := AsValidationError(errors.notFound); ok {
		t.Error("Expected validation error assertion to fail for not found error")
	}
}

// testNotFoundErrorAssertions tests AsNotFoundError function
func testNotFoundErrorAssertions(t *testing.T, errors testErrorTypes) {
	t.Helper()
	
	if nfe, ok := AsNotFoundError(errors.notFound); !ok || nfe != errors.notFound {
		t.Error("Expected not found error assertion to succeed")
	}
	if _, ok := AsNotFoundError(errors.validation); ok {
		t.Error("Expected not found error assertion to fail for validation error")
	}
}

// testConflictErrorAssertions tests AsConflictError function
func testConflictErrorAssertions(t *testing.T, errors testErrorTypes) {
	t.Helper()
	
	if ce, ok := AsConflictError(errors.conflict); !ok || ce != errors.conflict {
		t.Error("Expected conflict error assertion to succeed")
	}
	if _, ok := AsConflictError(errors.validation); ok {
		t.Error("Expected conflict error assertion to fail for validation error")
	}
}

// testInternalErrorAssertions tests AsInternalError function
func testInternalErrorAssertions(t *testing.T, errors testErrorTypes) {
	t.Helper()
	
	if ie, ok := AsInternalError(errors.internal); !ok || ie != errors.internal {
		t.Error("Expected internal error assertion to succeed")
	}
	if _, ok := AsInternalError(errors.validation); ok {
		t.Error("Expected internal error assertion to fail for validation error")
	}
}
