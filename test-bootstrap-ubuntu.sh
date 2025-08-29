#!/bin/bash
# Test bootstrap.sh in Ubuntu Docker container

set -euo pipefail

echo "🧪 Testing bootstrap.sh in Ubuntu Docker container..."

# Create test directory structure
TEST_DIR="/tmp/bootstrap-test-$$"
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"

echo "📁 Created test directory: $TEST_DIR"

# Create a minimal Go project
cat > go.mod <<EOF
module test-project

go 1.21
EOF

cat > main.go <<EOF
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
EOF

# Initialize git repo
git init
git add .
git commit -m "Initial commit"

echo "✅ Created minimal Go project with git"

# Run the bootstrap script in Ubuntu Docker
echo "🚀 Running bootstrap.sh in Ubuntu container..."

# Copy the bootstrap script to the test directory
cp /Users/larsartmann/projects/template-arch-lint/bootstrap.sh .

# Test in Ubuntu container
docker run --rm -v "$TEST_DIR:/workspace" -w /workspace ubuntu:22.04 bash -c "
    apt-get update -qq &&
    apt-get install -y curl git build-essential -qq &&
    
    # Install Go
    wget -q https://go.dev/dl/go1.21.5.linux-amd64.tar.gz &&
    tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz &&
    export PATH=/usr/local/go/bin:\$PATH &&
    
    # Verify Go installation  
    go version &&
    
    # Test our bootstrap script
    chmod +x bootstrap.sh &&
    ./bootstrap.sh
"

# Verify results
if [ $? -eq 0 ]; then
    echo "✅ Bootstrap script succeeded in Ubuntu Docker!"
    
    # Check if required files were created
    if [ -f "$TEST_DIR/.go-arch-lint.yml" ] && [ -f "$TEST_DIR/.golangci.yml" ] && [ -f "$TEST_DIR/justfile" ]; then
        echo "✅ Configuration files successfully created"
    else
        echo "❌ Some configuration files missing"
        ls -la "$TEST_DIR"
    fi
else
    echo "❌ Bootstrap script failed in Ubuntu Docker"
    exit 1
fi

# Cleanup
rm -rf "$TEST_DIR"
echo "🧹 Cleaned up test directory"

echo "🎉 Ubuntu Docker test completed successfully!"