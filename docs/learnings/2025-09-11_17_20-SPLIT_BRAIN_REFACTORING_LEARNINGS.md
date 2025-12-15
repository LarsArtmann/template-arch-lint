# Learning: Split Brain Refactoring with Brutal Honesty Assessment

Date: 2025-09-11 17:20
Difficulty: Advanced
Time Investment: 4 hours
Session Context: template-arch-lint User entity refactoring

## Problem Context

The User entity had a critical "split brain" pattern - duplicate representation of the same data:

- Domain entity had `Email` and `Name` value objects (proper domain modeling)
- Same entity had `email string` and `name string` fields (primitive obsession)
- JSON marshaling only exposed the primitive fields, making value objects invisible
- Tests were inconsistent about which fields to use, causing silent failures

This created a dangerous situation where domain logic and persistence logic were operating on different data representations.

## Key Insights

### 1. Split Brain Detection Patterns

**Before**: Identified the split brain through inconsistent test behavior

```go
// DANGEROUS: Two ways to represent the same concept
type User struct {
    ID    UserID    `json:"id" db:"id"`
    Email Email     `json:"-"` // Value object (proper domain)
    Name  UserName  `json:"-"` // Value object (proper domain)
    email string    `json:"email" db:"email"` // Primitive (persistence)
    name  string    `json:"name" db:"name"`   // Primitive (persistence)
}
```

**After**: Single source of truth with custom JSON marshaling

```go
// CLEAN: Single representation with proper JSON handling
type User struct {
    ID    UserID   `json:"id" db:"id"`
    Email Email    `json:"email" db:"email"`
    Name  UserName `json:"name" db:"name"`
}

func (u User) MarshalJSON() ([]byte, error) {
    return json.Marshal(struct {
        ID    string `json:"id"`
        Email string `json:"email"`
        Name  string `json:"name"`
    }{
        ID:    u.ID.String(),
        Email: u.Email.String(),
        Name:  u.Name.String(),
    })
}
```

### 2. Custom JSON Marshaling for Value Objects

**Key Learning**: Value objects with private fields need custom JSON marshaling to be API-friendly.

**Implementation Pattern**:

```go
// Value object with validation
type Email struct {
    value string // Private field for encapsulation
}

// JSON marshaling exposes the value
func (e Email) MarshalJSON() ([]byte, error) {
    return json.Marshal(e.value)
}

// String representation
func (e Email) String() string {
    return e.value
}
```

**Critical Gotcha**: Without custom marshaling, value objects serialize as `{}` empty objects, breaking API contracts.

### 3. Ghost System Detection Techniques

**Discovery**: Found `UserQueryService` with elaborate interfaces but zero actual usage.

**Detection Pattern**:

```bash
# Find interface definitions
rg "type.*QueryService.*interface" --type go

# Find implementations
rg "func.*QueryService" --type go

# Find actual usage (the smoking gun)
rg "QueryService" --type go | grep -v "type\|func\|interface"
```

**Result**: 200+ lines of code with zero business value - pure architectural theater.

### 4. Template Project vs Production Application Scope Management

**Critical Insight**: Template projects require different thinking than production applications.

**Template Project Constraints**:

- Purpose: Demonstrate architectural patterns, not solve business problems
- Audience: Developers learning Clean Architecture + DDD patterns
- Success criteria: Clear examples of linting rules and boundaries
- Anti-pattern: Over-engineering for imaginary future requirements

**Production Application Approach**:

- Purpose: Solve real business problems efficiently
- Audience: End users and business stakeholders
- Success criteria: User value delivery and business metrics
- Pattern: Start simple, evolve based on actual needs

### 5. Brutal Honesty Assessment Framework

**Self-Assessment Questions**:

1. "Am I building what the project actually needs?"
2. "Is this architectural purity providing real value?"
3. "Would a new developer understand this better with or without this complexity?"
4. "Am I solving real problems or creating impressive-looking abstractions?"

**Red Flags Identified**:

- Claiming "success" before running tests (amateur mistake)
- Building CQRS for systems with no HTTP layer
- Creating 25+ TODOs for features that will never be implemented
- Elaborate domain services for simple value object validation

## Before vs After Understanding

**Before**: "Perfect domain modeling requires elaborate patterns and maximum abstraction"
**After**: "Effective domain modeling balances business value with architectural clarity - especially in template projects"

**Before**: "Value objects must be completely encapsulated with private fields"
**After**: "Value objects need thoughtful JSON marshaling to be API-friendly while maintaining encapsulation"

**Before**: "More abstraction layers = better architecture"  
**After**: "Right-sized abstractions for the problem context = better architecture"

## Practical Applications

### For Template Projects

- Keep examples focused on demonstrating architectural boundaries
- Avoid over-engineering for imaginary scale requirements
- Prioritize clarity and educational value over enterprise patterns
- Document why patterns are useful, not just how to implement them

### For Production Applications

- Start with simpler value objects (public fields) and evolve as needed
- Add custom JSON marshaling only when API contracts require it
- Build abstractions based on actual business requirements
- Regular "ghost system audits" to remove unused code

### For Refactoring Sessions

- Always run tests before claiming success
- Identify and eliminate split brain patterns immediately
- Question every abstraction: "What problem does this solve?"
- Time-box architectural purity efforts based on project context

## Code Examples

### Split Brain Elimination Pattern

```go
// BEFORE: Split brain nightmare
type User struct {
    Email Email  `json:"-"`      // Domain representation
    email string `json:"email"`   // API representation
}

// AFTER: Single source of truth
type User struct {
    Email Email `json:"email"`   // Both domain AND API
}

func (u User) MarshalJSON() ([]byte, error) {
    // Custom marshaling bridges the gap
    return json.Marshal(struct {
        Email string `json:"email"`
    }{
        Email: u.Email.String(),
    })
}
```

### Ghost System Detection

```bash
# 1. Find interfaces
rg "type.*Service.*interface"

# 2. Find implementations
rg "struct.*Service"

# 3. Find usage (critical step)
rg "Service" | grep -v "type\|func\|struct\|interface"

# If step 3 returns nothing = GHOST SYSTEM
```

## Performance/Quality Impact

**Metrics from Session**:

- Fixed 21 compilation errors across service test files
- Eliminated 200+ lines of unused code (UserQueryService)
- Reduced cognitive complexity by removing dual representation
- Improved test reliability by eliminating field confusion
- Maintained 91 existing tests with zero behavior changes

**Quality Benefits**:

- Eliminated silent data inconsistency bugs
- Simplified mental model for new developers
- Improved API contract reliability
- Reduced maintenance burden

## Related Technologies

- Go's `json` package custom marshaling interfaces
- Value object patterns from DDD
- Clean Architecture dependency rules
- Test-driven refactoring approaches
- Static analysis tools for dead code detection

## Gotchas and Pitfalls

### Value Object JSON Marshaling

- **Mistake**: Forgetting custom `MarshalJSON()` implementation
- **Result**: APIs return `{}` instead of actual values
- **Solution**: Always implement custom marshaling for value objects with private fields

### Split Brain Prevention

- **Warning Sign**: Multiple fields representing the same business concept
- **Red Flag**: Tests using different fields inconsistently
- **Prevention**: Single source of truth with transformation functions

### Template Project Scope Creep

- **Trap**: Adding enterprise patterns "for completeness"
- **Reality Check**: "Will copying this config file help developers?"
- **Focus**: Demonstrate linting rules and boundaries, not business logic

### Ghost System Creation

- **How it happens**: Building interfaces before implementations
- **Detection**: Regular usage audits with grep/ripgrep
- **Prevention**: Code only when there's a concrete consumer

## Success Criteria

- Split brain patterns eliminated from codebase
- Value objects work seamlessly with JSON APIs
- Ghost systems identified and removed
- Template project scope maintained appropriately
- Brutal honesty assessment prevents over-engineering

This learning session transformed both the codebase and my approach to architectural refactoring, emphasizing practical value over architectural purity.
