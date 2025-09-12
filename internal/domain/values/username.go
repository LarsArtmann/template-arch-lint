// Package values provides domain value objects with validation.
package values

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
)

// UserName represents a validated username value object.
type UserName struct {
	value string
}

// usernameRegex provides basic username validation pattern.
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// reservedUsernames contains usernames that are not allowed.
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

// NewUserName creates a new UserName value object with validation.
func NewUserName(username string) (UserName, error) {
	if err := validateUserNameFormat(username); err != nil {
		return UserName{}, err
	}

	return UserName{
		value: strings.TrimSpace(username),
	}, nil
}

// String returns the string representation of the username.
func (u UserName) String() string {
	return u.value
}

// Value returns the username value for database storage.
func (u UserName) Value() string {
	return u.value
}

// Length returns the length of the username.
func (u UserName) Length() int {
	return len(u.value)
}

// Equals compares two UserName value objects.
func (u UserName) Equals(other UserName) bool {
	return u.value == other.value
}

// IsEmpty checks if the username is empty.
func (u UserName) IsEmpty() bool {
	return u.value == ""
}

// IsReserved checks if the username is in the reserved list.
func (u UserName) IsReserved() bool {
	return reservedUsernames[strings.ToLower(u.value)]
}

// HasValidCharacters checks if username contains only allowed characters.
func (u UserName) HasValidCharacters() bool {
	return usernameRegex.MatchString(u.value)
}

// validateUserNameFormat enforces business rules for username validation.
func validateUserNameFormat(username string) error {
	if username == "" {
		return errors.NewRequiredFieldError("username")
	}

	normalized := strings.TrimSpace(username)

	if err := validateUsernameLength(normalized); err != nil {
		return err
	}

	if err := validateUsernameWhitespace(username, normalized); err != nil {
		return err
	}

	if err := validateUsernameCharacters(normalized); err != nil {
		return err
	}

	if err := validateUsernameEdges(normalized); err != nil {
		return err
	}

	if err := validateUsernameContent(normalized); err != nil {
		return err
	}

	return nil
}

// validateUsernameLength checks length constraints.
func validateUsernameLength(normalized string) error {
	if len(normalized) < 2 {
		return errors.NewValidationError("username", "username too short (minimum 2 characters)")
	}
	if len(normalized) > 50 {
		return errors.NewValidationError("username", "username too long (maximum 50 characters)")
	}
	return nil
}

// validateUsernameWhitespace checks for leading/trailing whitespace.
func validateUsernameWhitespace(username, normalized string) error {
	if username != normalized {
		return errors.NewValidationError("username", "username cannot have leading or trailing spaces")
	}
	return nil
}

// validateUsernameCharacters validates allowed characters.
func validateUsernameCharacters(normalized string) error {
	for _, char := range normalized {
		if !isValidUsernameChar(char) {
			return errors.NewValidationError("username", "name can only contain letters, numbers, dots, hyphens, underscores, apostrophes, commas, and spaces")
		}
	}
	return nil
}

// isValidUsernameChar checks if a character is allowed in usernames.
func isValidUsernameChar(char rune) bool {
	return isASCIIAlphanumeric(char) || isAllowedPunctuation(char) || unicode.IsLetter(char)
}

// isASCIIAlphanumeric checks if character is ASCII letter or digit.
func isASCIIAlphanumeric(char rune) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
}

// isAllowedPunctuation checks if character is allowed punctuation.
func isAllowedPunctuation(char rune) bool {
	switch char {
	case '-', '_', '.', ' ', '\'', ',':
		return true
	default:
		return false
	}
}

// validateUsernameEdges checks start/end character restrictions and consecutive characters.
func validateUsernameEdges(normalized string) error {
	firstChar := normalized[0]
	lastChar := normalized[len(normalized)-1]

	if firstChar == '.' || firstChar == '-' || firstChar == '_' {
		return errors.NewValidationError("username", "name cannot start with dot, hyphen, or underscore")
	}

	if lastChar == '.' || lastChar == '-' || lastChar == '_' {
		return errors.NewValidationError("username", "name cannot end with dot, hyphen, or underscore")
	}

	if strings.Contains(normalized, "..") || strings.Contains(normalized, "--") ||
		strings.Contains(normalized, "__") {
		return errors.NewValidationError("username", "name cannot contain consecutive dots, hyphens, or underscores")
	}

	return nil
}

// validateUsernameContent validates username content rules (letters, reserved names, numbers).
func validateUsernameContent(normalized string) error {
	if err := validateHasLetter(normalized); err != nil {
		return err
	}

	if err := validateNotReserved(normalized); err != nil {
		return err
	}

	if err := validateNotAllNumbers(normalized); err != nil {
		return err
	}

	return nil
}

// validateHasLetter ensures username contains at least one letter.
func validateHasLetter(normalized string) error {
	for _, char := range normalized {
		if unicode.IsLetter(char) {
			return nil
		}
	}
	return errors.NewValidationError("username", "username must contain at least one letter")
}

// validateNotReserved checks against reserved usernames.
func validateNotReserved(normalized string) error {
	lowercased := strings.ToLower(strings.ReplaceAll(normalized, " ", ""))
	if reservedUsernames[lowercased] {
		return errors.NewValidationError("username", fmt.Sprintf("name '%s' is reserved and cannot be used", normalized))
	}
	return nil
}

// validateNotAllNumbers ensures username is not all numbers.
func validateNotAllNumbers(normalized string) error {
	for _, char := range normalized {
		if !unicode.IsDigit(char) {
			return nil
		}
	}
	return errors.NewValidationError("username", "username cannot be all numbers")
}
