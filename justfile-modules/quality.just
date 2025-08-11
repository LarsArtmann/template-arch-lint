# 📝 CODE QUALITY MODULE
# Importable justfile module for code quality enforcement
# Usage: import "path/to/quality.just"

# Tool version
GOLANGCI_VERSION := "v2.3.1"

# Configuration paths (can be overridden by importing justfile)
QUALITY_CONFIG := env_var_or_default("QUALITY_CONFIG", ".golangci.yml")

# Install code quality tools
install-quality-tools:
    @echo "📦 Installing golangci-lint {{GOLANGCI_VERSION}}..."
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{{GOLANGCI_VERSION}}
    @echo "✅ Quality tools installed!"

# Run code quality linting
lint-quality:
    @echo "📝 CODE QUALITY LINTING"
    @echo "Using config: {{QUALITY_CONFIG}}"
    @if [ -f "{{QUALITY_CONFIG}}" ]; then \
        golangci-lint run --config={{QUALITY_CONFIG}}; \
    else \
        echo "⚠️ Config file {{QUALITY_CONFIG}} not found, using defaults"; \
        golangci-lint run; \
    fi

# Auto-fix code quality issues
fix-quality:
    @echo "🔧 Auto-fixing quality issues..."
    gofmt -w -s .
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    else \
        go install golang.org/x/tools/cmd/goimports@latest; \
        goimports -w .; \
    fi
    @golangci-lint run --fix --config={{QUALITY_CONFIG}} 2>/dev/null || true
    @echo "✅ Auto-fix completed!"

# Generate quality reports
report-quality:
    @echo "📊 Generating quality reports..."
    @mkdir -p {{REPORTS_DIR}}
    @golangci-lint run --out-format json > {{REPORTS_DIR}}/quality.json 2>/dev/null || true
    @golangci-lint run --out-format checkstyle > {{REPORTS_DIR}}/checkstyle.xml 2>/dev/null || true
    @golangci-lint run --out-format junit-xml > {{REPORTS_DIR}}/junit.xml 2>/dev/null || true
    @echo "✅ Quality reports generated in {{REPORTS_DIR}}/"

# Verify quality setup  
verify-quality-setup:
    @echo "🔍 Verifying quality setup..."
    @which golangci-lint > /dev/null && echo "✅ golangci-lint installed" || echo "❌ golangci-lint missing"
    @[ -f "{{QUALITY_CONFIG}}" ] && echo "✅ Config file found: {{QUALITY_CONFIG}}" || echo "⚠️ Config file missing: {{QUALITY_CONFIG}}"