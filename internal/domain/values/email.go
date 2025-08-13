// Package values provides domain value objects with validation.
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
	if err := validateEmailNotEmpty(email); err != nil {
		return err
	}

	normalized := strings.ToLower(strings.TrimSpace(email))
	
	if err := validateEmailLength(normalized); err != nil {
		return err
	}

	if err := validateEmailBasicFormat(normalized); err != nil {
		return err
	}

	return validateEmailParts(normalized)
}

func validateEmailNotEmpty(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	return nil
}

func validateEmailLength(email string) error {
	if len(email) > 254 {
		return fmt.Errorf("email too long (max 254 characters)")
	}
	if len(email) < 5 {
		return fmt.Errorf("email too short (min 5 characters)")
	}
	return nil
}

func validateEmailBasicFormat(email string) error {
	if strings.Contains(email, " ") {
		return fmt.Errorf("email cannot contain spaces")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	if strings.Contains(email, "..") {
		return fmt.Errorf("email cannot contain consecutive dots")
	}

	if strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
		return fmt.Errorf("email cannot start or end with dot")
	}
	return nil
}

func validateEmailParts(email string) error {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("email must contain exactly one @ symbol")
	}

	localPart, domain := parts[0], parts[1]
	
	if err := validateEmailLocalPart(localPart); err != nil {
		return err
	}

	return validateEmailDomain(domain)
}

func validateEmailLocalPart(localPart string) error {
	if len(localPart) == 0 {
		return fmt.Errorf("email local part cannot be empty")
	}
	if len(localPart) > 64 {
		return fmt.Errorf("email local part too long (max 64 characters)")
	}
	return nil
}

func validateEmailDomain(domain string) error {
	if len(domain) == 0 {
		return fmt.Errorf("email domain cannot be empty")
	}
	if len(domain) > 253 {
		return fmt.Errorf("email domain too long (max 253 characters)")
	}
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("email domain must contain at least one dot")
	}
	return nil
}
