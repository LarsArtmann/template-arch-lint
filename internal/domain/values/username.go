// UserName value object with validation and business rules
package values

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// UserName represents a validated username value object
type UserName struct {
	value string
}

// usernameRegex provides basic username validation pattern
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// reservedUsernames contains usernames that are not allowed
var reservedUsernames = map[string]bool{
	"admin":      true,
	"root":       true,
	"system":     true,
	"user":       true,
	"guest":      true,
	"anonymous":  true,
	"null":       true,
	"undefined":  true,
	"api":        true,
	"www":        true,
	"mail":       true,
	"email":      true,
	"support":    true,
	"help":       true,
	"info":       true,
	"contact":    true,
	"noreply":    true,
	"no-reply":   true,
	"postmaster": true,
	"webmaster":  true,
	"hostmaster": true,
	"abuse":      true,
	"security":   true,
	"privacy":    true,
	"legal":      true,
	"billing":    true,
	"sales":      true,
	"marketing":  true,
}

// NewUserName creates a new UserName value object with validation
func NewUserName(username string) (UserName, error) {
	if err := validateUserNameFormat(username); err != nil {
		return UserName{}, err
	}

	return UserName{
		value: strings.TrimSpace(username),
	}, nil
}

// String returns the string representation of the username
func (u UserName) String() string {
	return u.value
}

// Value returns the username value for database storage
func (u UserName) Value() string {
	return u.value
}

// Length returns the length of the username
func (u UserName) Length() int {
	return len(u.value)
}

// Equals compares two UserName value objects
func (u UserName) Equals(other UserName) bool {
	return u.value == other.value
}

// IsEmpty checks if the username is empty
func (u UserName) IsEmpty() bool {
	return u.value == ""
}

// IsReserved checks if the username is in the reserved list
func (u UserName) IsReserved() bool {
	return reservedUsernames[strings.ToLower(u.value)]
}

// HasValidCharacters checks if username contains only allowed characters
func (u UserName) HasValidCharacters() bool {
	return usernameRegex.MatchString(u.value)
}

// validateUserNameFormat enforces business rules for username validation
func validateUserNameFormat(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Trim whitespace for validation
	normalized := strings.TrimSpace(username)

	// Business rule: Length constraints
	if len(normalized) < 2 {
		return fmt.Errorf("username too short (minimum 2 characters)")
	}

	if len(normalized) > 50 {
		return fmt.Errorf("username too long (maximum 50 characters)")
	}

	// Business rule: No leading/trailing whitespace in original
	if username != normalized {
		return fmt.Errorf("username cannot have leading or trailing spaces")
	}

	// Business rule: Character validation - allow spaces for display names
	validChars := true
	for _, char := range normalized {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.' || char == ' ') {
			validChars = false
			break
		}
	}
	
	if !validChars {
		return fmt.Errorf("name can only contain letters, numbers, dots, hyphens, underscores, and spaces")
	}

	// Business rule: Cannot start or end with special characters (but allow spaces in middle)
	firstChar := normalized[0]
	lastChar := normalized[len(normalized)-1]

	if firstChar == '.' || firstChar == '-' || firstChar == '_' {
		return fmt.Errorf("name cannot start with dot, hyphen, or underscore")
	}

	if lastChar == '.' || lastChar == '-' || lastChar == '_' {
		return fmt.Errorf("name cannot end with dot, hyphen, or underscore")
	}

	// Business rule: No excessive consecutive special characters (relaxed for display names)
	if strings.Contains(normalized, "..") || strings.Contains(normalized, "--") ||
		strings.Contains(normalized, "__") {
		return fmt.Errorf("name cannot contain consecutive dots, hyphens, or underscores")
	}

	// Business rule: Must contain at least one letter
	hasLetter := false
	for _, char := range normalized {
		if unicode.IsLetter(char) {
			hasLetter = true
			break
		}
	}

	if !hasLetter {
		return fmt.Errorf("username must contain at least one letter")
	}

	// Business rule: Check against reserved usernames (less restrictive for display names)
	lowercased := strings.ToLower(strings.ReplaceAll(normalized, " ", ""))
	if reservedUsernames[lowercased] {
		return fmt.Errorf("name '%s' is reserved and cannot be used", normalized)
	}

	// Business rule: Cannot be all numbers
	allNumbers := true
	for _, char := range normalized {
		if !unicode.IsDigit(char) {
			allNumbers = false
			break
		}
	}

	if allNumbers {
		return fmt.Errorf("username cannot be all numbers")
	}

	return nil
}
