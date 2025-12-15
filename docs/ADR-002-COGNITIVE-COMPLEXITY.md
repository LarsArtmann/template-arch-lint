# ADR-002: Cognitive Complexity Management

## Status

Accepted

## Context

Beyond cyclomatic complexity, we need to manage cognitive complexity - how difficult code is for humans to understand. This includes nested structures, logical operators, and control flow patterns.

## Decision

We will enforce cognitive complexity limits and establish patterns that reduce mental overhead when reading code.

### Key Principles

1. **Minimize Nesting**: Prefer early returns and guard clauses
2. **Limit Logical Operators**: Break complex boolean expressions into named variables
3. **Reduce Branching**: Use polymorphism or data structures instead of large switch statements
4. **Clear Control Flow**: Avoid complex loop structures and deeply nested conditions

### Strategies for Cognitive Complexity Reduction

#### 1. Early Returns (Guard Clauses)

```go
// Bad - Nested complexity
func ProcessOrder(order Order) error {
    if order.IsValid() {
        if order.HasItems() {
            if order.Customer.IsActive() {
                // ... process logic
                return nil
            } else {
                return errors.New("inactive customer")
            }
        } else {
            return errors.New("no items")
        }
    } else {
        return errors.New("invalid order")
    }
}

// Good - Guard clauses
func ProcessOrder(order Order) error {
    if !order.IsValid() {
        return errors.New("invalid order")
    }
    if !order.HasItems() {
        return errors.New("no items")
    }
    if !order.Customer.IsActive() {
        return errors.New("inactive customer")
    }

    // ... process logic
    return nil
}
```

#### 2. Named Boolean Variables

```go
// Bad - Complex boolean logic
if user.Age >= 18 && user.HasVerifiedEmail && (user.Country == "US" || user.Country == "CA") && !user.IsBanned {
    // ... logic
}

// Good - Named conditions
isAdult := user.Age >= 18
hasVerifiedEmail := user.HasVerifiedEmail
isFromNorthAmerica := user.Country == "US" || user.Country == "CA"
isNotBanned := !user.IsBanned

if isAdult && hasVerifiedEmail && isFromNorthAmerica && isNotBanned {
    // ... logic
}
```

#### 3. Replace Switch with Polymorphism

```go
// Bad - Large switch statement
func ProcessPayment(payment Payment) error {
    switch payment.Type {
    case "credit_card":
        // ... credit card logic
    case "paypal":
        // ... paypal logic
    case "bank_transfer":
        // ... bank transfer logic
    // ... many more cases
    }
}

// Good - Interface-based approach
type PaymentProcessor interface {
    Process(payment Payment) error
}

func (p *PaymentService) Process(payment Payment) error {
    processor := p.getProcessor(payment.Type)
    return processor.Process(payment)
}
```

#### 4. Functional Programming Patterns

```go
// Bad - Imperative with complex loops
func ProcessUsers(users []User) []ProcessedUser {
    var result []ProcessedUser
    for _, user := range users {
        if user.IsActive() && user.HasSubscription() {
            processed := ProcessedUser{
                ID:   user.ID,
                Name: strings.ToUpper(user.Name),
                // ... transformation logic
            }
            result = append(result, processed)
        }
    }
    return result
}

// Good - Functional approach
func ProcessUsers(users []User) []ProcessedUser {
    activeSubscribers := lo.Filter(users, func(u User, _ int) bool {
        return u.IsActive() && u.HasSubscription()
    })

    return lo.Map(activeSubscribers, func(u User, _ int) ProcessedUser {
        return ProcessedUser{
            ID:   u.ID,
            Name: strings.ToUpper(u.Name),
            // ... transformation logic
        }
    })
}
```

## Consequences

- Positive: Code is easier to understand and reason about
- Positive: Reduced mental fatigue when reading code
- Positive: Fewer bugs due to clearer logic flow
- Negative: May require more functions and abstractions
- Negative: Learning curve for functional programming patterns

## Monitoring

- Code reviews focus on readability and comprehension
- Cognitive complexity metrics tracked alongside cyclomatic complexity
- Team feedback on code clarity during reviews
