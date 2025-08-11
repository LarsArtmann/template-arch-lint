# 🚀 Quick Start: Template-Justfile Integration

## TL;DR - Copy & Paste Workflow

### 1. Add Template to Your Project
```bash
# In your Go project root
git subtree add --prefix=arch-lint-tools https://github.com/your-org/template-arch-lint master --squash
```

### 2. Create Your Justfile
```just
# Set the correct path for arch-lint tools
export ARCH_LINT_ROOT := "arch-lint-tools"

# Import all modules
import "arch-lint-tools/justfile-modules/setup.just"
import "arch-lint-tools/justfile-modules/arch-lint.just"
import "arch-lint-tools/justfile-modules/quality.just"

# Default recipe
default: help

# Show help
help:
    @echo "🔧 Project Linting Commands"
    @echo ""
    @just --list --unsorted
```

### 3. Run Complete Setup
```bash
# Install tools and deploy configurations
just setup-project clean standard

# Or for stricter settings
just setup-project clean strict
```

### 4. Start Linting
```bash
# Run all linting checks
just lint-quality           # Code quality
just lint-architecture      # Architecture validation

# Auto-fix issues
just fix-quality

# Generate reports
just report-quality
just report-architecture
```

## Available Templates

### Architecture Patterns
- `clean` - Clean architecture with domain/application/infrastructure layers
- `hexagonal` - Hexagonal architecture pattern

### Quality Levels  
- `standard` - Essential linters for most projects
- `strict` - Maximum enforcement with zero tolerance

## Common Commands

```bash
# List available templates
just list-templates

# Verify setup
just verify-setup

# Install tools only
just install-all-tools

# Update from template
git subtree pull --prefix=arch-lint-tools <repo-url> master --squash
```

## Requirements

- Go 1.23+
- Just command runner
- Git (for subtree operations)

## That's It! 

Your project now has enterprise-grade linting with modular justfile commands.