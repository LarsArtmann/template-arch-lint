// Package validation provides reusable validation utilities for the application.
package validation

import (
	"context"
	"fmt"
	"html"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/samber/lo"
)

// Validator represents a validation function that can be composed.
type Validator[T any] func(ctx context.Context, value T) error

// ValidatorFunc is a helper to create validators with custom logic.
type ValidatorFunc[T any] func(value T) error

// ToValidator converts a ValidatorFunc to a Validator.
func ToValidator[T any](fn ValidatorFunc[T]) Validator[T] {
	return func(_ context.Context, value T) error {
		return fn(value)
	}
}

// Chain combines multiple validators into a single validator.
func Chain[T any](validators ...Validator[T]) Validator[T] {
	return func(ctx context.Context, value T) error {
		for _, validator := range validators {
			if err := validator(ctx, value); err != nil {
				return err
			}
		}
		return nil
	}
}

// StringValidators provides common string validation functions.
type StringValidators struct{}

// NewStringValidators creates a new StringValidators instance.
func NewStringValidators() *StringValidators {
	return &StringValidators{}
}

// NotEmpty validates that a string is not empty or only whitespace.
func (sv *StringValidators) NotEmpty(fieldName string) ValidatorFunc[string] {
	return func(value string) error {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s cannot be empty", fieldName)
		}
		return nil
	}
}

// MinLength validates that a string has a minimum length.
func (sv *StringValidators) MinLength(fieldName string, minLen int) ValidatorFunc[string] {
	return func(value string) error {
		if utf8.RuneCountInString(value) < minLen {
			return fmt.Errorf("%s must be at least %d characters long", fieldName, minLen)
		}
		return nil
	}
}

// MaxLength validates that a string doesn't exceed maximum length.
func (sv *StringValidators) MaxLength(fieldName string, maxLen int) ValidatorFunc[string] {
	return func(value string) error {
		if utf8.RuneCountInString(value) > maxLen {
			return fmt.Errorf("%s must not exceed %d characters", fieldName, maxLen)
		}
		return nil
	}
}

// LengthRange validates that a string is within a length range.
func (sv *StringValidators) LengthRange(fieldName string, minLen, maxLen int) ValidatorFunc[string] {
	return func(value string) error {
		length := utf8.RuneCountInString(value)
		if length < minLen || length > maxLen {
			return fmt.Errorf("%s must be between %d and %d characters long", fieldName, minLen, maxLen)
		}
		return nil
	}
}

// Regex validates that a string matches a regular expression.
func (sv *StringValidators) Regex(fieldName string, pattern *regexp.Regexp, message string) ValidatorFunc[string] {
	return func(value string) error {
		if !pattern.MatchString(value) {
			if message == "" {
				return fmt.Errorf("%s has invalid format", fieldName)
			}
			return fmt.Errorf("%s %s", fieldName, message)
		}
		return nil
	}
}

// Email validates that a string is a valid email address.
func (sv *StringValidators) Email(fieldName string) ValidatorFunc[string] {
	return func(value string) error {
		if _, err := mail.ParseAddress(value); err != nil {
			return fmt.Errorf("%s must be a valid email address", fieldName)
		}
		return nil
	}
}

// AlphaNumeric validates that a string contains only alphanumeric characters.
func (sv *StringValidators) AlphaNumeric(fieldName string) ValidatorFunc[string] {
	return func(value string) error {
		for _, r := range value {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return fmt.Errorf("%s must contain only letters and numbers", fieldName)
			}
		}
		return nil
	}
}

// NoSpecialChars validates that a string doesn't contain special characters.
func (sv *StringValidators) NoSpecialChars(fieldName string, allowedChars ...rune) ValidatorFunc[string] {
	allowed := lo.SliceToMap(allowedChars, func(r rune) (rune, bool) {
		return r, true
	})

	return func(value string) error {
		for _, r := range value {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !allowed[r] {
				return fmt.Errorf("%s contains invalid characters", fieldName)
			}
		}
		return nil
	}
}

// SanitizationFunctions provides input sanitization utilities.
type SanitizationFunctions struct{}

// NewSanitizationFunctions creates a new SanitizationFunctions instance.
func NewSanitizationFunctions() *SanitizationFunctions {
	return &SanitizationFunctions{}
}

// TrimWhitespace removes leading and trailing whitespace.
func (sf *SanitizationFunctions) TrimWhitespace(value string) string {
	return strings.TrimSpace(value)
}

// EscapeHTML escapes HTML special characters.
func (sf *SanitizationFunctions) EscapeHTML(value string) string {
	return html.EscapeString(value)
}

// RemoveExtraSpaces collapses multiple consecutive spaces into single spaces.
func (sf *SanitizationFunctions) RemoveExtraSpaces(value string) string {
	spaceRegex := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(spaceRegex.ReplaceAllString(value, " "))
}

// StripNonPrintable removes non-printable characters.
func (sf *SanitizationFunctions) StripNonPrintable(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, value)
}

// Truncate safely truncates a string to a maximum length.
func (sf *SanitizationFunctions) Truncate(value string, maxLen int) string {
	if utf8.RuneCountInString(value) <= maxLen {
		return value
	}

	runes := []rune(value)
	if maxLen > 3 {
		return string(runes[:maxLen-3]) + "..."
	}
	return string(runes[:maxLen])
}

// NormalizeWhitespace normalizes various types of whitespace to regular spaces.
func (sf *SanitizationFunctions) NormalizeWhitespace(value string) string {
	// Replace various whitespace characters with regular spaces
	normalized := strings.ReplaceAll(value, "\t", " ")
	normalized = strings.ReplaceAll(normalized, "\n", " ")
	normalized = strings.ReplaceAll(normalized, "\r", " ")

	// Remove extra spaces
	return sf.RemoveExtraSpaces(normalized)
}

// ValidationResult represents the result of a validation operation.
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// IsValid returns true if the validation passed.
func (vr ValidationResult) IsValid() bool {
	return vr.Valid
}

// HasErrors returns true if there are validation errors.
func (vr ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// FirstError returns the first validation error, or empty string if none.
func (vr ValidationResult) FirstError() string {
	if len(vr.Errors) > 0 {
		return vr.Errors[0]
	}
	return ""
}

// AllErrors returns all validation errors as a single string.
func (vr ValidationResult) AllErrors() string {
	return strings.Join(vr.Errors, "; ")
}

// ValidateWithResult validates a value and returns a ValidationResult.
func ValidateWithResult[T any](ctx context.Context, value T, validators ...Validator[T]) ValidationResult {
	var errors []string

	for _, validator := range validators {
		if err := validator(ctx, value); err != nil {
			errors = append(errors, err.Error())
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// Common pre-compiled regex patterns for performance.
var (
	// UsernamePattern allows alphanumeric characters, hyphens, and underscores.
	UsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// UUIDPattern matches UUID v4 format.
	UUIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// PrebuiltValidators contains commonly used validation functions.
type PrebuiltValidators struct {
	String       *StringValidators
	Sanitization *SanitizationFunctions
}

// NewPrebuiltValidators creates a new instance with all validator types.
func NewPrebuiltValidators() *PrebuiltValidators {
	return &PrebuiltValidators{
		String:       NewStringValidators(),
		Sanitization: NewSanitizationFunctions(),
	}
}

// Username validates a username using common rules.
func (pv *PrebuiltValidators) Username(fieldName string) ValidatorFunc[string] {
	return func(value string) error {
		if err := pv.String.NotEmpty(fieldName)(value); err != nil {
			return err
		}
		if err := pv.String.LengthRange(fieldName, 3, 30)(value); err != nil {
			return err
		}
		if err := pv.String.Regex(fieldName, UsernamePattern, "must contain only letters, numbers, hyphens, and underscores")(value); err != nil {
			return err
		}
		return nil
	}
}

// SimplePassword validates a password with basic requirements.
func (pv *PrebuiltValidators) SimplePassword(fieldName string) ValidatorFunc[string] {
	return func(value string) error {
		if err := pv.String.NotEmpty(fieldName)(value); err != nil {
			return err
		}
		if err := pv.String.MinLength(fieldName, 8)(value); err != nil {
			return err
		}

		// Check for at least one letter and one number
		hasLetter := false
		hasDigit := false
		for _, r := range value {
			if unicode.IsLetter(r) {
				hasLetter = true
			}
			if unicode.IsDigit(r) {
				hasDigit = true
			}
		}

		if !hasLetter || !hasDigit {
			return fmt.Errorf("%s must contain at least one letter and one number", fieldName)
		}

		return nil
	}
}

// StrongPassword validates a password with strict requirements.
func (pv *PrebuiltValidators) StrongPassword(fieldName string) ValidatorFunc[string] {
	return func(value string) error {
		if err := pv.String.NotEmpty(fieldName)(value); err != nil {
			return err
		}
		if err := pv.String.LengthRange(fieldName, 12, 128)(value); err != nil {
			return err
		}

		// Check for uppercase, lowercase, digit, and special character
		hasUpper := false
		hasLower := false
		hasDigit := false
		hasSpecial := false
		specialChars := "@$!%*?&"

		for _, r := range value {
			if unicode.IsUpper(r) {
				hasUpper = true
			} else if unicode.IsLower(r) {
				hasLower = true
			} else if unicode.IsDigit(r) {
				hasDigit = true
			} else if strings.ContainsRune(specialChars, r) {
				hasSpecial = true
			}
		}

		if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
			return fmt.Errorf("%s must contain uppercase, lowercase, number, and special character", fieldName)
		}

		return nil
	}
}

// UUID validates that a string is a valid UUID v4.
func (pv *PrebuiltValidators) UUID(fieldName string) ValidatorFunc[string] {
	return func(value string) error {
		if err := pv.String.NotEmpty(fieldName)(value); err != nil {
			return err
		}
		if err := pv.String.Regex(fieldName, UUIDPattern, "must be a valid UUID")(value); err != nil {
			return err
		}
		return nil
	}
}
