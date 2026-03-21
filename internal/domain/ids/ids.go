// Package ids provides branded, strongly-typed identifiers using go-composable-business-types.
//
// This package replaces the legacy UserID implementation with compile-time type-safe IDs
// that prevent mixing different entity identifiers (e.g., passing a SessionID where a UserID is expected).
//
// # Usage
//
//	import "github.com/LarsArtmann/template-arch-lint/internal/domain/ids"
//
//	// Create IDs with validation
//	userID, err := ids.NewUserID("user-123")
//	sessionID, err := ids.NewSessionID("sess-456")
//
//	// Generate new IDs
//	userID := ids.GenerateUserID()
//
//	// Type-safe comparison
//	if userID.Equal(otherUserID) { ... }
//
//	// Zero value check
//	if userID.IsZero() { ... }
//
//	// Serialization (automatic)
//	json.Marshal(userID)  // "user-123"
//	db.Exec("INSERT ...", userID)  // SQL driver.Value
package ids

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/LarsArtmann/template-arch-lint/pkg/errors"
	"github.com/larsartmann/go-composable-business-types/id"
)

// ID generation and validation constraints.
const (
	idByteLength = 16
	idMinLength  = 2
	idMaxLength  = 100
)

// Brand types provide compile-time distinctness for different entity IDs.
// These are phantom types - they have no runtime representation.

// UserBrand distinguishes UserID from other ID types.
type UserBrand struct{}

// SessionBrand distinguishes SessionID from other ID types.
type SessionBrand struct{}

// ID type aliases for convenient use throughout the codebase.

// UserID is a branded identifier for users. It cannot be accidentally
// mixed with SessionID or other entity IDs at compile time.
type UserID = id.ID[UserBrand, string]

// SessionID is a branded identifier for user sessions.
type SessionID = id.ID[SessionBrand, string]

// Constructor functions with validation.

// NewUserID creates a new UserID with validation.
// Returns an error if the ID format is invalid.
func NewUserID(value string) (UserID, error) {
	err := validateUserID(value)
	if err != nil {
		return UserID{}, err
	}

	return id.NewID[UserBrand](strings.TrimSpace(value)), nil
}

// GenerateUserID creates a new randomly generated UserID.
// Uses crypto/rand for security. Format: "user_<32 hex chars>".
func GenerateUserID() (UserID, error) {
	bytes := make([]byte, idByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return UserID{}, fmt.Errorf("failed to generate random ID: %w", err)
	}

	return id.NewID[UserBrand](fmt.Sprintf("user_%x", bytes)), nil
}

// MustGenerateUserID creates a new UserID or panics on failure.
// Use only in contexts where failure is impossible (e.g., tests, init).
func MustGenerateUserID() UserID {
	id, err := GenerateUserID()
	if err != nil {
		panic(err)
	}

	return id
}

// NewSessionID creates a new SessionID with validation.
func NewSessionID(value string) (SessionID, error) {
	err := validateSessionID(value)
	if err != nil {
		return SessionID{}, err
	}

	return id.NewID[SessionBrand](strings.TrimSpace(value)), nil
}

// GenerateSessionID creates a new randomly generated SessionID.
func GenerateSessionID() (SessionID, error) {
	bytes := make([]byte, idByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return SessionID{}, fmt.Errorf("failed to generate session ID: %w", err)
	}

	return id.NewID[SessionBrand](fmt.Sprintf("sess_%x", bytes)), nil
}

// MustGenerateSessionID creates a new SessionID or panics on failure.
func MustGenerateSessionID() SessionID {
	id, err := GenerateSessionID()
	if err != nil {
		panic(err)
	}

	return id
}

// Validation functions.

func validateUserID(id string) error {
	if id == "" {
		return newValidationError("user ID is required")
	}

	normalized := strings.TrimSpace(id)
	if normalized != id {
		return newValidationError("user ID cannot have leading or trailing whitespace")
	}

	if strings.ContainsAny(normalized, " \t\n\r") {
		return newValidationError("user ID cannot contain whitespace")
	}

	if len(normalized) < idMinLength {
		return newValidationError("user ID too short (minimum 2 characters)")
	}

	if len(normalized) > idMaxLength {
		return newValidationError("user ID too long (maximum 100 characters)")
	}

	for _, char := range normalized {
		if !isValidIDChar(char) {
			return newValidationError(
				"user ID can only contain letters, numbers, hyphens, and underscores",
			)
		}
	}

	return nil
}

func validateSessionID(id string) error {
	if id == "" {
		return newValidationError("session ID is required")
	}

	normalized := strings.TrimSpace(id)
	if normalized != id {
		return newValidationError("session ID cannot have leading or trailing whitespace")
	}

	if len(normalized) < idMinLength {
		return newValidationError("session ID too short (minimum 2 characters)")
	}

	if len(normalized) > idMaxLength {
		return newValidationError("session ID too long (maximum 100 characters)")
	}

	return nil
}

func isValidIDChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= 'A' && char <= 'Z') ||
		(char >= '0' && char <= '9') ||
		char == '-' || char == '_'
}

// newValidationError creates a validation error with the standard error type.
func newValidationError(message string) error {
	return errors.NewValidationError("id", message)
}

// IsGenerated reports whether id appears to be a generated UserID.
// Generated IDs have the format "user_<32 hex chars>" (37 characters total).
func IsGeneratedUserID(id UserID) bool {
	value := id.Get()

	return strings.HasPrefix(value, "user_") && len(value) == 37
}

// IsGeneratedSessionID reports whether id appears to be a generated SessionID.
func IsGeneratedSessionID(id SessionID) bool {
	value := id.Get()

	return strings.HasPrefix(value, "sess_") && len(value) == 37
}
