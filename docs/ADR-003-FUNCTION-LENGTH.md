# ADR-003: Function Length Management

## Status
Accepted

## Context
Long functions are difficult to understand, test, and maintain. We need clear guidelines for managing function length while maintaining code clarity and avoiding premature abstraction.

## Decision
We will enforce a maximum function length of 50 lines through golangci-lint, with a preference for functions under 30 lines.

### Rationale
- Functions over 50 lines typically do too many things
- Shorter functions are easier to test in isolation
- Better readability and comprehension
- Easier to debug and modify
- Forces better separation of concerns

### Guidelines for Function Length

#### Target Sizes
- **Ideal**: 1-15 lines (single responsibility)
- **Good**: 16-30 lines (focused functionality)
- **Acceptable**: 31-50 lines (complex but cohesive)
- **Refactor Required**: 50+ lines

#### Strategies for Length Reduction

#### 1. Extract Helper Functions
```go
// Bad - Long function
func CreateUserAccount(req CreateUserRequest) (*User, error) {
    // Email validation (5 lines)
    if req.Email == "" {
        return nil, errors.New("email required")
    }
    if !strings.Contains(req.Email, "@") {
        return nil, errors.New("invalid email format")
    }
    // ... more email validation
    
    // Password validation (8 lines)
    if len(req.Password) < 8 {
        return nil, errors.New("password too short")
    }
    // ... more password validation
    
    // User creation (10 lines)
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    // ... user creation logic
    
    // Database operations (15 lines)
    // ... database logic
    
    return user, nil
}

// Good - Extracted functions
func CreateUserAccount(req CreateUserRequest) (*User, error) {
    if err := validateCreateUserRequest(req); err != nil {
        return nil, err
    }
    
    user, err := buildUserFromRequest(req)
    if err != nil {
        return nil, err
    }
    
    return persistUser(user)
}

func validateCreateUserRequest(req CreateUserRequest) error {
    if err := validateEmail(req.Email); err != nil {
        return err
    }
    return validatePassword(req.Password)
}
```

#### 2. Use Functional Programming
```go
// Bad - Long imperative function
func ProcessOrderItems(items []OrderItem) ([]ProcessedItem, error) {
    var result []ProcessedItem
    var totalValue float64
    
    for _, item := range items {
        if item.Quantity <= 0 {
            continue
        }
        if item.Price < 0 {
            return nil, errors.New("negative price")
        }
        
        processed := ProcessedItem{
            ID:       item.ID,
            Name:     strings.ToUpper(item.Name),
            Quantity: item.Quantity,
            Price:    item.Price,
            Total:    item.Price * float64(item.Quantity),
        }
        
        totalValue += processed.Total
        
        if totalValue > 10000 {
            return nil, errors.New("order value too high")
        }
        
        result = append(result, processed)
    }
    
    return result, nil
}

// Good - Functional approach
func ProcessOrderItems(items []OrderItem) ([]ProcessedItem, error) {
    validItems := filterValidItems(items)
    
    if err := validateItemPrices(validItems); err != nil {
        return nil, err
    }
    
    processed := transformItems(validItems)
    
    if err := validateTotalValue(processed); err != nil {
        return nil, err
    }
    
    return processed, nil
}
```

#### 3. Configuration Objects for Complex Initialization
```go
// Bad - Long constructor
func NewUserService(db *sql.DB, cache Redis, logger Logger, 
                   emailService EmailService, validator Validator,
                   encryptor Encryptor, metrics Metrics) *UserService {
    // 30+ lines of initialization
}

// Good - Configuration object
type UserServiceConfig struct {
    DB           *sql.DB
    Cache        Redis
    Logger       Logger
    EmailService EmailService
    Validator    Validator
    Encryptor    Encryptor
    Metrics      Metrics
}

func NewUserService(config UserServiceConfig) *UserService {
    return &UserService{
        db:           config.DB,
        cache:        config.Cache,
        logger:       config.Logger,
        emailService: config.EmailService,
        validator:    config.Validator,
        encryptor:    config.Encryptor,
        metrics:      config.Metrics,
    }
}
```

## When NOT to Extract
- Don't extract single-use functions that don't improve clarity
- Don't create functions just to meet line count if it reduces readability
- Keep cohesive logic together even if it approaches the limit

## Consequences
- Positive: More maintainable and testable functions
- Positive: Better separation of concerns
- Positive: Easier code navigation and understanding
- Negative: May create more functions and complexity
- Negative: Risk of premature abstraction

## Monitoring
- golangci-lint funlen linter enforces this rule
- Code reviews check for appropriate function decomposition
- Metrics on average function length across codebase