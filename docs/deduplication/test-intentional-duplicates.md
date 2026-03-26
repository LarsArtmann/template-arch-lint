# Test Code Duplication Analysis

## Overview

Test files contain intentional, repeated patterns for maintainability:

### 1. Gomega Assertion Patterns

Repeated test assertion patterns like:

```go
gomega.Expect(user.GetEmail().Value()).To(gomega.Equal(testEmail))
gomega.Expect(user.GetName().Value()).To(gomega.Equal(testName))
gomega.Expect(jsonMap["id"]).To(gomega.Equal(userID))
```

**Why intentional**: These are structural test setup patterns with clear, balanced calls. Refactoring to extract helpers would not improve maintainability - it would increase cognitive load understanding table-driven tests.

### 2. Mock Repository Setup

Repeated mock repository implementations:

```go
func (m *mockRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
    m.calls.Add("FindByEmail", email)
    if m.closeOnCall["FindByEmail"] {
        m.close()
    }
    return m.FindByEmailData, m.FindByEmailErr
}
```

**Why intentional**: Each mock typically needs unique setup for each test case. Table-driven setups require explicit testing table, which adds unnecessary complexity for simple assertions.

### 3. Failing Repository Patterns

Errors are intentionally repeated for explicit error testing:

```go
FailingUserRepository{
    FindByEmailFunc: func(ctx context.Context, email string) (*entities.User, error) {
        return nil, errors.New("database error")
    },
}
```

**Why intentional**: Each test should be self-contained and independent. Extracting helpers would break test isolation.

## Analysis Findings

- **Total test file duplicates**: 32 clone groups (50% of total)
- **Estimated tokens**: 200+ (intentional patterns)
- **ROI**: Negative - extract helpers would reduce code readability without benefit
- **Recommendation**: Do NOT refactor test code

## When to Consider Test Deduplication

Only consider if:

1. Test setup code exceeds 25 lines (patterns require complex table-driven setups)
2. Same test logic repeated >3 times across test files
3. Extracted helper improves test readability (not obfuscates it)

## Current State

All production code refactors are complete. Test code remains untouched intentionally.
