# 🤝 Contributing to Template-Arch-Lint

Thank you for your interest in contributing to **template-arch-lint**! This project helps developers build better Go applications with enterprise-grade architecture enforcement and zero-tolerance quality standards.

## 🎯 **Vision & Mission**

**Vision**: Make architectural excellence and code quality automatic for every Go developer and team.

**Mission**: Provide the most comprehensive, copy-paste ready linting template that prevents technical debt and enforces Clean Architecture principles from day one.

---

## 📋 **Table of Contents**

- [🚀 Quick Start for Contributors](#-quick-start-for-contributors)
- [🎯 Ways to Contribute](#-ways-to-contribute)
- [🛠️ Development Setup](#️-development-setup)
- [📝 Coding Standards](#-coding-standards)
- [🧪 Testing Guidelines](#-testing-guidelines)
- [📖 Documentation Standards](#-documentation-standards)
- [🔄 Pull Request Process](#-pull-request-process)
- [🏗️ Architecture Guidelines](#️-architecture-guidelines)
- [🎉 Recognition & Rewards](#-recognition--rewards)

---

## 🚀 **Quick Start for Contributors**

### 1. **Fork & Clone**

```bash
# Fork on GitHub, then clone your fork
git clone https://github.com/yourusername/template-arch-lint.git
cd template-arch-lint

# Add upstream remote
git remote add upstream https://github.com/LarsArtmann/template-arch-lint.git
```

### 2. **Install Development Tools**

```bash
# Install all required tools
just install

# Verify installation
just lint
just test

# Check all systems working
just ci
```

### 3. **Create Feature Branch**

```bash
# Create a descriptive branch name
git checkout -b feature/add-performance-benchmarks
git checkout -b fix/golangci-lint-memory-usage
git checkout -b docs/architecture-decision-records
```

### 4. **Make Your Changes**

```bash
# Follow our development workflow
just format          # Format code
just fix             # Auto-fix issues
just lint            # Validate quality
just test            # Run tests
just ci              # Full validation
```

### 5. **Submit Pull Request**

```bash
# Push to your fork
git push origin feature/your-feature

# Create PR on GitHub with:
# - Clear title and description
# - Link to related issues
# - Screenshots/demos if applicable
```

---

## 🎯 **Ways to Contribute**

### 🏆 **High-Impact Contributions**

#### **1. 🏗️ Architecture Patterns**

- **Event Sourcing Templates** - Add `.go-arch-lint.yml` configurations
- **CQRS Patterns** - Command/Query separation enforcement
- **Microservices Boundaries** - Service isolation rules
- **Hexagonal Architecture** - Ports and adapters validation

#### **2. 🔍 Linting Enhancements**

- **New Linters Integration** - Research and add valuable linters
- **Custom Rules** - Project-specific architectural rules
- **Performance Optimizations** - Make linting faster
- **False Positive Reduction** - Improve accuracy

#### **3. 🧪 Testing Infrastructure**

- **Property-Based Testing** - Add QuickCheck-style tests
- **Mutation Testing** - Add go-mutesting integration
- **Performance Benchmarks** - CPU, memory, and I/O benchmarks
- **Integration Test Patterns** - End-to-end testing examples

#### **4. 📚 Documentation & Education**

- **Architecture Decision Records** - Document design decisions
- **Video Tutorials** - Architecture enforcement walkthroughs
- **Case Studies** - Real-world implementation examples
- **Best Practices Guides** - Industry-specific recommendations

#### **5. 🚀 DevOps & Automation**

- **CI/CD Templates** - GitHub Actions, GitLab CI, Jenkins
- **Docker Optimizations** - Multi-stage builds, security scanning
- **Deployment Patterns** - Kubernetes, serverless configurations
- **Monitoring Integration** - Observability stack setups

### 💡 **Medium-Impact Contributions**

#### **Bug Fixes & Improvements**

- Fix linter configuration issues
- Improve error messages and user experience
- Optimize performance bottlenecks
- Enhance cross-platform compatibility

#### **Tool Integrations**

- IDE plugins (VS Code, GoLand, Vim)
- Pre-commit hook enhancements
- Git hook automations
- Build tool integrations

#### **Community & Support**

- Answer questions in Discussions
- Review pull requests
- Improve issue templates
- Create troubleshooting guides

### 🛠️ **Beginner-Friendly Contributions**

#### **Documentation**

- Fix typos and grammar
- Improve code examples
- Add missing documentation
- Translate documentation

#### **Examples & Demos**

- Add more example projects
- Create demo applications
- Improve existing examples
- Add edge case demonstrations

#### **Testing**

- Add test cases for edge scenarios
- Improve test coverage
- Add integration tests
- Create test utilities

---

## 🛠️ **Development Setup**

### **Prerequisites**

```bash
# Required tools (auto-installed via `just install`)
- Go 1.21+ (https://golang.org/dl/)
- just command runner (https://github.com/casey/just)
- Git (https://git-scm.com/)

# Recommended tools
- VS Code with Go extension
- Docker & Docker Compose
- golangci-lint v2.3.1+
- go-arch-lint v1.12.0+
```

### **Environment Setup**

```bash
# 1. Clone and setup
git clone https://github.com/yourusername/template-arch-lint.git
cd template-arch-lint

# 2. Install all tools
just install

# 3. Verify everything works
just ci

# 4. Run the example application
just run
# Visit http://localhost:8080
```

### **Development Workflow Commands**

```bash
# 📝 Code Quality
just format          # Format code (gofumpt + goimports)
just fix             # Auto-fix linting issues
just lint            # Run all linters
just lint-arch       # Architecture validation only
just lint-code       # Code quality only
just lint-security   # Security linting only

# 🧪 Testing
just test            # Run all tests
just test-unit       # Unit tests only
just test-integration # Integration tests only
just coverage        # Generate coverage report

# 🏗️ Building & Running
just build           # Build application
just run             # Run application
just docker-test     # Test Docker build

# 📊 Analysis & Reporting
just report          # Generate comprehensive reports
just stats           # Show project statistics
just deps-check      # Check dependency vulnerabilities

# 🧹 Maintenance
just clean           # Clean generated files
just update-deps     # Update Go dependencies
just update-tools    # Update development tools
```

---

## 📝 **Coding Standards**

### **🏗️ Architecture Standards**

#### **Clean Architecture Principles**

```go
// ✅ GOOD: Domain entities with no infrastructure dependencies
package entities

type User struct {
    id       UserID
    email    Email
    name     UserName
    created  time.Time
    modified time.Time
}

func (u *User) UpdateEmail(newEmail Email) error {
    // Business logic only - no database, HTTP, etc.
    if u.email == newEmail {
        return ErrEmailUnchanged
    }
    u.email = newEmail
    u.modified = time.Now()
    return nil
}

// ❌ BAD: Domain importing infrastructure
import "database/sql"  // Violates Clean Architecture
```

#### **Dependency Inversion**

```go
// ✅ GOOD: Repository interface in domain
package repositories

type UserRepository interface {
    Save(ctx context.Context, user *entities.User) error
    FindByID(ctx context.Context, id values.UserID) (*entities.User, error)
}

// ✅ GOOD: Infrastructure implements domain interface
package persistence

type sqlUserRepository struct {
    db *sql.DB
}

func (r *sqlUserRepository) Save(ctx context.Context, user *entities.User) error {
    // Implementation details here
}
```

### **🔒 Type Safety Standards**

#### **Forbidden Patterns**

```go
// ❌ BANNED: interface{} erases type safety
var data interface{}
json.Unmarshal(body, &data)

// ❌ BANNED: any erases type safety
func Process(input any) any {
    return input
}

// ❌ BANNED: panic causes runtime crashes
if err != nil {
    panic(err)  // Use proper error handling instead
}

// ✅ GOOD: Specific types with proper error handling
type UserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func ProcessUser(req UserRequest) (*User, error) {
    if err := validator.Validate(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    // Process with type safety
    return &User{Name: req.Name, Email: req.Email}, nil
}
```

### **🛡️ Error Handling Standards**

#### **Error Wrapping & Context**

```go
// ✅ GOOD: Proper error wrapping with context
func (s *UserService) CreateUser(ctx context.Context, email string) (*User, error) {
    user, err := entities.NewUser(email)
    if err != nil {
        return nil, fmt.Errorf("failed to create user entity: %w", err)
    }

    if err := s.repo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to save user %s: %w", user.ID, err)
    }

    return user, nil
}

// ❌ BAD: No error wrapping, loses context
func (s *UserService) CreateUser(ctx context.Context, email string) (*User, error) {
    user, err := entities.NewUser(email)
    if err != nil {
        return nil, err  // No context about what failed
    }

    if err := s.repo.Save(ctx, user); err != nil {
        return nil, err  // No context about which user or operation
    }

    return user, nil
}
```

### **📊 Complexity Standards**

#### **Function Size Limits**

```go
// ✅ GOOD: Function under 50 lines, single responsibility
func (s *UserService) ValidateAndCreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    if err := s.validateRequest(req); err != nil {
        return nil, err
    }

    user, err := s.createUserEntity(req)
    if err != nil {
        return nil, err
    }

    return s.saveUser(ctx, user)
}

// Supporting functions keep complexity low
func (s *UserService) validateRequest(req CreateUserRequest) error { /* ... */ }
func (s *UserService) createUserEntity(req CreateUserRequest) (*User, error) { /* ... */ }
func (s *UserService) saveUser(ctx context.Context, user *User) (*User, error) { /* ... */ }

// ❌ BAD: Monolithic function over 50 lines
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // 100+ lines of mixed validation, creation, persistence logic
    // High cognitive complexity, hard to test, maintain
}
```

#### **Cyclomatic Complexity Limits**

```go
// ✅ GOOD: Complexity under 10, uses table-driven patterns
func ValidateEmail(email string) error {
    validations := []struct {
        condition bool
        error     error
    }{
        {email == "", ErrEmailEmpty},
        {len(email) > 255, ErrEmailTooLong},
        {!emailRegex.MatchString(email), ErrEmailInvalid},
        {strings.Contains(email, ".."), ErrEmailConsecutiveDots},
    }

    for _, v := range validations {
        if v.condition {
            return v.error
        }
    }

    return nil
}

// ❌ BAD: High complexity with nested conditions
func ValidateEmail(email string) error {
    if email == "" {
        return ErrEmailEmpty
    } else if len(email) > 255 {
        return ErrEmailTooLong
    } else if !emailRegex.MatchString(email) {
        return ErrEmailInvalid
    } else if strings.Contains(email, "..") {
        return ErrEmailConsecutiveDots
    } // ... many more nested conditions
    return nil
}
```

### **📝 Documentation Standards**

#### **Package Documentation**

```go
// ✅ GOOD: Comprehensive package documentation
// Package entities contains the core business entities for the user domain.
//
// This package implements Domain-Driven Design (DDD) principles with rich
// domain models that encapsulate business logic and invariants. Entities
// in this package have no dependencies on infrastructure concerns.
//
// Key entities:
//   - User: Represents a system user with identity and behavior
//   - Email: Value object ensuring email validity
//   - UserID: Strongly-typed identifier preventing confusion
//
// Example usage:
//   user, err := entities.NewUser(userID, email, name)
//   if err != nil {
//       return fmt.Errorf("invalid user data: %w", err)
//   }
package entities
```

#### **Function Documentation**

```go
// ✅ GOOD: Complete function documentation
// CreateUser creates a new user entity with validation and business rules.
//
// The function enforces domain invariants:
//   - Email must be valid and unique
//   - Name must be between 2-100 characters
//   - UserID must follow system format
//
// Parameters:
//   - id: Unique identifier for the user
//   - email: Valid email address string
//   - name: User's display name
//
// Returns:
//   - *User: Created user entity with generated timestamps
//   - error: Validation or business rule violation
//
// Example:
//   user, err := entities.NewUser(id, "john@example.com", "John Doe")
//   if err != nil {
//       log.Printf("User creation failed: %v", err)
//       return err
//   }
func NewUser(id UserID, email string, name string) (*User, error) {
    // Implementation...
}

// ❌ BAD: Minimal or missing documentation
// CreateUser creates user
func NewUser(id UserID, email string, name string) (*User, error) {
    // No context, parameters, returns, or examples
}
```

---

## 🧪 **Testing Guidelines**

### **🎯 Testing Philosophy**

#### **Test Pyramid**

```
           ╭─────────────╮
          ╱  E2E Tests    ╲     ← Few, expensive, realistic
         ╱    (10%)       ╲
        ╱─────────────────╲
       ╱ Integration Tests ╲    ← Some, medium cost, focused
      ╱      (20%)         ╲
     ╱───────────────────────╲
    ╱    Unit Tests (70%)     ╲  ← Many, fast, isolated
   ╱─────────────────────────╲
```

#### **Testing Standards**

- **70% Unit Tests**: Fast, isolated, cover business logic
- **20% Integration Tests**: Component interactions, database operations
- **10% End-to-End Tests**: Full user workflows

### **🔧 Unit Testing Patterns**

#### **Table-Driven Tests**

```go
func TestUserValidation(t *testing.T) {
    tests := []struct {
        name          string
        email         string
        username      string
        expectedError error
    }{
        {
            name:          "valid user",
            email:         "test@example.com",
            username:      "testuser",
            expectedError: nil,
        },
        {
            name:          "invalid email",
            email:         "invalid-email",
            username:      "testuser",
            expectedError: ErrInvalidEmail,
        },
        {
            name:          "empty username",
            email:         "test@example.com",
            username:      "",
            expectedError: ErrEmptyUsername,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := entities.NewUser(tt.email, tt.username)

            if tt.expectedError != nil {
                assert.ErrorIs(t, err, tt.expectedError)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

#### **BDD-Style Tests (Ginkgo/Gomega)**

```go
var _ = Describe("User Entity", func() {
    Describe("Creating a new user", func() {
        Context("with valid data", func() {
            It("should create user successfully", func() {
                // Given
                email := "test@example.com"
                username := "testuser"

                // When
                user, err := entities.NewUser(email, username)

                // Then
                Expect(err).ToNot(HaveOccurred())
                Expect(user.Email()).To(Equal(email))
                Expect(user.Username()).To(Equal(username))
                Expect(user.ID()).ToNot(BeEmpty())
            })
        })

        Context("with invalid email", func() {
            It("should return validation error", func() {
                // Given
                invalidEmail := "not-an-email"
                username := "testuser"

                // When
                user, err := entities.NewUser(invalidEmail, username)

                // Then
                Expect(err).To(HaveOccurred())
                Expect(user).To(BeNil())
                Expect(err).To(MatchError(ContainSubstring("invalid email")))
            })
        })
    })
})
```

### **🏗️ Integration Testing**

#### **Repository Testing**

```go
func TestSQLUserRepository(t *testing.T) {
    // Setup in-memory database for tests
    db := setupTestDB(t)
    defer db.Close()

    repo := persistence.NewSQLUserRepository(db, logger)

    t.Run("SaveAndFindUser", func(t *testing.T) {
        // Given
        user := createTestUser(t)

        // When
        err := repo.Save(ctx, user)
        require.NoError(t, err)

        found, err := repo.FindByID(ctx, user.ID)

        // Then
        require.NoError(t, err)
        assert.Equal(t, user.ID, found.ID)
        assert.Equal(t, user.Email, found.Email)
    })
}
```

#### **HTTP Handler Testing**

```go
func TestUserHandler_CreateUser(t *testing.T) {
    // Setup
    mockService := &MockUserService{}
    handler := handlers.NewUserHandler(mockService, logger)
    router := setupTestRouter(handler)

    t.Run("ValidRequest", func(t *testing.T) {
        // Given
        requestBody := `{"email": "test@example.com", "username": "testuser"}`
        req := httptest.NewRequest("POST", "/users", strings.NewReader(requestBody))
        req.Header.Set("Content-Type", "application/json")

        w := httptest.NewRecorder()

        // When
        router.ServeHTTP(w, req)

        // Then
        assert.Equal(t, http.StatusCreated, w.Code)

        var response UserResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)
        assert.Equal(t, "test@example.com", response.Email)
    })
}
```

### **🎭 Test Helpers & Builders**

#### **Test Data Builders**

```go
// Builder pattern for test data
type UserBuilder struct {
    email    string
    username string
    id       string
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        email:    "default@example.com",
        username: "defaultuser",
        id:       "default-id",
    }
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.email = email
    return b
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
    b.username = username
    return b
}

func (b *UserBuilder) Build() *entities.User {
    user, _ := entities.NewUser(b.id, b.email, b.username)
    return user
}

// Usage in tests
func TestUserService_CreateUser(t *testing.T) {
    // Given
    user := NewUserBuilder().
        WithEmail("specific@example.com").
        WithUsername("specificuser").
        Build()

    // Test implementation...
}
```

---

## 📖 **Documentation Standards**

### **📋 Documentation Types**

#### **1. API Documentation**

- Use GoDoc format for all public APIs
- Include examples for complex functions
- Document error conditions and return values
- Add usage examples for packages

#### **2. Architecture Decision Records (ADRs)**

```markdown
# ADR-001: Use Clean Architecture Pattern

## Status: Accepted

## Context

We need a way to organize code that scales from small teams to enterprise applications while maintaining testability and flexibility.

## Decision

We will use Clean Architecture with Domain-Driven Design principles.

## Consequences

- ✅ Better testability through dependency inversion
- ✅ Framework independence
- ✅ Clear separation of concerns
- ❌ Additional complexity for small projects
- ❌ Learning curve for developers new to DDD
```

#### **3. Code Examples**

- Every package should have working examples
- Examples should demonstrate real use cases
- Include both simple and complex scenarios
- Show error handling patterns

### **📝 Writing Guidelines**

#### **Documentation Style**

- Use clear, concise language
- Write for developers at different experience levels
- Include code examples for complex concepts
- Use bullet points and numbered lists for clarity
- Add diagrams for architectural concepts

#### **Code Comments**

```go
// ✅ GOOD: Explains the "why" not just the "what"
// validateBusinessRules ensures user data meets domain requirements.
// This validation happens before persistence to catch business rule
// violations early and provide meaningful error messages to users.
func (u *User) validateBusinessRules() error {
    // Implementation...
}

// ❌ BAD: States the obvious
// validateBusinessRules validates business rules
func (u *User) validateBusinessRules() error {
    // Implementation...
}
```

---

## 🔄 **Pull Request Process**

### **📋 Pre-PR Checklist**

```bash
# Before submitting your PR, ensure:
□ Code follows all coding standards
□ All tests pass: `just test`
□ Linting passes: `just lint`
□ Documentation is updated
□ CHANGELOG.md is updated (if applicable)
□ Commit messages follow conventions
□ Branch is up-to-date with main
```

### **📝 PR Template**

When creating a PR, please use this template:

```markdown
## Description

Brief description of what this PR does and why.

## Type of Change

- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring

## Related Issues

Fixes #123
Closes #456

## Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed
- [ ] All tests pass

## Documentation

- [ ] Code comments updated
- [ ] README.md updated (if applicable)
- [ ] API documentation updated
- [ ] Architecture docs updated (if applicable)

## Screenshots/Demos

(If applicable, add screenshots or demo videos)

## Checklist

- [ ] Code follows project style guidelines
- [ ] Self-review of code completed
- [ ] Code is commented where necessary
- [ ] Documentation reflects changes
- [ ] No new warnings or errors introduced
- [ ] Tests cover edge cases
```

### **🔍 Review Process**

#### **Review Criteria**

1. **Architecture Compliance**
   - Follows Clean Architecture principles
   - Respects domain boundaries
   - Uses proper dependency direction

2. **Code Quality**
   - Passes all linting rules
   - Has appropriate test coverage
   - Follows coding standards

3. **Documentation**
   - Public APIs are documented
   - Complex logic is explained
   - Examples are provided where helpful

4. **Testing**
   - Tests cover happy path and edge cases
   - Tests are isolated and deterministic
   - Integration tests verify component interactions

#### **Review Timeline**

- **Initial Response**: Within 24 hours
- **Full Review**: Within 72 hours
- **Follow-up**: Within 24 hours of updates

### **🚀 Merge Requirements**

- ✅ All CI checks pass
- ✅ At least one approving review from maintainer
- ✅ All conversations resolved
- ✅ Branch is up-to-date with main
- ✅ No merge conflicts

---

## 🏗️ **Architecture Guidelines**

### **🎯 Clean Architecture Principles**

#### **Dependency Rule**

```
Outer layers can depend on inner layers, never the reverse:

Infrastructure → Application → Domain
```

#### **Layer Responsibilities**

**Domain Layer (Core)**

- Business entities and value objects
- Domain services and business logic
- Repository interfaces (contracts)
- Domain events and specifications
- **NO** dependencies on infrastructure

**Application Layer (Use Cases)**

- Application services (orchestration)
- Use case implementations
- DTO mappings and validation
- **CAN** depend on domain layer

**Infrastructure Layer (Adapters)**

- Repository implementations
- External service integrations
- Database access and persistence
- **CAN** depend on domain and application layers

### **🔧 Component Organization**

#### **Package Structure**

```
internal/
├── domain/                 # Core business logic
│   ├── entities/          # Business entities
│   ├── repositories/      # Repository interfaces
│   ├── services/          # Domain services
│   ├── values/            # Value objects
│   └── errors/            # Domain-specific errors
├── application/           # Use cases and orchestration
│   ├── handlers/          # HTTP handlers
│   ├── dto/               # Data transfer objects
│   └── services/          # Application services
└── infrastructure/        # External concerns
    ├── persistence/       # Database implementations
    ├── http/              # HTTP clients
    └── messaging/         # Message queue implementations
```

#### **Naming Conventions**

- **Packages**: lowercase, single word when possible
- **Files**: lowercase with underscores (`user_service.go`)
- **Types**: PascalCase (`UserService`)
- **Functions**: PascalCase for public, camelCase for private
- **Constants**: UPPER_SNAKE_CASE
- **Variables**: camelCase

---

## 🎉 **Recognition & Rewards**

### **🏆 Contributor Recognition**

#### **Contributor Levels**

1. **First-Time Contributor** 🌟
   - GitHub badge on profile
   - Welcome package with stickers
   - Mention in release notes

2. **Regular Contributor** ⭐
   - Listed in CONTRIBUTORS.md
   - Access to contributor Discord channel
   - Early access to new features

3. **Core Contributor** 🚀
   - Repository collaborator access
   - Voting rights on major decisions
   - Conference speaking opportunities

4. **Maintainer** 👑
   - Full repository access
   - Decision-making authority
   - Travel sponsorship for conferences

### **🎁 Rewards Program**

#### **Contribution Types & Rewards**

- **Bug Fixes**: GitHub sponsor shoutout + stickers
- **New Features**: T-shirt + LinkedIn recommendation
- **Documentation**: Certificate of contribution
- **Major Contributions**: Hoodie + conference ticket sponsorship

#### **Special Recognition**

- **Monthly Contributor Award**: $100 gift card
- **Annual Contributor Award**: Conference speaking slot + travel sponsorship
- **Lifetime Achievement**: Custom trophy + permanent recognition

### **📢 Community Involvement**

#### **Communication Channels**

- **GitHub Discussions**: Design discussions and Q&A
- **Discord Server**: Real-time contributor chat
- **Monthly Calls**: Contributor sync meetings
- **Annual Summit**: In-person contributor meetup

#### **Learning Opportunities**

- **Mentorship Program**: Pair new contributors with experienced ones
- **Workshop Sessions**: Architecture and Go best practices
- **Code Review Sessions**: Learn from reviewing others' code
- **Conference Sponsorship**: Support for speaking at events

---

## 📞 **Getting Help**

### **🤔 Questions & Support**

#### **Where to Ask**

1. **GitHub Discussions**: Design questions, feature discussions
2. **GitHub Issues**: Bug reports, feature requests
3. **Discord #contributors**: Real-time help and chat
4. **Email**: template-arch-lint@lars.software for private matters

#### **Response Times**

- **Discord**: Usually within a few hours during business hours
- **GitHub Issues**: Within 24 hours for bugs, 72 hours for features
- **Email**: Within 48 hours

### **📚 Learning Resources**

#### **Architecture & Design**

- [Clean Architecture by Robert Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design Quickly](https://www.infoq.com/minibooks/domain-driven-design-quickly/)
- [Go Clean Architecture Examples](https://github.com/bxcodec/go-clean-arch)

#### **Go Best Practices**

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

#### **Testing in Go**

- [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests/)
- [Advanced Go Testing](https://about.sourcegraph.com/go/advanced-testing-in-go)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)

---

## 🙏 **Thank You**

Your contributions make this project better for the entire Go community. Whether you're fixing a typo, adding a feature, or helping other contributors, every contribution matters.

Together, we're building a future where architectural excellence and code quality are automatic for every Go developer! 🚀

---

**Made with ❤️ by the Go community**

_Last updated: $(date)_
