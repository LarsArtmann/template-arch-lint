# 🚨 CRITICAL DEPGUARD CONFIGURATION FIX

## Problem Identified
The `.golangci.yml` references a `"Main"` depguard rule that doesn't exist, causing 68 valid project imports to be blocked.

## Configuration Analysis
```yaml
# Current .golangci.yml lines 779-788
depguard:
  rules:
    main:  # ← This rule is used
      list-mode: allow
      # ... allowlist entries

# But output shows: "is not allowed from list 'Main'"
# Note the capitalization mismatch!
```

## Solution Options

### Option 1: Create Proper "Main" Rule (RECOMMENDED)
Add a proper "Main" rule for production code.

### Option 2: Use Existing "main" Rule
Change references to match existing lowercase rule.

### Option 3: Split Production/Test Rules
Use "main" and "tests" rules with proper file matching.

## Decision Required
Which approach should be implemented? This affects 68 critical import violations.