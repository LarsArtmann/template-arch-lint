# Code Style Guidelines

## Line Length Management

### Overview
This document establishes guidelines for managing line length in Go code to maintain readability while adhering to modern development practices.

### Line Length Limits
- **Soft Limit**: 100 characters (preferred)
- **Hard Limit**: 120 characters (enforced by linter)
- **Comments**: 80 characters (for better readability)

### Rationale
- Modern monitors can comfortably display 120+ characters
- Allows for reasonable function signatures and struct definitions
- Balances readability with practical coding needs
- Prevents excessive horizontal scrolling

## Line Breaking Strategies

### 1. Function Signatures
```go
// Good - Parameters on new lines when needed
func CreateUserWithCompleteProfile(
    ctx context.Context,
    userDetails UserCreationDetails,
    profileInfo ProfileInformation,
    preferences UserPreferences,
) (*User, error) {
    // implementation
}

// Good - Short signatures on single line
func CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    // implementation
}

// Bad - Forced single line that's too long
func CreateUserWithCompleteProfile(ctx context.Context, userDetails UserCreationDetails, profileInfo ProfileInformation, preferences UserPreferences) (*User, error) {
    // implementation
}
```

### 2. Struct Definitions
```go
// Good - Readable field alignment
type UserProfile struct {
    ID          UserID    `json:"id" db:"id"`
    Email       Email     `json:"email" db:"email"`
    Username    Username  `json:"username" db:"username"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
    Preferences struct {
        Theme        string `json:"theme"`
        Notifications bool   `json:"notifications"`
        Language     string `json:"language"`
    } `json:"preferences" db:"preferences"`
}

// Good - Break complex embedded structs
type ComplexUserProfile struct {
    ID       UserID `json:"id" db:"id"`
    Email    Email  `json:"email" db:"email"`
    Settings UserSettings
    Metadata UserMetadata
}

type UserSettings struct {
    Theme         string `json:"theme"`
    Notifications bool   `json:"notifications"`
    Language      string `json:"language"`
}
```

### 3. Function Calls with Many Parameters
```go
// Good - Named parameters pattern
func processUserRegistration() error {
    user, err := userService.CreateUser(CreateUserParams{
        Email:       "user@example.com",
        Password:    "securePassword123",
        FirstName:   "John",
        LastName:    "Doe",
        DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
        Preferences: defaultUserPreferences(),
    })
    return err
}

// Good - Multiple lines for clarity
func configureUserService() *UserService {
    return NewUserService(
        db,
        cache,
        logger.WithField("service", "user"),
        emailService,
        validator,
    )
}

// Bad - Long parameter list
func processUserRegistration() error {
    user, err := userService.CreateUser("user@example.com", "securePassword123", "John", "Doe", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), defaultUserPreferences())
    return err
}
```

### 4. Chain Calls and Fluent Interfaces
```go
// Good - Readable chaining
user, err := userQuery.
    WithEmail("user@example.com").
    WithActiveStatus(true).
    WithCreatedAfter(lastWeek).
    IncludeProfile().
    IncludePreferences().
    First()

// Good - Break complex chains
query := userQuery.
    WithEmail("user@example.com").
    WithActiveStatus(true)

if includeProfile {
    query = query.IncludeProfile()
}

user, err := query.First()
```

### 5. Conditional Statements
```go
// Good - Break complex conditions
isEligibleUser := user.IsActive() && 
                 user.HasVerifiedEmail() && 
                 user.SubscriptionStatus == "premium"

if isEligibleUser {
    // process premium user
}

// Good - Multiple condition blocks
if user.IsActive() &&
   user.HasVerifiedEmail() &&
   user.SubscriptionStatus == "premium" &&
   user.LastLoginWithin(30*24*time.Hour) {
    // process eligible user
}

// Bad - Unreadable single line
if user.IsActive() && user.HasVerifiedEmail() && user.SubscriptionStatus == "premium" && user.LastLoginWithin(30*24*time.Hour) && user.CountryCode == "US" {
    // process user
}
```

### 6. String Concatenation and Formatting
```go
// Good - Multi-line string formatting
message := fmt.Sprintf(
    "User %s (%s) has successfully completed registration "+
    "with subscription level %s on %s",
    user.Username,
    user.Email,
    user.SubscriptionLevel,
    user.CreatedAt.Format("2006-01-02"),
)

// Good - Template strings for complex content
const emailTemplate = `
Dear %s,

Your account has been created successfully with the following details:
- Email: %s
- Username: %s
- Subscription: %s

Best regards,
The Team
`

emailBody := fmt.Sprintf(emailTemplate, 
    user.FirstName, 
    user.Email, 
    user.Username, 
    user.SubscriptionLevel,
)
```

### 7. Error Messages
```go
// Good - Readable error construction
return fmt.Errorf(
    "failed to create user account for %s: %w",
    req.Email,
    err,
)

// Good - Multi-line error context
return errors.New(
    "user validation failed: email must be valid, " +
    "password must be at least 8 characters, " +
    "and username must be unique",
)

// Good - Error wrapping with context
if err != nil {
    return fmt.Errorf(
        "database operation failed while creating user %s "+
        "in transaction %s: %w",
        user.Email,
        transactionID,
        err,
    )
}
```

## Comments and Documentation

### Line Length for Comments
```go
// Good - Comments under 80 characters
// CreateUser creates a new user account with the provided details.
// Returns the created user or an error if validation fails.
func CreateUser(req CreateUserRequest) (*User, error) {
    // implementation
}

// Good - Break long comments appropriately
// ProcessUserRegistration handles the complete user registration flow
// including validation, account creation, email verification setup,
// and initial preference configuration.
func ProcessUserRegistration(req RegistrationRequest) error {
    // implementation
}

// Bad - Long comment line
// ProcessUserRegistration handles the complete user registration flow including validation, account creation, email verification setup, and initial preference configuration.
func ProcessUserRegistration(req RegistrationRequest) error {
    // implementation
}
```

## Tools and Automation

### Linting Configuration
The project uses golangci-lint with line length enforcement:
```yaml
linters-settings:
  lll:
    line-length: 120
    tab-width: 4
```

### Editor Configuration
Recommended editor settings:
```
# .editorconfig
[*.go]
max_line_length = 120
tab_width = 4
indent_style = tab
```

### Formatting Tools
- Use `gofmt` for basic formatting
- Use `goimports` for import organization
- Use `golangci-lint` for comprehensive style checking

## Exceptions

### When to Exceed Line Limits
1. **URLs and file paths** that cannot be broken
2. **Generated code** that should not be manually modified
3. **Test data** where breaking would reduce clarity
4. **Regular expressions** where breaking affects readability

### Example Exceptions
```go
// Acceptable - URL cannot be broken meaningfully
const apiEndpoint = "https://api.example.com/v1/users/profile/detailed-information/with-preferences"

// Acceptable - File path
const configPath = "/etc/myapp/configuration/database/connection-settings/production.yaml"

// Acceptable - Test data table
var testCases = []struct{
    input    string
    expected UserValidationResult
    message  string
}{
    {"valid@email.com", UserValidationResult{Valid: true}, "standard email should be valid"},
    // ...
}
```

## Best Practices Summary

1. **Prioritize Readability**: Choose line breaks that improve code comprehension
2. **Use Named Parameters**: For functions with many parameters, use struct parameters
3. **Break at Logical Points**: Break lines at logical separators (commas, operators)
4. **Consistent Indentation**: Maintain consistent indentation when breaking lines
5. **Consider Context**: Some long lines are more readable than artificially broken ones
6. **Use Tools**: Leverage automated formatting and linting tools
7. **Team Consistency**: Follow team conventions even if personal preference differs