// Package errors provides typed error definitions for domain operations
package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents a typed error code
type ErrorCode string

const (
	// Validation error codes
	ValidationErrorCode ErrorCode = "VALIDATION_ERROR"
	RequiredFieldCode   ErrorCode = "REQUIRED_FIELD"
	InvalidFormatCode   ErrorCode = "INVALID_FORMAT"

	// Business logic error codes
	NotFoundErrorCode ErrorCode = "NOT_FOUND"
	ConflictErrorCode ErrorCode = "CONFLICT"

	// System error codes
	InternalErrorCode ErrorCode = "INTERNAL_ERROR"
)

// DomainError represents the base interface for all domain errors
type DomainError interface {
	error
	Code() ErrorCode
	HTTPStatus() int
	Details() map[string]interface{}
}

// ValidationError represents validation failures in domain entities
type ValidationError struct {
	code    ErrorCode
	message string
	field   string
	details map[string]interface{}
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		code:    ValidationErrorCode,
		message: message,
		field:   field,
		details: map[string]interface{}{
			"field": field,
		},
	}
}

// NewRequiredFieldError creates a validation error for required fields
func NewRequiredFieldError(field string) *ValidationError {
	return &ValidationError{
		code:    RequiredFieldCode,
		message: fmt.Sprintf("%s cannot be empty", field),
		field:   field,
		details: map[string]interface{}{
			"field": field,
		},
	}
}

func (e *ValidationError) Error() string {
	return e.message
}

func (e *ValidationError) Code() ErrorCode {
	return e.code
}

func (e *ValidationError) HTTPStatus() int {
	return http.StatusBadRequest
}

func (e *ValidationError) Details() map[string]interface{} {
	return e.details
}

func (e *ValidationError) Field() string {
	return e.field
}

// NotFoundError represents resources that cannot be found
type NotFoundError struct {
	code     ErrorCode
	message  string
	resource string
	id       string
	details  map[string]interface{}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		code:     NotFoundErrorCode,
		message:  fmt.Sprintf("%s with id '%s' not found", resource, id),
		resource: resource,
		id:       id,
		details: map[string]interface{}{
			"resource": resource,
			"id":       id,
		},
	}
}

func (e *NotFoundError) Error() string {
	return e.message
}

func (e *NotFoundError) Code() ErrorCode {
	return e.code
}

func (e *NotFoundError) HTTPStatus() int {
	return http.StatusNotFound
}

func (e *NotFoundError) Details() map[string]interface{} {
	return e.details
}

func (e *NotFoundError) Resource() string {
	return e.resource
}

func (e *NotFoundError) ID() string {
	return e.id
}

// ConflictError represents business rule conflicts
type ConflictError struct {
	code    ErrorCode
	message string
	details map[string]interface{}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string, details map[string]interface{}) *ConflictError {
	if details == nil {
		details = make(map[string]interface{})
	}

	return &ConflictError{
		code:    ConflictErrorCode,
		message: message,
		details: details,
	}
}

func (e *ConflictError) Error() string {
	return e.message
}

func (e *ConflictError) Code() ErrorCode {
	return e.code
}

func (e *ConflictError) HTTPStatus() int {
	return http.StatusConflict
}

func (e *ConflictError) Details() map[string]interface{} {
	return e.details
}

// InternalError represents system-level errors
type InternalError struct {
	code    ErrorCode
	message string
	cause   error
	details map[string]interface{}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *InternalError {
	return &InternalError{
		code:    InternalErrorCode,
		message: message,
		cause:   cause,
		details: make(map[string]interface{}),
	}
}

func (e *InternalError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

func (e *InternalError) Code() ErrorCode {
	return e.code
}

func (e *InternalError) HTTPStatus() int {
	return http.StatusInternalServerError
}

func (e *InternalError) Details() map[string]interface{} {
	return e.details
}

func (e *InternalError) Cause() error {
	return e.cause
}

// Unwrap implements error unwrapping for InternalError
func (e *InternalError) Unwrap() error {
	return e.cause
}

// IsDomainError checks if an error is a domain error
func IsDomainError(err error) bool {
	_, ok := err.(DomainError)
	return ok
}

// AsValidationError attempts to cast error to ValidationError
func AsValidationError(err error) (*ValidationError, bool) {
	ve, ok := err.(*ValidationError)
	return ve, ok
}

// AsNotFoundError attempts to cast error to NotFoundError
func AsNotFoundError(err error) (*NotFoundError, bool) {
	nfe, ok := err.(*NotFoundError)
	return nfe, ok
}

// AsConflictError attempts to cast error to ConflictError
func AsConflictError(err error) (*ConflictError, bool) {
	ce, ok := err.(*ConflictError)
	return ce, ok
}

// AsInternalError attempts to cast error to InternalError
func AsInternalError(err error) (*InternalError, bool) {
	ie, ok := err.(*InternalError)
	return ie, ok
}
