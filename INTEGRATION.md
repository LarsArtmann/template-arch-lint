# 🚀 Template-Justfile Integration Guide

This document explains how to integrate **Template Architecture Lint** with your **template-justfile** project using justfile imports and git subtrees.

## 📦 Integration Methods

### Method 1: Git Subtree (Recommended)

Add template-arch-lint as a git subtree to your template-justfile project:

```bash
# Add as subtree
git subtree add --prefix=linting/arch-lint https://github.com/LarsArtmann/template-arch-lint.git main --squash

# Update subtree when needed
git subtree pull --prefix=linting/arch-lint https://github.com/LarsArtmann/template-arch-lint.git main --squash
```

### Method 2: Direct Clone

Clone template-arch-lint as a dependency:

```bash
# Clone to vendor directory
git clone https://github.com/LarsArtmann/template-arch-lint.git vendor/arch-lint

# Or as git submodule
git submodule add https://github.com/LarsArtmann/template-arch-lint.git linting/arch-lint
```

## 🔧 Usage in Your Justfile

### Import Modules

```bash
# Set path to arch-lint installation
ARCH_LINT_ROOT := "./linting/arch-lint"

# Import specific modules
import ARCH_LINT_ROOT + "/justfile-modules/arch-lint.just"
import ARCH_LINT_ROOT + "/justfile-modules/quality.just" 
import ARCH_LINT_ROOT + "/justfile-modules/setup.just"
```

### Use Imported Commands

```bash
# Your project justfile
default: help

# Setup commands (from setup.just)
setup-clean: 
    @just setup-project clean standard

setup-strict:
    @just setup-project clean strict

# Linting commands (from arch-lint.just and quality.just)  
lint:
    @just lint-architecture
    @just lint-quality

# Install commands (from setup.just)
install: install-all-tools
```

## 📁 Available Modules

### `arch-lint.just`
- `install-arch-tools` - Install go-arch-lint
- `lint-architecture` - Run architecture validation
- `report-architecture` - Generate architecture reports
- `verify-arch-setup` - Verify architecture setup

### `quality.just`  
- `install-quality-tools` - Install golangci-lint
- `lint-quality` - Run code quality linting
- `fix-quality` - Auto-fix quality issues
- `report-quality` - Generate quality reports
- `verify-quality-setup` - Verify quality setup

### `setup.just`
- `install-all-tools` - Install all tools
- `setup-arch <pattern>` - Setup architecture (clean/hexagonal) 
- `setup-quality <strictness>` - Setup quality (standard/strict)
- `setup-project <pattern> <strictness>` - Complete setup
- `verify-setup` - Verify complete setup
- `list-templates` - Show available templates

## 🗂️ Configuration Templates

### Architecture Patterns
- `clean` - Clean Architecture pattern
- `hexagonal` - Hexagonal Architecture (Ports & Adapters)

### Quality Levels
- `standard` - Balanced quality enforcement
- `strict` - Maximum strictness, zero tolerance

## 📋 Setup Examples

### Quick Start
```bash
# Install tools and setup Clean Architecture with standard quality
just setup-project clean standard

# Install tools and setup Hexagonal Architecture with strict quality  
just setup-project hexagonal strict
```

### Step by Step
```bash
# 1. Install tools
just install-all-tools

# 2. Setup architecture
just setup-arch clean

# 3. Setup quality
just setup-quality strict

# 4. Verify everything works
just verify-setup
```

## 🔄 Environment Variables

Customize paths and behavior with environment variables:

```bash
# Configuration file paths
export ARCH_CONFIG=".go-arch-lint.yml"
export QUALITY_CONFIG=".golangci.yml"
export REPORTS_DIR="./reports"

# Template root (for setup commands)
export ARCH_LINT_ROOT="/path/to/arch-lint"
```

## 📊 Complete Example

See `examples/integration-justfile` for a complete example showing how to integrate all modules into your project.

## 🔧 Customization

### Custom Architecture Patterns

Create your own templates in `configs/templates/`:

```bash
# Add custom template
cp .go-arch-lint.yml configs/templates/.go-arch-lint.microservices.yml

# Use in setup
just setup-arch microservices
```

### Custom Quality Configurations

```bash  
# Add custom quality config
cp .golangci.yml configs/templates/.golangci.enterprise.yml

# Use in setup
just setup-quality enterprise
```

## 🚀 Benefits

✅ **Modular** - Import only what you need  
✅ **Flexible** - Multiple architecture patterns and quality levels  
✅ **Smart Defaults** - Works out of the box  
✅ **Customizable** - Override paths and configurations  
✅ **Automated** - Complete setup with one command  
✅ **Maintainable** - Update via git subtree