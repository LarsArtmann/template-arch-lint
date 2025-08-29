#!/bin/bash
# 🧪 Simplified BDD Tests for Bootstrap Script
# Essential behavior-driven tests for bootstrap.sh core functionality

set -uo pipefail  # More permissive than full strict mode

# Colors and formatting
readonly GREEN='\033[0;32m'
readonly RED='\033[0;31m'
readonly BLUE='\033[0;34m'
readonly CYAN='\033[0;36m'
readonly YELLOW='\033[1;33m'
readonly BOLD='\033[1m'
readonly NC='\033[0m'

# Test results
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

log_test() {
    echo -e "\n${CYAN}🧪 TEST: $1${NC}"
    ((TESTS_TOTAL++))
}

assert_success() {
    local test_name="$1"
    local exit_code="$2"
    
    if [[ $exit_code -eq 0 ]]; then
        echo -e "  ${GREEN}✅ PASS:${NC} $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}❌ FAIL:${NC} $test_name (exit code: $exit_code)"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_name")
    fi
}

assert_failure() {
    local test_name="$1"
    local exit_code="$2"
    
    if [[ $exit_code -ne 0 ]]; then
        echo -e "  ${GREEN}✅ PASS:${NC} $test_name (correctly failed)"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}❌ FAIL:${NC} $test_name (should have failed)"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_name")
    fi
}

assert_contains() {
    local test_name="$1"
    local content="$2"
    local expected="$3"
    
    if [[ "$content" == *"$expected"* ]]; then
        echo -e "  ${GREEN}✅ PASS:${NC} $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "  ${RED}❌ FAIL:${NC} $test_name (missing: $expected)"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_name")
    fi
    ((TESTS_TOTAL++))
}

# Test execution with error handling
run_test_command() {
    local cmd="$1"
    set +e  # Allow commands to fail for testing
    local output
    output=$(eval "$cmd" 2>&1)
    local exit_code=$?
    set -e
    echo "$exit_code|$output"
}

# ============================================================================
# CORE BDD TESTS
# ============================================================================

echo -e "${BOLD}${BLUE}🧪 BOOTSTRAP BDD TEST SUITE (Simplified)${NC}"
echo -e "${BLUE}Testing core bootstrap.sh functionality${NC}\n"

# Test 1: Help flag functionality
log_test "Bootstrap --help flag works correctly"
result=$(run_test_command "./bootstrap.sh --help")
exit_code="${result%%|*}"
output="${result#*|}"

assert_success "Help flag execution" "$exit_code"
assert_contains "Help shows usage" "$output" "USAGE:"
assert_contains "Help shows options" "$output" "OPTIONS:"
assert_contains "Help shows examples" "$output" "EXAMPLES:"

# Test 2: Diagnostic mode in current environment
log_test "Bootstrap --diagnose works in current environment"
result=$(run_test_command "./bootstrap.sh --diagnose")
exit_code="${result%%|*}"
output="${result#*|}"

assert_success "Diagnostic mode execution" "$exit_code"
assert_contains "Diagnose checks environment" "$output" "COMPREHENSIVE ENVIRONMENT DIAGNOSTICS"
assert_contains "Diagnose shows summary" "$output" "DIAGNOSTIC SUMMARY"

# Test 3: Verbose mode provides debug info  
log_test "Bootstrap --diagnose --verbose provides debug information"
result=$(run_test_command "./bootstrap.sh --diagnose --verbose")
exit_code="${result%%|*}"
output="${result#*|}"

assert_success "Verbose mode execution" "$exit_code"
assert_contains "Verbose shows debug" "$output" "DEBUG:"
assert_contains "Verbose shows modes" "$output" "Modes:"

# Test 4: Invalid flag handling
log_test "Bootstrap handles invalid flags gracefully"
result=$(run_test_command "./bootstrap.sh --invalid-nonexistent-flag")
exit_code="${result%%|*}"
output="${result#*|}"

assert_failure "Invalid flag rejection" "$exit_code"
assert_contains "Invalid flag error" "$output" "Unknown option"
assert_contains "Invalid flag help" "$output" "Use --help"

# Test 5: Multiple flags work together
log_test "Bootstrap handles multiple flags correctly"
result=$(run_test_command "./bootstrap.sh --diagnose --verbose")
exit_code="${result%%|*}"
output="${result#*|}"

assert_success "Multi-flag execution" "$exit_code" 
assert_contains "Multi-flag verbose" "$output" "Verbose mode enabled"
assert_contains "Multi-flag diagnose" "$output" "DIAGNOSTICS"

# Test 6: Error handling consolidation test
log_test "Bootstrap error handling provides actionable suggestions"
# Create a temporary directory without git to test error handling
temp_dir="/tmp/bdd-no-git-$$"
mkdir -p "$temp_dir"
cd "$temp_dir"
cp /Users/larsartmann/projects/template-arch-lint/bootstrap.sh .
chmod +x bootstrap.sh

# Create go.mod but no git repo
cat > go.mod <<EOF
module test-project
go 1.21
EOF

result=$(run_test_command "./bootstrap.sh --diagnose")
exit_code="${result%%|*}"
output="${result#*|}"

cd /Users/larsartmann/projects/template-arch-lint
rm -rf "$temp_dir"

assert_failure "No git repo error" "$exit_code"
assert_contains "No git error message" "$output" "No .git directory found"
assert_contains "Git fix suggestion" "$output" "Fix: Run 'git init'"

# ============================================================================
# TEST SUMMARY
# ============================================================================

echo -e "\n${BOLD}${BLUE}📊 BDD TEST RESULTS${NC}"
echo -e "${BLUE}$(printf '=%.0s' $(seq 1 20))${NC}"

echo -e "\nSummary:"
echo -e "  Total Tests: $TESTS_TOTAL"
echo -e "  ${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "  ${RED}Failed: $TESTS_FAILED${NC}"

if [[ $TESTS_FAILED -eq 0 ]]; then
    echo -e "\n${BOLD}${GREEN}🎉 ALL BDD TESTS PASSED!${NC}"
    echo -e "${GREEN}Bootstrap script core functionality verified.${NC}"
    
    success_rate=$((TESTS_PASSED * 100 / TESTS_TOTAL))
    echo -e "${GREEN}Success Rate: ${success_rate}%${NC}"
    exit 0
else
    echo -e "\n${BOLD}${RED}❌ SOME BDD TESTS FAILED${NC}"
    echo -e "${RED}Failed tests:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}• $test${NC}"
    done
    
    echo -e "\n${YELLOW}💡 Action Required:${NC}"
    echo -e "  Review failed tests and fix bootstrap.sh issues"
    exit 1
fi