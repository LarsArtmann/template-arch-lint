// Package errors provides typed error definitions for domain operations
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

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
)

// DomainError represents the base interface for all domain errors.
type DomainError interface {
	error
	Code() ErrorCode
	HTTPStatus() int
	Details() ErrorDetails
}

// ValidationError represents validation failures in domain entities.
type ValidationError struct {
	code    ErrorCode
	message string
	field   string
	details ErrorDetails
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		code:    ValidationErrorCode,
		message: message,
		field:   field,
		details: ErrorDetails{
			Field: field,
		},
	}
}

// NewRequiredFieldError creates a validation error for required fields.
func NewRequiredFieldError(field string) *ValidationError {
	return &ValidationError{
		code:    RequiredFieldCode,
		message: fmt.Sprintf("%s cannot be empty", field),
		field:   field,
		details: ErrorDetails{
			Field: field,
		},
	}
}

func (e *ValidationError) Error() string {
	return e.message
}

// Code returns the error code for the validation error.
func (e *ValidationError) Code() ErrorCode {
	return e.code
}

// HTTPStatus returns the HTTP status code for the validation error.
func (e *ValidationError) HTTPStatus() int {
	return http.StatusBadRequest
}

// Details returns the error details for the validation error.
func (e *ValidationError) Details() ErrorDetails {
	return e.details
}

// Field returns the field name that caused the validation error.
func (e *ValidationError) Field() string {
	return e.field
}

// NotFoundError represents resources that cannot be found.
type NotFoundError struct {
	code     ErrorCode
	message  string
	resource string
	id       string
	details  ErrorDetails
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		code:     NotFoundErrorCode,
		message:  fmt.Sprintf("%s with id '%s' not found", resource, id),
		resource: resource,
		id:       id,
		details: ErrorDetails{
			Resource: resource,
			ID:       id,
		},
	}
}

func (e *NotFoundError) Error() string {
	return e.message
}

// Code returns the error code for the not found error.
func (e *NotFoundError) Code() ErrorCode {
	return e.code
}

// HTTPStatus returns the HTTP status code for the not found error.
func (e *NotFoundError) HTTPStatus() int {
	return http.StatusNotFound
}

// Details returns the error details for the not found error.
func (e *NotFoundError) Details() ErrorDetails {
	return e.details
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
	code    ErrorCode
	message string
	details ErrorDetails
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string, details ErrorDetails) *ConflictError {
	return &ConflictError{
		code:    ConflictErrorCode,
		message: message,
		details: details,
	}
}

func (e *ConflictError) Error() string {
	return e.message
}

// Code returns the error code for the conflict error.
func (e *ConflictError) Code() ErrorCode {
	return e.code
}

// HTTPStatus returns the HTTP status code for the conflict error.
func (e *ConflictError) HTTPStatus() int {
	return http.StatusConflict
}

// Details returns the error details for the conflict error.
func (e *ConflictError) Details() ErrorDetails {
	return e.details
}

// InternalError represents system-level errors.
type InternalError struct {
	code    ErrorCode
	message string
	cause   error
	details ErrorDetails
}

// NewInternalError creates a new internal error.
func NewInternalError(message string, cause error) *InternalError {
	return &InternalError{
		code:    InternalErrorCode,
		message: message,
		cause:   cause,
		details: ErrorDetails{},
	}
}

func (e *InternalError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Code returns the error code for the internal error.
func (e *InternalError) Code() ErrorCode {
	return e.code
}

// HTTPStatus returns the HTTP status code for the internal error.
func (e *InternalError) HTTPStatus() int {
	return http.StatusInternalServerError
}

// Details returns the error details for the internal error.
func (e *InternalError) Details() ErrorDetails {
	return e.details
}

// Cause returns the underlying cause of the internal error.
func (e *InternalError) Cause() error {
	return e.cause
}

// Unwrap implements error unwrapping for InternalError.
func (e *InternalError) Unwrap() error {
	return e.cause
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
