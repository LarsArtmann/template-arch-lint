# PARTS.md

> Analysis of extractable components from `template-arch-lint` project.
> Version: 1.0 | Date: 2026-02-26

---

## Executive Summary

This document analyzes components within `template-arch-lint` that could be extracted as standalone libraries/SDKs. Each component is evaluated against existing alternatives in the Go ecosystem to determine unique value and extraction viability.

| Component | Extract? | Unique Value | Recommendation |
|-----------|----------|--------------|----------------|
| `pkg/errors` | **No** | Low | Use `cockroachdb/errors` + `samber/mo` Result |
| `pkg/linter-plugins` | **Yes** | High | Extract as `go-arch-linters` |
| `template-configs/` | **No** | Medium | Keep as copy-paste templates |
| `internal/domain/values` | **No** | Low | Use `go-playground/validator` |
| `internal/testhelpers` | **No** | None | Too sparse, not ready |

---

## 1. pkg/errors — Centralized Error Handling

### Current Implementation

**Location:** `pkg/errors/errors.go` (502 lines)

**Features:**
- Typed error codes: `ValidationError`, `NotFoundError`, `ConflictError`, `InternalError`, `DatabaseError`, `NetworkError`, `ConfigurationError`
- `DomainError` and `InfrastructureError` interfaces
- `ErrorDetails` struct with field, resource, ID, value, reason, extra
- HTTP status mapping built-in (`HTTPStatus()` method)
- `IsRetryable()` pattern for infrastructure errors
- Helper constructors: `NewValidationError()`, `NewRequiredFieldError()`, `NewNotFoundError()`, etc.

### Existing Alternatives

| Library | Stars | Features | Gaps |
|---------|-------|----------|------|
| `cockroachdb/errors` | 1.5k+ | Stack traces, error wrapping, PII-safe details, domains, HTTP code wrappers | No built-in HTTP status mapping |
| `samber/oops` | 500+ | Context, assertion, stack traces, source fragments | Newer, less battle-tested |
| `samber/mo` Result | 2k+ | Functional Result[T] type, Either, Option | Not error-specific, functional pattern |
| `larsartmann/uniflow` | - | Railway Oriented Programming, errors as values | Private/personal project |

### Comparison Analysis

```
Our pkg/errors:
+ HTTP status mapping built-in (DomainError.HTTPStatus())
+ Typed error codes with consistent structure
+ Domain/Infrastructure error distinction
+ IsRetryable pattern for resilience decisions
- No stack traces
- No error wrapping/chaining
- No PII-safe details
- No Sentry integration

cockroachdb/errors:
+ Stack traces with source fragments
+ Error wrapping with context
+ PII-safe details via SafeFormatter
+ Domain-based error categorization
+ Sentry.io integration
+ HTTP code wrappers
- No built-in HTTP status interface
- No IsRetryable pattern
```

### Unique Value Proposition

**Low.** Our implementation provides HTTP status mapping and retryable patterns, but these can be added as thin wrappers around `cockroachdb/errors`:

```go
// Instead of our custom errors, use:
import "github.com/cockroachdb/errors"

type HTTPError interface {
    error
    HTTPStatus() int
}

func WithHTTPStatus(err error, code int) HTTPError {
    return &httpErrorWrapper{cause: err, code: code}
}
```

### Recommendation

**Do NOT extract.** Align with `HOW_TO_GOLANG.md` policy:

> Use `larsartmann/uniflow` or `cockroachdb/errors`.

**Action:** Replace `pkg/errors` with:
1. `cockroachdb/errors` for error creation and wrapping
2. `samber/mo` Result[T] for functional error handling
3. Add thin HTTP status wrapper if needed

---

## 2. pkg/linter-plugins — Custom golangci-lint Plugin Suite

### Current Implementation

**Location:** `pkg/linter-plugins/template-arch-lint/main.go` (47 lines + implementations)

**Analyzers:**

| Analyzer | Purpose | Lines |
|----------|---------|-------|
| `filename-validator` | Validates Go file naming conventions | ~60 |
| `cmd-single-main` | Enforces exactly one main.go in cmd/ | ~80 |
| `import-cycle-detector` | Detects circular dependencies via AST | ~100 |
| `code-duplication-detector` | Finds duplicate code blocks | ~90 |

**Technical Details:**
- Uses `golang.org/x/tools/go/analysis` framework
- Compatible with golangci-lint v2 plugin system
- Entry point: `func New(conf any) ([]*analysis.Analyzer, error)`

### Existing Alternatives

| Tool | Purpose | Gaps |
|------|---------|------|
| `go-arch-lint` | Architecture boundary enforcement | No cmd-single-main, no duplication detection |
| `arch-go` | Architecture testing (newer) | Limited adoption, different approach |
| `dupl` | Code duplication detection | No integration with our other analyzers |
| `goimports` | Import management | No cycle detection |
| Google's custom plugins | `sliceofpointers`, `fmtpercentv` | Different focus, not architecture |

### Plugin Ecosystem Analysis

From Sourcegraph research:
- **Google/go-github** maintains custom golangci-lint plugins
- Plugin pattern: `github.com/golangci/plugin-module-register/register`
- Few public architecture-focused plugins exist
- `go-arch-lint` is the main architecture tool but has different scope

### Unique Value Proposition

**High.** This plugin suite provides:

1. **cmd-single-main**: Unique in ecosystem. Enforces clean architecture by preventing command proliferation
2. **filename-validator**: Consistent naming across projects
3. **import-cycle-detector**: AST-based, integrates with golangci-lint workflow
4. **code-duplication-detector**: Integrated with lint pipeline (vs standalone `dupl`)

**Combined value:** All four analyzers in a single plugin with consistent configuration.

### Recommendation

**EXTRACT as `go-arch-linters`.**

**Proposed Structure:**
```
github.com/larsartmann/go-arch-linters/
├── cmd-single-main/       # Single entry point enforcement
├── filename-validator/    # Naming convention enforcement
├── import-cycle/          # Circular dependency detection
├── code-duplication/      # Duplicate code detection
├── all.go                 # Combined plugin entry point
└── .golangci.yml          # Example configuration
```

**Distribution Options:**
1. **Custom golangci-lint binary**: `golangci-lint custom` with `.custom-gcl.yml`
2. **Go plugin**: `.so` file loaded at runtime
3. **Module import**: Direct analyzer use in tests

**Action:**
1. Extract to standalone repository
2. Add comprehensive documentation
3. Provide example `.custom-gcl.yml`
4. Add to golangci-lint plugin registry

---

## 3. template-configs/ — Configuration Distribution System

### Current Implementation

**Location:** `template-configs/`

**Files:**
- `.go-arch-lint.yml` (348 lines) — Architecture boundary rules
- `.golangci.yml` (131 lines) — 99+ linters configuration
- `justfile` (1153 lines) — Task automation commands

**Approach:** Copy-paste to new projects, not installed as dependency.

### Existing Alternatives

| Approach | Examples | Tradeoffs |
|----------|----------|-----------|
| **Copy-paste configs** | Our approach | Simple, no versioning issues, requires manual sync |
| **Shared config package** | `golangci-lint` shared configs | Versioning, but limited flexibility |
| **Config inheritance** | Some linters support | Complex, harder to debug |
| **Renovate/Dependabot** | Auto-update configs | Only works if configs are packages |

### Comparison Analysis

```
Copy-paste approach:
+ No dependency version conflicts
+ Full customization per project
+ Easy to understand and debug
- Manual sync required
- No automatic updates
- Drift between projects

Package approach:
+ Automatic updates via go.mod
+ Single source of truth
- Version conflicts possible
- Less flexibility
- Harder to debug
```

### Unique Value Proposition

**Medium.** The configurations themselves are valuable (99+ linters, architecture rules, comprehensive justfile), but the distribution mechanism (copy-paste) is intentional, not a limitation.

The value is in the **content** of the configs, not a library.

### Recommendation

**Do NOT extract as library.** Keep as copy-paste templates.

**Rationale:**
1. Configs need project-specific customization
2. Linter versions change; pinning causes conflicts
3. Architecture rules are project-specific
4. Justfile commands vary by project

**Action:**
1. Keep `template-configs/` as reference
2. Add `README.md` explaining copy-paste process
3. Document sync process for updates
4. Consider GitHub template repository for bootstrapping

---

## 4. internal/domain/values — Value Object Pattern with Validation

### Current Implementation

**Location:** `internal/domain/values/`

**Examples:**
- `email.go` (185 lines) — Email value object with comprehensive validation
- `username.go` — UserName value object
- `user_id.go` — UserID value object

**Pattern:**
```go
type Email struct {
    value string
}

func NewEmail(email string) (Email, error) {
    if err := validateEmailFormat(email); err != nil {
        return Email{}, err
    }
    return Email{value: email}, nil
}
```

### Existing Alternatives

| Library | Approach | Features |
|---------|----------|----------|
| `go-playground/validator` | Struct tags | 100+ validators, custom validators, cross-field |
| `ozzo-validation` | Fluent API | Conditional validation, built-in rules |
| `asaskevich/govalidator` | Struct tags + functions | String validators, sanitizers |
| `samber/mo` | Functional types | Option[T], Result[T], Either[L,R] |

### Comparison Analysis

```
Our value objects:
+ Type safety at domain level
+ Business rules encapsulated
+ Integration with centralized errors
+ Immutable by design
- Manual validation code
- No reusable validation rules
- Boilerplate per value type

go-playground/validator:
+ Declarative struct tags
+ 100+ built-in validators
+ Custom validator support
+ Cross-field validation
- Runtime validation (not compile-time)
- No type-level enforcement

samber/mo Option[T]:
+ Functional composition
+ Null-safe operations
+ Type-safe at compile time
- No business validation
```

### Unique Value Proposition

**Low.** The value object pattern is well-established. Our implementation:

- Doesn't provide unique validation capabilities
- Tightly coupled to our error types
- Duplicates logic available in validators

### Recommendation

**Do NOT extract.** Use existing validation libraries.

**Per `HOW_TO_GOLANG.md`:**
> Use `go-playground/validator` for non-Huma code.

**Alternative approach:**
```go
// Instead of custom value objects, use:
type Email struct {
    Value string `validate:"required,email,max=254"`
}

func NewEmail(email string) (Email, error) {
    e := Email{Value: email}
    if err := validate.Struct(e); err != nil {
        return Email{}, err
    }
    return e, nil
}
```

---

## 5. internal/testhelpers — Testing Utilities

### Current Implementation

**Location:** `internal/testhelpers/`

**Status:** Sparse. `suite.go` is an empty stub.

### Existing Alternatives

| Library | Features |
|---------|----------|
| `onsi/ginkgo/v2` | BDD testing, parallel execution, reporters |
| `onsi/gomega` | Rich matchers, async assertions |
| `testcontainers-go` | Integration testing with real dependencies |
| `data-dog/go-sqlmock` | Database mocking |

### Recommendation

**Do NOT extract.** Not ready for extraction.

**Action:**
1. Develop testing utilities as needed
2. Use Ginkgo/Gomega per `HOW_TO_GOLANG.md`
3. Consider extracting only when mature

---

## Summary: Extraction Roadmap

### Immediate Actions

| Priority | Action | Effort | Impact |
|----------|--------|--------|--------|
| 1 | Extract `pkg/linter-plugins` → `go-arch-linters` | Medium | High |
| 2 | Replace `pkg/errors` with `cockroachdb/errors` | Low | Medium |
| 3 | Document `template-configs/` copy-paste process | Low | Medium |

### Future Considerations

| Component | Condition for Extraction |
|-----------|--------------------------|
| `pkg/errors` | If HTTP status wrapper becomes reusable across 3+ projects |
| `internal/domain/values` | If generic value object builder pattern emerges |
| `internal/testhelpers` | When mature with 5+ reusable utilities |

---

## Appendix: Library Policy Alignment

Per `HOW_TO_GOLANG.md`, the following libraries should be used instead of custom implementations:

| Category | Required Library | Replaces |
|----------|-----------------|----------|
| Error Handling | `cockroachdb/errors`, `larsartmann/uniflow` | `pkg/errors` |
| Functional Types | `samber/lo`, `samber/mo` | Custom patterns |
| Validation | `go-playground/validator` | Custom validators |
| Testing | `onsi/ginkgo/v2`, `onsi/gomega` | `testify` (banned) |
| DI | `samber/do/v2` | Manual DI |

---

## References

- [golangci-lint Plugin Documentation](https://golangci-lint.run/plugins/go-plugins/)
- [go/analysis Framework](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- [cockroachdb/errors GitHub](https://github.com/cockroachdb/errors)
- [samber/mo GitHub](https://github.com/samber/mo)
- [go-playground/validator GitHub](https://github.com/go-playground/validator)
