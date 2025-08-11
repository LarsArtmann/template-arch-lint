// UserID value object for type-safe user identification
package values

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// UserID represents a unique user identifier value object
type UserID struct {
	value string
}

// NewUserID creates a new UserID value object with validation
func NewUserID(id string) (UserID, error) {
	if err := validateUserIDFormat(id); err != nil {
		return UserID{}, err
	}

	return UserID{
		value: strings.TrimSpace(id),
	}, nil
}

// GenerateUserID creates a new random UserID
func GenerateUserID() (UserID, error) {
	// Generate a random ID using crypto/rand for security
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return UserID{}, fmt.Errorf("failed to generate random ID: %w", err)
	}

	// Convert to hex string with user prefix
	id := fmt.Sprintf("user_%x", bytes)

	return UserID{
		value: id,
	}, nil
}

// String returns the string representation of the user ID
func (u UserID) String() string {
	return u.value
}

// String returns the user ID value for database storage
func (u UserID) StringValue() string {
	return u.value
}

// Equals compares two UserID value objects
func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}

// IsEmpty checks if the user ID is empty
func (u UserID) IsEmpty() bool {
	return u.value == ""
}

// IsGenerated checks if this looks like a generated ID
func (u UserID) IsGenerated() bool {
	return strings.HasPrefix(u.value, "user_") && len(u.value) == 37 // "user_" + 32 hex chars
}

// validateUserIDFormat enforces business rules for user ID validation
func validateUserIDFormat(id string) error {
	if id == "" {
		return fmt.Errorf("user ID cannot be empty")
	}

	// Trim whitespace for validation
	normalized := strings.TrimSpace(id)

	// Business rule: No leading/trailing whitespace in original
	if id != normalized {
		return fmt.Errorf("user ID cannot have leading or trailing spaces")
	}

	// Business rule: Length constraints
	if len(normalized) < 1 {
		return fmt.Errorf("user ID too short")
	}

	if len(normalized) > 100 {
		return fmt.Errorf("user ID too long (maximum 100 characters)")
	}

	// Business rule: No whitespace allowed anywhere
	if strings.Contains(normalized, " ") || strings.Contains(normalized, "\t") ||
		strings.Contains(normalized, "\n") || strings.Contains(normalized, "\r") {
		return fmt.Errorf("user ID cannot contain whitespace")
	}

	// Business rule: Basic character validation - allow alphanumeric, hyphen, underscore
	for _, char := range normalized {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return fmt.Errorf("user ID can only contain letters, numbers, hyphens, and underscores")
		}
	}

	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (u UserID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.value)
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (u *UserID) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	userID, err := NewUserID(value)
	if err != nil {
		return err
	}

	*u = userID
	return nil
}

// Value implements the driver.Valuer interface for database storage
func (u UserID) Value() (driver.Value, error) {
	return u.value, nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (u *UserID) Scan(value interface{}) error {
	if value == nil {
		*u = UserID{}
		return nil
	}

	switch v := value.(type) {
	case string:
		userID, err := NewUserID(v)
		if err != nil {
			return err
		}
		*u = userID
		return nil
	case []byte:
		userID, err := NewUserID(string(v))
		if err != nil {
			return err
		}
		*u = userID
		return nil
	default:
		return fmt.Errorf("cannot scan %T into UserID", value)
	}
}
