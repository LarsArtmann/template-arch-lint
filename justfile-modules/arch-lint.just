# 🏗️ ARCHITECTURE LINTING MODULE
# Importable justfile module for architecture enforcement
# Usage: import "path/to/arch-lint.just"

# Tool version
GO_ARCH_LINT_VERSION := "v1.12.0"

# Configuration paths (can be overridden by importing justfile)
ARCH_CONFIG := env_var_or_default("ARCH_CONFIG", ".go-arch-lint.yml")

# Install architecture linting tools
install-arch-tools:
    @echo "📦 Installing go-arch-lint {{GO_ARCH_LINT_VERSION}}..."
    go install github.com/fe3dback/go-arch-lint@{{GO_ARCH_LINT_VERSION}}
    @echo "✅ Architecture tools installed!"

# Run architecture validation
lint-architecture:
    @echo "🏗️ ARCHITECTURE VALIDATION"
    @echo "Using config: {{ARCH_CONFIG}}"
    @if [ -f "{{ARCH_CONFIG}}" ]; then \
        go-arch-lint check --config={{ARCH_CONFIG}}; \
    else \
        echo "⚠️ Config file {{ARCH_CONFIG}} not found"; \
        go-arch-lint check; \
    fi

# Generate architecture reports
report-architecture:
    @echo "📊 Generating architecture reports..."
    @mkdir -p {{env_var_or_default("REPORTS_DIR", "reports")}}
    @go-arch-lint check --json > {{env_var_or_default("REPORTS_DIR", "reports")}}/architecture.json 2>/dev/null || true
    @go-arch-lint graph > {{env_var_or_default("REPORTS_DIR", "reports")}}/dependencies.dot 2>/dev/null || true
    @echo "✅ Reports generated in {{env_var_or_default('REPORTS_DIR', 'reports')}}/"

# Verify architecture setup
verify-arch-setup:
    @echo "🔍 Verifying architecture setup..."
    @which go-arch-lint > /dev/null && echo "✅ go-arch-lint installed" || echo "❌ go-arch-lint missing"
    @[ -f "{{ARCH_CONFIG}}" ] && echo "✅ Config file found: {{ARCH_CONFIG}}" || echo "⚠️ Config file missing: {{ARCH_CONFIG}}"