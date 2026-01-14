// Service error utilities for reducing duplication in domain services
package errors

import "fmt"

// WrapServiceError wraps an error with service operation context.
// This uses centralized error types for consistency.
func WrapServiceError(operation string, err error) error {
	if err == nil {
		return nil
	}

	return NewInternalError("failed to "+operation, err)
}

// WrapRepoError wraps repository errors with consistent messaging.
func WrapRepoError(operation, entity string, err error) error {
	if err == nil {
		return nil
	}

	return NewInternalError(fmt.Sprintf("failed to %s %s", operation, entity), err)
}

// WrapValidationError wraps validation errors with consistent messaging.
func WrapValidationError(entity string, err error) error {
	if err == nil {
		return nil
	}

	return NewValidationError(entity, fmt.Sprintf("validation failed: %v", err))
}

// WrapBusinessRuleError wraps business rule validation errors.
func WrapBusinessRuleError(rule string, err error) error {
	if err == nil {
		return nil
	}

	return NewValidationError(rule, fmt.Sprintf("business rule validation failed: %v", err))
}
