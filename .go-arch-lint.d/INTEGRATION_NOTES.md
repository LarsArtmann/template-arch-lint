# 🚫 ERROR CENTRALIZATION PATTERNS FOR GOLANGCI-LINT

# These patterns from .go-arch-lint.d/ have been integrated into .golangci.yml

# Use the forbidigo linter to enforce these patterns

## PATTERNS INTEGRATED:

# Direct error creation bans:

- p: 'errors\.New\('
  msg: "🚨 BANNED: Direct error creation. Use pkg/errors predefined types instead"
  # Note: Allow in pkg/errors only (manual enforcement needed)
- p: 'fmt\.Errorf\('
  msg: "🚨 BANNED: Direct error formatting. Use pkg/errors predefined types instead"

  # Note: Allow in pkg/errors only (manual enforcement needed)

- p: 'error\(\s\*{'
  msg: "🚨 BANNED: Anonymous error structs. Use pkg/errors Error type instead"
  # Note: Allow in pkg/errors only (manual enforcement needed)

## USAGE NOTES:

1. The forbidigo patterns are global - they apply to ALL files
2. go-arch-lint.d/ had "allowed_paths" which golangci-lint doesn't support
3. Manual code review needed to ensure pkg/errors can still create errors
4. Consider adding these exemptions to pkg/errors files: //nolint:forbidigo

## ALTERNATIVE APPROACHES:

### Option 1: Use custom linter plugin

- Create custom golangci-lint plugin with path-aware checking
- More complex but provides precise control

### Option 2: Pre-commit hooks with path filtering

- Use scripts to check only specific directories
- More flexible enforcement

### Option 3: Keep go-arch-lint separate

- Run go-arch-lint for architectural validation
- Run golangci-lint for code quality
- Different tools for different concerns
