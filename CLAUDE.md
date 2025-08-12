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