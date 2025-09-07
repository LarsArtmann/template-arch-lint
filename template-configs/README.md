# 📁 Template Configuration Files

**Copy these 4 files to your Go project root for instant enterprise-grade linting**

---

## 📋 Files Included

### 1. `.golangci.yml` (13KB)
**Enterprise Go Linting Configuration**
- 40+ linters with maximum strictness
- Type safety enforcement (bans `interface{}`, `any`, `panic()`)
- Security scanning (gosec, vulnerability detection)
- Code quality gates (function length, complexity, file size)
- Modern Go 1.25 features and best practices

### 2. `.go-arch-lint.yml` (2KB)
**Clean Architecture Boundary Validation**
- Domain-driven design enforcement
- Dependency inversion validation
- Import cycle prevention
- Layer isolation (domain → application → infrastructure)

### 3. `justfile` (50KB)
**Development Workflow Automation**
- 30+ standardized commands
- `just lint` - Run all quality checks
- `just fix` - Auto-fix formatting issues
- `just test` - Run tests with coverage
- `just build` - Build with validations
- `just security-audit` - Complete security scan

### 4. `QUICK_START.md` (7KB)
**5-Minute Setup Guide**
- Installation instructions
- Usage examples
- Common issues and solutions
- Customization guide

---

## 🚀 Installation

```bash
# Copy all 4 files to your Go project root
cp template-configs/.golangci.yml /path/to/your/project/
cp template-configs/.go-arch-lint.yml /path/to/your/project/
cp template-configs/justfile /path/to/your/project/
cp template-configs/QUICK_START.md /path/to/your/project/
```

## 🎯 Usage

```bash
cd /path/to/your/project/
just lint          # Run all quality checks
just fix           # Auto-fix issues
just test          # Run tests
just build         # Build with validation
```

---

## 📈 Results

- **Setup Time:** 8 hours → 5 minutes (96x faster)
- **Code Quality:** Manual review → 40+ automated checks
- **Architecture:** Ad-hoc → Clean architecture enforcement
- **Security:** None → Automated vulnerability scanning
- **Type Safety:** Optional → Strictly enforced

---

## 🔗 Source

These configurations come from: https://github.com/LarsArtmann/template-arch-lint

**Transform your Go development workflow in 5 minutes!**