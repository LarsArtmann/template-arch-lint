// Package errors provides centralized error definitions for the entire project
// ALL error definitions MUST be in this package - no exceptions
package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// baseError provides common error functionality to reduce duplication.
type baseError struct {
	code    ErrorCode
	message string
	details ErrorDetails
}

// Error implements the error interface.
func (e *baseError) Error() string {
	return e.message
}

// Code returns the error code.
func (e *baseError) Code() ErrorCode {
	return e.code
}

// Details returns the error details.
func (e *baseError) Details() ErrorDetails {
	return e.details
}

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
	// InvalidFormatCode represents invalid format validation errors.
	InvalidFormatCode ErrorCode = "INVALID_FORMAT"

	// NotFoundErrorCode represents resource not found errors.
	NotFoundErrorCode ErrorCode = "NOT_FOUND"
	// ConflictErrorCode represents business rule conflict errors.
	ConflictErrorCode ErrorCode = "CONFLICT"

	// InternalErrorCode represents internal system errors.
	InternalErrorCode ErrorCode = "INTERNAL_ERROR"

	// DatabaseErrorCode represents database errors.
	DatabaseErrorCode ErrorCode = "DATABASE_ERROR"
	// NetworkErrorCode represents network errors.
	NetworkErrorCode ErrorCode = "NETWORK_ERROR"
	// ConfigurationErrorCode represents configuration errors.
	ConfigurationErrorCode ErrorCode = "CONFIGURATION_ERROR"
	// AuthorizationErrorCode represents authorization errors.
	AuthorizationErrorCode ErrorCode = "AUTHORIZATION_ERROR"
)

// DomainError represents the base interface for all domain errors.
type DomainError interface {
	error
	Code() ErrorCode
	HTTPStatus() int
	Details() ErrorDetails
}

// InfrastructureError represents infrastructure-layer technical errors.
type InfrastructureError interface {
	error
	Code() ErrorCode
	HTTPStatus() int
	Details() ErrorDetails
	IsRetryable() bool
}

// ValidationError represents validation failures in domain entities.
type ValidationError struct {
	baseError

	field string
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		baseError: baseError{
			code:    ValidationErrorCode,
			message: message,
			details: ErrorDetails{
				Field: field,
			},
		},
		field: field,
	}
}

// NewRequiredFieldError creates a validation error for required fields.
func NewRequiredFieldError(field string) *ValidationError {
	return &ValidationError{
		baseError: baseError{
			code:    RequiredFieldCode,
			message: field + " cannot be empty",
			details: ErrorDetails{
				Field: field,
			},
		},
		field: field,
	}
}

// HTTPStatus returns the HTTP status code for the validation error.
func (e *ValidationError) HTTPStatus() int {
	return http.StatusBadRequest
}

// Field returns the field name that caused the validation error.
func (e *ValidationError) Field() string {
	return e.field
}

// NotFoundError represents resources that cannot be found.
type NotFoundError struct {
	baseError

	resource string
	id       string
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		baseError: baseError{
			code:    NotFoundErrorCode,
			message: fmt.Sprintf("%s with id '%s' not found", resource, id),
			details: ErrorDetails{
				Resource: resource,
				ID:       id,
			},
		},
		resource: resource,
		id:       id,
	}
}

// HTTPStatus returns the HTTP status code for the not found error.
func (e *NotFoundError) HTTPStatus() int {
	return http.StatusNotFound
}

// Resource returns the resource name that was not found.
func (e *NotFoundError) Resource() string {
	return e.resource
}

// ID returns the ID of the resource that was not found.
func (e *NotFoundError) ID() string {
	return e.id
}

// ConflictError represents business rule conflicts.
type ConflictError struct {
	baseError
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string, details ErrorDetails) *ConflictError {
	return &ConflictError{
		baseError: baseError{
			code:    ConflictErrorCode,
			message: message,
			details: details,
		},
	}
}

// HTTPStatus returns the HTTP status code for the conflict error.
func (e *ConflictError) HTTPStatus() int {
	return http.StatusConflict
}

// InternalError represents system-level errors.
type InternalError struct {
	baseError

	cause error
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, cause error) *InternalError {
	return &InternalError{
		baseError: baseError{
			code:    InternalErrorCode,
			message: message,
			details: ErrorDetails{},
		},
		cause: cause,
	}
}

func (e *InternalError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}

	return e.message
}

// HTTPStatus returns the HTTP status code for internal error.
func (e *InternalError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// IsRetryable returns whether the internal error can be retried.
func (e *InternalError) IsRetryable() bool {
	return false
}

// Cause returns the underlying cause of the internal error.
func (e *InternalError) Cause() error {
	return e.cause
}

// Unwrap implements error unwrapping for InternalError.
func (e *InternalError) Unwrap() error {
	return e.cause
}

// DatabaseError represents database operation errors.
type DatabaseError struct {
	baseError

	operation string
	retryable bool
}

// NewDatabaseError creates a new database error.
func NewDatabaseError(operation string, cause error, retryable bool) *DatabaseError {
	return &DatabaseError{
		baseError: baseError{
			code:    DatabaseErrorCode,
			message: fmt.Sprintf("database %s failed: %v", operation, cause),
			details: ErrorDetails{
				Extra: map[string]string{
					"operation": operation,
					"retryable": strconv.FormatBool(retryable),
				},
			},
		},
		operation: operation,
		retryable: retryable,
	}
}

// HTTPStatus returns the HTTP status code for database error.
func (e *DatabaseError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// Operation returns the database operation that failed.
func (e *DatabaseError) Operation() string {
	return e.operation
}

// IsRetryable returns whether the database operation can be retried.
func (e *DatabaseError) IsRetryable() bool {
	return e.retryable
}

// NetworkError represents network operation errors.
type NetworkError struct {
	baseError

	service   string
	retryable bool
}

// NewNetworkError creates a new network error.
func NewNetworkError(service string, cause error, retryable bool) *NetworkError {
	return &NetworkError{
		baseError: baseError{
			code:    NetworkErrorCode,
			message: fmt.Sprintf("network service %s failed: %v", service, cause),
			details: ErrorDetails{
				Extra: map[string]string{
					"service":   service,
					"retryable": strconv.FormatBool(retryable),
				},
			},
		},
		service:   service,
		retryable: retryable,
	}
}

// HTTPStatus returns the HTTP status code for network error.
func (e *NetworkError) HTTPStatus() int {
	return http.StatusServiceUnavailable
}

// Service returns the network service that failed.
func (e *NetworkError) Service() string {
	return e.service
}

// IsRetryable returns whether the network operation can be retried.
func (e *NetworkError) IsRetryable() bool {
	return e.retryable
}

// ConfigurationError represents configuration errors.
type ConfigurationError struct {
	baseError

	configKey string
}

// NewConfigurationError creates a new configuration error.
func NewConfigurationError(key, message string) *ConfigurationError {
	return &ConfigurationError{
		baseError: baseError{
			code:    ConfigurationErrorCode,
			message: fmt.Sprintf("configuration error for %s: %s", key, message),
			details: ErrorDetails{
				Extra: map[string]string{
					"config_key": key,
				},
			},
		},
		configKey: key,
	}
}

// HTTPStatus returns the HTTP status code for configuration error.
func (e *ConfigurationError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// ConfigKey returns the configuration key that caused the error.
func (e *ConfigurationError) ConfigKey() string {
	return e.configKey
}

// IsRetryable returns whether the configuration error can be retried.
func (e *ConfigurationError) IsRetryable() bool {
	return false
}

// IsDomainError checks if an error is a domain error.
func IsDomainError(err error) bool {
	var domainErr DomainError

	return errors.As(err, &domainErr)
}

// AsValidationError attempts to cast error to ValidationError.
func AsValidationError(err error) (*ValidationError, bool) {
	var ve *ValidationError
	ok := errors.As(err, &ve)

	return ve, ok
}

// AsNotFoundError attempts to cast error to NotFoundError.
func AsNotFoundError(err error) (*NotFoundError, bool) {
	var nfe *NotFoundError
	ok := errors.As(err, &nfe)

	return nfe, ok
}

// AsConflictError attempts to cast error to ConflictError.
func AsConflictError(err error) (*ConflictError, bool) {
	var ce *ConflictError
	ok := errors.As(err, &ce)

	return ce, ok
}

// AsInternalError attempts to cast error to InternalError.
func AsInternalError(err error) (*InternalError, bool) {
	var ie *InternalError
	ok := errors.As(err, &ie)

	return ie, ok
}

// AsDatabaseError attempts to cast error to DatabaseError.
func AsDatabaseError(err error) (*DatabaseError, bool) {
	var de *DatabaseError
	ok := errors.As(err, &de)

	return de, ok
}

// AsNetworkError attempts to cast error to NetworkError.
func AsNetworkError(err error) (*NetworkError, bool) {
	var ne *NetworkError
	ok := errors.As(err, &ne)

	return ne, ok
}

// AsConfigurationError attempts to cast error to ConfigurationError.
func AsConfigurationError(err error) (*ConfigurationError, bool) {
	var ce *ConfigurationError
	ok := errors.As(err, &ce)

	return ce, ok
}

// IsInfrastructureError checks if error is an infrastructure error.
func IsInfrastructureError(err error) bool {
	var infraErr InfrastructureError

	return errors.As(err, &infraErr)
}

// IsRetryableError checks if error can be retried.
func IsRetryableError(err error) bool {
	var de *DatabaseError
	if errors.As(err, &de) {
		return de.IsRetryable()
	}

	var ne *NetworkError
	if errors.As(err, &ne) {
		return ne.IsRetryable()
	}

	return false
}

// NewDomainValidationError creates a new domain validation error (alias for NewValidationError).
func NewDomainValidationError(field, reason string) *ValidationError {
	return NewValidationError(field, fmt.Sprintf("validation failed for %s: %s", field, reason))
}

// NewDomainNotFoundError creates a new domain not found error (alias for NewNotFoundError).
func NewDomainNotFoundError(resource, id string) *NotFoundError {
	return NewNotFoundError(resource, id)
}

// NewDomainConflictError creates a new domain conflict error (alias for NewConflictError).
func NewDomainConflictError(resource, reason string) *ConflictError {
	return NewConflictError(resource+" conflict", ErrorDetails{
		Extra: map[string]string{
			"resource": resource,
			"reason":   reason,
		},
	})
}

// NewInfrastructureError creates a new infrastructure error based on the type.
func NewInfrastructureError(resource, operation string, cause error) InfrastructureError {
	resourceLower := strings.ToLower(resource)
	operationLower := strings.ToLower(operation)

	// Try to determine specific error type based on the context
	if isDatabaseOperation(resourceLower, operationLower) {
		return NewDatabaseError(operation, cause, false)
	}

	if isNetworkOperation(resourceLower, operationLower) {
		return NewNetworkError(resource, cause, true)
	}

	if strings.Contains(resourceLower, "config") || strings.Contains(operationLower, "configuration") {
		return NewConfigurationError(resource, cause.Error())
	}

	// Default to internal error
	return NewInternalError(fmt.Sprintf("%s %s failed", resource, operation), cause)
}

func isDatabaseOperation(resource, operation string) bool {
	return strings.Contains(resource, "database") ||
		strings.Contains(operation, "query") ||
		strings.Contains(operation, "insert") ||
		strings.Contains(operation, "update") ||
		strings.Contains(operation, "delete")
}

func isNetworkOperation(resource, operation string) bool {
	return strings.Contains(resource, "network") ||
		strings.Contains(operation, "http") ||
		strings.Contains(operation, "request")
}
