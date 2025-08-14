# 📚 Example Project - Clean Architecture with Linting

This example demonstrates how to apply the Template Architecture Lint configurations to a fresh Go project. It shows both correct implementations and common violations.

## 🎯 What This Example Shows

- ✅ **Proper Clean Architecture setup** with linting enforcement
- ✅ **Configuration files** copied from the main template
- ✅ **Correct patterns** that pass all linting rules
- ❌ **Common violations** and how to fix them
- 🔧 **Step-by-step setup** for new projects

## 🏗️ Project Structure

```
example/
├── README.md                    # This guide
├── go.mod                      # Go module
├── .golangci.yml               # Code quality linting (copied from template)
├── .go-arch-lint.yml           # Architecture boundary enforcement
├── justfile                    # Development commands
├── cmd/
│   └── api/
│       └── main.go             # Application entry point
└── internal/
    ├── domain/                 # Domain layer (pure business logic)
    │   ├── entities/          # Business entities
    │   ├── services/          # Domain services
    │   ├── values/            # Value objects
    │   └── repositories/      # Repository interfaces
    ├── application/           # Application layer
    │   └── handlers/          # HTTP handlers
    └── infrastructure/        # Infrastructure layer
        └── persistence/       # Database implementations
```

## 🚀 Quick Start

### 1. Setup and Installation

```bash
cd example/
go mod init github.com/yourname/example-project
go mod tidy

# Install required tools
just install

# Build the project
just build
```

### 2. Verify Linting Works

```bash
# Run all linting (should pass)
just lint

# Run specific linting
just lint-arch          # Architecture boundaries
just lint-code          # Code quality rules
```

### 3. Test the Application

```bash
# Run tests
just test

# Start the application
just run
# Visit: http://localhost:8090
```

## ✅ Correct Patterns (Pass Linting)

### Domain Layer (Pure Business Logic)

**Value Object Example:**
```go
// internal/domain/values/product_id.go
package values

import (
    "errors"
    "strings"
)

type ProductID struct {
    value string
}

func NewProductID(id string) (ProductID, error) {
    id = strings.TrimSpace(id)
    if id == "" {
        return ProductID{}, errors.New("product ID cannot be empty")
    }
    if len(id) > 50 {
        return ProductID{}, errors.New("product ID too long")
    }
    return ProductID{value: id}, nil
}

func (p ProductID) String() string {
    return p.value
}
```

**Entity Example:**
```go
// internal/domain/entities/product.go
package entities

import (
    "time"
    "github.com/yourname/example-project/internal/domain/values"
)

type Product struct {
    ID          values.ProductID
    Name        string
    Price       int64 // cents
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func NewProduct(id values.ProductID, name string, price int64) (*Product, error) {
    if name == "" {
        return nil, errors.New("product name cannot be empty")
    }
    if price < 0 {
        return nil, errors.New("product price cannot be negative")
    }
    
    now := time.Now()
    return &Product{
        ID:        id,
        Name:      name,
        Price:     price,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}
```

**Repository Interface (in Domain):**
```go
// internal/domain/repositories/product_repository.go
package repositories

import (
    "context"
    "github.com/yourname/example-project/internal/domain/entities"
    "github.com/yourname/example-project/internal/domain/values"
)

type ProductRepository interface {
    Save(ctx context.Context, product *entities.Product) error
    FindByID(ctx context.Context, id values.ProductID) (*entities.Product, error)
    List(ctx context.Context) ([]*entities.Product, error)
    Delete(ctx context.Context, id values.ProductID) error
}
```

### Application Layer

**HTTP Handler Example:**
```go
// internal/application/handlers/product_handler.go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/yourname/example-project/internal/domain/services"
)

type ProductHandler struct {
    productService *services.ProductService
}

func NewProductHandler(productService *services.ProductService) *ProductHandler {
    return &ProductHandler{
        productService: productService,
    }
}

func (h *ProductHandler) ListProducts(c *gin.Context) {
    products, err := h.productService.ListProducts(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"products": products})
}
```

### Infrastructure Layer

**Repository Implementation:**
```go
// internal/infrastructure/persistence/memory_product_repository.go
package persistence

import (
    "context"
    "sync"
    "github.com/yourname/example-project/internal/domain/entities"
    "github.com/yourname/example-project/internal/domain/repositories"
    "github.com/yourname/example-project/internal/domain/values"
)

type MemoryProductRepository struct {
    products map[string]*entities.Product
    mutex    sync.RWMutex
}

func NewMemoryProductRepository() repositories.ProductRepository {
    return &MemoryProductRepository{
        products: make(map[string]*entities.Product),
    }
}

func (r *MemoryProductRepository) Save(ctx context.Context, product *entities.Product) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    r.products[product.ID.String()] = product
    return nil
}
```

## ❌ Common Violations (Fail Linting)

### Architecture Violations

```go
// ❌ WRONG: Domain depending on infrastructure
package entities

import (
    "github.com/gin-gonic/gin"  // ❌ Domain cannot import web framework
    "database/sql"              // ❌ Domain cannot import database
)

// ❌ WRONG: Infrastructure types in domain
type Product struct {
    ID    string
    Name  string
    DB    *sql.DB  // ❌ Database dependency in domain entity
}
```

**Fix:**
```go
// ✅ CORRECT: Pure domain entity
package entities

import (
    "time"
    "github.com/yourname/example-project/internal/domain/values"
)

type Product struct {
    ID        values.ProductID  // ✅ Domain value object
    Name      string
    CreatedAt time.Time        // ✅ Standard library only
}
```

### Code Quality Violations

```go
// ❌ WRONG: Multiple violations
func CreateProduct(data interface{}) error {  // ❌ interface{} usage
    if data == nil {
        panic("data is nil")  // ❌ panic usage
    }
    
    // ❌ Function too long (>50 lines)
    // ... 60 lines of code ...
    
    return nil  // ❌ Missing error wrapping
}
```

**Fix:**
```go
// ✅ CORRECT: Proper types and error handling
func CreateProduct(name string, price int64) (*entities.Product, error) {
    if name == "" {
        return nil, fmt.Errorf("product name cannot be empty")
    }
    
    id, err := values.NewProductID(generateID())
    if err != nil {
        return nil, fmt.Errorf("failed to create product ID: %w", err)
    }
    
    product, err := entities.NewProduct(id, name, price)
    if err != nil {
        return nil, fmt.Errorf("failed to create product: %w", err)
    }
    
    return product, nil
}
```

## 🔧 Configuration Files

### Architecture Linting (`.go-arch-lint.yml`)

```yaml
version: 1

modules:
  - name: domain
    path: "internal/domain/**"
    depends_on: []
    may_not_depend_on:
      - "internal/infrastructure/**"
      - "internal/application/**"
      - "github.com/gin-gonic/**"
      - "database/sql"

  - name: application  
    path: "internal/application/**"
    depends_on:
      - "internal/domain/**"
    may_not_depend_on:
      - "internal/infrastructure/**"

  - name: infrastructure
    path: "internal/infrastructure/**"
    depends_on:
      - "internal/domain/**"
      - "internal/application/**"
```

### Code Quality (`.golangci.yml`)

```yaml
# Copied from template - enables 32+ linters
# Key rules:
# - No interface{} usage
# - No panic() usage  
# - Functions max 50 lines
# - Cyclomatic complexity max 10
# - All errors must be wrapped
```

## 🧪 Testing Examples

### Domain Test (Correct Pattern)

```go
// internal/domain/entities/product_test.go
package entities_test

import (
    "testing"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "github.com/yourname/example-project/internal/domain/entities"
    "github.com/yourname/example-project/internal/domain/values"
)

func TestProduct(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Product Entity Suite")
}

var _ = Describe("Product", func() {
    Describe("NewProduct", func() {
        It("should create a valid product", func() {
            id, err := values.NewProductID("prod-123")
            Expect(err).NotTo(HaveOccurred())
            
            product, err := entities.NewProduct(id, "Test Product", 1000)
            
            Expect(err).NotTo(HaveOccurred())
            Expect(product.Name).To(Equal("Test Product"))
            Expect(product.Price).To(Equal(int64(1000)))
        })
        
        It("should reject empty name", func() {
            id, _ := values.NewProductID("prod-123")
            
            _, err := entities.NewProduct(id, "", 1000)
            
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(ContainSubstring("name cannot be empty"))
        })
    })
})
```

## 🚀 Commands Reference

```bash
# 🏗️ Development
just build              # Build application
just run                # Start server (port 8090)
just clean              # Clean build artifacts

# 🔍 Quality Assurance  
just lint               # Run all linting
just lint-arch          # Architecture boundaries only
just lint-code          # Code quality only
just fix                # Auto-fix issues

# 🧪 Testing
just test               # Run all tests
just test-watch         # Run tests in watch mode
just coverage           # Coverage analysis

# 📊 Performance
just bench              # Run benchmarks
just profile            # Generate performance profiles
```

## 🎯 Learning Exercises

### Exercise 1: Fix Architecture Violations

Try adding these violations and then fix them:

```go
// Add to internal/domain/entities/product.go
import "github.com/gin-gonic/gin"  // Should fail lint-arch
```

### Exercise 2: Fix Code Quality Issues

```go
// Add to any file
func BadFunction(data interface{}) {  // Should fail lint-code
    panic("something went wrong")     // Should fail lint-code
}
```

### Exercise 3: Add New Domain Entity

Create a `Category` entity following the same patterns as `Product`.

## 📚 Next Steps

1. **Copy configurations** to your real project
2. **Adapt the architecture** to your domain
3. **Run linting** regularly during development
4. **Set up CI/CD** using the GitHub Actions templates
5. **Add benchmarks** for performance-critical code

## 🔗 Resources

- **Main Template**: `../` (parent directory)
- **Linting Configurations**: `.golangci.yml`, `.go-arch-lint.yml`
- **Documentation**: `../docs/USAGE.md`
- **CI/CD Templates**: `../.github/workflows/`