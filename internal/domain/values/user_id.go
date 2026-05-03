// Package values provides domain value objects with validation.
package values

import (
	"github.com/LarsArtmann/template-arch-lint/internal/domain/ids"
)

// UserID is a type alias for the branded UserID from the ids package.
//
// DEPRECATED: Use ids.UserID directly. This alias is maintained for backward
// compatibility during migration. Will be removed in a future version.
//
// The new UserID provides:
//   - Compile-time type safety (cannot mix UserID with SessionID)
//   - Built-in JSON, SQL, and binary serialization
//   - Zero-allocation operations
//   - Generic ID capabilities via go-branded-id
//
// Example migration:
//
//	// Old (custom struct):
//	import "github.com/LarsArtmann/template-arch-lint/internal/domain/values"
//	userID, err := values.NewUserID("user-123")
//
//	// New (branded type):
//	import "github.com/LarsArtmann/template-arch-lint/internal/domain/ids"
//	userID, err := ids.NewUserID("user-123")
//
// See internal/domain/ids/ids.go for the new implementation.
type UserID = ids.UserID

// NewUserID creates a new UserID value object with validation.
//
// DEPRECATED: Use ids.NewUserID instead.
func NewUserID(id string) (UserID, error) {
	return ids.NewUserID(id)
}

// GenerateUserID creates a new random UserID.
//
// DEPRECATED: Use ids.GenerateUserID instead.
func GenerateUserID() (UserID, error) {
	return ids.GenerateUserID()
}

// MustGenerateUserID creates a new UserID or panics on failure.
//
// DEPRECATED: Use ids.MustGenerateUserID instead.
func MustGenerateUserID() UserID {
	return ids.MustGenerateUserID()
}

// IsEmpty checks if the user ID is empty.
// This is a backward-compatibility wrapper around IsZero().
//
// DEPRECATED: Use id.IsZero() instead.
func IsEmpty(id UserID) bool {
	return id.IsZero()
}

// Equals compares two UserID value objects.
// This is a backward-compatibility wrapper.
//
// DEPRECATED: Use id.Equal() instead.
func Equals(a, b UserID) bool {
	return a.Equal(b)
}

// IsGenerated reports whether id appears to be a generated UserID.
func IsGenerated(id UserID) bool {
	return ids.IsGeneratedUserID(id)
}
