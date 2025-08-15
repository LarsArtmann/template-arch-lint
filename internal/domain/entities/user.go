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
type User struct {
	ID       values.UserID `json:"id"`
	Email    string        `json:"email"` // String for JSON serialization
	Name     string        `json:"name"`  // String for JSON serialization
	Created  time.Time     `json:"created"`
	Modified time.Time     `json:"modified"`

	// Internal value objects for domain logic
	emailVO values.Email
	nameVO  values.UserName
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
func NewUserFromStrings(id, email, name string) (*User, error) {
	userID, err := values.NewUserID(id)
	if err != nil {
		return nil, errors.NewValidationError("user ID", err.Error())
	}

	return NewUser(userID, email, name)
}

// Validate ensures the user is in a valid state using value objects.
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
func (u *User) initEmailVO() {
	if u.emailVO.IsEmpty() && u.Email != "" {
		if emailVO, err := values.NewEmail(u.Email); err == nil {
			u.emailVO = emailVO
		}
	}
}

// initNameVO initializes name value object if needed.
func (u *User) initNameVO() {
	if u.nameVO.IsEmpty() && u.Name != "" {
		if nameVO, err := values.NewUserName(u.Name); err == nil {
			u.nameVO = nameVO
		}
	}
}

// GetEmail returns the email value object.
func (u *User) GetEmail() values.Email {
	u.initEmailVO()
	return u.emailVO
}

// GetUserName returns the username value object.
func (u *User) GetUserName() values.UserName {
	u.initNameVO()
	return u.nameVO
}

// SetEmail updates the email with validation.
func (u *User) SetEmail(email string) error {
	emailVO, err := values.NewEmail(email)
	if err != nil {
		return errors.NewValidationError("email", err.Error())
	}

	u.emailVO = emailVO
	u.Email = emailVO.Value()
	u.Modified = time.Now()
	return nil
}

// SetName updates the name with validation.
func (u *User) SetName(name string) error {
	nameVO, err := values.NewUserName(name)
	if err != nil {
		return errors.NewValidationError("name", err.Error())
	}

	u.nameVO = nameVO
	u.Name = nameVO.Value()
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
