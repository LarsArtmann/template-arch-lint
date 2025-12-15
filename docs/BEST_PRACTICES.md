# 🏆 Best Practices Guide - Template Architecture Lint

This guide consolidates enterprise-grade best practices for Go development, Clean Architecture, and code quality enforcement demonstrated in this template project.

## 📋 Table of Contents

1. [🏗️ Clean Architecture Principles](#️-clean-architecture-principles)
2. [🔍 Code Quality Standards](#-code-quality-standards)
3. [🧪 Testing Strategy](#-testing-strategy)
4. [📊 Performance Optimization](#-performance-optimization)
5. [🔧 Development Workflow](#-development-workflow)
6. [🚀 CI/CD Best Practices](#-cicd-best-practices)
7. [📚 Documentation Standards](#-documentation-standards)
8. [🛡️ Security Practices](#️-security-practices)

---

## 🏗️ Clean Architecture Principles

### Layer Separation Rules

**✅ DO:**

```go
// Domain layer - Pure business logic
package entities

import (
    "time"
    "github.com/yourproject/internal/domain/values"  // OK - domain to domain
)

type User struct {
    ID        values.UserID
    Email     string
    CreatedAt time.Time  // OK - standard library
}
```

**❌ DON'T:**

```go
// Domain layer with infrastructure dependency
package entities

import (
    "database/sql"        // ❌ Domain cannot depend on infrastructure
    "github.com/gin-gonic/gin"  // ❌ Domain cannot depend on web framework
)

type User struct {
    ID string
    DB *sql.DB  // ❌ Database dependency in domain
}
```

### Dependency Inversion Pattern

**✅ Correct Pattern:**

```go
// 1. Interface in domain layer
package repositories

type UserRepository interface {
    Save(ctx context.Context, user *entities.User) error
    FindByID(ctx context.Context, id values.UserID) (*entities.User, error)
}

// 2. Implementation in infrastructure layer
package persistence

type SQLUserRepository struct {
    db *sql.DB
}

func (r *SQLUserRepository) Save(ctx context.Context, user *entities.User) error {
    // Implementation details
}

// 3. Injection in main/container
func main() {
    db := setupDatabase()
    userRepo := persistence.NewSQLUserRepository(db)  // Concrete implementation
    userService := services.NewUserService(userRepo)  // Depends on interface
}
```

### Value Objects Best Practices

**✅ Immutable Value Objects:**

```go
type Email struct {
    value string
}

func NewEmail(email string) (Email, error) {
    if err := validateEmail(email); err != nil {
        return Email{}, fmt.Errorf("invalid email: %w", err)
    }
    return Email{value: strings.ToLower(email)}, nil
}

func (e Email) String() string {
    return e.value  // Read-only access
}

func (e Email) Domain() string {
    parts := strings.Split(e.value, "@")
    return parts[1]
}
```

### Entity Best Practices

**✅ Entities with Business Logic:**

```go
type User struct {
    ID        values.UserID
    Email     values.Email
    Status    UserStatus
    CreatedAt time.Time
    UpdatedAt time.Time
}

func NewUser(id values.UserID, email values.Email) (*User, error) {
    // Business rules and validation
    if id.IsEmpty() {
        return nil, errors.New("user ID cannot be empty")
    }

    now := time.Now().UTC()
    return &User{
        ID:        id,
        Email:     email,
        Status:    StatusActive,
        CreatedAt: now,
        UpdatedAt: now,
    }, nil
}

func (u *User) Deactivate() error {
    if u.Status == StatusDeactivated {
        return errors.New("user already deactivated")
    }

    u.Status = StatusDeactivated
    u.UpdatedAt = time.Now().UTC()
    return nil
}
```

---

## 🔍 Code Quality Standards

### Error Handling Excellence

**✅ Comprehensive Error Wrapping:**

```go
func (s *UserService) CreateUser(ctx context.Context, email string) (*entities.User, error) {
    // Validate input
    emailVO, err := values.NewEmail(email)
    if err != nil {
        return nil, fmt.Errorf("invalid email provided: %w", err)
    }

    // Check business rules
    exists, err := s.userRepo.ExistsByEmail(ctx, emailVO)
    if err != nil {
        return nil, fmt.Errorf("failed to check user existence: %w", err)
    }

    if exists {
        return nil, fmt.Errorf("user with email %s already exists", email)
    }

    // Create entity
    user, err := entities.NewUser(s.generateUserID(), emailVO)
    if err != nil {
        return nil, fmt.Errorf("failed to create user entity: %w", err)
    }

    // Persist
    if err := s.userRepo.Save(ctx, user); err != nil {
        return nil, fmt.Errorf("failed to save user: %w", err)
    }

    return user, nil
}
```

### Function Design Principles

**✅ Single Responsibility Functions:**

```go
// Good: Single purpose, clear responsibility
func (s *UserService) validateUserEmail(email string) error {
    if email == "" {
        return errors.New("email cannot be empty")
    }

    if !emailRegex.MatchString(email) {
        return errors.New("email format is invalid")
    }

    if len(email) > maxEmailLength {
        return errors.New("email too long")
    }

    return nil
}

// Good: Pure function with clear inputs/outputs
func calculateUserAge(birthDate time.Time) int {
    now := time.Now()
    age := now.Year() - birthDate.Year()

    if now.YearDay() < birthDate.YearDay() {
        age--
    }

    return age
}
```

### Type Safety Best Practices

**✅ Strong Typing Over Primitives:**

```go
// Good: Type-safe domain concepts
type UserID struct{ value string }
type Email struct{ value string }
type Amount struct{ cents int64 }

func ProcessPayment(userID UserID, amount Amount) error {
    // Cannot accidentally mix up parameters
    return paymentService.Process(userID, amount)
}

// Bad: Primitive obsession
func ProcessPayment(userID string, amount int64) error {
    // Easy to mix up parameters
    return paymentService.Process(amount, userID)  // ❌ Wrong order
}
```

### Concurrency Best Practices

**✅ Safe Concurrent Patterns:**

```go
type UserCache struct {
    users map[string]*entities.User
    mutex sync.RWMutex
}

func (c *UserCache) Get(id string) (*entities.User, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    user, exists := c.users[id]
    if !exists {
        return nil, false
    }

    // Return copy to prevent external mutation
    userCopy := *user
    return &userCopy, true
}

func (c *UserCache) Set(user *entities.User) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    // Store copy to prevent external mutation
    userCopy := *user
    c.users[user.ID.String()] = &userCopy
}
```

---

## 🧪 Testing Strategy

### Test Organization Hierarchy

```
1. Unit Tests (Fast, Isolated)
   └── Domain entities, value objects, services

2. Integration Tests (Medium, Real Dependencies)
   └── Repository implementations, HTTP handlers

3. End-to-End Tests (Slow, Full System)
   └── Complete user workflows

4. Performance Tests (Benchmarks)
   └── Critical path performance validation
```

### Domain Testing Best Practices

**✅ BDD-Style Domain Tests:**

```go
var _ = Describe("User Entity", func() {
    Describe("Creating a new user", func() {
        Context("with valid data", func() {
            It("should create user successfully", func() {
                userID, _ := values.NewUserID("user-123")
                email, _ := values.NewEmail("test@example.com")

                user, err := entities.NewUser(userID, email)

                Expect(err).NotTo(HaveOccurred())
                Expect(user.ID).To(Equal(userID))
                Expect(user.Email).To(Equal(email))
                Expect(user.Status).To(Equal(entities.StatusActive))
            })
        })

        Context("with invalid data", func() {
            It("should reject empty user ID", func() {
                emptyID := values.UserID{}
                email, _ := values.NewEmail("test@example.com")

                _, err := entities.NewUser(emptyID, email)

                Expect(err).To(HaveOccurred())
                Expect(err.Error()).To(ContainSubstring("user ID cannot be empty"))
            })
        })
    })
})
```

### Repository Testing Patterns

**✅ Test Against Real Implementations:**

```go
var _ = Describe("SQLUserRepository", func() {
    var (
        repo repositories.UserRepository
        db   *sql.DB
        ctx  context.Context
    )

    BeforeEach(func() {
        // Use real SQLite database for integration tests
        db = setupTestDatabase()
        repo = persistence.NewSQLUserRepository(db, logger)
        ctx = context.Background()
    })

    AfterEach(func() {
        db.Close()
    })

    Describe("Save and FindByID", func() {
        It("should persist and retrieve user", func() {
            user := createTestUser()

            err := repo.Save(ctx, user)
            Expect(err).NotTo(HaveOccurred())

            retrieved, err := repo.FindByID(ctx, user.ID)
            Expect(err).NotTo(HaveOccurred())
            Expect(retrieved.Email).To(Equal(user.Email))
        })
    })
})
```

### HTTP Handler Testing

**✅ Complete Request/Response Testing:**

```go
func TestUserHandler_CreateUser(t *testing.T) {
    // Setup
    userService := &mockUserService{}
    handler := handlers.NewUserHandler(userService)
    router := setupTestRouter(handler)

    tests := []struct {
        name           string
        body           string
        expectedStatus int
        expectedBody   string
    }{
        {
            name:           "valid user creation",
            body:           `{"email": "test@example.com", "name": "Test User"}`,
            expectedStatus: http.StatusCreated,
            expectedBody:   "user created successfully",
        },
        {
            name:           "invalid email",
            body:           `{"email": "invalid", "name": "Test User"}`,
            expectedStatus: http.StatusBadRequest,
            expectedBody:   "invalid email",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("POST", "/users", strings.NewReader(tt.body))
            req.Header.Set("Content-Type", "application/json")
            resp := httptest.NewRecorder()

            router.ServeHTTP(resp, req)

            assert.Equal(t, tt.expectedStatus, resp.Code)
            assert.Contains(t, resp.Body.String(), tt.expectedBody)
        })
    }
}
```

---

## 📊 Performance Optimization

### Memory Management Best Practices

**✅ Efficient Memory Patterns:**

```go
// Pre-allocate slices when size is known
func ProcessUsers(userCount int) []*entities.User {
    users := make([]*entities.User, 0, userCount)  // Pre-allocate capacity

    for i := 0; i < userCount; i++ {
        user := createUser(i)
        users = append(users, user)
    }

    return users
}

// Use object pools for frequently created objects
var userPool = sync.Pool{
    New: func() interface{} {
        return &entities.User{}
    },
}

func ProcessUserWithPool() {
    user := userPool.Get().(*entities.User)
    defer userPool.Put(user)

    // Use user object
    processUser(user)
}

// Stream processing for large datasets
func ProcessLargeDataset(data io.Reader) error {
    scanner := bufio.NewScanner(data)
    scanner.Split(bufio.ScanLines)

    for scanner.Scan() {
        if err := processLine(scanner.Text()); err != nil {
            return fmt.Errorf("failed to process line: %w", err)
        }
    }

    return scanner.Err()
}
```

### Database Optimization Patterns

**✅ Efficient Query Patterns:**

```go
// Batch operations instead of N+1 queries
func (r *SQLUserRepository) SaveBatch(ctx context.Context, users []*entities.User) error {
    if len(users) == 0 {
        return nil
    }

    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO users (id, email, name, created_at)
        VALUES (?, ?, ?, ?)`)
    if err != nil {
        return fmt.Errorf("failed to prepare statement: %w", err)
    }
    defer stmt.Close()

    for _, user := range users {
        if _, err := stmt.ExecContext(ctx, user.ID, user.Email, user.Name, user.CreatedAt); err != nil {
            return fmt.Errorf("failed to insert user %s: %w", user.ID, err)
        }
    }

    return tx.Commit()
}

// Use proper indexes
func (r *SQLUserRepository) FindByEmailPrefix(ctx context.Context, prefix string) ([]*entities.User, error) {
    // Ensure email column has index: CREATE INDEX idx_users_email ON users(email)
    rows, err := r.db.QueryContext(ctx, `
        SELECT id, email, name, created_at
        FROM users
        WHERE email LIKE ?
        ORDER BY email
        LIMIT 100`, prefix+"%")

    if err != nil {
        return nil, fmt.Errorf("failed to query users: %w", err)
    }
    defer rows.Close()

    return r.scanUsers(rows)
}
```

### Benchmarking Best Practices

**✅ Comprehensive Benchmarks:**

```go
func BenchmarkUserService_CreateUser(b *testing.B) {
    service := setupBenchmarkService(b)
    ctx := context.Background()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        email := fmt.Sprintf("user%d@example.com", i)
        _, err := service.CreateUser(ctx, email)
        if err != nil {
            b.Fatalf("CreateUser failed: %v", err)
        }
    }
}

func BenchmarkUserRepository_BatchInsert(b *testing.B) {
    repo := setupBenchmarkRepository(b)
    ctx := context.Background()

    batchSizes := []int{10, 100, 1000}

    for _, size := range batchSizes {
        b.Run(fmt.Sprintf("BatchSize%d", size), func(b *testing.B) {
            users := createTestUsers(size)

            b.ResetTimer()
            b.ReportAllocs()

            for i := 0; i < b.N; i++ {
                if err := repo.SaveBatch(ctx, users); err != nil {
                    b.Fatalf("SaveBatch failed: %v", err)
                }
            }
        })
    }
}
```

---

## 🔧 Development Workflow

### Git Workflow Best Practices

**✅ Conventional Commits:**

```bash
# Use clear, conventional commit messages
git commit -m "feat(auth): add user registration with email validation

- Implement user registration endpoint
- Add email validation with regex
- Include duplicate email check
- Add comprehensive test coverage

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"

# Use feature branches
git checkout -b feature/user-registration
git checkout -b fix/email-validation-bug
git checkout -b refactor/user-service-cleanup
```

**✅ Branch Management:**

```bash
# Use git-town for branch management
git town append feature/user-auth    # Create feature branch
git town sync                        # Sync with remote
git town ship                        # Merge and cleanup

# Keep commits atomic and focused
git add internal/domain/entities/user.go
git commit -m "feat(domain): add User entity with business validation"

git add internal/domain/services/user_service.go
git commit -m "feat(services): implement UserService with creation logic"
```

### Code Review Guidelines

**✅ Review Checklist:**

1. **Architecture Compliance**
   - [ ] Domain layer doesn't import infrastructure
   - [ ] Interfaces defined in domain, implemented in infrastructure
   - [ ] No circular dependencies

2. **Code Quality**
   - [ ] Functions under 50 lines
   - [ ] Cyclomatic complexity under 10
   - [ ] All errors properly wrapped
   - [ ] No `interface{}` or `panic()` usage

3. **Testing**
   - [ ] Unit tests for business logic
   - [ ] Integration tests for repositories
   - [ ] Edge cases covered
   - [ ] Performance considerations

4. **Documentation**
   - [ ] Public functions documented
   - [ ] Complex business rules explained
   - [ ] README updated if needed

### Development Environment Setup

**✅ Essential Tools Configuration:**

```bash
# Install core tools
brew install just golangci-lint
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/fdaines/arch-go@latest

# Setup pre-commit hooks
pre-commit install

# Configure VS Code settings.json
{
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.formatTool": "goimports",
    "go.testFlags": ["-v", "-race"],
    "files.associations": {
        "*.templ": "html"
    }
}

# Setup justfile aliases
alias jb="just build"
alias jt="just test"
alias jl="just lint"
alias jr="just run"
```

---

## 🚀 CI/CD Best Practices

### Pipeline Design Principles

**✅ Fast Feedback Loops:**

```yaml
# Stage 1: Fast quality checks (< 2 minutes)
- Linting (golangci-lint, go-arch-lint)
- Unit tests
- Security scanning (gosec)

# Stage 2: Integration validation (< 5 minutes)
- Integration tests
- Build verification
- Template generation

# Stage 3: Comprehensive validation (< 15 minutes)
- E2E tests
- Performance benchmarks
- Coverage analysis
- Container building
```

**✅ Quality Gates:**

```yaml
quality_gates:
  required_checks:
    - architecture_compliance: pass
    - code_quality_score: ">= 90%"
    - test_coverage: ">= 80%"
    - security_scan: pass
    - performance_regression: none

  blocking_conditions:
    - panic_usage_detected: true
    - interface_any_usage: true
    - circular_dependencies: true
    - missing_error_wrapping: true
```

### Deployment Best Practices

**✅ Blue-Green Deployment Pattern:**

```yaml
# Health checks before traffic switch
health_checks:
  - endpoint: /health/live
    expected_status: 200
    timeout: 5s
    retries: 3

  - endpoint: /health/ready
    expected_status: 200
    timeout: 10s
    retries: 5

# Performance validation
performance_checks:
  - endpoint: /performance/stats
    max_response_time: 100ms
    max_memory_usage: 512MB
    max_goroutines: 1000

# Rollback triggers
rollback_conditions:
  - error_rate: "> 1%"
  - response_time_p95: "> 500ms"
  - memory_usage: "> 1GB"
```

---

## 📚 Documentation Standards

### Code Documentation Best Practices

**✅ Self-Documenting Code:**

```go
// UserRegistrationService handles the complete user registration workflow
// including validation, duplication checks, and email verification.
type UserRegistrationService struct {
    userRepo        repositories.UserRepository
    emailService    EmailService
    eventPublisher  EventPublisher
}

// RegisterUser creates a new user account with email verification.
// It performs the following steps:
//   1. Validates user input and business rules
//   2. Checks for existing users with the same email
//   3. Creates user entity with generated ID
//   4. Persists user to database
//   5. Sends verification email
//   6. Publishes user registration event
//
// Returns the created user or an error if registration fails.
// Common errors include invalid email, duplicate email, or persistence failures.
func (s *UserRegistrationService) RegisterUser(ctx context.Context, req RegistrationRequest) (*entities.User, error) {
    // Implementation with clear step-by-step logic
}
```

### Architecture Documentation

**✅ Decision Records:**

```markdown
# ADR-001: Domain Layer Isolation

## Status: Accepted

## Context
We need to ensure domain business logic remains pure and testable
without external dependencies.

## Decision
Domain layer (entities, services, value objects) cannot import:
- Infrastructure packages (database, HTTP, external APIs)
- Framework-specific code (gin, echo, etc.)
- Implementation details

## Consequences
- ✅ Domain logic is framework-agnostic and easily testable
- ✅ Business rules are centralized and consistent
- ❌ Requires dependency inversion patterns
- ❌ More initial setup complexity

## Enforcement
- go-arch-lint configuration prevents violations
- CI/CD pipeline blocks merges with violations
- Code review checklist includes architecture compliance
```

---

## 🛡️ Security Practices

### Secure Coding Standards

**✅ Input Validation & Sanitization:**

```go
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*entities.User, error) {
    // Validate all inputs
    if err := s.validateCreateUserRequest(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }

    // Sanitize input data
    email := strings.TrimSpace(strings.ToLower(req.Email))
    name := strings.TrimSpace(req.Name)

    // Business rule validation
    if err := s.enforceBusinessRules(email, name); err != nil {
        return nil, fmt.Errorf("business rule violation: %w", err)
    }

    // Create with validated data
    return s.createUserEntity(email, name)
}

func (s *UserService) validateCreateUserRequest(req CreateUserRequest) error {
    var errs []string

    if req.Email == "" {
        errs = append(errs, "email is required")
    } else if !isValidEmail(req.Email) {
        errs = append(errs, "email format is invalid")
    }

    if req.Name == "" {
        errs = append(errs, "name is required")
    } else if len(req.Name) > maxNameLength {
        errs = append(errs, "name too long")
    }

    if len(errs) > 0 {
        return fmt.Errorf("validation errors: %s", strings.Join(errs, ", "))
    }

    return nil
}
```

**✅ Database Security:**

```go
// Always use parameterized queries
func (r *SQLUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
    row := r.db.QueryRowContext(ctx, `
        SELECT id, email, name, created_at, updated_at
        FROM users
        WHERE email = ?`, email)  // ✅ Parameterized query

    var user entities.User
    if err := row.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, repositories.ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to scan user: %w", err)
    }

    return &user, nil
}

// Connection security
func setupDatabase() *sql.DB {
    db, err := sql.Open("sqlite3", "file:app.db?_journal_mode=WAL&_foreign_keys=on")
    if err != nil {
        log.Fatal(err)
    }

    // Set connection limits
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(10)
    db.SetConnMaxLifetime(5 * time.Minute)

    return db
}
```

### Secret Management

**✅ Environment-Based Configuration:**

```go
type Config struct {
    Database struct {
        DSN string `env:"DATABASE_DSN,required"`
    }

    JWT struct {
        SecretKey string `env:"JWT_SECRET_KEY,required"`
    }

    Email struct {
        APIKey string `env:"EMAIL_API_KEY,required"`
    }
}

// Never log secrets
func (c *Config) String() string {
    return fmt.Sprintf("Config{Database: %s, JWT: [REDACTED], Email: [REDACTED]}",
                      maskDSN(c.Database.DSN))
}

func maskDSN(dsn string) string {
    // Mask password in DSN for logging
    re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
    return re.ReplaceAllString(dsn, "://$1:[REDACTED]@")
}
```

---

## 🎯 Summary & Quick Reference

### Essential Commands Checklist

```bash
# Daily development workflow
just build                 # Build application
just lint                  # Run all linting
just test                  # Run all tests
just coverage 80           # Check coverage threshold
just run                   # Start development server

# Quality assurance
just lint-arch             # Architecture boundaries
just lint-code             # Code quality rules
just lint-security         # Security scanning
just fix                   # Auto-fix issues

# Performance monitoring
just bench                 # Run benchmarks
just profile-cpu           # CPU profiling
just profile-heap          # Memory profiling
```

### Architecture Quick Validation

```bash
# Check for common violations
grep -r "interface{}" internal/domain/     # Should be empty
grep -r "panic(" internal/                 # Should be empty
grep -r "github.com/gin" internal/domain/ # Should be empty

# Verify layer separation
just lint-arch              # Must pass
go list -deps ./internal/domain/entities | grep infrastructure  # Should be empty
```

### Code Quality Metrics

- **Function Length**: Max 50 lines
- **Cyclomatic Complexity**: Max 10
- **Test Coverage**: Min 80%
- **Architecture Violations**: 0
- **Security Issues**: 0
- **Performance Regressions**: 0

---

This best practices guide represents battle-tested patterns for enterprise Go development. Follow these practices to build maintainable, scalable, and robust applications.

For specific implementation examples, see:

- 📁 **[Example Project](../example/)** - Working demonstration
- 📊 **[Profiling Guide](./PROFILING.md)** - Performance optimization
- 📚 **[Usage Guide](./USAGE.md)** - Complete instructions
