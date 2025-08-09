// Domain entity example - good architectural design
package entities

import (
	"errors"
	"time"
)

// UserID represents a unique user identifier
type UserID string

// User represents a domain entity
type User struct {
	ID       UserID    `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// NewUser creates a new user with validation
func NewUser(id UserID, email, name string) (*User, error) {
	if id == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}

	now := time.Now()
	return &User{
		ID:       id,
		Email:    email,
		Name:     name,
		Created:  now,
		Modified: now,
	}, nil
}

// Validate ensures the user is in a valid state
func (u *User) Validate() error {
	if u.ID == "" {
		return errors.New("user ID cannot be empty")
	}
	if u.Email == "" {
		return errors.New("email cannot be empty")
	}
	if u.Name == "" {
		return errors.New("name cannot be empty")
	}
	return nil
}
