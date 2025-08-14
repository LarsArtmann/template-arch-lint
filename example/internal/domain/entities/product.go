// Package entities contains domain entities with business identity.
package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/values"
)

// Product represents a product in our domain.
// This is an entity - has identity and can change over time.
type Product struct {
	ID        values.ProductID
	Name      string
	Price     int64 // Price in cents to avoid floating point issues
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProduct creates a new Product with validation.
// This is a factory function that ensures domain invariants.
func NewProduct(id values.ProductID, name string, price int64) (*Product, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return nil, errors.New("product name cannot be empty")
	}

	if len(name) > 100 {
		return nil, errors.New("product name cannot exceed 100 characters")
	}

	if price < 0 {
		return nil, errors.New("product price cannot be negative")
	}

	if price > 10000000 { // $100,000 in cents
		return nil, errors.New("product price cannot exceed $100,000")
	}

	now := time.Now().UTC()

	return &Product{
		ID:        id,
		Name:      name,
		Price:     price,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateName updates the product name with validation.
func (p *Product) UpdateName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return errors.New("product name cannot be empty")
	}

	if len(name) > 100 {
		return errors.New("product name cannot exceed 100 characters")
	}

	p.Name = name
	p.UpdatedAt = time.Now().UTC()

	return nil
}

// UpdatePrice updates the product price with validation.
func (p *Product) UpdatePrice(price int64) error {
	if price < 0 {
		return errors.New("product price cannot be negative")
	}

	if price > 10000000 { // $100,000 in cents
		return errors.New("product price cannot exceed $100,000")
	}

	p.Price = price
	p.UpdatedAt = time.Now().UTC()

	return nil
}

// FormattedPrice returns the price formatted as dollars.
// This is a computed property - doesn't modify state.
func (p *Product) FormattedPrice() string {
	dollars := p.Price / 100
	cents := p.Price % 100
	return fmt.Sprintf("$%d.%02d", dollars, cents)
}

// IsExpensive returns true if the product is considered expensive.
// Business rule: products over $500 are expensive.
func (p *Product) IsExpensive() bool {
	const expensiveThreshold = 50000 // $500 in cents
	return p.Price > expensiveThreshold
}
