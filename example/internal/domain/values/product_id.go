// Package values contains immutable value objects for the domain.
package values

import (
	"errors"
	"strings"
)

// ProductID represents a unique product identifier.
// This is a value object - immutable and validated.
type ProductID struct {
	value string
}

// NewProductID creates a new ProductID with validation.
// Returns an error if the ID is invalid.
func NewProductID(id string) (ProductID, error) {
	id = strings.TrimSpace(id)

	if id == "" {
		return ProductID{}, errors.New("product ID cannot be empty")
	}

	if len(id) > 50 {
		return ProductID{}, errors.New("product ID cannot exceed 50 characters")
	}

	// Basic format validation - adjust for your needs
	if strings.Contains(id, " ") {
		return ProductID{}, errors.New("product ID cannot contain spaces")
	}

	return ProductID{value: id}, nil
}

// String returns the string representation of the ProductID.
func (p ProductID) String() string {
	return p.value
}

// Equals checks if two ProductIDs are equal.
func (p ProductID) Equals(other ProductID) bool {
	return p.value == other.value
}

// IsEmpty checks if the ProductID is empty (zero value).
func (p ProductID) IsEmpty() bool {
	return p.value == ""
}
