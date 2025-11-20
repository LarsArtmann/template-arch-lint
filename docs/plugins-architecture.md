# Go Plugin Architecture with Independent go-arch-lint

This document explains the plugin architecture exclusion pattern in `.go-arch-lint.yml` and demonstrates how each plugin should have its own architectural validation.

## 🏗️ Plugin Architecture Pattern

### Why Exclude Plugins from Parent Architecture?

```
parent-project/
├── .go-arch-lint.yml          # Main project architecture rules
├── go.mod                     # Main project dependencies
├── internal/                  # Clean Architecture layers
└── pkg/linter-plugins/        # ✅ Public API - plugin extensibility
    ├── auth-plugin/           # ✅ Independent Go module
    │   ├── go.mod            # Plugin-specific dependencies
    │   ├── .go-arch-lint.yml  # Plugin-specific architecture rules
    │   └── internal/          # Plugin's own architecture
    ├── logging-plugin/        # ✅ Independent Go module
    │   ├── go.mod
    │   ├── .go-arch-lint.yml
    │   └── internal/
    └── metrics-plugin/        # ✅ Independent Go module
        ├── go.mod
        ├── .go-arch-lint.yml
        └── internal/
```

### Benefits of Independent Plugin Architecture

1. **Autonomous Design**: Each plugin can optimize its own architecture for its specific domain
2. **Dependency Isolation**: Plugin dependencies don't pollute the main project
3. **Tailored Rules**: Different plugins have different architectural needs
4. **Independent Development**: Plugins can be developed, tested, and released separately
5. **Clean Boundaries**: Prevents architectural "leakage" between plugins and main project

## 📋 Plugin Examples and Architectures

### Example 1: Auth Plugin
```yaml
# plugins/auth-plugin/.go-arch-lint.yml
version: 3
components:
  auth-domain:
    in: internal/domain/**
  auth-handlers:
    in: internal/handlers/**
  auth-providers:
    in: internal/providers/**

deps:
  auth-domain:
    anyVendorDeps: true
    mayDependOn: []
  
  auth-handlers:
    mayDependOn:
      - auth-domain
      - auth-providers
```

### Example 2: Logging Plugin
```yaml
# plugins/logging-plugin/.go-arch-lint.yml
version: 3
components:
  logging-core:
    in: internal/core/**
  logging-adapters:
    in: internal/adapters/**

deps:
  logging-core:
    anyVendorDeps: true
    mayDependOn: []
  
  logging-adapters:
    mayDependOn:
      - logging-core
```

### Example 3: Database Plugin
```yaml
# plugins/database-plugin/.go-arch-lint.yml
version: 3
components:
  db-repositories:
    in: internal/repositories/**
  db-migrations:
    in: internal/migrations/**

deps:
  db-repositories:
    anyVendorDeps: true
    mayDependOn:
      - db-migrations
```

## 🔧 Implementation Guidelines

### 1. Plugin Structure
```
plugins/my-plugin/
├── go.mod                    # Plugin module definition
├── .go-arch-lint.yml         # Plugin-specific architecture
├── README.md                 # Plugin documentation
├── internal/                 # Plugin architecture
│   ├── domain/              # Plugin business logic
│   ├── application/         # Plugin use cases
│   └── infrastructure/      # Plugin external deps
└── api/                      # Plugin public interfaces
```

### 2. Plugin go.mod Example
```go
module github.com/yourorg/project/plugins/auth-plugin

go 1.21

require (
    github.com/golang-jwt/jwt/v5 v5.0.0
    golang.org/x/crypto v0.12.0
)
```

### 3. Plugin Architecture Rules
Each plugin should:
- Define its own Clean Architecture layers
- Use appropriate dependencies for its domain
- Have independent architectural validation
- Follow consistent naming conventions

## 🚀 Plugin Integration Patterns

### 1. Interface-Based Integration
```go
// In main project
type AuthProvider interface {
    Authenticate(ctx context.Context, token string) (*User, error)
}

// Plugin registers implementation
func RegisterAuthPlugin(provider AuthProvider) {
    authServiceProvider = provider
}
```

### 2. Plugin Discovery
```go
// Main project discovers and loads plugins
func LoadPlugins() error {
    plugins := []string{
        "auth-plugin",
        "logging-plugin", 
        "metrics-plugin",
    }
    
    for _, plugin := range plugins {
        if err := loadPlugin(plugin); err != nil {
            return fmt.Errorf("failed to load %s: %w", plugin, err)
        }
    }
    return nil
}
```

## 📊 Architectural Validation Strategy

### Main Project Validation
```bash
# Main project - plugins excluded
cd /project
just lint-arch    # Validates main architecture only
```

### Individual Plugin Validation
```bash
# Each plugin validates independently
cd plugins/auth-plugin
just lint-arch    # Validates auth plugin architecture

cd plugins/logging-plugin  
just lint-arch    # Validates logging plugin architecture
```

### CI/CD Pipeline
```yaml
# .github/workflows/architecture.yml
jobs:
  architecture-main:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Validate Main Architecture
        run: just lint-arch
  
  architecture-plugins:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        plugin: [auth-plugin, logging-plugin, metrics-plugin]
    steps:
      - uses: actions/checkout@v3
      - name: Validate Plugin Architecture
        working-directory: plugins/${{ matrix.plugin }}
        run: just lint-arch
```

## 🎯 Best Practices

### 1. Clear Plugin Boundaries
- Each plugin should solve one specific problem
- Avoid circular dependencies between plugins
- Use well-defined interfaces for communication

### 2. Consistent Architecture
- Follow Clean Architecture in each plugin
- Use similar naming conventions
- Maintain consistent documentation

### 3. Independent Testing
- Each plugin has its own test suite
- Plugin tests validate plugin-specific architecture
- Integration tests verify plugin interactions

### 4. Version Management
- Each plugin has independent versioning
- Maintain compatibility matrices
- Document breaking changes clearly

## 🚨 Common Pitfalls to Avoid

### 1. Plugin Proliferation
- Don't create plugins for simple utilities
- Consider if functionality belongs in main project
- Balance modularity with complexity

### 2. Inconsistent Architectures
- Don't mix architectural patterns between plugins
- Maintain consistency in Clean Architecture implementation
- Review and standardize plugin architectures regularly

### 3. Dependency Hell
- Minimize plugin interdependencies
- Use interface contracts to reduce coupling
- Consider dependency injection for plugin coordination

This pattern enables scalable, maintainable plugin architectures while preserving strict architectural validation both at the project and plugin levels.