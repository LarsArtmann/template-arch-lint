// Package entities contains the core domain entities for the application.
package entities

import (
	"encoding/json"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

// UserID represents a unique user identifier (alias for convenience).
type UserID = values.UserID

// User represents a domain entity with value objects.
// REFACTORED: Split brain eliminated - using ONLY value objects for type safety
// JSON serialization handled by custom MarshalJSON/UnmarshalJSON methods
// Direct field access prevented through private fields - validation guaranteed.
type User struct {
	ID       values.UserID `json:"id"`
	Created  time.Time     `json:"created"`
	Modified time.Time     `json:"modified"`

	// Private value objects - single source of truth, type safe
	email values.Email    // Private - access through GetEmail() only
	name  values.UserName // Private - access through GetUserName() only
}

// NewUser creates a new user with validation using value objects.
func NewUser(id values.UserID, email, name string) (*User, error) {
	// Validate and create value objects
	emailVO, err := values.NewEmail(email)
	if err != nil {
		return nil, errors.NewValidationError("email", err.Error())
	}

	nameVO, err := values.NewUserName(name)
	if err != nil {
		return nil, errors.NewValidationError("name", err.Error())
	}

	if id.IsEmpty() {
		return nil, errors.NewRequiredFieldError("user ID")
	}

	now := time.Now()
	return &User{
		ID:       id,
		Created:  now,
		Modified: now,
		email:    emailVO, // Single source of truth - value object only
		name:     nameVO,  // Single source of truth - value object only
	}, nil
}

// NewUserFromStrings creates a new user with string ID (for backward compatibility).
// TODO: DEPRECATION CANDIDATE - Remove this once all callers use values.UserID directly
// TODO: TYPE SAFETY - Prefer NewUser with proper value objects over string conversion.
func NewUserFromStrings(id, email, name string) (*User, error) {
	userID, err := values.NewUserID(id)
	if err != nil {
		return nil, errors.NewValidationError("user ID", err.Error())
	}

	return NewUser(userID, email, name)
}

// Validate ensures the user is in a valid state using value objects.
// REFACTORED: Split brain eliminated - direct validation of value objects with no performance overhead.
func (u *User) Validate() error {
	if u.ID.IsEmpty() {
		return errors.NewRequiredFieldError("user ID")
	}

	// Validate email using value object
	email := u.GetEmail()
	if email.IsEmpty() {
		return errors.NewRequiredFieldError("email")
	}

	// Validate name using value object
	name := u.GetUserName()
	if name.IsEmpty() {
		return errors.NewRequiredFieldError("name")
	}

	return nil
}

// Split brain initialization methods removed - no longer needed
// Value objects are now created once during construction and stored as single source of truth

// GetEmail returns the email value object directly.
// REFACTORED: No lazy initialization needed - value object created during construction.
func (u *User) GetEmail() values.Email {
	return u.email
}

// GetUserName returns the username value object directly.
// REFACTORED: No lazy initialization needed - value object created during construction.
func (u *User) GetUserName() values.UserName {
	return u.name
}

// GetCreatedAt returns the creation timestamp.
func (u *User) GetCreatedAt() time.Time {
	return u.Created
}

// GetUpdatedAt returns the modification timestamp.
func (u *User) GetUpdatedAt() time.Time {
	return u.Modified
}

// SetEmail updates the email with validation.
// REFACTORED: Split brain eliminated - only updates single value object field.
func (u *User) SetEmail(email string) error {
	emailVO, err := values.NewEmail(email)
	if err != nil {
		return errors.NewValidationError("email", err.Error())
	}

	// Single source of truth - no synchronization needed
	u.email = emailVO
	u.Modified = time.Now()
	return nil
}

// SetName updates the name with validation.
// REFACTORED: Split brain eliminated - only updates single value object field.
func (u *User) SetName(name string) error {
	nameVO, err := values.NewUserName(name)
	if err != nil {
		return errors.NewValidationError("name", err.Error())
	}

	// Single source of truth - no synchronization needed
	u.name = nameVO
	u.Modified = time.Now()
	return nil
}

// EmailDomain returns the domain part of the user's email.
func (u *User) EmailDomain() string {
	return u.GetEmail().Domain()
}

// IsEmailValid checks if the user's email is valid.
func (u *User) IsEmailValid() bool {
	email := u.GetEmail()
	return !email.IsEmpty()
}

// IsNameReserved checks if the username is reserved.
func (u *User) IsNameReserved() bool {
	return u.GetUserName().IsReserved()
}

// MarshalJSON implements custom JSON marshaling for User.
// Converts value objects to their string representations for JSON serialization.
func (u *User) MarshalJSON() ([]byte, error) {
	// Create a temporary struct for JSON marshaling with string fields
	type userJSON struct {
		ID       string    `json:"id"`
		Email    string    `json:"email"`
		Name     string    `json:"name"`
		Created  time.Time `json:"created"`
		Modified time.Time `json:"modified"`
	}

	// Convert value objects to strings
	temp := userJSON{
		ID:       u.ID.String(),
		Email:    u.email.String(),
		Name:     u.name.String(),
		Created:  u.Created,
		Modified: u.Modified,
	}

	return json.Marshal(temp)
}

// UnmarshalJSON implements custom JSON unmarshaling for User.
// Converts JSON string values to internal value objects with validation.
func (u *User) UnmarshalJSON(data []byte) error {
	// Create a temporary struct for JSON unmarshaling
	type userJSON struct {
		ID       string    `json:"id"`
		Email    string    `json:"email"`
		Name     string    `json:"name"`
		Created  time.Time `json:"created"`
		Modified time.Time `json:"modified"`
	}

	var temp userJSON
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Validate and create value objects
	userID, err := values.NewUserID(temp.ID)
	if err != nil {
		return errors.NewValidationError("id", err.Error())
	}

	email, err := values.NewEmail(temp.Email)
	if err != nil {
		return errors.NewValidationError("email", err.Error())
	}

	name, err := values.NewUserName(temp.Name)
	if err != nil {
		return errors.NewValidationError("name", err.Error())
	}

	// Set the fields
	u.ID = userID
	u.email = email
	u.name = name
	u.Created = temp.Created
	u.Modified = temp.Modified

	return nil
}
