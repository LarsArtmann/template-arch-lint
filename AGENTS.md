# AGENTS.md

This file provides guidance to AI coding assistants working in this repository.

## 🎯 What This Project IS and IS NOT

### ✅ **What This Project IS:**

- **Go Linting Template**: Demonstrates enterprise-grade architecture and code quality enforcement
- **Reference Implementation**: Shows Clean Architecture + DDD patterns in Go
- **Configuration Library**: Provides `.go-arch-lint.yml` and a custom golangci-lint plugin for copy/paste use; `.golangci.yml` is generated and maintained by [golangci-lint-auto-configure](https://github.com/LarsArtmann/golangci-lint-auto-configure)
- **Simple HTTP Demo**: Basic net/http server with in-memory user CRUD and sqlc/SQLite scaffolding
- **Educational Resource**: Learn proper Go architecture boundaries and functional programming patterns

### ❌ **What This Project IS NOT:**

- **Production Application**: Not meant for real business use - it's a template/demo
- **Framework or Library**: Not installable via `go get` - copy configurations instead
- **Enterprise Platform**: Despite having monitoring/Docker/K8s - these are demos of over-engineering
- **Complex Business Domain**: User CRUD is intentionally simple to focus on architecture

### 🎯 **Core Purpose:**

Copy the architecture configuration (`.go-arch-lint.yml`) to your real projects and generate `.golangci.yml` with golangci-lint-auto-configure to enforce architectural boundaries and code quality. The Go code demonstrates how to structure projects following these rules.

## 🏗️ High-Level Architecture Understanding

### Layer Structure (Dependency Flow: Infrastructure → Application → Domain)

```
cmd/                    # Single entry point (main.go, enforced by the custom plugin)

internal/application/   # HTTP layer
└── handlers/           # HTTP request handlers (user_handler.go, errorhandler.go)

internal/domain/        # Pure business logic (NO external dependencies)
├── entities/           # Business entities (user.go)
├── services/           # Domain services (user_service.go, user_query_service.go)
├── repositories/       # Repository interfaces
├── values/             # Value objects (email.go, username.go, port.go, log_level.go)
└── ids/                # Branded IDs (UserID via go-branded-id)

internal/infrastructure/ # External concerns
├── database.go         # SQLite connection scaffolding
└── db/                 # SQLC-generated type-safe SQL code

internal/config/         # Viper-based configuration
internal/testhelpers/    # Test builders and helpers
sql/sqlite/
├── schema/             # Database schema files
└── queries/            # SQL query files for SQLC
```

### Key Architectural Patterns Demonstrated

- **Clean Architecture**: Strict dependency rules enforced by go-arch-lint
- **Domain-Driven Design**: Rich domain entities with value objects
- **Functional Programming**: Heavy use of samber/lo for Map/Filter/Reduce operations
- **Result Pattern**: `mo.Result[T]` (samber/mo) for railway-oriented error handling in domain services
- **Value Objects**: Email, UserName, UserID with validation in domain/values
- **Repository Pattern**: Domain interfaces implemented by infrastructure
- **Standard library HTTP**: net/http routing via larsartmann/httputil (no framework)

## Essential Commands

### Core Development Commands

```bash
# Setup (Nix recommended)
nix develop              # Go 1.27 toolchain + lint tools
nix build                # Build demo binary → ./result/bin/template-arch-lint
# Without Nix: go.mod requires Go 1.27.1, so set GOTOOLCHAIN=auto

# Primary workflow
go-arch-lint check       # Architecture boundary validation
golangci-lint run        # Code quality (config maintained by golangci-lint-auto-configure)
go test ./... -race      # BDD tests (Ginkgo/Gomega) with race detection
GOTOOLCHAIN=auto go build ./...   # Build

# Custom plugin (cmd-single, import cycles, duplication, filenames)
golangci-lint custom     # Build custom binary from .custom-gcl.yml
./custom-golangci-lint run

# Security
govulncheck ./...                 # CVE scanning
semgrep --config .semgrep.yml    # 10 custom security rules

# Pre-commit hooks
pre-commit install       # Installs hooks from .pre-commit-config.yaml
```

### Specialized Linting Commands

```bash
# Architecture & Design
go-arch-lint check                  # Architecture boundary validation
go-arch-lint graph                  # Generate architecture graph (see --help)
scripts/check-cmd-single.sh         # CMD single main.go enforcement
scripts/compare-arch-configs.sh     # Compare strict vs default arch config

# Code Quality
golangci-lint run                                   # Code quality (100+ linters)
golangci-lint custom && ./custom-golangci-lint run  # With custom plugin

# Formatting & Generation
golangci-lint fmt        # Run configured formatters
gofumpt -w .             # Or format directly
goimports -w -local github.com/LarsArtmann/template-arch-lint .
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
- **Automated Validation**: Enforced via the custom golangci-lint plugin and `scripts/check-cmd-single.sh`

**Examples:**

```bash
scripts/check-cmd-single.sh                          # Standalone check
golangci-lint custom && ./custom-golangci-lint run   # Plugin-based check
```

**Violation Examples:**

- ❌ Multiple main files: `cmd/server/main.go` + `cmd/cli/main.go`
- ❌ No main files: Empty `cmd/` directory
- ✅ Single main file: `cmd/server/main.go` only

**Consolidation Suggestions:**

- Use CLI frameworks like [Cobra](https://pkg.go.dev/github.com/spf13/cobra) for subcommands
- Create single main with multiple modes: `server start`, `server migrate`
- Move additional tools to separate packages/repositories

**Shipped**: This constraint IS a custom golangci-lint plugin: [`pkg/linter-plugins/template-arch-lint`](pkg/linter-plugins/template-arch-lint), wired via `.custom-gcl.yml` alongside `import-cycle-detector`, `code-duplication-detector`, and `filename-validator`.

### Code Quality Enforcement (`.golangci.yml`)

- **99+ linters enabled** including cutting-edge tools:
  - `nilaway`: Uber's nil panic prevention (2024-2025)
  - `godox`: TODO/FIXME/HACK detection
  - `forbidigo`: Bans `interface{}`, `any`, `panic()`, and `fmt.Print*`
  - `gomnd`: Magic number detection
  - `maligned`: Struct alignment optimization
  - `gochecknoinits`: No init functions
  - `gochecknoglobals`: No global variables
  - `nilnil`: Prevent (nil, nil) return pattern bugs (2026)
  - `exptostd`: Modernize code by replacing golang.org/x/exp with stdlib (2026)
  - `gocheckcompilerdirectives`: Validate build tags and compiler directives (2026)
  - `iotamixing`: Ensure clean const declarations with iota (2026)
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

- **net/http + larsartmann/httputil**: Standard library HTTP server helpers (no framework)
- **charm.land/log/v2**: Structured logging
- **go-branded-id**: Branded ID types (UserID)
- **go-playground/validator/v10**: Input validation
- **sqlc**: Type-safe SQL code generation (build-time tool)
- **samber/lo**: Functional programming utilities (Map, Filter, Reduce)
- **samber/mo**: Monads incl. `mo.Result[T]` for railway-oriented error handling
- **viper**: Configuration management
- **Ginkgo/Gomega**: BDD testing framework

### Important Implementation Patterns

#### Value Objects (`internal/domain/values/`)

- Enforce validation and type safety at domain level
- Examples: `Email`, `UserName`, `UserID` with business rules

#### Repository Pattern

- **Interfaces** in `internal/domain/repositories/`
- **In-memory implementations** used by handlers and tests
- SQLC-generated persistence in `internal/infrastructure/db/`

#### Result Pattern (`mo.Result[T]`, samber/mo)

- Functional error handling without exceptions
- Chain operations with success/failure paths
- See `GetUserEmailsWithResult` in `internal/domain/services/user_service.go`

### Architecture Graph Organization

**📁 Graph Location: `docs/graphs/` (not polluting project root!)**

```
docs/graphs/
├── README.md                     # Graph documentation
├── index.md                      # Index of all graphs
├── flow/                         # Flow graphs (execution flow)
│   └── architecture-flow.svg     # Main flow graph
└── dependency-injection/         # DI graphs (component dependencies)
    └── architecture-di.svg       # Dependencies graph
```

**Usage Examples:**

```bash
go-arch-lint graph        # Regenerate (see --help for output options)

# View organized graphs
open docs/graphs/index.md  # See all available graphs
```

#### Functional Programming with samber/lo

- Heavy use of `lo.Map()`, `lo.Filter()`, `lo.Reduce()`
- See `internal/domain/services/user_service.go` for examples

## Common Development Workflows

### Before Committing Code

```bash
go-arch-lint check   # Architecture boundaries
golangci-lint run    # Code quality
go test ./... -race  # Tests
```

### Adding New Features

```bash
go-arch-lint check   # Verify architecture compliance
golangci-lint run    # Check code quality
go test ./...        # Test your changes
```

### Security Review

```bash
govulncheck ./...                 # CVE scan
semgrep --config .semgrep.yml    # Custom security rules
```

## Architecture Violations You'll Encounter

Common violations and their meanings:

- `domain-entities cannot depend on infrastructure` - Keep domain pure
- `🚨 BANNED: interface{} erases type safety` - Use specific types
- `🚨 BANNED: panic() crashes programs` - Return errors instead
- `Function too long (max 50 lines)` - Split into smaller functions
- `Cyclomatic complexity too high (max 10)` - Simplify logic

## Important Configuration Files

- **`.go-arch-lint.yml`**: Architecture boundary rules (+ `.go-arch-lint-strict.yml` strict variant)
- **`.golangci.yml`**: Code quality linters (generated/maintained by golangci-lint-auto-configure)
- **`.custom-gcl.yml`**: Custom plugin wiring (cmd-single, cycles, duplication, filenames)
- **`.semgrep.yml`**: 10 custom security rules
- **`sqlc.yaml`**: Type-safe SQL generation
- **`.pre-commit-config.yaml`**: Git hook configuration
- **`flake.nix`**: Development environment and build

## Database Setup

- **SQLite** scaffolding in `internal/infrastructure/database.go` (no SQLite driver in go.mod — add `mattn/go-sqlite3` or `modernc.org/sqlite` before opening a real DB)
- **In-memory repositories** power the demo and tests
- **SQLC** for type-safe queries
- Schema in `sql/sqlite/schema/`
- Queries in `sql/sqlite/queries/`

## Project-Specific Notes

### SQLC Integration

- Always run `sqlc generate` after modifying SQL files
- Generated code goes to `internal/infrastructure/db/`
- Custom type mappings configured in `sqlc.yaml` (e.g., `users.id` → `internal/domain/ids.UserID`)

### HTTP Handlers

- Handlers in `internal/application/handlers/` (user, query, error)
- Server wiring in `cmd/main.go` via larsartmann/httputil

### Testing Strategy

- BDD tests with Ginkgo/Gomega
- Test helpers in `internal/testhelpers/`
- Builder pattern for test data
- Parallel test execution enabled

### Error Handling

- `mo.Result[T]` (samber/mo) for railway-oriented domain error handling
- Centralized HTTP error responses in `internal/application/handlers/errorhandler.go`

## Important Implementation Guidelines

### When Adding New Code

1. Respect layer boundaries (domain must stay pure)
2. Use value objects for domain primitives
3. Prefer functional programming patterns with samber/lo
4. Write BDD-style tests with Ginkgo
5. Run `go-arch-lint check` and `golangci-lint run` before committing

### When Modifying Architecture

1. Update `.go-arch-lint.yml` for new components
2. Regenerate architecture graph with `go-arch-lint graph`
3. Ensure no circular dependencies
4. Maintain dependency inversion principle

### When Working with Database

1. Write SQL in `sql/sqlite/queries/`
2. Run `sqlc generate` to create type-safe code
3. Implement repository interfaces from domain layer
4. Use in-memory repositories for testing

## Quick Troubleshooting

- **"Tool not found"**: `nix develop` (or install go-arch-lint/golangci-lint manually)
- **Build fails with "requires go >= 1.27.1"**: run with `GOTOOLCHAIN=auto`
- **Architecture violations**: Check dependency direction (Infrastructure → Application → Domain)
- **Too many linting errors**: `golangci-lint-auto-configure configure` regenerates a clean config, then address remaining issues
- **Performance issues**: Run linters individually instead of the full config
