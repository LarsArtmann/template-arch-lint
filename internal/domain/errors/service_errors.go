// Service error utilities for reducing duplication in domain services
package errors

import "fmt"

// WrapServiceError wraps an error with a service operation context.
// This reduces duplication of fmt.Errorf("failed to %s: %w", operation, err) patterns.
func WrapServiceError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// WrapRepoError wraps repository errors with consistent messaging.
func WrapRepoError(operation, entity string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("failed to %s %s: %w", operation, entity, err)
}

// WrapValidationError wraps validation errors with consistent messaging.
func WrapValidationError(entity string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s validation failed: %w", entity, err)
}

// WrapBusinessRuleError wraps business rule validation errors.
func WrapBusinessRuleError(rule string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s validation failed: %w", rule, err)
}
