// Package errors provides enhanced error handling utilities with context and debugging support.
package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	domainErrors "github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/samber/lo"
)

// ContextKey represents a type for context keys.
type ContextKey string

const (
	// CorrelationIDKey is used to store correlation IDs in context.
	CorrelationIDKey ContextKey = "correlation_id"
	// UserIDKey is used to store user IDs in context.
	UserIDKey ContextKey = "user_id"
	// RequestIDKey is used to store request IDs in context.
	RequestIDKey ContextKey = "request_id"
	// OperationKey is used to store operation names in context.
	OperationKey ContextKey = "operation"
)

// StackFrame represents a single frame in a stack trace.
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// ErrorContext contains contextual information about an error.
type ErrorContext struct {
	CorrelationID string                 `json:"correlation_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	RequestID     string                 `json:"request_id,omitempty"`
	Operation     string                 `json:"operation,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
	StackTrace    []StackFrame           `json:"stack_trace,omitempty"`
}

// WrappedError represents an error with additional context and debugging information.
type WrappedError struct {
	originalError error
	message       string
	context       ErrorContext
	cause         error
}

// NewWrappedError creates a new wrapped error with context.
func NewWrappedError(err error, message string) *WrappedError {
	if err == nil {
		return nil
	}

	return &WrappedError{
		originalError: err,
		message:       message,
		context: ErrorContext{
			Timestamp: time.Now().UTC(),
		},
		cause: err,
	}
}

// WithContext adds context information to the error.
func (we *WrappedError) WithContext(ctx context.Context) *WrappedError {
	if we == nil {
		return nil
	}

	// Extract information from context
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		we.context.CorrelationID = correlationID
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		we.context.UserID = userID
	}
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		we.context.RequestID = requestID
	}
	if operation, ok := ctx.Value(OperationKey).(string); ok {
		we.context.Operation = operation
	}

	return we
}

// WithOperation adds operation information to the error.
func (we *WrappedError) WithOperation(operation string) *WrappedError {
	if we == nil {
		return nil
	}
	we.context.Operation = operation
	return we
}

// WithExtra adds extra debugging information to the error.
func (we *WrappedError) WithExtra(key string, value interface{}) *WrappedError {
	if we == nil {
		return nil
	}
	if we.context.Extra == nil {
		we.context.Extra = make(map[string]interface{})
	}
	we.context.Extra[key] = value
	return we
}

// WithStackTrace captures the current stack trace.
func (we *WrappedError) WithStackTrace(skip int) *WrappedError {
	if we == nil {
		return nil
	}

	var frames []StackFrame
	for i := skip; i < skip+10; i++ { // Capture up to 10 frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		// Skip internal runtime functions
		funcName := fn.Name()
		if strings.Contains(funcName, "runtime.") {
			continue
		}

		frames = append(frames, StackFrame{
			Function: funcName,
			File:     file,
			Line:     line,
		})
	}

	we.context.StackTrace = frames
	return we
}

// Error implements the error interface.
func (we *WrappedError) Error() string {
	if we.message != "" {
		return fmt.Sprintf("%s: %v", we.message, we.originalError)
	}
	return we.originalError.Error()
}

// Unwrap implements error unwrapping.
func (we *WrappedError) Unwrap() error {
	return we.cause
}

// Context returns the error context.
func (we *WrappedError) Context() ErrorContext {
	return we.context
}

// Original returns the original error.
func (we *WrappedError) Original() error {
	return we.originalError
}

// IsDomainError checks if the wrapped error is a domain error.
func (we *WrappedError) IsDomainError() bool {
	return domainErrors.IsDomainError(we.originalError)
}

// ToDomainError attempts to convert to a domain error.
func (we *WrappedError) ToDomainError() (domainErrors.DomainError, bool) {
	var domainErr domainErrors.DomainError
	ok := errors.As(we.originalError, &domainErr)
	return domainErr, ok
}

// ToJSON serializes the error to JSON for logging.
func (we *WrappedError) ToJSON() ([]byte, error) {
	data := struct {
		Error   string       `json:"error"`
		Message string       `json:"message,omitempty"`
		Context ErrorContext `json:"context"`
	}{
		Error:   we.originalError.Error(),
		Message: we.message,
		Context: we.context,
	}

	return json.Marshal(data)
}

// ErrorFactory provides convenient methods for creating common error types.
type ErrorFactory struct{}

// NewErrorFactory creates a new ErrorFactory instance.
func NewErrorFactory() *ErrorFactory {
	return &ErrorFactory{}
}

// Wrap wraps an error with a message and captures stack trace.
func (ef *ErrorFactory) Wrap(err error, message string, args ...interface{}) *WrappedError {
	if err == nil {
		return nil
	}

	formattedMessage := message
	if len(args) > 0 {
		formattedMessage = fmt.Sprintf(message, args...)
	}

	return NewWrappedError(err, formattedMessage).WithStackTrace(2)
}

// WrapWithContext wraps an error with a message and context.
func (ef *ErrorFactory) WrapWithContext(ctx context.Context, err error, message string, args ...interface{}) *WrappedError {
	if err == nil {
		return nil
	}

	wrappedErr := ef.Wrap(err, message, args...)
	return wrappedErr.WithContext(ctx)
}

// Database creates database-related errors.
func (ef *ErrorFactory) Database(err error, operation string) *WrappedError {
	return ef.Wrap(err, "database operation failed").
		WithOperation(operation).
		WithStackTrace(2)
}

// Validation creates validation-related errors.
func (ef *ErrorFactory) Validation(err error, field string) *WrappedError {
	return ef.Wrap(err, "validation failed").
		WithExtra("field", field).
		WithStackTrace(2)
}

// External creates external service-related errors.
func (ef *ErrorFactory) External(err error, service string) *WrappedError {
	return ef.Wrap(err, "external service error").
		WithExtra("service", service).
		WithStackTrace(2)
}

// Authentication creates authentication-related errors.
func (ef *ErrorFactory) Authentication(err error, reason string) *WrappedError {
	return ef.Wrap(err, "authentication failed").
		WithExtra("reason", reason).
		WithStackTrace(2)
}

// Authorization creates authorization-related errors.
func (ef *ErrorFactory) Authorization(err error, resource string, action string) *WrappedError {
	return ef.Wrap(err, "authorization failed").
		WithExtra("resource", resource).
		WithExtra("action", action).
		WithStackTrace(2)
}

// Configuration creates configuration-related errors.
func (ef *ErrorFactory) Configuration(err error, config string) *WrappedError {
	return ef.Wrap(err, "configuration error").
		WithExtra("config", config).
		WithStackTrace(2)
}

// ErrorChain represents a chain of errors that can occur during processing.
type ErrorChain struct {
	errors []error
}

// NewErrorChain creates a new error chain.
func NewErrorChain() *ErrorChain {
	return &ErrorChain{
		errors: make([]error, 0),
	}
}

// Add adds an error to the chain.
func (ec *ErrorChain) Add(err error) *ErrorChain {
	if err != nil {
		ec.errors = append(ec.errors, err)
	}
	return ec
}

// HasErrors returns true if the chain contains any errors.
func (ec *ErrorChain) HasErrors() bool {
	return len(ec.errors) > 0
}

// Count returns the number of errors in the chain.
func (ec *ErrorChain) Count() int {
	return len(ec.errors)
}

// Errors returns all errors in the chain.
func (ec *ErrorChain) Errors() []error {
	return lo.Map(ec.errors, func(err error, _ int) error {
		return err
	})
}

// FirstError returns the first error in the chain, or nil if none.
func (ec *ErrorChain) FirstError() error {
	if len(ec.errors) == 0 {
		return nil
	}
	return ec.errors[0]
}

// ToError converts the error chain to a single error.
func (ec *ErrorChain) ToError() error {
	if len(ec.errors) == 0 {
		return nil
	}
	if len(ec.errors) == 1 {
		return ec.errors[0]
	}

	messages := lo.Map(ec.errors, func(err error, _ int) string {
		return err.Error()
	})

	return fmt.Errorf("multiple errors occurred: %s", strings.Join(messages, "; "))
}

// ErrorGroup represents a group of errors with context about their source.
type ErrorGroup struct {
	errors map[string][]error
}

// NewErrorGroup creates a new error group.
func NewErrorGroup() *ErrorGroup {
	return &ErrorGroup{
		errors: make(map[string][]error),
	}
}

// Add adds an error to a specific category.
func (eg *ErrorGroup) Add(category string, err error) *ErrorGroup {
	if err != nil {
		if eg.errors[category] == nil {
			eg.errors[category] = make([]error, 0)
		}
		eg.errors[category] = append(eg.errors[category], err)
	}
	return eg
}

// HasErrors returns true if the group contains any errors.
func (eg *ErrorGroup) HasErrors() bool {
	return len(eg.errors) > 0
}

// HasCategory returns true if the group has errors in the specified category.
func (eg *ErrorGroup) HasCategory(category string) bool {
	return len(eg.errors[category]) > 0
}

// Categories returns all categories that have errors.
func (eg *ErrorGroup) Categories() []string {
	return lo.Keys(eg.errors)
}

// ErrorsInCategory returns all errors in a specific category.
func (eg *ErrorGroup) ErrorsInCategory(category string) []error {
	return eg.errors[category]
}

// AllErrors returns all errors from all categories.
func (eg *ErrorGroup) AllErrors() []error {
	var allErrors []error
	for _, categoryErrors := range eg.errors {
		allErrors = append(allErrors, categoryErrors...)
	}
	return allErrors
}

// ToError converts the error group to a single error.
func (eg *ErrorGroup) ToError() error {
	if !eg.HasErrors() {
		return nil
	}

	var parts []string
	for category, categoryErrors := range eg.errors {
		messages := lo.Map(categoryErrors, func(err error, _ int) string {
			return err.Error()
		})
		parts = append(parts, fmt.Sprintf("%s: %s", category, strings.Join(messages, ", ")))
	}

	return fmt.Errorf("errors occurred in multiple categories: %s", strings.Join(parts, "; "))
}

// ErrorRecovery provides utilities for error recovery and fallback handling.
type ErrorRecovery struct{}

// NewErrorRecovery creates a new ErrorRecovery instance.
func NewErrorRecovery() *ErrorRecovery {
	return &ErrorRecovery{}
}

// Try executes a function and captures any panic as an error.
func (er *ErrorRecovery) Try(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = NewWrappedError(v, "panic recovered").WithStackTrace(2)
			case string:
				err = NewWrappedError(fmt.Errorf("%s", v), "panic recovered").WithStackTrace(2)
			default:
				err = NewWrappedError(fmt.Errorf("panic: %v", v), "panic recovered").WithStackTrace(2)
			}
		}
	}()

	return fn()
}

// WithRetry executes a function with retry logic on specific errors.
func (er *ErrorRecovery) WithRetry(fn func() error, maxRetries int, shouldRetry func(error) bool) error {
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if i < maxRetries && shouldRetry(err) {
				time.Sleep(time.Duration(i+1) * 100 * time.Millisecond) // Simple backoff
				continue
			}
			return err
		}
		return nil
	}

	return lastErr
}

// ErrorAggregator collects and aggregates errors over time.
type ErrorAggregator struct {
	errors    []error
	threshold int
}

// NewErrorAggregator creates a new ErrorAggregator.
func NewErrorAggregator(threshold int) *ErrorAggregator {
	return &ErrorAggregator{
		errors:    make([]error, 0),
		threshold: threshold,
	}
}

// Add adds an error to the aggregator.
func (ea *ErrorAggregator) Add(err error) bool {
	if err != nil {
		ea.errors = append(ea.errors, err)
	}
	return len(ea.errors) >= ea.threshold
}

// Errors returns all collected errors.
func (ea *ErrorAggregator) Errors() []error {
	return lo.Map(ea.errors, func(err error, _ int) error {
		return err
	})
}

// Count returns the number of collected errors.
func (ea *ErrorAggregator) Count() int {
	return len(ea.errors)
}

// Reset clears all collected errors.
func (ea *ErrorAggregator) Reset() {
	ea.errors = ea.errors[:0]
}

// ShouldTrigger returns true if the error count meets or exceeds the threshold.
func (ea *ErrorAggregator) ShouldTrigger() bool {
	return len(ea.errors) >= ea.threshold
}

// IgnoreError is a utility function to ignore specific errors.
func IgnoreError(err error, errorsToIgnore ...error) error {
	if err == nil {
		return nil
	}

	for _, ignoreErr := range errorsToIgnore {
		if errors.Is(err, ignoreErr) {
			return nil
		}
	}

	return err
}

// FormatErrorChain formats a chain of errors for display.
func FormatErrorChain(err error) string {
	if err == nil {
		return ""
	}

	var parts []string
	for err != nil {
		parts = append(parts, err.Error())
		err = errors.Unwrap(err)
	}

	return strings.Join(parts, " -> ")
}
