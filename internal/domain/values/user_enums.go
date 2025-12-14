// Package values provides domain value objects with validation.
package values

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

// UserStatus represents a user's account status.
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending  UserStatus = "pending"
)

// AllUserStatuses returns all valid user statuses.
func AllUserStatuses() []UserStatus {
	return []UserStatus{
		UserStatusActive,
		UserStatusInactive,
		UserStatusSuspended,
		UserStatusPending,
	}
}

// IsValid checks if the user status is valid.
func (s UserStatus) IsValid() bool {
	validStatuses := AllUserStatuses()
	for _, status := range validStatuses {
		if s == status {
			return true
		}
	}
	return false
}

// String returns the string representation of user status.
func (s UserStatus) String() string {
	return string(s)
}

// IsActive checks if user is in active status.
func (s UserStatus) IsActive() bool {
	return s == UserStatusActive
}

// CanLogin checks if user can login based on status.
func (s UserStatus) CanLogin() bool {
	return s == UserStatusActive || s == UserStatusPending
}

// MarshalJSON implements json.Marshaler interface.
func (s UserStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (s *UserStatus) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	
	status := UserStatus(strings.ToLower(str))
	if !status.IsValid() {
		return errors.NewDomainValidationError("user_status", fmt.Sprintf("invalid status: %s", str))
	}
	
	*s = status
	return nil
}

// Scan implements the Scanner interface for database compatibility.
func (s *UserStatus) Scan(value interface{}) error {
	if value == nil {
		*s = UserStatusInactive
		return nil
	}
	
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return errors.NewDomainValidationError("user_status", "cannot scan non-string value")
	}
	
	status := UserStatus(strings.ToLower(str))
	if !status.IsValid() {
		return errors.NewDomainValidationError("user_status", fmt.Sprintf("invalid status: %s", str))
	}
	
	*s = status
	return nil
}

// Value implements the driver Valuer interface for database compatibility.
func (s UserStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// UserRole represents a user's role in the system.
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
	UserRoleGuest UserRole = "guest"
)

// AllUserRoles returns all valid user roles.
func AllUserRoles() []UserRole {
	return []UserRole{
		UserRoleAdmin,
		UserRoleUser,
		UserRoleGuest,
	}
}

// IsValid checks if the user role is valid.
func (r UserRole) IsValid() bool {
	validRoles := AllUserRoles()
	for _, role := range validRoles {
		if r == role {
			return true
		}
	}
	return false
}

// String returns the string representation of user role.
func (r UserRole) String() string {
	return string(r)
}

// IsAdmin checks if user has admin privileges.
func (r UserRole) IsAdmin() bool {
	return r == UserRoleAdmin
}

// CanModerate checks if user can moderate content.
func (r UserRole) CanModerate() bool {
	return r == UserRoleAdmin
}

// CanEdit checks if user can edit content.
func (r UserRole) CanEdit() bool {
	return r == UserRoleAdmin || r == UserRoleUser
}

// MarshalJSON implements json.Marshaler interface.
func (r UserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (r *UserRole) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	
	role := UserRole(strings.ToLower(str))
	if !role.IsValid() {
		return errors.NewDomainValidationError("user_role", fmt.Sprintf("invalid role: %s", str))
	}
	
	*r = role
	return nil
}

// Scan implements the Scanner interface for database compatibility.
func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = UserRoleGuest
		return nil
	}
	
	var str string
	switch v := value.(type) {
	case string:
		str = v
	case []byte:
		str = string(v)
	default:
		return errors.NewDomainValidationError("user_role", "cannot scan non-string value")
	}
	
	role := UserRole(strings.ToLower(str))
	if !role.IsValid() {
		return errors.NewDomainValidationError("user_role", fmt.Sprintf("invalid role: %s", str))
	}
	
	*r = role
	return nil
}

// Value implements the driver Valuer interface for database compatibility.
func (r UserRole) Value() (driver.Value, error) {
	return string(r), nil
}