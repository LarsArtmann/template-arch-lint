# 📚 Template Architecture Lint - Usage Guide

This guide provides comprehensive instructions for using the Template Architecture Lint project, whether you're adopting the linting configurations, learning from the architecture, or extending the codebase.

## 🎯 Quick Start

### Prerequisites

**Required Tools:**

- Go 1.21+ (tested with 1.21-1.24)
- Just command runner (`brew install just` or `cargo install just`)
- Git with recent version

**Recommended Tools:**

- Docker (for containerized development)
- VS Code with Go extension
- golangci-lint v2.0+
- go-arch-lint

### Installation

1. **Clone the repository:**

   ```bash
   git clone https://github.com/LarsArtmann/template-arch-lint.git
   cd template-arch-lint
   ```

2. **Install dependencies and tools:**

   ```bash
   just install
   ```

3. **Build the project:**

   ```bash
   just build
   ```

4. **Verify installation:**

   ```bash
   just test
   just lint
   ```

5. **Start the application:**
   ```bash
   just run
   # Visit http://localhost:8080
   ```

## 🏗️ Using This Template

### Option 1: Copy Linting Configurations (Recommended)

This is the primary use case - copy the linting configurations to your existing projects:

```bash
# Copy essential configuration files to your project
cp .go-arch-lint.yml /path/to/your/project/
cp .golangci.yml /path/to/your/project/
cp justfile /path/to/your/project/
cp sqlc.yaml /path/to/your/project/  # if using sqlc
```

**Key files to copy:**

- `.go-arch-lint.yml` - Architecture boundary enforcement
- `.golangci.yml` - 32+ linters with maximum strictness
- `justfile` - Development workflow automation
- `.pre-commit-config.yaml` - Quality gates
- `.github/workflows/` - CI/CD pipeline templates

### Option 2: Use as Project Template

Fork or use as template for new Go projects:

1. **Use GitHub's template feature** (recommended)
2. **Or fork the repository:**

   ```bash
   git clone https://github.com/LarsArtmann/template-arch-lint.git my-new-project
   cd my-new-project
   rm -rf .git
   git init
   ```

3. **Customize for your project:**

   ```bash
   # Update module name in go.mod
   go mod edit -module github.com/yourname/your-project

   # Find and replace import paths
   find . -name "*.go" -exec sed -i 's|github.com/LarsArtmann/template-arch-lint|github.com/yourname/your-project|g' {} \;
   ```

## ⚡ Essential Commands

### Development Workflow

```bash
# 🏗️ Build and Development
just build              # Build application
just run                # Start development server
just dev                # Start with auto-reload
just clean              # Clean build artifacts

# 🧪 Testing
just test               # Run all tests
just test-unit          # Unit tests only
just test-integration   # Integration tests only
just test-race          # Race condition detection
just coverage 80        # Coverage analysis with 80% threshold

# 🔍 Code Quality
just lint               # Complete linting suite (architecture + code + files)
just lint-arch          # Architecture boundaries only
just lint-code          # Code quality only
just lint-security      # Security-focused linting
just fix                # Auto-fix formatting issues

# ⚡ Performance
just bench              # Run all benchmarks
just bench-cpu          # CPU-focused benchmarks
just bench-memory       # Memory allocation benchmarks
just profile-cpu        # CPU profiling (requires running server)
```

### Architecture Enforcement

The project enforces Clean Architecture boundaries:

```bash
# ✅ Architecture validation
just lint-arch

# Example violations you'll catch:
# ❌ domain-entities cannot depend on infrastructure
# ❌ domain-services must not import gin
# ❌ interface{} usage detected (use concrete types)
```

### Database Operations

```bash
# 🗄️ Database (SQLite with SQLC)
just db-migrate         # Run database migrations
just db-reset           # Reset database
just db-seed            # Populate with test data
just sqlc-generate      # Generate type-safe SQL code
```

## 📊 Understanding the Linting

### Architecture Linting (`.go-arch-lint.yml`)

Enforces Clean Architecture + Domain-Driven Design:

```yaml
# Layer dependency rules (strictly enforced)
infrastructure:
  - can_depend_on: [domain, application]
application:
  - can_depend_on: [domain]
domain:
  - cannot_depend_on: [infrastructure, application]
```

**Key principles enforced:**

- Domain layer is pure (no external dependencies)
- Infrastructure depends on domain contracts
- Value objects are immutable
- Repository interfaces in domain layer

### Code Quality Linting (`.golangci.yml`)

32+ active linters with maximum strictness:

```bash
# 🚨 Violations you'll catch:
# ❌ Function too long (max 50 lines)
# ❌ Cyclomatic complexity too high (max 10)
# ❌ interface{} usage (type erasure)
# ❌ panic() usage (return errors instead)
# ❌ Magic numbers and strings
# ❌ Missing error handling
# ❌ Unused variables and imports
```

## 🏛️ Architecture Overview

### Layer Structure

```
cmd/server/                    # Application entry point
├── main.go                   # Server bootstrap

internal/
├── application/              # Application layer
│   ├── handlers/            # HTTP handlers (Gin + HTMX)
│   └── middleware/          # HTTP middleware
│
├── domain/                   # Domain layer (pure business logic)
│   ├── entities/            # Business entities
│   ├── services/            # Domain services
│   ├── values/              # Value objects
│   └── repositories/        # Repository interfaces
│
├── infrastructure/           # Infrastructure layer
│   └── persistence/         # Database implementations
│
├── config/                   # Configuration
└── container/               # Dependency injection
```

### Technology Stack

**Backend:**

- **Go 1.21+** - Core language
- **Gin** - HTTP web framework
- **templ** - Type-safe HTML templates
- **HTMX** - Frontend interactivity
- **SQLite** - Database
- **SQLC** - Type-safe SQL generation

**Quality & Tooling:**

- **golangci-lint v2** - Code quality (32+ linters)
- **go-arch-lint** - Architecture boundaries
- **Ginkgo/Gomega** - BDD testing
- **pprof** - Performance profiling
- **GitHub Actions** - CI/CD

## 🚀 CI/CD Pipeline

### GitHub Actions Workflows

The project includes comprehensive CI/CD:

```bash
# 🔄 Automated workflows
.github/workflows/
├── lint.yml          # Linting (architecture + code quality)
├── test.yml          # Testing (unit + integration + coverage)
├── benchmark.yml     # Performance benchmarking
├── ci.yml           # Complete CI pipeline
└── status.yml       # Status checks
```

**Features:**

- ✅ Multi-version Go testing (1.21-1.24)
- ✅ Architecture boundary validation
- ✅ Code quality enforcement
- ✅ Race condition detection
- ✅ Security scanning (gosec, trivy)
- ✅ Performance regression detection
- ✅ Coverage tracking with Codecov
- ✅ Automated dependency updates

### Quality Gates

**Pre-commit hooks:**

```bash
# Install pre-commit hooks
pre-commit install

# Runs automatically on git commit:
# ✅ golangci-lint
# ✅ go-arch-lint
# ✅ go fmt/goimports
# ✅ templ formatting
```

## 📊 Performance Monitoring

### Built-in Profiling

Start the server and access profiling endpoints:

```bash
# Start server with profiling enabled
APP_ENVIRONMENT=development just run

# Available endpoints (development only):
curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof
curl http://localhost:8080/debug/pprof/heap -o heap.prof
curl http://localhost:8080/performance/stats  # Runtime statistics
```

### Benchmarking

```bash
# 📊 Comprehensive benchmarks
just bench              # All benchmarks
just bench-baseline     # Create baseline
just bench-compare      # Compare with baseline
just bench-report       # Generate analysis report

# 🎯 Specialized benchmarks
just bench-cpu          # CPU performance
just bench-memory       # Memory allocation
just bench-stress       # Stress testing
just bench-profile      # With pprof integration
```

## 🔧 Configuration

### Environment Configuration

```bash
# Development
export APP_ENVIRONMENT=development
export APP_SERVER_PORT=8080
export APP_LOGGING_LEVEL=debug

# Production
export APP_ENVIRONMENT=production
export APP_SERVER_PORT=80
export APP_LOGGING_LEVEL=info
```

### Database Configuration

```bash
# SQLite (default)
export DATABASE_DRIVER=sqlite3
export DATABASE_DSN=./app.db

# Connection pooling
export DATABASE_MAX_OPEN_CONNS=25
export DATABASE_MAX_IDLE_CONNS=25
```

## 🎨 Frontend (HTMX + templ)

### Template Development

```bash
# 🖼️ Template workflow
just templ              # Generate templ templates
just dev                # Auto-reload on template changes

# Template structure:
templates/
├── components/         # Reusable components
├── pages/             # Full page templates
└── layouts/           # Layout templates
```

### HTMX Integration

The project demonstrates modern HTMX patterns:

```html
<!-- Dynamic user list with search -->
<div
  hx-get="/users/search"
  hx-trigger="keyup changed delay:500ms"
  hx-target="#user-list"
  hx-swap="outerHTML"
>
  <input type="search" name="query" placeholder="Search users..." />
</div>

<!-- Inline editing -->
<div hx-get="/users/123/edit-inline" hx-trigger="click" hx-swap="outerHTML">Edit User</div>
```

## 🚨 Common Issues & Solutions

### Linting Violations

**Architecture violations:**

```bash
# ❌ Error: domain cannot depend on infrastructure
# Solution: Use dependency inversion - define interfaces in domain

# ❌ Error: interface{} usage detected
# Solution: Use concrete types or generics

# ❌ Error: Function too long (>50 lines)
# Solution: Extract smaller functions or use just lint-fix
```

**Performance issues:**

```bash
# 📊 Memory leaks detected
just profile-heap       # Capture heap profile
just bench-memory       # Check allocation patterns

# 🖥️ CPU bottlenecks
just profile-cpu        # Capture CPU profile
just bench-cpu          # Benchmark CPU-intensive operations
```

### Development Issues

**Build problems:**

```bash
# Dependency issues
just clean && go mod download && just build

# Template generation issues
templ generate --watch

# Database issues
just db-reset
```

## 📚 Learning Resources

### Understanding the Architecture

1. **Start with domain layer** (`internal/domain/`) - Pure business logic
2. **Study value objects** - Immutable types with validation
3. **Review repository pattern** - Interface in domain, implementation in infrastructure
4. **Examine handlers** - Clean separation of HTTP concerns

### Linting Configuration Deep-dive

```bash
# 🔍 Understand linting rules
just lint-arch --explain         # Architecture rule explanations
golangci-lint linters            # Available linters
just lint-strict                 # Maximum strictness mode
```

### Testing Patterns

The project demonstrates:

- **BDD testing** with Ginkgo/Gomega
- **Table-driven tests** for value objects
- **Integration tests** with real SQLite
- **Benchmark tests** with memory allocation tracking
- **Architecture tests** validating layer boundaries

## 🎯 Next Steps

### For Learning

1. Run `just lint` and fix violations to understand code quality rules
2. Study the Clean Architecture implementation
3. Experiment with HTMX patterns in templates
4. Use performance profiling to optimize your code

### For Production Use

1. Copy linting configurations to your existing projects
2. Adapt the CI/CD pipeline to your needs
3. Customize the architecture layers for your domain
4. Set up monitoring and alerting based on the performance endpoints

### For Contribution

1. Check `CONTRIBUTING.md` for development guidelines
2. Run the full test suite: `just ci`
3. Follow the established patterns and linting rules
4. Add benchmarks for performance-critical code

---

## 📖 Additional Documentation

- 📊 **[Profiling Guide](./PROFILING.md)** - Detailed performance analysis
- 🏗️ **[Architecture Tests](../architecture_test.go)** - Boundary validation examples
- ⚙️ **[Configuration](../internal/config/)** - Application configuration
- 🔧 **[Justfile](../justfile)** - All available commands
- 🚀 **[CI/CD](../.github/workflows/)** - GitHub Actions workflows

For questions or issues, check the [GitHub repository](https://github.com/LarsArtmann/template-arch-lint) or create an issue.
