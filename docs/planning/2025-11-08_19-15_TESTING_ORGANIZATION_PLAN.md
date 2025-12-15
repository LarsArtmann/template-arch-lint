# Test Organization Plan

## Current Test Files Analysis

Based on the project structure and justfile, I can see:

### Existing Test Infrastructure:

1. **`bootstrap-test`** command in justfile (line 180) - downloads and runs BDD tests
2. **Test scripts** that need to be organized:
   - `test-bootstrap-simple-bdd.sh` - BDD tests for bootstrap functionality
   - `test-path-verification.sh` - PATH verification tests
   - `test-complete-workflow.sh` - Complete workflow tests
   - `test_rate_limiting.sh` - Rate limiting tests (seems misplaced)
   - `test_graceful_shutdown.sh` - Graceful shutdown tests (seems misplaced)

### Issues Identified:

1. **Scattered tests**: Test files are not organized in a proper directory structure
2. **Misplaced tests**: Some tests (rate limiting, graceful shutdown) seem to be for a web server, not a linting template
3. **Inconsistent naming**: Some files use hyphens, others use underscores
4. **Missing integration**: Tests are not properly integrated into the CI workflow

## Proposed Test Organization Structure

```
tests/
├── bootstrap/                 # Bootstrap script tests
│   ├── bdd/                  # BDD tests
│   │   ├── bootstrap-bdd.sh
│   │   └── features/        # BDD feature files
│   ├── integration/          # Integration tests
│   │   ├── path-verification.sh
│   │   └── complete-workflow.sh
│   └── unit/                 # Unit tests for individual functions
│       └── bootstrap-unit.sh
├── linting/                  # Linting configuration tests
│   ├── architecture/         # go-arch-lint tests
│   ├── golangci/            # golangci-lint tests
│   └── security/            # Security linting tests
├── scripts/                 # Test utility scripts
│   ├── test-runner.sh       # Main test runner
│   ├── setup-test-env.sh    # Test environment setup
│   └── cleanup-test-env.sh  # Test environment cleanup
└── integration/             # Full integration tests
    ├── ci-pipeline.sh       # CI pipeline simulation
    └── end-to-end.sh        # End-to-end tests
```

## Test Categories and Responsibilities

### 1. Bootstrap Tests (`tests/bootstrap/`)

- **BDD Tests**: Behavior-driven tests for bootstrap script functionality
- **Integration Tests**: PATH verification, complete workflow tests
- **Unit Tests**: Individual function testing

### 2. Linting Tests (`tests/linting/`)

- **Architecture Tests**: go-arch-lint configuration validation
- **Code Quality Tests**: golangci-lint rule validation
- **Security Tests**: Security linting rule validation

### 3. Integration Tests (`tests/integration/`)

- **CI Pipeline Tests**: Simulate complete CI pipeline
- **End-to-End Tests**: Full workflow from bootstrap to linting

### 4. Test Scripts (`tests/scripts/`)

- **Test Runner**: Unified test execution script
- **Environment Setup**: Prepare test environments
- **Environment Cleanup**: Clean up after tests

## Justfile Integration Plan

### New Test Commands:

```bash
just test              # Run all tests
just test-bootstrap    # Run bootstrap tests only
just test-linting       # Run linting configuration tests
just test-integration   # Run integration tests
just test-unit          # Run unit tests only
just test-clean         # Clean up test artifacts
```

### Enhanced CI Integration:

- Update `ci` command to include comprehensive test suite
- Add test coverage reporting
- Add test result archiving

## Implementation Steps

1. **Create Directory Structure**: Set up the proposed test directory structure
2. **Organize Existing Tests**: Move and rename existing test files
3. **Create Test Runner**: Implement unified test execution script
4. **Update Justfile**: Add new test commands and update CI integration
5. **Create Missing Tests**: Add tests for linting configurations
6. **Documentation**: Update README with test organization information

## File Renaming Convention

- Use hyphens for file names (consistent with project style)
- Use descriptive names that indicate test type and scope
- Group related tests in subdirectories

## Test Dependencies

- **BDD Framework**: Use existing BDD approach or standardize on a framework
- **Test Utilities**: Create shared test utility functions
- **Mock Data**: Prepare test data and mock environments
- **CI Integration**: Ensure tests run properly in CI environment

This organization will provide:

- Clear separation of test concerns
- Easy maintenance and extensibility
- Better CI/CD integration
- Comprehensive test coverage for all project aspects
  </file>
