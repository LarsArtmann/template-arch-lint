// Email value object with validation and business rules
package values

import (
	"fmt"
	"regexp"
	"strings"
)

// Email represents a validated email address value object
type Email struct {
	value string
}

// emailRegex provides basic email validation pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewEmail creates a new Email value object with validation
func NewEmail(email string) (Email, error) {
	if err := validateEmailFormat(email); err != nil {
		return Email{}, err
	}

	return Email{
		value: strings.ToLower(strings.TrimSpace(email)),
	}, nil
}

// String returns the string representation of the email
func (e Email) String() string {
	return e.value
}

// Value returns the email value for database storage
func (e Email) Value() string {
	return e.value
}

// Domain returns the domain part of the email
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// LocalPart returns the local part of the email (before @)
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// Equals compares two Email value objects
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// IsEmpty checks if the email is empty
func (e Email) IsEmpty() bool {
	return e.value == ""
}

// validateEmailFormat enforces business rules for email validation
func validateEmailFormat(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Trim whitespace and convert to lowercase for validation
	normalized := strings.ToLower(strings.TrimSpace(email))

	// Check length constraints
	if len(normalized) > 254 {
		return fmt.Errorf("email too long (max 254 characters)")
	}

	if len(normalized) < 5 {
		return fmt.Errorf("email too short (min 5 characters)")
	}

	// Business rule: No spaces allowed
	if strings.Contains(normalized, " ") {
		return fmt.Errorf("email cannot contain spaces")
	}

	// Basic format validation
	if !emailRegex.MatchString(normalized) {
		return fmt.Errorf("invalid email format")
	}

	// Business rule: Check for consecutive dots
	if strings.Contains(normalized, "..") {
		return fmt.Errorf("email cannot contain consecutive dots")
	}

	// Business rule: Cannot start or end with dot
	if strings.HasPrefix(normalized, ".") || strings.HasSuffix(normalized, ".") {
		return fmt.Errorf("email cannot start or end with dot")
	}

	// Split and validate parts
	parts := strings.Split(normalized, "@")
	if len(parts) != 2 {
		return fmt.Errorf("email must contain exactly one @ symbol")
	}

	localPart, domain := parts[0], parts[1]

	// Validate local part
	if len(localPart) == 0 {
		return fmt.Errorf("email local part cannot be empty")
	}

	if len(localPart) > 64 {
		return fmt.Errorf("email local part too long (max 64 characters)")
	}

	// Validate domain
	if len(domain) == 0 {
		return fmt.Errorf("email domain cannot be empty")
	}

	if len(domain) > 253 {
		return fmt.Errorf("email domain too long (max 253 characters)")
	}

	// Domain must contain at least one dot
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("email domain must contain at least one dot")
	}

	return nil
}
