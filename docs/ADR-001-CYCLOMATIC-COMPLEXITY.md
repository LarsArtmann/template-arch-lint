# ADR-001: Cyclomatic Complexity Management

## Status

Accepted

## Context

The project enforces strict cyclomatic complexity limits to maintain code readability and testability. We need to establish clear guidelines for managing complexity in Go applications.

## Decision

We will enforce a maximum cyclomatic complexity of 10 per function through golangci-lint configuration.

### Rationale

- Functions with complexity > 10 become difficult to test thoroughly
- High complexity indicates multiple responsibilities within a single function
- Lower complexity improves code maintainability and debugging

### Strategies for Complexity Reduction

1. **Extract Methods**: Break large functions into smaller, focused functions
2. **Early Returns**: Use guard clauses to reduce nesting levels
3. **Table-Driven Tests**: Replace complex conditional logic with data structures
4. **Strategy Pattern**: Replace complex switch statements with interfaces
5. **Functional Programming**: Use samber/lo functions (Map, Filter, Reduce) to reduce imperative complexity

### Examples

#### Bad (High Complexity)

```go
func ProcessUser(user User) error {
    if user.Email == "" {
        return errors.New("email required")
    }
    if user.Name == "" {
        return errors.New("name required")
    }
    if len(user.Name) < 2 {
        return errors.New("name too short")
    }
    if len(user.Name) > 50 {
        return errors.New("name too long")
    }
    // ... more validation logic
    // ... business logic
    // ... persistence logic
    return nil
}
```

#### Good (Low Complexity)

```go
func ProcessUser(user User) error {
    if err := validateUser(user); err != nil {
        return err
    }
    return persistUser(user)
}

func validateUser(user User) error {
    validations := []func(User) error{
        validateEmail,
        validateName,
        validateNameLength,
    }

    for _, validate := range validations {
        if err := validate(user); err != nil {
            return err
        }
    }
    return nil
}
```

## Consequences

- Positive: More maintainable and testable code
- Positive: Easier to understand individual functions
- Positive: Better separation of concerns
- Negative: May require more functions and interfaces
- Negative: Initial development may be slower

## Monitoring

- golangci-lint gocyclo linter enforces this rule
- CI/CD pipeline fails on complexity violations
- Code reviews should specifically check for complexity patterns
