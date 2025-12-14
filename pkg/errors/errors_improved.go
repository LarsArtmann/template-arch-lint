// Package errors provides centralized error definitions for entire project
// ALL error definitions MUST be in this package - no exceptions
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Semantic error interfaces for layered error handling
// These provide type safety while maintaining centralization

// DomainError represents domain-layer business logic errors
type DomainError interface {
	error
	IsDomainError()
}

// InfrastructureError represents infrastructure-layer technical errors
type InfrastructureError interface {
	error
	IsInfrastructureError()
}

// ValidationError represents input validation errors
type ValidationError interface {
	error
	IsValidationError()
}

// baseError provides common error functionality to reduce duplication.
type baseError struct {
	code    ErrorCode
	message string
	details ErrorDetails
}

// Error implements error interface.
func (e *baseError) Error() string {
	return e.message
}

// Code returns error code.
func (e *baseError) Code() ErrorCode {
	return e.code
}

// Details returns error details.
func (e *baseError) Details() ErrorDetails {
	return e.details
}

// Domain-specific error types

type domainError struct {
	*baseError
}

func (e *domainError) IsDomainError() {}

type infrastructureError struct {
	*baseError
}

func (e *infrastructureError) IsInfrastructureError() {}

type validationError struct {
	*baseError
}

func (e *validationError) IsValidationError() {}

// ErrorCode represents a typed error code.
type ErrorCode string

// ErrorDetails represents strongly typed error details.
type ErrorDetails struct {
	Field    string            `json:"field,omitempty"`
	Resource string            `json:"resource,omitempty"`
	ID       string            `json:"id,omitempty"`
	Value    string            `json:"value,omitempty"`
	Reason   string            `json:"reason,omitempty"`
	Extra    map[string]string `json:"extra,omitempty"`
}

const (
	// ValidationErrorCode represents validation errors.
	ValidationErrorCode ErrorCode = "VALIDATION_ERROR"
	// RequiredFieldCode represents required field validation errors.
	RequiredFieldCode ErrorCode = "REQUIRED_FIELD"
	// NotFoundCode represents resource not found errors.
	NotFoundCode ErrorCode = "NOT_FOUND"
	// ConflictCode represents conflict errors.
	ConflictCode ErrorCode = "CONFLICT"
	// DatabaseErrorCode represents database errors.
	DatabaseErrorCode ErrorCode = "DATABASE_ERROR"
	// NetworkErrorCode represents network errors.
	NetworkErrorCode ErrorCode = "NETWORK_ERROR"
	// ConfigurationErrorCode represents configuration errors.
	ConfigurationErrorCode ErrorCode = "CONFIGURATION_ERROR"
	// AuthorizationErrorCode represents authorization errors.
	AuthorizationErrorCode ErrorCode = "AUTHORIZATION_ERROR"
	// InternalErrorCode represents unexpected internal errors.
	InternalErrorCode ErrorCode = "INTERNAL_ERROR"
)

// HTTP status code mapping
var errorHTTPStatus = map[ErrorCode]int{
	ValidationErrorCode:    http.StatusBadRequest,
	RequiredFieldCode:      http.StatusBadRequest,
	NotFoundCode:           http.StatusNotFound,
	ConflictCode:           http.StatusConflict,
	DatabaseErrorCode:      http.StatusInternalServerError,
	NetworkErrorCode:       http.StatusServiceUnavailable,
	ConfigurationErrorCode: http.StatusInternalServerError,
	AuthorizationErrorCode: http.StatusUnauthorized,
	InternalErrorCode:      http.StatusInternalServerError,
}

// HTTPStatus returns appropriate HTTP status code for error.
func HTTPStatus(err error) int {
	if baseErr, ok := err.(*baseError); ok {
		if status, exists := errorHTTPStatus[baseErr.code]; exists {
			return status
		}
	}
	return http.StatusInternalServerError
}

// Error creation functions with semantic typing

// NewDomainValidationError creates a new domain validation error.
func NewDomainValidationError(field, reason string) ValidationError {
	return &validationError{
		&baseError{
			code:    ValidationErrorCode,
			message: fmt.Sprintf("validation failed for %s: %s", field, reason),
			details: ErrorDetails{
				Field:  field,
				Reason: reason,
			},
		},
	}
}

// NewInfrastructureError creates a new infrastructure error.
func NewInfrastructureError(resource, operation string, cause error) InfrastructureError {
	message := fmt.Sprintf("%s %s failed", resource, operation)
	if cause != nil {
		message = fmt.Sprintf("%s: %s", message, cause.Error())
	}

	return &infrastructureError{
		&baseError{
			code:    DatabaseErrorCode,
			message: message,
			details: ErrorDetails{
				Resource: resource,
				Extra:    map[string]string{"operation": operation},
			},
		},
	}
}

// NewDomainNotFoundError creates a new domain not found error.
func NewDomainNotFoundError(resource, id string) DomainError {
	return &domainError{
		&baseError{
			code:    NotFoundCode,
			message: fmt.Sprintf("%s with id %s not found", resource, id),
			details: ErrorDetails{
				Resource: resource,
				ID:       id,
			},
		},
	}
}

// NewDomainConflictError creates a new domain conflict error.
func NewDomainConflictError(resource, reason string) DomainError {
	return &domainError{
		&baseError{
			code:    ConflictCode,
			message: fmt.Sprintf("%s conflict: %s", resource, reason),
			details: ErrorDetails{
				Resource: resource,
				Reason:   reason,
			},
		},
	}
}

// Legacy error functions for backward compatibility
// These will be deprecated in favor of semantic versions

// NewValidationError creates a new validation error.
func NewValidationError(field, reason string) error {
	return NewDomainValidationError(field, reason)
}

// NewRequiredFieldError creates a new required field error.
func NewRequiredFieldError(field string) error {
	return NewDomainValidationError(field, "required field")
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resource, id string) error {
	return NewDomainNotFoundError(resource, id)
}

// NewConflictError creates a new conflict error.
func NewConflictError(resource, reason string) error {
	return NewDomainConflictError(resource, reason)
}

// IsDomainError checks if error is a domain error.
func IsDomainError(err error) bool {
	var domainErr DomainError
	return errors.As(err, &domainErr)
}

// IsInfrastructureError checks if error is an infrastructure error.
func IsInfrastructureError(err error) bool {
	var infraErr InfrastructureError
	return errors.As(err, &infraErr)
}

// IsValidationError checks if error is a validation error.
func IsValidationError(err error) bool {
	var validationErr ValidationError
	return errors.As(err, &validationErr)
}
