#!/bin/bash
# Template Architecture Lint - Quick Install
# Extracts the essential architecture config using git subtree

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
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
cp .lint-config/.go-arch-lint-strict.yml . 2>/dev/null || true

echo "🧹 Cleaning up temporary directory..."
rm -rf .lint-config

echo -e "${GREEN}✅ Architecture config installed!${NC}"
echo ""
echo "📝 Files added:"
echo "  • .go-arch-lint.yml         (Architecture boundaries)"
echo "  • .go-arch-lint-strict.yml  (Strict variant, optional)"
echo ""

echo "🚀 Next steps:"
echo "  1. go-arch-lint check       (Validate architecture)"
echo "  2. Adapt component paths in .go-arch-lint.yml to your layout"
if command -v golangci-lint-auto-configure >/dev/null 2>&1; then
	echo "  3. golangci-lint-auto-configure configure   (Generate code-quality config)"
else
	echo -e "  ${YELLOW}3. Install golangci-lint-auto-configure to generate .golangci.yml${NC}"
	echo "     go install github.com/larsartmann/golangci-lint-auto-configure/cmd/golangci-lint-auto-configure@latest"
	echo "     golangci-lint-auto-configure configure"
fi
echo ""
echo -e "${BOLD}Happy linting! 🎉${NC}"
