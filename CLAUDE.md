# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **template-arch-lint** - a comprehensive Go architecture and code quality enforcement template demonstrating enterprise-grade patterns. The project serves as both a working application and a reference template for Go projects requiring strict architectural boundaries and code quality standards.

**Core Purpose**: Demonstrate and enforce Clean Architecture, Domain-Driven Design (DDD), and maximum code quality through automated linting and architectural validation.

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
# Run single test file
go test ./internal/domain/services/ -v

# Run tests with race detection
go test ./... -v -race

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

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

## Architecture Overview

This project implements **Clean Architecture with Domain-Driven Design** using the following structure:

### Layer Dependencies (Strictly Enforced)
```
Infrastructure → Application → Domain
```

### Key Architectural Components

**Domain Layer** (`internal/domain/`):
- **Entities**: Core business objects with behavior (`entities/user.go`)
- **Value Objects**: Type-safe primitives (`values/user_id.go`, `values/email.go`)
- **Repositories**: Abstract data access interfaces (`repositories/user_repository.go`)
- **Services**: Complex business logic coordination (`services/user_service.go`)
- **Errors**: Typed domain errors (`errors/domain_errors.go`)

**Application Layer** (`internal/application/`):
- **Handlers**: HTTP request/response handling with gin + templ
- **Middleware**: Cross-cutting concerns (logging, metrics, SLA tracking)

**Infrastructure Layer** (`internal/infrastructure/`):
- **Persistence**: SQLC-generated database implementations
- **Repositories**: Concrete repository implementations

**Database Layer** (`internal/db/`):
- **SQLC Generated**: Type-safe database operations
- **Schema**: SQL DDL in `sql/schema/`
- **Queries**: SQL DML in `sql/queries/`

### Technology Stack Integration

**HTTP & Templates**:
- **gin-gonic/gin**: HTTP routing and middleware
- **a-h/templ**: Type-safe HTML templates (generate with `templ generate`)
- **HTMX**: Progressive enhancement for dynamic UI

**Data & Configuration**:
- **sqlc**: Type-safe SQL code generation
- **SQLite**: Database (via `github.com/mattn/go-sqlite3`)
- **spf13/viper**: Configuration management with hot-reloading

**Functional Programming**:
- **samber/lo**: Functional utilities (Filter, Map, Reduce patterns extensively used)
- **samber/mo**: Monads and functional abstractions (Result, Option, Either patterns)
- **samber/do**: Dependency injection container

**Testing**:
- **Ginkgo/Gomega**: BDD-style testing framework
- **testify**: Assertions and mocking

**Observability**:
- **OpenTelemetry**: Distributed tracing and metrics
- **Prometheus**: Metrics collection (port 2112)

## Critical Configuration Files

### Architecture Enforcement
- **`.go-arch-lint.yml`**: Defines component boundaries and dependency rules
- **`.golangci.yml`**: Comprehensive code quality rules with maximum strictness
- **`justfile`**: Complete build and linting automation

### Database & Code Generation
- **`sqlc.yaml`**: Type-safe SQL generation with custom UserID value object mapping
- **`go.mod`**: Go 1.24+ with key dependencies

### Template System
- **`web/templates/`**: templ template components
  - `layouts/`: Base page layouts
  - `pages/`: Full page templates  
  - `components/`: Reusable UI components

## Development Patterns

### Domain Modeling
- Value objects enforce invariants at compile time
- Entities contain business logic, not just data
- Repository interfaces defined in domain, implemented in infrastructure
- Use samber/lo for functional operations on collections

### Error Handling
- Typed domain errors using custom error types
- Railway Oriented Programming with Result/Option patterns
- No panic() usage - return errors explicitly

### Testing Strategy
- BDD tests using Ginkgo/Gomega for business logic
- Unit tests for value objects and entities
- Integration tests for repository implementations
- Table-driven tests for complex scenarios

### SQLC Integration
- Custom type mapping for UserID value objects
- Strict query validation rules (no SELECT *, require WHERE for DELETE)
- JSON tags for API serialization
- Prepared statements for performance

## Architecture Violations to Avoid

The `.go-arch-lint.yml` strictly enforces:
- Domain layer cannot import infrastructure packages
- Application handlers cannot directly import database packages
- Infrastructure implements domain interfaces, never the reverse
- Value objects and entities remain pure (minimal external dependencies)

## Observability & Monitoring

- **Metrics**: Prometheus metrics on port 2112 (`/metrics`)
- **Health Checks**: `/health/live` and `/health/ready` endpoints
- **Profiling**: pprof endpoints available in development
- **Tracing**: OpenTelemetry integration for distributed tracing

## Container & Deployment

- **Docker**: Multi-stage build with distroless runtime (~20MB image)
- **Kubernetes**: Complete manifests in `k8s/` directory
- **CI/CD**: GitHub Actions with security scanning and quality gates

This codebase prioritizes architectural correctness, type safety, and functional programming patterns while maintaining practical simplicity for real-world usage.