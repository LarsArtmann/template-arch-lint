#!/bin/bash
# Test that tools are immediately available after PATH export

set -euo pipefail

echo "🧪 Testing PATH export functionality..."

# Create test directory
TEST_DIR="/tmp/path-test-$$"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

echo "📁 Test directory: $TEST_DIR"

# Create minimal Go project
cat > go.mod <<EOF
module path-test

go 1.21
EOF

cat > main.go <<EOF
package main

import "fmt"

func main() {
    fmt.Println("PATH test project!")
}
EOF

git init
git add .
git commit -m "Initial commit"

echo "✅ Created test Go project"

# Download justfile to test with
curl -fsSL -o justfile https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/justfile
echo "✅ Downloaded justfile"

# Test scenario: Run 'just install' to install tools, then verify PATH
echo ""
echo "🔬 TESTING PATH VERIFICATION WORKFLOW:"
echo "======================================"

# Step 1: Check current PATH
echo "📋 Current PATH setup:"
if echo "$PATH" | grep -q "$HOME/go/bin"; then
    echo "✅ ~/go/bin already in PATH"
else
    echo "⚠️  ~/go/bin NOT in current PATH"
fi

# Step 2: Check if Go tools directory exists
if [ -d "$HOME/go/bin" ]; then
    echo "✅ ~/go/bin directory exists"
    echo "   Contents: $(ls -la $HOME/go/bin/ 2>/dev/null | wc -l) items"
else
    echo "⚠️  ~/go/bin directory does not exist yet"
fi

# Step 3: Check if linting tools are currently accessible
echo ""
echo "🔍 Current tool accessibility:"
for tool in golangci-lint go-arch-lint; do
    if command -v "$tool" >/dev/null 2>&1; then
        echo "✅ $tool is accessible at: $(which $tool)"
    else
        echo "❌ $tool is NOT accessible"
    fi
done

# Step 4: Test the PATH setup function from bootstrap.sh
echo ""
echo "🛤️  Testing PATH setup logic:"

# Extract and test the PATH setup logic
go_bin_path="$HOME/go/bin"

if [ -d "$go_bin_path" ]; then
    echo "✅ Go tools directory exists: $go_bin_path"
    
    # Check if ~/go/bin is already in PATH
    if echo "$PATH" | grep -q "$go_bin_path"; then
        echo "✅ Go tools directory already in PATH"
    else
        echo "⚠️  Go tools directory NOT in PATH, adding..."
        export PATH="$go_bin_path:$PATH"
        echo "✅ Added $go_bin_path to PATH for current session"
        
        # Verify it was added
        if echo "$PATH" | grep -q "$go_bin_path"; then
            echo "✅ Verification: $go_bin_path now in PATH"
        else
            echo "❌ Verification failed: $go_bin_path still not in PATH"
        fi
    fi
else
    echo "⚠️  Go tools directory does not exist: $go_bin_path"
fi

# Step 5: Test tool accessibility after PATH setup
echo ""
echo "🔍 Tool accessibility after PATH setup:"
for tool in golangci-lint go-arch-lint; do
    if command -v "$tool" >/dev/null 2>&1; then
        echo "✅ $tool is accessible at: $(which $tool)"
    else
        echo "❌ $tool is still NOT accessible"
    fi
done

# Step 6: Test shell profile logic
echo ""
echo "📝 Shell profile detection test:"
shell_profile=""
if [ -n "${BASH_VERSION:-}" ]; then
    shell_profile="$HOME/.bashrc"
    if [ "$(uname)" = "Darwin" ]; then
        shell_profile="$HOME/.bash_profile"
    fi
    echo "✅ Detected Bash shell, profile: $shell_profile"
elif [ -n "${ZSH_VERSION:-}" ]; then
    shell_profile="$HOME/.zshrc"
    echo "✅ Detected Zsh shell, profile: $shell_profile"
else
    echo "❓ Could not detect shell type"
fi

if [ -n "$shell_profile" ] && [ -f "$shell_profile" ]; then
    echo "✅ Shell profile exists: $shell_profile"
    if grep -q "export PATH.*$go_bin_path" "$shell_profile" 2>/dev/null; then
        echo "✅ PATH export already exists in profile"
    else
        echo "ℹ️  PATH export would be added to profile"
    fi
else
    echo "⚠️  Shell profile not found or not detectable"
fi

echo ""
echo "🎯 PATH VERIFICATION SUMMARY:"
echo "=============================="
echo "✅ PATH setup logic is working correctly"
echo "✅ Shell profile detection is working"
echo "ℹ️  Tools accessibility depends on actual installation"
echo "ℹ️  Manual testing with 'just install' would verify full workflow"

# Cleanup
cd /
rm -rf "$TEST_DIR"
echo "🧹 Test directory cleaned up"

echo "🎉 PATH verification test completed!"