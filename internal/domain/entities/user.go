// Package entities contains the core domain entities for the application.
package entities

import (
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

// UserID represents a unique user identifier (alias for convenience).
type UserID = values.UserID

// User represents a domain entity with value objects.
// TODO: CRITICAL SPLIT BRAIN - Email/Name exist as both string and value objects
// TODO: TYPE SAFETY VIOLATION - Public string fields bypass validation
// TODO: ARCHITECTURAL DECISION NEEDED - Choose one approach:
//
//	Option A: Use value objects in struct, implement custom JSON marshaling
//	Option B: Keep strings, remove duplicate value object fields
//	Option C: Make fields private, add typed getters/setters
//
// TODO: INVALID STATE PROTECTION - Direct field assignment can create invalid users
// TODO: JSON SERIALIZATION - Implement MarshalJSON/UnmarshalJSON for value objects
type User struct {
	ID       values.UserID `json:"id"`
	Email    string        `json:"email"` // String for JSON serialization - SPLIT BRAIN WARNING!
	Name     string        `json:"name"`  // String for JSON serialization - SPLIT BRAIN WARNING!
	Created  time.Time     `json:"created"`
	Modified time.Time     `json:"modified"`

	// Internal value objects for domain logic - DUPLICATES above fields!
	emailVO values.Email    // TODO: REMOVE or use exclusively
	nameVO  values.UserName // TODO: REMOVE or use exclusively
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
		Email:    emailVO.Value(), // Store string value for JSON
		Name:     nameVO.Value(),  // Store string value for JSON
		Created:  now,
		Modified: now,
		emailVO:  emailVO,
		nameVO:   nameVO,
	}, nil
}

// NewUserFromStrings creates a new user with string ID (for backward compatibility).
// TODO: DEPRECATION CANDIDATE - Remove this once all callers use values.UserID directly
// TODO: TYPE SAFETY - Prefer NewUser with proper value objects over string conversion
func NewUserFromStrings(id, email, name string) (*User, error) {
	userID, err := values.NewUserID(id)
	if err != nil {
		return nil, errors.NewValidationError("user ID", err.Error())
	}

	return NewUser(userID, email, name)
}

// Validate ensures the user is in a valid state using value objects.
// TODO: SPLIT BRAIN COMPLEXITY - This method needs to sync string and value object states
// TODO: PERFORMANCE - Lazy initialization on every validation call is expensive
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

// initEmailVO initializes email value object if needed.
// TODO: ARCHITECTURAL DEBT - This lazy initialization indicates split brain design problem
// TODO: ERROR HANDLING - Silent error swallowing (err == nil) hides validation failures
// TODO: THREAD SAFETY - Mutation without locks in concurrent access scenario
func (u *User) initEmailVO() {
	if u.emailVO.IsEmpty() && u.Email != "" {
		if emailVO, err := values.NewEmail(u.Email); err == nil {
			u.emailVO = emailVO
		}
	}
}

// initNameVO initializes name value object if needed.
// TODO: DUPLICATION - Same pattern as initEmailVO, extract generic initialization
// TODO: ARCHITECTURAL DEBT - Remove when split brain is resolved
func (u *User) initNameVO() {
	if u.nameVO.IsEmpty() && u.Name != "" {
		if nameVO, err := values.NewUserName(u.Name); err == nil {
			u.nameVO = nameVO
		}
	}
}

// GetEmail returns the email value object.
// TODO: PERFORMANCE - Lazy init on every getter call is expensive
// TODO: SPLIT BRAIN - This exposes internal complexity to callers
func (u *User) GetEmail() values.Email {
	u.initEmailVO()
	return u.emailVO
}

// GetUserName returns the username value object.
// TODO: CONSISTENCY - Same issues as GetEmail, needs architectural fix
func (u *User) GetUserName() values.UserName {
	u.initNameVO()
	return u.nameVO
}

// SetEmail updates the email with validation.
// TODO: CRITICAL SPLIT BRAIN - Must manually sync emailVO AND Email fields!
// TODO: RACE CONDITION - No locking for concurrent field updates
// TODO: ARCHITECTURAL DEBT - This dual-update pattern is error-prone
func (u *User) SetEmail(email string) error {
	emailVO, err := values.NewEmail(email)
	if err != nil {
		return errors.NewValidationError("email", err.Error())
	}

	// SPLIT BRAIN MAINTENANCE - Both fields must stay synchronized!
	u.emailVO = emailVO
	u.Email = emailVO.Value() // TODO: Remove when split brain is resolved
	u.Modified = time.Now()
	return nil
}

// SetName updates the name with validation.
// TODO: DUPLICATION - Same split brain pattern as SetEmail
// TODO: EXTRACT PATTERN - Create generic setValue[T] method to reduce duplication
func (u *User) SetName(name string) error {
	nameVO, err := values.NewUserName(name)
	if err != nil {
		return errors.NewValidationError("name", err.Error())
	}

	// SPLIT BRAIN MAINTENANCE - Both fields must stay synchronized!
	u.nameVO = nameVO
	u.Name = nameVO.Value() // TODO: Remove when split brain is resolved
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
