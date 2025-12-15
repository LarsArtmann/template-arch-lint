# 🧪 Test Organization

This directory contains organized test scripts for the bootstrap.sh functionality.

## Directory Structure

```
tests/
├── bootstrap/           # Bootstrap script specific tests
│   ├── test-bootstrap-bdd.sh        # Comprehensive BDD tests
│   ├── test-bootstrap-local.sh      # Local environment tests
│   ├── test-bootstrap-simple-bdd.sh # Simplified BDD tests
│   └── test-bootstrap-ubuntu.sh     # Ubuntu Docker container tests
├── integration/        # Integration and component tests
│   ├── test-path-verification.sh    # PATH and tool verification
│   └── test-simple-bootstrap.sh     # Simple bootstrap functionality
└── workflow/          # End-to-end workflow tests
    └── test-complete-workflow.sh    # Complete bootstrap → usage workflow
```

## Test Categories

### Bootstrap Tests (`tests/bootstrap/`)

- **BDD Tests**: Comprehensive behavior-driven development tests
- **Local Tests**: Local environment validation
- **Simple BDD**: Essential BDD functionality tests
- **Ubuntu Tests**: Docker container compatibility

### Integration Tests (`tests/integration/`)

- **PATH Verification**: Tool installation and PATH setup
- **Simple Bootstrap**: Basic bootstrap functionality validation

### Workflow Tests (`tests/workflow/`)

- **Complete Workflow**: End-to-end bootstrap → immediate usage validation

## Running Tests

### Individual Tests

```bash
# Run specific bootstrap test
./tests/bootstrap/test-bootstrap-bdd.sh

# Run integration test
./tests/integration/test-path-verification.sh

# Run workflow test
./tests/workflow/test-complete-workflow.sh
```

### All Tests in Category

```bash
# Run all bootstrap tests
for test in tests/bootstrap/*.sh; do echo "Running $test"; "$test"; done

# Run all integration tests
for test in tests/integration/*.sh; do echo "Running $test"; "$test"; done

# Run all workflow tests
for test in tests/workflow/*.sh; do echo "Running $test"; "$test"; done
```

### All Tests

```bash
# Run all tests
find tests -name "*.sh" -executable -exec echo "Running {}" \; -exec {} \;
```

## Test Dependencies

All test scripts require:

- `bash` (tested with Bash 4+)
- `git` (for repository operations)
- `curl` (for downloading configuration files)
- `docker` (for Ubuntu container tests)

## Test Output

Tests provide detailed output including:

- ✅ Success indicators
- ❌ Failure indicators with actionable error messages
- 📊 Summary statistics
- 💡 Recommendations for failed tests

## Contributing

When adding new tests:

1. Place in appropriate category directory
2. Follow existing naming convention: `test-<description>.sh`
3. Make executable: `chmod +x tests/<category>/test-<description>.sh`
4. Update this README with new test description
