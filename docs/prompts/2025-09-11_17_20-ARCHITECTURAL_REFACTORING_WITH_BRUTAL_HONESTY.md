# Prompt: Architectural Refactoring with Brutal Honesty Assessment

Date: 2025-09-11 17:20
Context: Clean Architecture + DDD refactoring
Status: Successful
Reusability: High - applicable to any architectural refactoring

## Original Request Template

"Refactor [COMPONENT/ENTITY] to eliminate architectural violations and improve domain modeling, but apply brutal honesty assessment to prevent over-engineering for the project context."

## Refined Prompt Framework

### Phase 1: Pre-Refactoring Analysis

"Before making any changes, perform this analysis:

1. **Split Brain Detection**
   - Scan for duplicate field representations (value objects + primitives)
   - Identify inconsistent data access patterns in tests
   - Look for fields with identical business meaning but different types

2. **Ghost System Audit**
   - Find interfaces: `rg 'type.*Service.*interface'`
   - Find implementations: `rg 'struct.*Service'`
   - Find usage: `rg 'Service' | grep -v 'type|func|struct|interface'`
   - Flag any interfaces with zero concrete usage

3. **Project Context Assessment**
   - Is this a template/demo project or production application?
   - What's the actual business value being delivered?
   - Who is the target audience for this code?

4. **Scope Reality Check**
   - Will anyone actually use the patterns I'm about to build?
   - Am I solving real problems or creating impressive abstractions?
   - What's the simplest solution that demonstrates the architectural principle?"

### Phase 2: Execution Strategy with Checkpoints

"Execute refactoring with these mandatory checkpoints:

1. **Split Brain Elimination**

   ```
   For each entity with dual representation:
   - Remove primitive duplicate fields
   - Implement custom JSON marshaling for value objects
   - Update all tests to use single field representation
   - Verify JSON serialization works correctly
   ```

2. **Value Object Integration**

   ```
   For each value object:
   - Ensure proper encapsulation (private fields)
   - Implement String() method for debugging
   - Implement MarshalJSON() for API compatibility
   - Add validation in constructor functions
   ```

3. **Ghost System Removal**

   ```
   For each unused interface/service:
   - Confirm zero usage with grep/ripgrep
   - Remove interface definition
   - Remove implementation
   - Remove associated test files
   - Update documentation/comments
   ```

4. **Test-First Validation**
   ```
   MANDATORY: Run tests at each checkpoint
   - Never claim success without green tests
   - Fix compilation errors immediately
   - Verify behavior unchanged (refactoring only)
   - Update test assertions to match new field structure
   ```

### Phase 3: Brutal Honesty Assessment

"After implementation, answer these questions truthfully:

**Architecture Questions:**

- Did I build what the project actually needed?
- Is this architectural purity providing measurable value?
- Would a new developer understand this better with less complexity?
- Am I solving real problems or showcasing theoretical knowledge?

**Scope Questions:**

- Does this refactoring align with project purpose (template vs production)?
- Did I add enterprise patterns that will never be used?
- Are my abstractions right-sized for the problem context?
- What TODOs did I create that will never be implemented?

**Quality Questions:**

- Did I test my changes before claiming success?
- Are there simpler solutions that achieve the same architectural goals?
- What assumptions did I make that should be validated?
- Did I create more maintenance burden than business value?"

## Results Framework

### Expected Outcomes

- **Split brains eliminated**: Single source of truth for all business concepts
- **Value objects integrated**: Proper JSON marshaling without API breaks
- **Ghost systems removed**: Zero unused interfaces or implementations
- **Tests passing**: All existing behavior preserved through refactoring
- **Scope appropriate**: Patterns match project context and audience needs

### Success Metrics

```bash
# Quantitative measures
- Compilation errors fixed: [COUNT]
- Tests maintained: [COUNT]
- Lines of dead code removed: [COUNT]
- Split brain patterns eliminated: [COUNT]
- Ghost systems removed: [COUNT]

# Qualitative measures
- API contracts maintained: [YES/NO]
- New developer comprehension: [IMPROVED/UNCHANGED/DEGRADED]
- Maintenance burden: [REDUCED/UNCHANGED/INCREASED]
- Business value alignment: [STRONG/MODERATE/WEAK]
```

## Execution Strategy Template

### 1. Discovery Phase Commands

```bash
# Split brain detection
rg "type.*struct" --type go -A 20 | grep -E "(Email|Name|ID).*string|Email.*Email"

# Ghost system audit
rg "type.*Service.*interface" --type go
rg "Service" --type go | grep -v "type\|func\|struct\|interface"

# Test coverage baseline
go test ./... -v | grep -E "PASS|FAIL" | wc -l
```

### 2. Implementation Checkpoints

```bash
# After each major change
go build ./...          # Compilation check
go test ./... -v        # Behavior preservation
just lint-arch         # Architecture compliance
git add . && git status # Change tracking
```

### 3. Validation Commands

```bash
# JSON marshaling verification
go test -v -run "TestUser.*JSON"

# Dead code confirmation
rg "QueryService|UnusedInterface" --type go

# Final architecture check
just lint
```

## Lessons Learned Integration

### Common Pitfalls to Avoid

1. **Premature Success Claims**: Always run tests before declaring victory
2. **Template Over-Engineering**: Question every enterprise pattern in demo projects
3. **Ghost System Creation**: Build interfaces only when you have concrete consumers
4. **Split Brain Tolerance**: Eliminate duplicate representations immediately
5. **JSON Marshaling Oversight**: Value objects need custom marshaling for APIs

### Context-Specific Adaptations

**For Template Projects:**

- Prioritize educational clarity over enterprise completeness
- Document WHY patterns are useful, not just HOW to implement
- Keep examples focused on demonstrating architectural boundaries
- Avoid building features that will never be implemented

**For Production Applications:**

- Start simple, evolve based on actual business requirements
- Add complexity only when justified by real use cases
- Regular ghost system audits to prevent architectural debt
- Measure actual impact of architectural decisions

### Quality Assurance Checklist

- [ ] All tests pass without modification to test logic
- [ ] JSON APIs return expected data format
- [ ] No duplicate field representations exist
- [ ] Zero unused interfaces or implementations
- [ ] Architecture boundaries properly enforced
- [ ] Code complexity appropriate for project context
- [ ] Documentation updated to reflect changes
- [ ] Brutal honesty assessment completed

## Related Patterns

- **Similar to**: Domain model refactoring, API contract evolution
- **Builds on**: Clean Architecture, Domain-Driven Design principles
- **Enables**: Type-safe value objects, maintainable domain models
- **Prevents**: Split brain patterns, ghost systems, over-engineering

## Reusable Pattern Summary

This prompt provides a systematic approach to architectural refactoring that balances domain modeling best practices with practical project constraints. The brutal honesty assessment prevents the common trap of over-engineering while ensuring that architectural improvements deliver real value.

Key innovations:

- Split brain detection and elimination techniques
- Ghost system identification and removal
- Context-appropriate scope management
- Test-first validation at every checkpoint
- Brutal honesty assessment framework

Use this prompt when refactoring domain models, eliminating architectural debt, or any time you need to balance architectural purity with practical delivery constraints.
