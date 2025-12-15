# Error Handling Best Practices

## Overview

This document establishes comprehensive error handling patterns for Go applications following Clean Architecture principles. Proper error handling is crucial for maintainable, debuggable, and reliable software.

## Core Principles

### 1. Errors are Values

Treat errors as first-class values, not exceptional conditions.

```go
// Good - Error as return value
func GetUser(id UserID) (*User, error) {
    user, err := repository.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user %s: %w", id, err)
    }
    return user, nil
}

// Bad - Using panic for business logic
func GetUser(id UserID) *User {
    user, err := repository.FindByID(id)
    if err != nil {
        panic("user not found") // Never do this
    }
    return user
}
```

### 2. Error Context and Wrapping

Always provide context when wrapping errors using `fmt.Errorf` with `%w` verb.

```go
// Good - Contextual error wrapping
func ProcessUserRegistration(req RegistrationRequest) error {
    if err := validateRegistrationRequest(req); err != nil {
        return fmt.Errorf("registration validation failed for %s: %w", req.Email, err)
    }

    user, err := createUserAccount(req)
    if err != nil {
        return fmt.Errorf("failed to create account for %s: %w", req.Email, err)
    }

    if err := sendWelcomeEmail(user); err != nil {
        // Note: This might not be fatal - consider logging instead
        return fmt.Errorf("account created but welcome email failed for %s: %w", user.Email, err)
    }

    return nil
}

// Bad - Losing error context
func ProcessUserRegistration(req RegistrationRequest) error {
    err := validateRegistrationRequest(req)
    if err != nil {
        return err // Lost: where validation failed
    }

    _, err = createUserAccount(req)
    if err != nil {
        return errors.New("failed") // Lost: original error details
    }

    return nil
}
```

### 3. Domain-Specific Error Types

Create meaningful error types for your domain.

```go
// Domain error types
type DomainError struct {
    Code    string
    Message string
    Cause   error
}

func (e DomainError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e DomainError) Unwrap() error {
    return e.Cause
}

// Specific domain errors
var (
    ErrUserNotFound = DomainError{
        Code:    "USER_NOT_FOUND",
        Message: "user does not exist",
    }

    ErrInvalidEmail = DomainError{
        Code:    "INVALID_EMAIL",
        Message: "email address is not valid",
    }

    ErrEmailAlreadyExists = DomainError{
        Code:    "EMAIL_EXISTS",
        Message: "email address is already registered",
    }
)

// Usage
func CreateUser(req CreateUserRequest) (*User, error) {
    if !isValidEmail(req.Email) {
        return nil, ErrInvalidEmail
    }

    existing, err := repository.FindByEmail(req.Email)
    if err != nil && !errors.Is(err, ErrUserNotFound) {
        return nil, fmt.Errorf("failed to check existing user: %w", err)
    }
    if existing != nil {
        return nil, ErrEmailAlreadyExists
    }

    // ... creation logic
    return user, nil
}
```

## Layer-Specific Error Handling

### Domain Layer Errors

Domain layer should define business-specific error types.

```go
// internal/domain/errors/user_errors.go
package errors

import "errors"

var (
    // Business rule violations
    ErrUserNotFound        = errors.New("user not found")
    ErrInvalidEmail        = errors.New("invalid email address")
    ErrWeakPassword        = errors.New("password does not meet security requirements")
    ErrUsernameUnavailable = errors.New("username is already taken")

    // Domain invariant violations
    ErrUserAlreadyActive   = errors.New("user is already active")
    ErrUserNotActive       = errors.New("user is not active")
    ErrInsufficientPermissions = errors.New("user lacks required permissions")
)

// UserValidationError for detailed validation feedback
type UserValidationError struct {
    Field   string
    Value   string
    Message string
}

func (e UserValidationError) Error() string {
    return fmt.Sprintf("validation failed for field '%s' with value '%s': %s",
        e.Field, e.Value, e.Message)
}

// Multiple validation errors
type ValidationErrors []UserValidationError

func (ve ValidationErrors) Error() string {
    if len(ve) == 0 {
        return "no validation errors"
    }
    if len(ve) == 1 {
        return ve[0].Error()
    }

    var msgs []string
    for _, err := range ve {
        msgs = append(msgs, err.Error())
    }
    return fmt.Sprintf("multiple validation errors: %s", strings.Join(msgs, "; "))
}
```

### Application Layer Error Handling

Application layer translates domain errors to appropriate responses.

```go
// internal/application/handlers/user_handler.go
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.respondWithError(c, http.StatusBadRequest, "invalid request format", err)
        return
    }

    user, err := h.userService.CreateUser(req)
    if err != nil {
        h.handleUserServiceError(c, err)
        return
    }

    h.respondWithSuccess(c, http.StatusCreated, user)
}

func (h *UserHandler) handleUserServiceError(c *gin.Context, err error) {
    // Check for specific domain errors
    var validationErrors ValidationErrors
    if errors.As(err, &validationErrors) {
        h.respondWithValidationErrors(c, validationErrors)
        return
    }

    switch {
    case errors.Is(err, ErrUserNotFound):
        h.respondWithError(c, http.StatusNotFound, "user not found", err)
    case errors.Is(err, ErrInvalidEmail):
        h.respondWithError(c, http.StatusBadRequest, "invalid email address", err)
    case errors.Is(err, ErrEmailAlreadyExists):
        h.respondWithError(c, http.StatusConflict, "email already registered", err)
    default:
        // Log detailed error but don't expose internal details
        h.logger.Error("unexpected error in user service", "error", err)
        h.respondWithError(c, http.StatusInternalServerError, "internal server error", nil)
    }
}

func (h *UserHandler) respondWithError(c *gin.Context, status int, message string, err error) {
    response := ErrorResponse{
        Error: ErrorDetail{
            Code:    http.StatusText(status),
            Message: message,
        },
        Timestamp: time.Now(),
        Path:      c.Request.URL.Path,
    }

    // Add correlation ID if available
    if correlationID := c.GetString("correlation_id"); correlationID != "" {
        response.CorrelationID = correlationID
    }

    // Log error with correlation ID
    if err != nil {
        h.logger.Error("request error",
            "error", err,
            "correlation_id", response.CorrelationID,
            "path", response.Path,
            "status", status,
        )
    }

    c.JSON(status, response)
}

func (h *UserHandler) respondWithValidationErrors(c *gin.Context, validationErrors ValidationErrors) {
    response := ValidationErrorResponse{
        Error: ErrorDetail{
            Code:    "VALIDATION_FAILED",
            Message: "request validation failed",
        },
        ValidationErrors: make([]FieldError, len(validationErrors)),
        Timestamp:        time.Now(),
        Path:             c.Request.URL.Path,
    }

    for i, ve := range validationErrors {
        response.ValidationErrors[i] = FieldError{
            Field:   ve.Field,
            Value:   ve.Value,
            Message: ve.Message,
        }
    }

    c.JSON(http.StatusBadRequest, response)
}
```

### Infrastructure Layer Error Handling

Infrastructure layer handles external system errors and translates them to domain errors.

```go
// internal/infrastructure/persistence/user_repository_sql.go
func (r *UserRepositorySQL) FindByID(ctx context.Context, id UserID) (*User, error) {
    const query = `SELECT id, email, username, created_at, updated_at FROM users WHERE id = $1`

    var user User
    err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
        &user.ID,
        &user.Email,
        &user.Username,
        &user.CreatedAt,
        &user.UpdatedAt,
    )

    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to query user by ID %s: %w", id, err)
    }

    return &user, nil
}

func (r *UserRepositorySQL) Create(ctx context.Context, user *User) error {
    const query = `
        INSERT INTO users (id, email, username, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `

    _, err := r.db.ExecContext(ctx, query,
        user.ID.String(),
        user.Email.String(),
        user.Username.String(),
        user.CreatedAt,
        user.UpdatedAt,
    )

    if err != nil {
        // Check for constraint violations
        if isUniqueConstraintViolation(err, "users_email_key") {
            return ErrEmailAlreadyExists
        }
        if isUniqueConstraintViolation(err, "users_username_key") {
            return ErrUsernameUnavailable
        }

        return fmt.Errorf("failed to create user %s: %w", user.Email, err)
    }

    return nil
}

// Helper function to check database constraint violations
func isUniqueConstraintViolation(err error, constraintName string) bool {
    var pgErr *pq.Error
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505" && strings.Contains(pgErr.Constraint, constraintName)
    }
    return false
}
```

## Error Response Patterns

### Standardized Error Response Structure

```go
// Standard error response
type ErrorResponse struct {
    Error         ErrorDetail `json:"error"`
    Timestamp     time.Time   `json:"timestamp"`
    Path          string      `json:"path"`
    CorrelationID string      `json:"correlation_id,omitempty"`
}

type ErrorDetail struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// Validation error response
type ValidationErrorResponse struct {
    Error            ErrorDetail  `json:"error"`
    ValidationErrors []FieldError `json:"validation_errors"`
    Timestamp        time.Time    `json:"timestamp"`
    Path             string       `json:"path"`
    CorrelationID    string       `json:"correlation_id,omitempty"`
}

type FieldError struct {
    Field   string `json:"field"`
    Value   string `json:"value"`
    Message string `json:"message"`
}
```

## Testing Error Handling

### Unit Testing Error Scenarios

```go
func TestUserService_CreateUser_ValidationErrors(t *testing.T) {
    tests := []struct {
        name          string
        request       CreateUserRequest
        expectedError error
    }{
        {
            name: "invalid email",
            request: CreateUserRequest{
                Email:    "invalid-email",
                Username: "validuser",
                Password: "validpassword123",
            },
            expectedError: ErrInvalidEmail,
        },
        {
            name: "weak password",
            request: CreateUserRequest{
                Email:    "user@example.com",
                Username: "validuser",
                Password: "weak",
            },
            expectedError: ErrWeakPassword,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewUserService(mockRepo, mockValidator)

            _, err := service.CreateUser(tt.request)

            assert.Error(t, err)
            assert.True(t, errors.Is(err, tt.expectedError))
        })
    }
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
    repo := NewUserRepositorySQL(testDB)
    nonExistentID := NewUserID()

    user, err := repo.FindByID(context.Background(), nonExistentID)

    assert.Nil(t, user)
    assert.True(t, errors.Is(err, ErrUserNotFound))
}
```

### Integration Testing with Error Scenarios

```go
func TestCreateUserHandler_DuplicateEmail(t *testing.T) {
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()

    // Create initial user
    createUserRequest := CreateUserRequest{
        Email:    "test@example.com",
        Username: "testuser",
        Password: "password123",
    }

    // First request should succeed
    resp1 := makeCreateUserRequest(t, server.URL, createUserRequest)
    assert.Equal(t, http.StatusCreated, resp1.StatusCode)

    // Second request with same email should fail
    resp2 := makeCreateUserRequest(t, server.URL, createUserRequest)
    assert.Equal(t, http.StatusConflict, resp2.StatusCode)

    var errorResponse ErrorResponse
    err := json.NewDecoder(resp2.Body).Decode(&errorResponse)
    assert.NoError(t, err)

    assert.Equal(t, "CONFLICT", errorResponse.Error.Code)
    assert.Contains(t, errorResponse.Error.Message, "email already registered")
    assert.NotEmpty(t, errorResponse.CorrelationID)
}
```

## Logging and Monitoring

### Structured Error Logging

```go
type ErrorLogger struct {
    logger *slog.Logger
}

func (el *ErrorLogger) LogError(ctx context.Context, err error, message string, fields ...any) {
    // Extract correlation ID from context
    correlationID := GetCorrelationID(ctx)

    // Build log fields
    logFields := []any{
        "error", err.Error(),
        "correlation_id", correlationID,
    }
    logFields = append(logFields, fields...)

    // Log with appropriate level based on error type
    if isDomainError(err) {
        el.logger.InfoContext(ctx, message, logFields...)
    } else {
        el.logger.ErrorContext(ctx, message, logFields...)
    }
}

func isDomainError(err error) bool {
    return errors.Is(err, ErrUserNotFound) ||
           errors.Is(err, ErrInvalidEmail) ||
           errors.Is(err, ErrEmailAlreadyExists)
    // Add other domain errors
}
```

### Error Metrics

```go
type ErrorMetrics struct {
    errorCounter *prometheus.CounterVec
}

func NewErrorMetrics() *ErrorMetrics {
    return &ErrorMetrics{
        errorCounter: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "http_errors_total",
                Help: "Total number of HTTP errors by type and endpoint",
            },
            []string{"endpoint", "error_type", "status_code"},
        ),
    }
}

func (em *ErrorMetrics) RecordError(endpoint, errorType string, statusCode int) {
    em.errorCounter.WithLabelValues(
        endpoint,
        errorType,
        fmt.Sprintf("%d", statusCode),
    ).Inc()
}
```

## Best Practices Summary

1. **Never Ignore Errors**: Always handle or explicitly document why an error is ignored
2. **Add Context**: Wrap errors with meaningful context using `fmt.Errorf`
3. **Use Typed Errors**: Create domain-specific error types for business logic
4. **Layer Boundaries**: Each layer should handle errors appropriately for its responsibility
5. **User-Friendly Messages**: Don't expose internal implementation details in user-facing error messages
6. **Log Appropriately**: Use structured logging with correlation IDs
7. **Test Error Paths**: Write tests for both success and error scenarios
8. **Monitor Errors**: Track error metrics and patterns for system health
9. **Fail Fast**: Validate inputs early and return errors immediately
10. **Recovery Patterns**: Know when to retry, when to degrade gracefully, and when to fail
