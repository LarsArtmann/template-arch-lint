# 🚀 Template-Arch-Lint Quick Start Guide

**5-Minute Setup: Enterprise-Grade Go Linting**

Transform your Go project with copy-paste configuration files that enforce architectural boundaries, type safety, and code quality.

---

## 📋 What You Get

✅ **40+ Linters** - Comprehensive code quality enforcement\
✅ **Architecture Rules** - Clean architecture boundary validation\
✅ **Type Safety** - Zero tolerance for `interface{}` and weak typing\
✅ **Security Scanning** - Automated vulnerability detection\
✅ **Performance Optimization** - Struct alignment, preallocation hints\
✅ **CMD Single Main** - Enforce single entry point architecture

---

## ⚡ 1-Minute Installation

### Option 1: Copy Essential Files (FASTEST)

```bash
# Copy these 3 files to your Go project root:
# 1. .golangci.yml (40+ linters configuration)
# 2. .go-arch-lint.yml (architecture rules)
# 3. justfile (automation commands)
```

### Option 2: Download & Extract

```bash
curl -s https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/install.sh | bash
```

---

## 📁 Essential Configuration Files

### 1. `.golangci.yml` - Copy this entire file to your project root

**Purpose:** 40+ linters with maximum strictness for enterprise-grade code quality

**Key Features:**

- 🚨 **Type Safety Enforcement** - Bans `interface{}`, `any`, `panic()`
- 🔒 **Security** - gosec, vulnerability scanning, injection prevention
- 📊 **Code Quality** - Function length (50 lines), complexity (10), file length (400 lines)
- 🚀 **Modern Go** - Latest Go 1.25 features and best practices

```yaml
# Copy the complete .golangci.yml from this repository
# It's 400+ lines of carefully tuned linter configuration
# Designed for zero-defect production systems
```

### 2. `.go-arch-lint.yml` - Copy this for architecture enforcement

**Purpose:** Clean Architecture boundary validation with dependency inversion

```yaml
# Copy the complete .go-arch-lint.yml from this repository
# Enforces domain-driven design patterns
# Prevents circular dependencies and architecture violations
```

### 3. `justfile` - Copy this for automation commands

**Purpose:** Standardized development workflow automation

```bash
# Copy the complete justfile from this repository
# Provides 30+ commands for linting, testing, building
# Examples: just lint, just fix, just test, just build
```

---

## 🎯 Basic Usage

### Run All Quality Checks

```bash
just lint          # Run all linters (architecture + code + security)
just fix           # Auto-fix formatting and simple violations
just test          # Run tests with coverage
just build         # Build with all validations
```

### Individual Linters

```bash
just lint-arch     # Architecture boundary validation only
just lint-code     # Code quality linting (40+ linters)
just security-audit # Complete security vulnerability scan
just lint-cmd-single # CMD single main.go enforcement
```

---

## 🔥 Enterprise Features

### Type Safety Enforcement

- **Bans `interface{}`** → Use specific types or generics
- **Bans `any`** → Strong typing required
- **Bans `panic()`** → Return errors instead
- **Bans print statements** → Use structured logging

### Architecture Validation

- **Domain Purity** → Domain layer cannot import infrastructure
- **Dependency Inversion** → Infrastructure depends on domain interfaces
- **Single Responsibility** → One main.go per project in `cmd/`
- **Import Cycles** → Automatic detection and prevention

### Code Quality Gates

- **Function Length** → Max 50 lines
- **Cyclomatic Complexity** → Max 10
- **File Length** → Max 400 lines
- **Cognitive Complexity** → Max 10

---

## 📊 Before/After Comparison

| Metric       | Before Template | After Template       |
| ------------ | --------------- | -------------------- |
| Setup Time   | 8 hours         | 5 minutes            |
| Linters      | 5-10 basic      | 40+ enterprise       |
| Architecture | Manual review   | Automated validation |
| Type Safety  | Optional        | Enforced             |
| Security     | None            | Automated scanning   |
| Code Quality | Inconsistent    | Standardized         |

---

## 🛠️ Customization

### Adjust Strictness Levels

```yaml
# In .golangci.yml, modify these settings:
funlen:
  lines: 50 # Increase for larger functions
  statements: 30 # Increase for more statements

cyclop:
  max-complexity: 10 # Increase for more complex logic

revive:
  rules:
    - name: file-length-limit
      arguments:
        - max: 400 # Increase for larger files
```

### Add Project-Specific Rules

```yaml
# In .go-arch-lint.yml, add your layers:
components:
  - name: "presentation"
    sourcePackages: ["./cmd", "./internal/handlers"]
  - name: "application"
    sourcePackages: ["./internal/application"]
  - name: "domain"
    sourcePackages: ["./internal/domain"]
  - name: "infrastructure"
    sourcePackages: ["./internal/infrastructure"]
```

---

## 🚨 Common Issues & Solutions

### Issue: Too Many Linting Errors

**Solution:** Start with `just fix` to auto-fix formatting, then address remaining issues incrementally

### Issue: Function Too Long Errors

**Solution:** Break functions into smaller, focused functions (<50 lines each)

### Issue: Architecture Violations

**Solution:** Check dependency direction - Infrastructure → Application → Domain

### Issue: Type Safety Violations

**Solution:** Replace `interface{}` with specific types, use generics for reusable code

---

## 🎯 Success Checklist

- [ ] `.golangci.yml` copied and `just lint-code` passes
- [ ] `.go-arch-lint.yml` copied and `just lint-arch` passes
- [ ] `justfile` copied and `just lint` runs successfully
- [ ] All functions <50 lines, files <400 lines
- [ ] No `interface{}`, `any`, or `panic()` in codebase
- [ ] Clean architecture boundaries respected
- [ ] Security scan shows no critical vulnerabilities

---

## 📈 ROI: 32x Development Speed Improvement

- **Setup:** 5 minutes vs 8 hours (96x faster)
- **Quality:** 40+ automated checks vs manual review
- **Architecture:** Automated validation vs architecture reviews
- **Security:** Automated scanning vs manual audits
- **Maintenance:** Self-documenting rules vs tribal knowledge

---

## 🔗 What's Next?

1. **Copy the 3 essential files** to your project
2. **Run `just lint`** to see current status
3. **Fix violations incrementally** using the error messages as guidance
4. **Customize rules** for your specific needs
5. **Share with your team** for organization-wide adoption

---

**🎯 GOAL: Transform Go project setup from 8 hours → 5 minutes with enterprise-grade quality**

Get the complete configuration files from: https://github.com/LarsArtmann/template-arch-lint
