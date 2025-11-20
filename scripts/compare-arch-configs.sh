#!/bin/bash
# Architecture Configuration Comparison Script
# Demonstrates difference between permissive vs strict vendor control

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 Architecture Configuration Comparison Demo${NC}"
echo "==================================="
echo

# Function to run check and capture output
run_check() {
    local config_file="$1"
    local description="$2"
    
    echo -e "${YELLOW}Testing: $description${NC}"
    echo "Config: $config_file"
    echo "----------------------------------------"
    
    if go-arch-lint check --arch-file "$config_file" 2>&1; then
        echo -e "${GREEN}✅ PASSED - No violations found${NC}"
    else
        local exit_code=$?
        echo -e "${RED}❌ FAILED - Violations detected (exit code: $exit_code)${NC}"
    fi
    echo
}

# Function to create test file with banned library
create_banned_test() {
    cat > /tmp/test-banned.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"           // Allowed
    "github.com/sirupsen/logrus"        // Banned per library-policy.yaml
    "github.com/dgrijalva/jwt-go"       // CVE-2020-26160 - Security risk
)

func main() {
    gin.New()
    logrus.Info("This should fail strict validation")
}
EOF
}

# Function to create test file with approved libraries only
create_approved_test() {
    cat > /tmp/test-approved.go << 'EOF'
package main

import (
    "context"
    "github.com/gin-gonic/gin"           // Approved
    "github.com/samber/lo"               // Approved
    "github.com/charmbracelet/log"       // Approved
)

func main() {
    gin.New()
    lo.Map([]int{1, 2, 3}, func(i int) int { return i * 2 })
    log.Info("This should pass strict validation")
}
EOF
}

# Function to cleanup test files
cleanup() {
    rm -f /tmp/test-banned.go /tmp/test-approved.go
}

# Cleanup on exit
trap cleanup EXIT

echo -e "${BLUE}📋 Creating test scenarios...${NC}"
echo

# Create test files
create_banned_test
create_approved_test

echo -e "${BLUE}🚫 Testing with BANNED libraries (permissive vs strict)${NC}"
echo "=========================================================="

# Test with permissive config
echo
run_check ".go-arch-lint.yml" "Permissive Configuration (anyVendorDeps: true)"

# Test with strict config  
echo
run_check ".go-arch-lint-strict.yml" "Strict Configuration (anyVendorDeps: false)"

echo -e "${BLUE}✅ Testing with APPROVED libraries only${NC}"
echo "============================================="

# Replace banned test with approved test
mv /tmp/test-approved.go /tmp/test-banned.go

# Test with permissive config
echo
run_check ".go-arch-lint.yml" "Permissive Configuration (anyVendorDeps: true)"

# Test with strict config
echo  
run_check ".go-arch-lint-strict.yml" "Strict Configuration (anyVendorDeps: false)"

echo -e "${BLUE}📊 Summary of Differences${NC}"
echo "========================="
echo
echo -e "${YELLOW}Permissive Configuration (.go-arch-lint.yml):${NC}"
echo "  ❌ Allows ANY vendor library (anyVendorDeps: true)"
echo "  ❌ No security control over dependencies"
echo "  ❌ Banned/deprecated libraries allowed"
echo "  ❌ Library-policy.yaml not enforced"
echo "  ❌ Deep scanning disabled (if configured)"
echo
echo -e "${GREEN}Strict Configuration (.go-arch-lint-strict.yml):${NC}"
echo "  ✅ Only APPROVED vendor libraries allowed"
echo "  ✅ Enforces library-policy.yaml recommendations"
echo "  ✅ Banned/deprecated libraries automatically blocked"
echo "  ✅ Security vulnerabilities prevented"
echo "  ✅ Deep scanning enabled for comprehensive analysis"
echo "  ✅ Explicit dependency control and visibility"
echo
echo -e "${BLUE}🎯 Recommendation:${NC}"
echo "Use strict configuration for production projects requiring:"
echo "  - Security compliance"
echo "  - Vendor dependency control" 
echo "  - Library policy enforcement"
echo "  - Architectural purity"
echo "  - Team consistency"
echo
echo -e "${GREEN}Files created:${NC}"
echo "  - .go-arch-lint-strict.yml (strict configuration)"
echo "  - docs/STRICT_VENDOR_CONTROL_EXAMPLE.md (comprehensive guide)"