# Error Centralization Migration Plan
# Priority: Work Required vs Impact Analysis

## 🎯 HIGH IMPACT / LOW WORK (Quick Wins - Do First!)
### Step 1: Move Existing Error Infrastructure (1 min)
- Move `internal/domain/errors/` → `pkg/errors/`
- Update go.mod for new import path
- Verify all files move correctly

### Step 2: Update go-arch-lint Config (2 min) 
- Remove `domain-errors` component from .go-arch-lint.yml
- Add `pkg-errors` component with proper dependencies
- Update commonComponents to reference pkg-errors

### Step 3: Fix Import Paths in Core Domain Files (5 min)
- Update imports in high-traffic files:
  - internal/domain/services/user_service.go
  - internal/domain/entities/user.go
  - internal/domain/values/*.go
- Test compilation after each batch

## 🎯 HIGH IMPACT / MEDIUM WORK (Critical Infrastructure)
### Step 4: Update All Application Layer Imports (10 min)
- Batch update: internal/application/**/**/*.go
- Focus on handlers and middleware files
- Run compilation checks

### Step 5: Update Infrastructure Layer Imports (8 min)
- Batch update: internal/infrastructure/**/**/*.go
- Repository implementations
- Database-specific code

### Step 6: Update Test Files (15 min)
- Convert test files using `errors.New()` to use centralized errors
- Update import paths in test files
- Preserve test intent while using consistent errors

## 🎯 MEDIUM IMPACT / MEDIUM WORK (Refinement)
### Step 7: Convert Direct fmt.Errorf() Usage (20 min)
- Find all `fmt.Errorf()` patterns
- Replace with appropriate centralized error types
- Focus on non-test production code first

### Step 8: Add Missing Error Types (10 min)
- Analyze current `fmt.Errorf()` usage for gaps
- Add missing error constructors to pkg/errors
- Ensure comprehensive error coverage

### Step 9: Update Documentation (5 min)
- Update CLAUDE.md to reflect pkg/errors structure
- Update error handling documentation
- Add migration notes to README

## 🎯 MEDIUM IMPACT / HIGH WORK (Advanced Features)
### Step 10: Integration with samber/mo Result<T> (25 min)
- Ensure pkg/errors works seamlessly with functional patterns
- Create utility functions for mo.Result ↔ pkg/errors conversion
- Update existing mo usage patterns

### Step 11: Add Error Context Enrichment (15 min)
- Add request ID, user ID, and context to errors
- Implement error correlation/logging integration
- Add error severity levels

### Step 12: Error Handling Automation (20 min)
- Create linter rules for error consistency
- Add pre-commit hooks for error validation
- Implement error documentation generation

## 🎯 LOW IMPACT / HIGH WORK (Future Enhancements)
### Step 13: Error Metrics and Monitoring (30 min)
- Add error rate monitoring
- Implement error pattern analysis
- Create error dashboards

### Step 14: Advanced Error Recovery (45 min)
- Implement retry mechanisms with specific error types
- Add circuit breaker patterns
- Create error classification for automated handling

## 🎯 CONTINUOUS IMPROVEMENT (Ongoing)
### Step 15: Review and Refine (Ongoing)
- Monthly error pattern reviews
- Performance impact analysis
- Developer experience feedback collection

---

## 🚀 EXECUTION STRATEGY

### Phase 1: Foundation (Steps 1-3)
**Goal**: Working centralized error system
**Time**: ~10 minutes
**Risk**: Low (just moving existing code)

### Phase 2: Migration (Steps 4-6) 
**Goal**: All code using pkg/errors
**Time**: ~33 minutes  
**Risk**: Medium (many import changes)

### Phase 3: Refinement (Steps 7-9)
**Goal**: Zero direct error creation
**Time**: ~35 minutes
**Risk**: Low (cleaning up patterns)

### Phase 4: Enhancement (Steps 10-14)
**Goal**: Advanced error features
**Time**: ~2 hours+
**Risk**: Medium (new features)

### Phase 5: Maintenance (Step 15)
**Goal**: Continuous improvement
**Time**: Ongoing
**Risk**: None

---

## 🔄 VERIFICATION CHECKPOINTS

After each phase:
1. `go build ./...` - Compilation check
2. `go test ./...` - Test suite passes
3. `just lint-arch` - Architecture compliance
4. `just lint` - Full linting compliance
5. Manual testing of error handling paths

---

## 🎯 SUCCESS METRICS

### Before Migration:
- 5 instances of `errors.New()`
- 26 instances of `fmt.Errorf()`
- Errors in internal/domain/errors/

### After Migration:
- 0 instances of direct error creation
- All errors in pkg/errors/
- 100% import path compliance
- No functional regressions