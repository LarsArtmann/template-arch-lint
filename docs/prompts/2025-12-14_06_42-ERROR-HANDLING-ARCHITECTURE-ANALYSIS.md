# Error Handling Architecture Analysis Prompt

## Prompt: Comprehensive Error Handling Architecture Analysis

**Purpose**: Analyze and improve error handling patterns in Go Clean Architecture projects  
**Use Case**: Evaluating error centralization vs layering approaches  
**Context**: Architectural decision-making for enterprise Go applications

---

## Prompt Content

You are a Senior Software Architect with 15+ years of experience building enterprise systems. I need a comprehensive analysis of error handling patterns in a Go project that implements Clean Architecture with strict layer enforcement using go-arch-lint.

## Current Context

- **Project**: Go Linting Template demonstrating enterprise-grade architecture
- **Architecture**: Clean Architecture with strict layer enforcement
- **Current Setup**: Centralized errors in `pkg/errors/` available to all components
- **Alternative**: Error layers where each architectural layer (domain, infrastructure, application) has its own error packages

## Analysis Requirements

### 1. Current Implementation Review

Examine the existing error handling approach:

- Analyze `.go-arch-lint.yml` configuration for error dependencies
- Review `pkg/errors/` structure and patterns
- Identify strengths and limitations of centralized approach
- Assess impact on layer boundaries and dependency direction

### 2. Alternative Evaluation

Evaluate error layering approach:

- Compare layer-specific error packages vs centralization
- Analyze impact on Clean Architecture principles
- Consider dependency rules and architectural boundaries
- Evaluate effect on domain purity and bounded context integrity

### 3. Architectural Trade-offs

Provide comprehensive analysis of:

- **Simplicity vs Precision**: Cognitive load vs semantic accuracy
- **Maintainability**: Long-term evolution costs of each approach
- **Developer Experience**: Onboarding, debugging, and testing implications
- **Operational Concerns**: Monitoring, alerting, and observability impact
- **Cross-cutting Concerns**: How errors span architectural boundaries

### 4. Implementation Patterns

Recommend specific Go patterns:

- Error creation and handling patterns
- Type-safe error detection and processing
- Error context preservation and propagation
- Integration with existing go-arch-lint rules

### 5. Decision Framework

Provide guidance on when to choose each approach:

- Team size and organization considerations
- System complexity and bounded context factors
- Operational requirements and compliance needs
- Long-term maintenance and evolution concerns

### 6. Hybrid Approaches

Explore middle-ground solutions:

- Interface-based semantic error layering
- Code generation for error boilerplate
- Semantic contracts within package boundaries
- Type-safe error handling without complexity explosion

## Expected Output

1. **Comprehensive Analysis**: Detailed comparison of both approaches
2. **Recommendations**: Clear guidance based on project context
3. **Implementation Patterns**: Specific Go code examples
4. **Decision Framework**: Criteria for choosing appropriate approach
5. **Architecture Documentation**: Updated architectural guidelines

## Key Questions to Address

- Does flat error centralization violate Clean Architecture principles?
- How do error layers impact domain purity and bounded context integrity?
- What are the operational implications of each approach?
- How do we maintain developer experience while ensuring architectural correctness?
- What hybrid approaches provide the best balance of simplicity and precision?

## Success Criteria

- Architecturally sound recommendations
- Practical implementation guidance
- Clear decision-making framework
- Context-aware solutions (template vs production)
- Maintainable long-term patterns

---

## Usage Instructions

### When to Use This Prompt

- **Architecture Reviews**: When evaluating error handling patterns in Go projects
- **Refactoring Decisions**: When considering error handling architecture changes
- **New Project Setup**: When establishing error handling patterns for new projects
- **Team Alignment**: When standardizing error handling across multiple teams
- **Performance Analysis**: When error handling impacts system performance

### Expected Deliverables

- Comprehensive error handling analysis document
- Implementation patterns with Go code examples
- Updated go-arch-lint configuration recommendations
- Decision framework for error handling architecture

### Prerequisites

- Understanding of Go programming language
- Familiarity with Clean Architecture principles
- Knowledge of Domain-Driven Design concepts
- Experience with enterprise system architecture
- Understanding of go-arch-lint and architectural enforcement

---

**Template Last Updated**: 2025-12-14  
**Session**: Error Handling Architecture Analysis  
**Project**: Go Linting Template
