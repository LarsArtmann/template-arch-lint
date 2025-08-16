# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🎯 What This Project IS and IS NOT

### ✅ **What This Project IS:**
- **Go Linting Template**: Demonstrates enterprise-grade architecture and code quality enforcement
- **Reference Implementation**: Shows Clean Architecture + DDD patterns in Go
- **Configuration Library**: Provides `.go-arch-lint.yml`, `.golangci.yml`, and `justfile` for copy/paste use
- **Simple HTMX Demo**: Basic web app with templ templates and SQLite database
- **Educational Resource**: Learn proper Go architecture boundaries and functional programming patterns

### ❌ **What This Project IS NOT:**
- **Production Application**: Not meant for real business use - it's a template/demo
- **Framework or Library**: Not installable via `go get` - copy configurations instead  
- **Enterprise Platform**: Despite having monitoring/Docker/K8s - these are demos of over-engineering
- **Complex Business Domain**: User CRUD is intentionally simple to focus on architecture

### 🎯 **Core Purpose:**
Copy the linting configurations (`.go-arch-lint.yml`, `.golangci.yml`, `justfile`) to your real projects to enforce architectural boundaries and code quality. The Go code demonstrates how to structure projects following these rules.

## 🏗️ High-Level Architecture Understanding

### Layer Structure (Dependency Flow: Infrastructure → Application → Domain)
```
web/templates/          # Templ templates for server-side rendering
├── components/         # Reusable UI components  
├── layouts/           # Page layouts
└── pages/             # Full page templates

internal/application/   # HTTP handlers & use case orchestration
├── handlers/          # HTTP request handlers (user_handler.go)
├── dto/               # Data transfer objects for HTTP
├── http/              # HTTP response helpers
└── middleware/        # Cross-cutting concerns

internal/domain/        # Pure business logic (NO external dependencies)
├── entities/          # Business entities (user.go with value objects)
├── services/          # Domain services (user_service.go)  
├── repositories/      # Repository interfaces (user_repository.go)
├── values/            # Value objects (email.go, username.go, user_id.go)
├── errors/            # Domain-specific errors
└── shared/            # Result pattern implementation

internal/infrastructure/ # External concerns
├── persistence/       # Repository implementations
└── repositories/      # Database-specific code

internal/db/            # SQLC-generated type-safe SQL code
sql/
├── schema/            # Database schema files
└── queries/           # SQL query files for SQLC
```

### Key Architectural Patterns Demonstrated
- **Clean Architecture**: Strict dependency rules enforced by go-arch-lint
- **Domain-Driven Design**: Rich domain entities with value objects
- **Functional Programming**: Heavy use of samber/lo for Map/Filter/Reduce operations
- **Result Pattern**: `internal/domain/shared/result.go` for error handling
- **Value Objects**: Email, UserName, UserID with validation in domain/values
- **Repository Pattern**: Domain interfaces implemented by infrastructure
- **HTMX + Templ**: Server-side rendering with progressive enhancement

## Essential Commands

### Development Workflow
```bash
# Install all required linting tools
just install

# Run complete linting suite (architecture + code quality + filenames)
just lint

# Auto-fix formatting and simple issues
just fix

# Generate templ templates and build application
just build

# Start the HTTP server (port 8080)
just run

# Development mode with auto-reload
just dev

# Run all tests with coverage
just test

# Complete CI/CD simulation
just ci
```

### Specialized Linting
```bash
# Architecture boundaries only
just lint-arch

# Code quality only  
just lint-code

# Problematic filenames only
just lint-files

# Maximum strictness mode
just lint-strict

# Security-focused linting
just lint-security
```

### Template Generation
```bash
# Generate a-h/templ templates
just templ

# Force regeneration after template changes
templ generate
```

### Testing & Development
```bash
# Run single test file or package
go test ./internal/domain/services/ -v
go test ./internal/domain/entities/ -v

# Run specific test function
go test ./internal/domain/services/ -v -run TestUserService_CreateUser

# Run tests with race detection
go test ./... -v -race

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
go test ./internal/domain/services/ -bench=.

# Configuration testing
just config-test
```

### Database Operations
```bash
# Generate type-safe SQL code from schema/queries
sqlc generate

# Verify sqlc configuration
sqlc verify

# SQLite database location
# Development: ./app.db
# Test: in-memory
```

## Architecture (What the Linting Enforces)

### Layer Dependencies (Strictly Enforced by `.go-arch-lint.yml`)
```
Infrastructure → Application → Domain
```

### Key Layers
- **Domain** (`internal/domain/`): Pure business logic, no external dependencies
- **Application** (`internal/application/`): HTTP handlers, orchestrates domain + infrastructure  
- **Infrastructure** (`internal/infrastructure/`): Database, external services
- **Database** (`internal/db/`): SQLC-generated type-safe SQL code

### Key Libraries Used
- **gin + templ + HTMX**: HTTP server with type-safe templates
- **sqlc**: Type-safe SQL code generation  
- **samber/lo**: Functional programming (Filter, Map, Reduce)
- **viper**: Configuration management
- **Ginkgo/Gomega**: BDD testing

## Important Files to Copy to Your Projects

- **`.go-arch-lint.yml`**: Architecture boundary enforcement  
- **`.golangci.yml`**: Maximum strictness code quality rules
- **`justfile`**: Complete linting and development automation
- **`sqlc.yaml`**: Type-safe SQL generation configuration

## What the Linting Enforces

- Domain layer cannot import infrastructure packages
- No `interface{}`, `any`, or `panic()` usage (`.golangci.yml`)
- Functional programming patterns with samber/lo
- Type-safe database operations via SQLC
- Value objects for domain primitives

## Architecture Violations You'll Get

Common violations the linters catch:
- `domain-entities cannot depend on infrastructure`
- `🚨 BANNED: interface{} erases type safety`
- `DELETE statements should include WHERE clauses`
- Function too long (max 50 lines)
- Cyclomatic complexity too high (max 10)

**Note**: The project includes extensive Docker/K8s/monitoring setup as examples of over-engineering to avoid in templates. Focus on the core linting configurations.

## 🔍 Code Understanding Guidelines

### Value Objects Pattern
Value objects enforce validation and type safety:
- `internal/domain/values/user_id.go` - Validated user identifiers
- `internal/domain/values/email.go` - Email validation with domain extraction
- `internal/domain/values/username.go` - Username validation with reserved word checking

### Repository Pattern Implementation
- **Interfaces**: `internal/domain/repositories/user_repository.go` (domain layer)
- **Implementations**: `internal/infrastructure/persistence/user_repository_sqlc.go` (infrastructure layer)
- **In-Memory**: `internal/domain/repositories/inmemory_user_repository.go` (for testing)

### SQLC Integration
- **Schema**: `sql/schema/001_users.sql` - Database structure  
- **Queries**: `sql/queries/users.sql` - Type-safe SQL operations
- **Generated**: `internal/db/` - Auto-generated Go code from SQLC
- **Config**: `sqlc.yaml` - Comprehensive SQLC configuration with custom type mappings

### Functional Programming with samber/lo
The codebase heavily uses functional programming patterns:
- `lo.Map()` - Transform slices
- `lo.Filter()` - Filter collections  
- `lo.Reduce()` - Aggregate data
- See `internal/domain/services/user_service.go` for extensive examples

### Error Handling Strategy
- **Domain Errors**: `internal/domain/errors/` - Typed error system
- **Result Pattern**: `internal/domain/shared/result.go` - Functional error handling
- **HTTP Responses**: `internal/application/http/response_helpers.go` - Standardized API responses

### Testing Architecture
- **Suite Pattern**: Uses Ginkgo/Gomega BDD testing framework
- **Test Helpers**: `internal/testhelpers/` - Comprehensive testing utilities
- **Builders**: `internal/testhelpers/domain/entities/builders.go` - Test data builders
- **Parallel Tests**: Configured for concurrent test execution
