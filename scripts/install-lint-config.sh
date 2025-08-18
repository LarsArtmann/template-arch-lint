#!/bin/bash
# Template Architecture Lint - Quick Install
# Extracts only the essential linting files using git subtree

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BOLD}${BLUE}🚀 Installing Template Architecture Lint${NC}"

# Check if we're in a git repo
if [ ! -d ".git" ]; then
    echo "❌ Not in a git repository. Please run from the root of a git project."
    exit 1
fi

# Check if go.mod exists  
if [ ! -f "go.mod" ]; then
    echo "❌ No go.mod found. Please run from the root of a Go project."
    exit 1
fi

echo "✅ Go project detected"

# Use git subtree to get only what we need
echo "📥 Pulling linting configuration..."
git subtree add --prefix=.lint-config https://github.com/LarsArtmann/template-arch-lint.git master --squash

echo "📋 Extracting essential files..."
cp .lint-config/.go-arch-lint.yml .
cp .lint-config/.golangci.yml .
cp .lint-config/justfile linting.just

echo "🧹 Cleaning up temporary directory..."
rm -rf .lint-config

echo -e "${GREEN}✅ Installation complete!${NC}"
echo ""
echo "📝 Files added:"
echo "  • .go-arch-lint.yml  (Architecture boundaries)"
echo "  • .golangci.yml      (Code quality rules)"  
echo "  • linting.just       (Development commands)"
echo ""
echo "🚀 Next steps:"
echo "  1. just install      (Install linting tools)"
echo "  2. just lint         (Run complete linting)"
echo "  3. just lint-arch    (Architecture only)"
echo ""
echo -e "${BOLD}Happy linting! 🎉${NC}"