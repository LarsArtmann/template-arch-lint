// Package services contains domain services that encapsulate business logic.
package services

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/values"
	serviceerrors "github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
)

// ProductService encapsulates business logic for products.
// This is a domain service - contains business rules and coordinates entities.
type ProductService struct {
	productRepo repositories.ProductRepository
}

// NewProductService creates a new ProductService.
func NewProductService(productRepo repositories.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

// CreateProduct creates a new product with business validation.
func (s *ProductService) CreateProduct(ctx context.Context, id values.ProductID, name string, price int64) (*entities.Product, error) {
	// Check if product already exists
	exists, err := s.productRepo.Exists(ctx, id)
	if err != nil {
		return nil, serviceerrors.WrapRepoError("check existence of", "product", err)
	}

	if exists {
		return nil, repositories.ErrProductAlreadyExists
	}

	// Create product entity (with domain validation)
	product, err := entities.NewProduct(id, name, price)
	if err != nil {
		return nil, serviceerrors.WrapServiceError("create product entity", err)
	}

	// Business rule: Premium products (>$1000) need special validation
	if s.isPremiumProduct(product) {
		if err := s.validatePremiumProduct(product); err != nil {
			return nil, serviceerrors.WrapBusinessRuleError("premium product", err)
		}
	}

	// Save to repository
	if err := s.productRepo.Save(ctx, product); err != nil {
		return nil, serviceerrors.WrapRepoError("save", "product", err)
	}

	return product, nil
}

// GetProduct retrieves a product by ID.
func (s *ProductService) GetProduct(ctx context.Context, id values.ProductID) (*entities.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, serviceerrors.WrapRepoError("get", "product", err)
	}

	return product, nil
}

// ListProducts retrieves all products.
func (s *ProductService) ListProducts(ctx context.Context) ([]*entities.Product, error) {
	products, err := s.productRepo.List(ctx)
	if err != nil {
		return nil, serviceerrors.WrapRepoError("list", "products", err)
	}

	return products, nil
}

// UpdateProduct updates an existing product.
func (s *ProductService) UpdateProduct(ctx context.Context, id values.ProductID, name string, price int64) (*entities.Product, error) {
	// Get existing product
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, serviceerrors.WrapRepoError("get for update", "product", err)
	}

	// Update fields with domain validation
	if err := product.UpdateName(name); err != nil {
		return nil, serviceerrors.WrapServiceError("update product name", err)
	}

	if err := product.UpdatePrice(price); err != nil {
		return nil, serviceerrors.WrapServiceError("update product price", err)
	}

	// Business rule validation for premium products
	if s.isPremiumProduct(product) {
		if err := s.validatePremiumProduct(product); err != nil {
			return nil, serviceerrors.WrapBusinessRuleError("premium product", err)
		}
	}

	// Save updated product
	if err := s.productRepo.Save(ctx, product); err != nil {
		return nil, serviceerrors.WrapRepoError("save updated", "product", err)
	}

	return product, nil
}

// DeleteProduct removes a product.
func (s *ProductService) DeleteProduct(ctx context.Context, id values.ProductID) error {
	if err := s.productRepo.Delete(ctx, id); err != nil {
		return serviceerrors.WrapRepoError("delete", "product", err)
	}

	return nil
}

// GetExpensiveProducts returns products that are considered expensive.
// This demonstrates a business rule implemented in the service.
func (s *ProductService) GetExpensiveProducts(ctx context.Context) ([]*entities.Product, error) {
	allProducts, err := s.productRepo.List(ctx)
	if err != nil {
		return nil, serviceerrors.WrapRepoError("get", "products", err)
	}

	var expensiveProducts []*entities.Product
	for _, product := range allProducts {
		if product.IsExpensive() {
			expensiveProducts = append(expensiveProducts, product)
		}
	}

	return expensiveProducts, nil
}

// isPremiumProduct checks if a product is premium (>$1000).
// This is a private business rule.
func (s *ProductService) isPremiumProduct(product *entities.Product) bool {
	const premiumThreshold = 100000 // $1000 in cents
	return product.Price > premiumThreshold
}

// validatePremiumProduct applies special validation rules for premium products.
// Business rule: Premium products must have detailed names.
func (s *ProductService) validatePremiumProduct(product *entities.Product) error {
	if len(product.Name) < 10 {
		return fmt.Errorf("premium products must have detailed names (minimum 10 characters)")
	}

	return nil
}
