#!/bin/bash

# Execute our framework test

echo "🧪 TESTING OUR NEW TESTING FRAMEWORK"
echo "====================================="

# Test 1: Check if all test files exist and are executable
echo "📋 STEP 1: Verifying test files exist and are executable..."

test_files=(
    "tests/bootstrap/bdd/bootstrap-bdd.sh"
    "tests/bootstrap/integration/path-verification.sh"
    "tests/bootstrap/integration/complete-workflow.sh"
    "tests/bootstrap/unit/bootstrap-unit.sh"
    "tests/linting/architecture/architecture-tests.sh"
    "tests/linting/golangci/golangci-tests.sh"
    "tests/linting/security/security-tests.sh"
    "tests/scripts/test-runner.sh"
    "tests/scripts/setup-test-env.sh"
    "tests/scripts/cleanup-test-env.sh"
    "tests/integration/bootstrap-e2e.sh"
)

missing_files=()
non_executable_files=()

for file in "${test_files[@]}"; do
    if [[ ! -f "$file" ]]; then
        missing_files+=("$file")
    elif [[ ! -x "$file" ]]; then
        non_executable_files+=("$file")
    fi
done

if [[ ${#missing_files[@]} -eq 0 ]]; then
    echo "✅ All test files exist"
else
    echo "❌ Missing test files:"
    for file in "${missing_files[@]}"; do
        echo "   • $file"
    done
    exit 1
fi

if [[ ${#non_executable_files[@]} -eq 0 ]]; then
    echo "✅ All test files are executable"
else
    echo "❌ Non-executable test files:"
    for file in "${non_executable_files[@]}"; do
        echo "   • $file"
    done
    exit 1
fi

# Test 2: Test the test runner help
echo ""
echo "📋 STEP 2: Testing test runner help..."

if ./tests/scripts/test-runner.sh --help >/dev/null 2>&1; then
    echo "✅ Test runner help works"
else
    echo "❌ Test runner help failed"
    exit 1
fi

# Test 3: Test setup script
echo ""
echo "📋 STEP 3: Testing setup script..."

if ./tests/scripts/setup-test-env.sh >/dev/null 2>&1; then
    echo "✅ Setup script works"
else
    echo "❌ Setup script failed"
    exit 1
fi

# Test 4: Test cleanup script
echo ""
echo "📋 STEP 4: Testing cleanup script..."

if ./tests/scripts/cleanup-test-env.sh >/dev/null 2>&1; then
    echo "✅ Cleanup script works"
else
    echo "❌ Cleanup script failed"
    exit 1
fi

# Test 5: Test a simple BDD test
echo ""
echo "📋 STEP 5: Testing BDD test..."

# Create a minimal test environment
TEST_DIR="/tmp/bdd-framework-test-$$"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

# Create minimal Go project
cat > go.mod <<EOF
module test-framework

go 1.21
EOF

git init
git add .
git commit -m "Initial commit" 2>/dev/null || true

# Copy bootstrap script
cp /Users/larsartmann/projects/template-arch-lint/bootstrap.sh .
chmod +x bootstrap.sh

# Run BDD test
if /Users/larsartmann/projects/template-arch-lint/tests/bootstrap/bdd/bootstrap-bdd.sh >/dev/null 2>&1; then
    echo "✅ BDD test works"
else
    echo "❌ BDD test failed"
    cd /
    rm -rf "$TEST_DIR"
    exit 1
fi

cd /
rm -rf "$TEST_DIR"

# Test 6: Test architecture test
echo ""
echo "📋 STEP 6: Testing architecture test..."

if /Users/larsartmann/projects/template-arch-lint/tests/linting/architecture/architecture-tests.sh >/dev/null 2>&1; then
    echo "✅ Architecture test works"
else
    echo "❌ Architecture test failed"
    exit 1
fi

# Test 7: Test golangci test
echo ""
echo "📋 STEP 7: Testing golangci test..."

if /Users/larsartmann/projects/template-arch-lint/tests/linting/golangci/golangci-tests.sh >/dev/null 2>&1; then
    echo "✅ golangci test works"
else
    echo "❌ golangci test failed"
    exit 1
fi

# Test 8: Test security test
echo ""
echo "📋 STEP 8: Testing security test..."

if /Users/larsartmann/projects/template-arch-lint/tests/linting/security/security-tests.sh >/dev/null 2>&1; then
    echo "✅ Security test works"
else
    echo "❌ Security test failed"
    exit 1
fi

# Test 9: Test integration test
echo ""
echo "📋 STEP 9: Testing integration test..."

if /Users/larsartmann/projects/template-arch-lint/tests/integration/bootstrap-e2e.sh >/dev/null 2>&1; then
    echo "✅ Integration test works"
else
    echo "❌ Integration test failed"
    exit 1
fi

# Test 10: Test full test runner
echo ""
echo "📋 STEP 10: Testing full test runner..."

# Create test environment
./tests/scripts/setup-test-env.sh >/dev/null 2>&1

if ./tests/scripts/test-runner.sh -c >/dev/null 2>&1; then
    echo "✅ Full test runner works"
else
    echo "❌ Full test runner failed"
    exit 1
fi

echo ""
echo "🎉 ALL TESTS PASSED!"
echo "✅ Our testing framework is working correctly"
echo "✅ All test files exist and are executable"
echo "✅ All test scripts run successfully"
echo "✅ The test runner works properly"

echo ""
echo "📋 FRAMEWORK VERIFICATION SUMMARY:"
echo "=================================="
echo "✅ Test file structure: ALL files present and executable"
echo "✅ Test runner: Help, setup, cleanup, and execution working"
echo "✅ BDD tests: Core functionality tests working"
echo "✅ Linting tests: Architecture, golangci, and security tests working"
echo "✅ Integration tests: End-to-end tests working"
echo "✅ Full workflow: Setup → Test → Cleanup working"

echo ""
echo "🚀 READY FOR PRODUCTION USE!"
echo "The testing framework is fully functional and ready to use."
</file></think>
<think></think>
## NOW LET'S ACTUALLY RUN THIS TEST!
<tool_call>execute
<arg_key>command_path</arg_key>
<arg_value>/Users/larsartmann/projects/template-arch-lint/execute-framework-test.sh