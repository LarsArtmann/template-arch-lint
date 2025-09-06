#!/bin/bash

# Comprehensive test suite for cmd/ single main enforcement
# Tests all scenarios safely using temporary directories

set -euo pipefail

# Test configuration
TEST_DIR="/tmp/cmd-single-test-$$"
SCRIPT_PATH="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/check-cmd-single.sh"
TESTS_PASSED=0
TESTS_FAILED=0

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Cleanup function
cleanup() {
    if [ -d "$TEST_DIR" ]; then
        rm -rf "$TEST_DIR"
    fi
}
trap cleanup EXIT

# Test result tracking
pass_test() {
    local test_name="$1"
    echo -e "${GREEN}✅ PASS${NC}: $test_name"
    ((TESTS_PASSED++))
}

fail_test() {
    local test_name="$1"
    local details="$2"
    echo -e "${RED}❌ FAIL${NC}: $test_name"
    echo -e "${RED}   Details: $details${NC}"
    ((TESTS_FAILED++))
}

# Test runner function
run_test() {
    local test_name="$1"
    local expected_exit_code="$2"
    local test_dir="$TEST_DIR/$test_name"
    
    echo -e "${BLUE}🧪 Testing${NC}: $test_name"
    
    # Create test directory
    mkdir -p "$test_dir"
    cd "$test_dir"
    
    # Run the script and capture result
    set +e
    output=$("$SCRIPT_PATH" 2>&1)
    actual_exit_code=$?
    set -e
    
    # Validate exit code
    if [ "$actual_exit_code" -eq "$expected_exit_code" ]; then
        pass_test "$test_name"
        echo -e "${CYAN}   Output preview:${NC} $(echo "$output" | head -n 1 | cut -c1-60)..."
    else
        fail_test "$test_name" "Expected exit code $expected_exit_code, got $actual_exit_code"
        echo -e "${RED}   Full output:${NC}"
        echo "$output" | sed 's/^/     /'
    fi
    
    echo ""
}

echo -e "${PURPLE}🎯 CMD SINGLE MAIN ENFORCEMENT - COMPREHENSIVE TEST SUITE${NC}"
echo -e "${CYAN}=========================================================${NC}"
echo ""

# Test 1: No cmd/ directory
echo -e "${YELLOW}📋 Phase 1: Directory Existence Tests${NC}"
run_test "no-cmd-directory" 1

# Test 2: Empty cmd/ directory  
mkdir -p "$TEST_DIR/empty-cmd-directory/cmd"
cd "$TEST_DIR/empty-cmd-directory"
run_test "empty-cmd-directory" 1

# Test 3: Single main.go file (success case)
test_dir="$TEST_DIR/single-main-success"
mkdir -p "$test_dir/cmd/server"
echo -e "package main\n\nfunc main() {\n\tprintln(\"Hello World\")\n}" > "$test_dir/cmd/server/main.go"
cd "$test_dir"
run_test "single-main-success" 0

echo -e "${YELLOW}📋 Phase 2: Multiple Files Tests${NC}"

# Test 4: Two main.go files (failure case)
test_dir="$TEST_DIR/two-main-files"
mkdir -p "$test_dir/cmd/server" "$test_dir/cmd/cli"
echo -e "package main\n\nfunc main() {\n\tprintln(\"Server\")\n}" > "$test_dir/cmd/server/main.go"
echo -e "package main\n\nfunc main() {\n\tprintln(\"CLI\")\n}" > "$test_dir/cmd/cli/main.go"
cd "$test_dir"
run_test "two-main-files" 1

# Test 5: Three main.go files (failure case)
test_dir="$TEST_DIR/three-main-files"
mkdir -p "$test_dir/cmd/server" "$test_dir/cmd/cli" "$test_dir/cmd/worker"
echo -e "package main\n\nfunc main() {\n\tprintln(\"Server\")\n}" > "$test_dir/cmd/server/main.go"
echo -e "package main\n\nfunc main() {\n\tprintln(\"CLI\")\n}" > "$test_dir/cmd/cli/main.go"  
echo -e "package main\n\nfunc main() {\n\tprintln(\"Worker\")\n}" > "$test_dir/cmd/worker/main.go"
cd "$test_dir"
run_test "three-main-files" 1

echo -e "${YELLOW}📋 Phase 3: Edge Cases Tests${NC}"

# Test 6: main.go with wrong package (warning test)
test_dir="$TEST_DIR/wrong-package"
mkdir -p "$test_dir/cmd/server"
echo -e "package server\n\nfunc main() {\n\tprintln(\"Wrong package\")\n}" > "$test_dir/cmd/server/main.go"
cd "$test_dir"
run_test "wrong-package" 0  # Should pass but with warning

# Test 7: main.go without main function (warning test)
test_dir="$TEST_DIR/no-main-func"
mkdir -p "$test_dir/cmd/server"
echo -e "package main\n\nfunc init() {\n\tprintln(\"No main function\")\n}" > "$test_dir/cmd/server/main.go"
cd "$test_dir"
run_test "no-main-func" 0  # Should pass but with warning

# Test 8: Symbolic link to main.go
if command -v ln >/dev/null 2>&1; then
    test_dir="$TEST_DIR/symlink-main"
    mkdir -p "$test_dir/cmd/server" "$test_dir/shared"
    echo -e "package main\n\nfunc main() {\n\tprintln(\"Symlinked main\")\n}" > "$test_dir/shared/main.go"
    ln -s "../../shared/main.go" "$test_dir/cmd/server/main.go"
    cd "$test_dir"
    run_test "symlink-main" 0
fi

# Test 9: Permission denied on cmd/ directory
test_dir="$TEST_DIR/permission-denied"
mkdir -p "$test_dir/cmd"
chmod 000 "$test_dir/cmd" 2>/dev/null || true
cd "$test_dir"
# This test might not work on all systems, so we'll skip if chmod doesn't work
if [ ! -r "cmd" ]; then
    run_test "permission-denied" 1
else
    echo -e "${YELLOW}⚠️  SKIP: permission-denied (chmod not effective on this system)${NC}"
fi
chmod 755 "$test_dir/cmd" 2>/dev/null || true

echo -e "${YELLOW}📋 Phase 4: Complex Structure Tests${NC}"

# Test 10: Nested subdirectories with single main.go
test_dir="$TEST_DIR/nested-structure"
mkdir -p "$test_dir/cmd/server/app"
echo -e "package main\n\nfunc main() {\n\tprintln(\"Nested main\")\n}" > "$test_dir/cmd/server/main.go"
# Add some non-main files
echo -e "package app\n\nfunc Helper() {}" > "$test_dir/cmd/server/app/helper.go"
mkdir -p "$test_dir/cmd/docs"
echo "README content" > "$test_dir/cmd/docs/README.md"
cd "$test_dir"
run_test "nested-structure" 0

# Test 11: Hidden files and unusual names  
test_dir="$TEST_DIR/hidden-files"
mkdir -p "$test_dir/cmd/server"
echo -e "package main\n\nfunc main() {\n\tprintln(\"Regular main\")\n}" > "$test_dir/cmd/server/main.go"
echo -e "package main\n\nfunc main() {\n\tprintln(\"Hidden main\")\n}" > "$test_dir/cmd/.hidden-main.go"
touch "$test_dir/cmd/not-main.go"
cd "$test_dir"
run_test "hidden-files" 0  # Should only find the regular main.go

# Summary
echo -e "${PURPLE}📊 TEST RESULTS SUMMARY${NC}"
echo -e "${CYAN}========================${NC}"
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
total_tests=$((TESTS_PASSED + TESTS_FAILED))
echo -e "Total tests:  ${BLUE}$total_tests${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}🎉 ALL TESTS PASSED! CMD single main enforcement is working correctly.${NC}"
    exit 0
else
    echo -e "\n${RED}💥 $TESTS_FAILED TESTS FAILED! Please review the failures above.${NC}"
    exit 1
fi