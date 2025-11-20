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

### Core Development Commands
```bash
# Installation & Setup
just install              # Install ALL linting tools (golangci-lint, go-arch-lint, etc.)
just install-hooks        # Install git pre-commit hooks (fast checks only)
just install-hooks-full   # Install comprehensive pre-commit hooks (includes architecture)

# Primary Workflow Commands  
just lint                 # Run ALL linters (architecture, code, security, dependencies)
just fix                  # Auto-fix formatting issues and simple violations
just test                 # Run all tests with coverage report
just build                # Build the application
just run                  # Start HTTP server on port 8080
just dev                  # Development mode with auto-reload
just ci                   # Complete CI/CD pipeline simulation

# Security & Vulnerability Scanning
just security-audit       # Complete security audit (all tools)
just lint-vulns          # Run govulncheck for CVE scanning
just lint-nilaway        # Uber's nil panic prevention (80% reduction)
just lint-licenses       # License compliance scanning (manual audit)
just lint-deps-advanced  # Advanced dependency vulnerability analysis
```

### Specialized Linting Commands
```bash
# Architecture & Design
just lint-arch           # Architecture boundary validation only
just lint-cmd-single     # CMD single main.go enforcement only
just graph               # Generate flow architecture graph (SVG)
just graph-di           # Generate dependency injection graph
just graph-vendor       # Generate graph with vendor dependencies
just graph-all          # Generate ALL architecture graphs
just graph-component <name> # Generate focused component graph
just graph-list-components   # List available components

# Code Quality
just lint-code           # Code quality linting (40+ linters)
just lint-strict         # Maximum strictness mode
just lint-files          # Filename compliance validation
just lint-cycles         # Import cycle detection
just lint-goroutines     # Goroutine leak detection (Uber's goleak)

# Formatting & Generation
just format              # gofumpt + goimports formatting
just templ               # Generate templ templates
templ generate           # Force regenerate templates
sqlc generate            # Generate type-safe SQL code
```

### Testing Commands
```bash
# Single test or package
go test ./internal/domain/services/ -v
go test ./internal/domain/entities/ -v

# Specific test function
go test ./internal/domain/services/ -v -run TestUserService_CreateUser

# Race detection
go test ./... -v -race

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Benchmarks
go test ./internal/domain/services/ -bench=.
```

## Critical Linting Configuration

### Architecture Enforcement (`.go-arch-lint.yml`)
- **Domain purity**: Domain layer cannot import infrastructure or application layers
- **Dependency inversion**: Infrastructure depends on domain interfaces
- **Clean architecture flow**: Infrastructure → Application → Domain

### CMD Single Main Enforcement (`lint-cmd-single`)
- **Single Entry Point**: Enforces exactly one `main.go` file in `cmd/` directory
- **Clean Architecture**: Prevents command proliferation and maintains single responsibility
- **Actionable Errors**: Provides specific consolidation suggestions when violations are found
- **Automated Validation**: Integrated into `just lint` pipeline for continuous enforcement

**Examples:**
```bash
just lint-cmd-single      # Check cmd/ single main constraint only
just lint                 # Includes cmd/ validation in full linting pipeline
```

**Violation Examples:**
- ❌ Multiple main files: `cmd/server/main.go` + `cmd/cli/main.go`
- ❌ No main files: Empty `cmd/` directory
- ✅ Single main file: `cmd/server/main.go` only

**Consolidation Suggestions:**
- Use CLI frameworks like [Cobra](https://pkg.go.dev/github.com/spf13/cobra) for subcommands
- Create single main with multiple modes: `server start`, `server migrate`
- Move additional tools to separate packages/repositories

**Future Enhancement**: This constraint will be available as a native golangci-lint plugin, providing deeper IDE integration and more sophisticated analysis. See `docs/planning/` for the plugin roadmap.

### Code Quality Enforcement (`.golangci.yml`)  
- **40+ linters enabled** including cutting-edge tools:
  - `nilaway`: Uber's nil panic prevention (2024-2025)
  - `godox`: TODO/FIXME/HACK detection
  - `forbidigo`: Bans `interface{}`, `any`, `panic()`, and `fmt.Print*`
  - `gomnd`: Magic number detection
  - `maligned`: Struct alignment optimization
  - `gochecknoinits`: No init functions
  - `gochecknoglobals`: No global variables
- **Function limits**: Max 50 lines, complexity 10
- **File limits**: Max 400 lines per file
- **Line length**: Max 120 characters

### Security Scanning (Built-in Tools)
- **10 custom security rules** for Go-specific vulnerabilities:
  - Hardcoded secrets detection
  - SQL injection prevention
  - Command injection risks
  - Path traversal vulnerabilities
  - Weak cryptography usage
  - Insecure TLS configurations

## Key Libraries and Patterns

### Core Dependencies
- **gin**: HTTP web framework
- **templ**: Type-safe HTML templates
- **HTMX**: Progressive enhancement for web UI
- **sqlc**: Type-safe SQL code generation
- **samber/lo**: Functional programming utilities (Map, Filter, Reduce)
- **samber/do**: Dependency injection
- **viper**: Configuration management
- **Ginkgo/Gomega**: BDD testing framework

### Important Implementation Patterns

#### Value Objects (`internal/domain/values/`)
- Enforce validation and type safety at domain level
- Examples: `Email`, `UserName`, `UserID` with business rules

#### Repository Pattern
- **Interfaces** in `internal/domain/repositories/`
- **Implementations** in `internal/infrastructure/persistence/`
- **In-memory versions** for testing

#### Result Pattern (`internal/domain/shared/result.go`)
- Functional error handling without exceptions
- Chain operations with success/failure paths

### Architecture Graph Organization

**📁 Graph Location: `docs/graphs/` (not polluting project root!)**

```
docs/graphs/
├── README.md                     # This documentation
├── index.md                      # Auto-generated index of all graphs
├── flow/                         # Flow graphs (execution flow)
│   └── architecture-flow.svg        # Main flow graph
├── dependency-injection/           # DI graphs (component dependencies)
│   └── architecture-di.svg         # Dependencies graph
├── vendor/                       # Vendor-inclusive graphs
│   └── architecture-with-vendors.svg # Including external dependencies
└── focused/                      # Component-focused graphs
    ├── domain-focused.svg           # Domain layer only
    ├── application-focused.svg      # Application layer only
    ├── infrastructure-focused.svg   # Infrastructure layer only
    └── cmd-focused.svg            # Command layer only
```

**Graph Types Explained:**
- **Flow graphs** (`just graph`): Show execution flow (reverse dependency injection)
- **DI graphs** (`just graph-di`): Show direct component dependencies
- **Vendor graphs** (`just graph-vendor`): Include external library dependencies
- **Focused graphs** (`just graph-component <name>`): Single component and its deps

**Usage Examples:**
```bash
# Generate all graphs (recommended for documentation)
just graph-all

# Generate specific graph types
just graph-di          # Dependency injection view
just graph-vendor      # Including external deps
just graph-component domain  # Focus on domain layer

# View organized graphs
open docs/graphs/index.md  # See all available graphs
```

#### Functional Programming with samber/lo
- Heavy use of `lo.Map()`, `lo.Filter()`, `lo.Reduce()`
- See `internal/domain/services/user_service.go` for examples

## Common Development Workflows

### Before Committing Code
```bash
just lint        # Run all quality checks
just fix         # Auto-fix formatting
just test        # Ensure tests pass
```

### Adding New Features
```bash
just lint-arch   # Verify architecture compliance
just lint-code   # Check code quality
just test        # Test your changes
```

### Security Review
```bash
just security-audit  # Complete security scan
cat gosec-report.json    # Review security findings
```

## Architecture Violations You'll Encounter

Common violations and their meanings:
- `domain-entities cannot depend on infrastructure` - Keep domain pure
- `🚨 BANNED: interface{} erases type safety` - Use specific types
- `🚨 BANNED: panic() crashes programs` - Return errors instead
- `Function too long (max 50 lines)` - Split into smaller functions
- `Cyclomatic complexity too high (max 10)` - Simplify logic

## Important Configuration Files

- **`.go-arch-lint.yml`**: Architecture boundary rules
- **`.golangci.yml`**: 40+ linters configuration
- **Built-in security**: gosec + govulncheck + NilAway (no external config needed)
- **`justfile`**: Task automation (30+ commands)
- **`sqlc.yaml`**: Type-safe SQL generation
- **`.pre-commit-config.yaml`**: Git hook configuration

## Database Setup

- **SQLite** for development (`./app.db`)
- **In-memory** for testing
- **SQLC** for type-safe queries
- Schema in `sql/schema/`
- Queries in `sql/queries/`

## Project-Specific Notes

### SQLC Integration
- Always run `sqlc generate` after modifying SQL files
- Generated code goes to `internal/db/`
- Custom type mappings configured in `sqlc.yaml`

### Templ Templates
- Run `just templ` or `templ generate` after template changes
- Templates in `web/templates/`
- Type-safe HTML generation

### Testing Strategy
- BDD tests with Ginkgo/Gomega
- Test helpers in `internal/testhelpers/`
- Builder pattern for test data
- Parallel test execution enabled

### Error Handling
- Domain errors in `internal/domain/errors/`
- Result pattern for functional error handling
- Standardized HTTP responses in `internal/application/http/`

## Important Implementation Guidelines

### When Adding New Code
1. Respect layer boundaries (domain must stay pure)
2. Use value objects for domain primitives
3. Prefer functional programming patterns with samber/lo
4. Write BDD-style tests with Ginkgo
5. Run `just lint` before committing

### When Modifying Architecture
1. Update `.go-arch-lint.yml` for new components
2. Regenerate architecture graph with `just graph`
3. Ensure no circular dependencies
4. Maintain dependency inversion principle

### When Working with Database
1. Write SQL in `sql/queries/`
2. Run `sqlc generate` to create type-safe code
3. Implement repository interfaces from domain layer
4. Use in-memory repositories for testing

## Quick Troubleshooting

- **"Tool not found"**: Run `just install`
- **Architecture violations**: Check dependency direction (Infrastructure → Application → Domain)
- **Too many linting errors**: Start with `just fix`, then address remaining issues
- **Test failures**: Check for goroutine leaks with `just lint-goroutines`
- **Performance issues**: Run linters individually instead of `just lint`