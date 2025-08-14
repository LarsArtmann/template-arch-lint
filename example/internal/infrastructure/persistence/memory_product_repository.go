// Package persistence contains repository implementations.
// This is the infrastructure layer - implements domain contracts.
package persistence

import (
	"context"
	"sync"

	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/values"
)

// MemoryProductRepository implements ProductRepository using in-memory storage.
// This is an adapter in hexagonal architecture - implements the port.
type MemoryProductRepository struct {
	products map[string]*entities.Product
	mutex    sync.RWMutex
}

// NewMemoryProductRepository creates a new in-memory product repository.
func NewMemoryProductRepository() repositories.ProductRepository {
	return &MemoryProductRepository{
		products: make(map[string]*entities.Product),
	}
}

// Save persists a product in memory.
func (r *MemoryProductRepository) Save(ctx context.Context, product *entities.Product) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Create a copy to avoid external modifications
	productCopy := *product
	r.products[product.ID.String()] = &productCopy

	return nil
}

// FindByID retrieves a product by ID.
func (r *MemoryProductRepository) FindByID(ctx context.Context, id values.ProductID) (*entities.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	product, exists := r.products[id.String()]
	if !exists {
		return nil, repositories.ErrProductNotFound
	}

	// Return a copy to avoid external modifications
	productCopy := *product
	return &productCopy, nil
}

// List retrieves all products.
func (r *MemoryProductRepository) List(ctx context.Context) ([]*entities.Product, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	products := make([]*entities.Product, 0, len(r.products))

	for _, product := range r.products {
		// Create copies to avoid external modifications
		productCopy := *product
		products = append(products, &productCopy)
	}

	return products, nil
}

// Delete removes a product by ID.
func (r *MemoryProductRepository) Delete(ctx context.Context, id values.ProductID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, exists := r.products[id.String()]
	if !exists {
		return repositories.ErrProductNotFound
	}

	delete(r.products, id.String())
	return nil
}

// Exists checks if a product exists.
func (r *MemoryProductRepository) Exists(ctx context.Context, id values.ProductID) (bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.products[id.String()]
	return exists, nil
}
