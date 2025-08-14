// Package handlers contains HTTP handlers for the application layer.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint-example/internal/domain/values"
)

// ProductHandler handles HTTP requests for products.
// This is part of the application layer - coordinates between HTTP and domain.
type ProductHandler struct {
	productService *services.ProductService
}

// NewProductHandler creates a new product handler.
func NewProductHandler(productService *services.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProductRequest represents the request body for creating a product.
type CreateProductRequest struct {
	ID    string `json:"id" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Price int64  `json:"price" binding:"required"`
}

// ProductResponse represents a product in API responses.
type ProductResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Price          int64  `json:"price"`
	FormattedPrice string `json:"formatted_price"`
	IsExpensive    bool   `json:"is_expensive"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// CreateProduct handles POST /products requests.
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	// Convert to domain value object
	productID, err := values.NewProductID(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid product ID",
			"details": err.Error(),
		})
		return
	}

	// Call domain service
	product, err := h.productService.CreateProduct(c.Request.Context(), productID, req.Name, req.Price)
	if err != nil {
		// Handle specific domain errors
		if err == repositories.ErrProductAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Product already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create product",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	response := h.toProductResponse(product)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": response,
	})
}

// ListProducts handles GET /products requests.
func (h *ProductHandler) ListProducts(c *gin.Context) {
	// Check for expensive filter
	expensiveOnly := c.Query("expensive") == "true"

	var products []*entities.Product
	var err error

	if expensiveOnly {
		products, err = h.productService.GetExpensiveProducts(c.Request.Context())
	} else {
		products, err = h.productService.ListProducts(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list products",
			"details": err.Error(),
		})
		return
	}

	// Convert to response format
	responses := make([]ProductResponse, len(products))
	for i, product := range products {
		responses[i] = h.toProductResponse(product)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Products retrieved successfully",
		"count":    len(responses),
		"products": responses,
	})
}

// GetProduct handles GET /products/:id requests.
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idParam := c.Param("id")

	productID, err := values.NewProductID(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid product ID",
			"details": err.Error(),
		})
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), productID)
	if err != nil {
		if err == repositories.ErrProductNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Product not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get product",
			"details": err.Error(),
		})
		return
	}

	response := h.toProductResponse(product)
	c.JSON(http.StatusOK, gin.H{
		"message": "Product retrieved successfully",
		"product": response,
	})
}

// toProductResponse converts a domain entity to API response format.
func (h *ProductHandler) toProductResponse(product *entities.Product) ProductResponse {
	return ProductResponse{
		ID:             product.ID.String(),
		Name:           product.Name,
		Price:          product.Price,
		FormattedPrice: product.FormattedPrice(),
		IsExpensive:    product.IsExpensive(),
		CreatedAt:      product.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      product.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
