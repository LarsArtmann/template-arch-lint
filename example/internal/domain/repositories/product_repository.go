// Package repositories contains domain repository interfaces.
// These are contracts that infrastructure must implement.
package repositories

import (
	"context"
	"errors"

	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/values"
)

// ProductRepository defines the contract for product persistence.
// This interface lives in the domain layer but is implemented in infrastructure.
// This is the "port" in hexagonal architecture.
type ProductRepository interface {
	// Save persists a product (create or update).
	Save(ctx context.Context, product *entities.Product) error

	// FindByID retrieves a product by its ID.
	// Returns ErrProductNotFound if not found.
	FindByID(ctx context.Context, id values.ProductID) (*entities.Product, error)

	// List retrieves all products.
	// Returns empty slice if no products exist.
	List(ctx context.Context) ([]*entities.Product, error)

	// Delete removes a product by ID.
	// Returns ErrProductNotFound if not found.
	Delete(ctx context.Context, id values.ProductID) error

	// Exists checks if a product exists by ID.
	Exists(ctx context.Context, id values.ProductID) (bool, error)
}

// Common repository errors
var (
	// ErrProductNotFound is returned when a product is not found.
	ErrProductNotFound = errors.New("product not found")

	// ErrProductAlreadyExists is returned when trying to create a product with existing ID.
	ErrProductAlreadyExists = errors.New("product already exists")
)
