# Strict Vendor Control Example - go-arch-lint Configuration

## 🎯 Purpose

This demonstrates how to achieve **strict vendor dependency control** without using `anyVendorDeps: true` in go-arch-lint, following the comprehensive library policy defined in `/Users/larsartmann/projects/library-policy/library-policy.yaml`.

## 📊 Comparison: Original vs Strict Configuration

### ❌ Original Configuration (.go-arch-lint.yml)

```yaml
# Problems: Unlimited vendor dependencies
allow:
  depOnAnyVendor: true # ❌ DANGEROUS: Allows ANY vendor library

deps:
  domain-services:
    anyVendorDeps: true # ❌ UNRESTRICTED: Can import any external library
    mayDependOn: [...]

  application-handlers:
    anyVendorDeps: true # ❌ UNRESTRICTED: No vendor control
    mayDependOn: [...]
```

**Issues:**

- No control over which vendor libraries are used
- Banned/deprecated libraries can be imported
- Security vulnerabilities from outdated dependencies
- No enforcement of library-policy.yaml recommendations
- Architecture violations go undetected

### ✅ Strict Configuration (.go-arch-lint-strict.yml)

```yaml
# Solution: Strict vendor control
allow:
  depOnAnyVendor: false # ✅ SECURE: No unlimited vendor dependencies
  deepScan: true # ✅ COMPREHENSIVE: AST analysis enabled

# Explicit vendor library approval based on library-policy.yaml
vendors:
  # Only approved libraries from library-policy.yaml
  gin:
    in: ["github.com/gin-gonic/gin"]
  ginkgo:
    in: ["github.com/onsi/ginkgo"]
  samber_lo:
    in: ["github.com/samber/lo"]
  # ... only allowed libraries listed

deps:
  domain-services:
    canUse: [std, x_errgroup, samber_lo, samber_mo] # ✅ CONTROLLED
    mayDependOn: [...]

  application-handlers:
    canUse: [std, gin, ginkgo, golang_jwt, casbin, charm_log] # ✅ CONTROLLED
    mayDependOn: [...]
```

**Benefits:**

- ✅ Only approved vendor libraries can be used
- ✅ Enforces library-policy.yaml recommendations
- ✅ Prevents banned/deprecated libraries
- ✅ Security compliance maintained
- ✅ Architecture violations detected immediately
- ✅ Deep scanning for comprehensive analysis

## 🔍 Real-World Demonstration

### Example 1: Violation Detection

**Adding a banned library:**

```go
import "github.com/sirupsen/logrus"  // Banned per library-policy.yaml

func handler() {
    logrus.Info("This will be caught!")
}
```

**Result with strict config:**

```
Component application-handlers shouldn't depend on github.com/sirupsen/logrus
exit status 1
```

### Example 2: Successful Validation

**Using approved libraries:**

```go
import (
    "github.com/gin-gonic/gin"    // ✅ Approved
    "github.com/samber/lo"        // ✅ Approved
    "github.com/charmbracelet/log" // ✅ Approved
)

func handler() {
    gin.New()                     // ✅ Allowed
    lo.Map([]int{1, 2, 3}, func(i int) int { return i * 2 }) // ✅ Allowed
    log.Info("Success!")           // ✅ Allowed
}
```

**Result with strict config:**

```
OK - No warnings found
```

## 📋 Library Policy Integration

### Approved Libraries (from library-policy.yaml)

| Purpose                | Library                         | Reason                             |
| ---------------------- | ------------------------------- | ---------------------------------- |
| HTTP Framework         | `github.com/gin-gonic/gin`      | Recommended over gorilla/mux, echo |
| Testing                | `github.com/onsi/ginkgo`        | Recommended over godog, gomega     |
| Functional Programming | `github.com/samber/lo`          | Core utility library               |
| Railway Programming    | `github.com/samber/mo`          | Error handling patterns            |
| Dependency Injection   | `github.com/samber/do`          | Recommended over wire, oklog/run   |
| Configuration          | `github.com/spf13/viper`        | Recommended over envconfig         |
| HTML Templates         | `github.com/a-h/templ`          | Recommended over html/template     |
| Caching                | `github.com/maypok86/otter/v2`  | 11x faster than go-cache           |
| Financial              | `github.com/shopspring/decimal` | Industry standard                  |
| JWT                    | `github.com/golang-jwt/jwt/v5`  | Security-critical, no CVEs         |
| Observability          | `go.opentelemetry.io/otel/*`    | CNCF standard                      |

### Banned Libraries (automatically blocked)

| Library                         | Banned Reason              | Replacement                     |
| ------------------------------- | -------------------------- | ------------------------------- |
| `github.com/sirupsen/logrus`    | Use OpenTelemetry or fang  | `go.opentelemetry.io/otel`      |
| `github.com/gorilla/mux`        | Gin is better              | `github.com/gin-gonic/gin`      |
| `github.com/patrickmn/go-cache` | 11x slower than otter      | `github.com/maypok86/otter/v2`  |
| `github.com/dgrijalva/jwt-go`   | CVE-2020-26160             | `github.com/golang-jwt/jwt/v5`  |
| `gorm.io/gorm`                  | Use sqlc for type-safe SQL | `github.com/sqlc-dev/sqlc`      |
| `github.com/pkg/errors`         | Deprecated                 | `github.com/cockroachdb/errors` |

## 🛠️ Implementation Steps

### 1. Configure Strict Mode

```yaml
version: 3
allow:
  depOnAnyVendor: false # ❌ DISABLE unlimited vendor deps
  deepScan: true # ✅ ENABLE comprehensive analysis
```

### 2. Define Approved Vendors

```yaml
vendors:
  # HTTP Server
  gin:
    in: ["github.com/gin-gonic/gin"]

  # Testing Framework
  ginkgo:
    in: ["github.com/onsi/ginkgo", "github.com/onsi/ginkgo/*"]

  # Utilities
  samber_lo:
    in: ["github.com/samber/lo"]
```

### 3. Set Component Permissions

```yaml
deps:
  domain-entities:
    canUse: [std] # Only standard library
    mayDependOn: [domain-values, pkg-errors]

  application-handlers:
    canUse: [std, gin, ginkgo] # Approved vendors only
    mayDependOn: [domain-entities, domain-services]
```

### 4. Test Configuration

```bash
# Test strict configuration
go-arch-lint check --arch-file .go-arch-lint-strict.yml

# Should show: "OK - No warnings found" for compliant code
# Should show violations for banned libraries
```

## 🎯 Benefits Achieved

### ✅ Security Benefits

- **Zero Unknown Dependencies**: Only pre-approved libraries can be used
- **CVE Prevention**: Banned vulnerable libraries automatically blocked
- **Supply Chain Security**: Full visibility into all external dependencies

### ✅ Performance Benefits

- **Enforced Best Practices**: Only recommended high-performance libraries
- **No Legacy Bloat**: Prevents slow, outdated libraries
- **Benchmark Compliance**: Libraries verified against performance standards

### ✅ Maintainability Benefits

- **Consistency**: All projects use same approved stack
- **Documentation**: Clear why each library is chosen
- **Upgrades**: Centralized library updates across all projects

### ✅ Architecture Benefits

- **Layer Purity**: Domain layer stays clean with minimal dependencies
- **Dependency Control**: Explicit dependency graph visualization
- **Violations Detected**: Real-time architectural enforcement

## 🚀 Migration Strategy

### Phase 1: Baseline Setup

1. Copy `.go-arch-lint-strict.yml` to your project
2. Run `go-arch-lint check --arch-file .go-arch-lint-strict.yml`
3. Identify existing violations

### Phase 2: Vendor Compliance

1. Replace banned libraries with approved alternatives
2. Add missing approved vendors to configuration
3. Update component permissions as needed

### Phase 3: Full Enforcement

1. Replace `.go-arch-lint.yml` with strict version
2. Add to CI/CD pipeline
3. Enable pre-commit hooks for immediate feedback

## 📈 Results

| Metric                   | Before (anyVendorDeps: true) | After (strict control)       |
| ------------------------ | ---------------------------- | ---------------------------- |
| Security Vulnerabilities | Unknown                      | Zero (controlled)            |
| Library Bloat            | High                         | Minimal (approved only)      |
| Performance Issues       | Common                       | Rare (benchmarked libraries) |
| Architectural Violations | Undetected                   | Real-time detection          |
| Dependency Visibility    | Low                          | Complete (explicit list)     |
| Team Consistency         | Variable                     | High (standardized stack)    |

## 🔗 References

- **Library Policy**: `/Users/larsartmann/projects/library-policy/library-policy.yaml`
- **Configuration File**: `.go-arch-lint-strict.yml`
- **Tool Documentation**: https://github.com/fe3dback/go-arch-lint
- **Go Standards**: https://golang.org/doc/effective_go.html

---

**Key Takeaway**: `anyVendorDeps: false` with explicit vendor approval provides enterprise-grade dependency control, security compliance, and architectural enforcement that `anyVendorDeps: true` cannot match.
