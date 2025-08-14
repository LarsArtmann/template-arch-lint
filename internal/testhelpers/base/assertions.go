// Package base provides fundamental test assertion patterns.
// These assertions eliminate repetitive validation code across test suites.
package base

import (
	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
)

// AssertionHelper provides common assertion patterns used across all tests.
// This interface allows different implementations for different testing scenarios.
type AssertionHelper interface {
	// AssertSuccess verifies that an operation completed successfully
	AssertSuccess(result any, err error)

	// AssertError verifies that an operation returned an error
	AssertError(result any, err error)

	// AssertValidationError verifies that a validation error occurred
	AssertValidationError(result any, err error)

	// AssertNotFound verifies that a not-found error occurred
	AssertNotFound(result any, err error)
}

// StandardAssertions implements common assertion patterns using Gomega.
type StandardAssertions struct{}

// NewStandardAssertions creates a new StandardAssertions helper.
func NewStandardAssertions() *StandardAssertions {
	return &StandardAssertions{}
}

// AssertSuccess verifies that an operation completed successfully.
// Result should not be nil and error should not have occurred.
func (a *StandardAssertions) AssertSuccess(result any, err error) {
	Expect(err).ToNot(HaveOccurred())
	Expect(result).ToNot(BeNil())
}

// AssertError verifies that an operation returned an error.
// Result should be nil and error should have occurred.
func (a *StandardAssertions) AssertError(result any, err error) {
	Expect(err).To(HaveOccurred())
	Expect(result).To(BeNil())
}

// AssertValidationError verifies that a validation error occurred.
// This checks for domain-specific validation errors.
func (a *StandardAssertions) AssertValidationError(result any, err error) {
	Expect(result).To(BeNil())
	Expect(err).To(HaveOccurred())
	_, isValidationError := errors.AsValidationError(err)
	Expect(isValidationError).To(BeTrue())
}

// AssertNotFound verifies that a not-found error occurred.
func (a *StandardAssertions) AssertNotFound(result any, err error) {
	Expect(result).To(BeNil())
	Expect(err).To(HaveOccurred())
}

// ValidationAssertions provides specialized validation error checking.
type ValidationAssertions struct {
	*StandardAssertions
}

// NewValidationAssertions creates validation-focused assertion helper.
func NewValidationAssertions() *ValidationAssertions {
	return &ValidationAssertions{
		StandardAssertions: NewStandardAssertions(),
	}
}

// AssertValidationErrorContains verifies validation error with specific message.
func (a *ValidationAssertions) AssertValidationErrorContains(result any, err error, expectedMessage string) {
	a.AssertValidationError(result, err)
	Expect(err.Error()).To(ContainSubstring(expectedMessage))
}

// AssertValidationErrorForField verifies validation error for a specific field.
func (a *ValidationAssertions) AssertValidationErrorForField(result any, err error, fieldName string) {
	a.AssertValidationErrorContains(result, err, fieldName)
}

// SuccessAssertions provides specialized success verification.
type SuccessAssertions struct {
	*StandardAssertions
}

// NewSuccessAssertions creates success-focused assertion helper.
func NewSuccessAssertions() *SuccessAssertions {
	return &SuccessAssertions{
		StandardAssertions: NewStandardAssertions(),
	}
}

// AssertSuccessWithValue verifies success and checks result value.
func (a *SuccessAssertions) AssertSuccessWithValue(result any, err error, expectedValue any) {
	a.AssertSuccess(result, err)
	Expect(result).To(Equal(expectedValue))
}

// AssertSuccessWithPredicate verifies success with custom predicate.
func (a *SuccessAssertions) AssertSuccessWithPredicate(result any, err error, predicate func(any) bool) {
	a.AssertSuccess(result, err)
	Expect(predicate(result)).To(BeTrue())
}

// Convenience functions for global use

var (
	// Global assertion helpers for convenient access.
	assert         = NewStandardAssertions()
	validateAssert = NewValidationAssertions()
	successAssert  = NewSuccessAssertions()
)

// AssertSuccess is a convenience function for successful operation verification.
func AssertSuccess(result any, err error) {
	assert.AssertSuccess(result, err)
}

// AssertError is a convenience function for error verification.
func AssertError(result any, err error) {
	assert.AssertError(result, err)
}

// AssertValidationError is a convenience function for validation error verification.
func AssertValidationError(result any, err error) {
	validateAssert.AssertValidationError(result, err)
}

// AssertValidationErrorForField is a convenience function for field validation error verification.
func AssertValidationErrorForField(result any, err error, fieldName string) {
	validateAssert.AssertValidationErrorForField(result, err, fieldName)
}

// AssertNotFound is a convenience function for not-found error verification.
func AssertNotFound(result any, err error) {
	assert.AssertNotFound(result, err)
}

// AssertSuccessWithValue is a convenience function for success verification with value checking.
func AssertSuccessWithValue(result any, err error, expectedValue any) {
	successAssert.AssertSuccessWithValue(result, err, expectedValue)
}
