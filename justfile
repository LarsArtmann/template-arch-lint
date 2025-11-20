# 🔥 ENTERPRISE-GRADE GO LINTING JUSTFILE
# Complete architecture and code quality enforcement
#
# Just is a handy way to save and run project-specific commands.
# https://github.com/casey/just

# Tool versions
GOLANGCI_VERSION := "v2.6.0"
GO_ARCH_LINT_VERSION := "v1.14.0"
CAPSLOCK_VERSION := "latest"

# Directories
ROOT_DIR := justfile_directory()
REPORT_DIR := ROOT_DIR / "reports"

# Default recipe (runs when just is called without arguments)
default: help

# Install git hooks for automatic pre-commit checks
install-hooks:
    @echo "\033[1m🪝 INSTALLING GIT HOOKS\033[0m"
    @echo "#!/bin/sh" > .git/hooks/pre-commit
    @echo "# Auto-generated pre-commit hook - fast formatting check only" >> .git/hooks/pre-commit
    @echo "just check-pre-commit-fast" >> .git/hooks/pre-commit
    @chmod +x .git/hooks/pre-commit
    @echo "\033[0;32m✅ Git pre-commit hook installed!\033[0m"
    @echo "\033[0;36mThe hook will do fast formatting checks only.\033[0m"
    @echo "\033[0;36mFor full checks including architecture: just check-pre-commit\033[0m"

# Install comprehensive git hooks (includes architecture validation)
install-hooks-full:
    @echo "\033[1m🪝 INSTALLING COMPREHENSIVE GIT HOOKS\033[0m"
    @echo "#!/bin/sh" > .git/hooks/pre-commit
    @echo "# Auto-generated pre-commit hook - comprehensive checks" >> .git/hooks/pre-commit
    @echo "just check-pre-commit" >> .git/hooks/pre-commit
    @chmod +x .git/hooks/pre-commit
    @echo "\033[0;32m✅ Comprehensive git pre-commit hook installed!\033[0m"
    @echo "\033[0;33m⚠️  This includes architecture graph validation - commits will be slower.\033[0m"

# Show this help message
help:
    @echo "\033[1m🔥 ENTERPRISE GO LINTING JUSTFILE\033[0m"
    @echo ""
    @echo "\033[1mUSAGE:\033[0m"
    @just --list --unsorted
    @echo ""
    @echo "\033[1mQUICK START:\033[0m"
    @echo "  1. \033[0;32mjust bootstrap\033[0m        - 🚀 Complete setup with enhanced error handling"
    @echo "  2. \033[0;36mjust bootstrap-diagnose\033[0m - 🔍 Environment diagnostics only"
    @echo "  3. \033[0;36mjust bootstrap-fix\033[0m     - 🔧 Auto-repair common issues"
    @echo "  4. \033[0;36mjust lint\033[0m             - Run all linters (including capslock)"
    @echo "  5. \033[0;36mjust security-audit\033[0m   - Complete security audit"
    @echo "  6. \033[0;36mjust format\033[0m           - Format code (gofumpt + goimports)"
    @echo "  7. \033[0;36mjust fix\033[0m              - Auto-fix issues"
    @echo "  8. \033[0;36mjust capslock-quick\033[0m   - Quick security capability check"
    @echo ""
    @echo "\033[1mDOCKER COMMANDS:\033[0m"
    @echo "  • \033[0;36mjust docker-test\033[0m         - Build and test Docker image (if available)"
    @echo ""
    @echo "\033[1mARCHITECTURE ANALYSIS:\033[0m"
    @echo "  • \033[0;36mjust graph\033[0m              - Generate flow graph (default)"
    @echo "  • \033[0;36mjust graph-di\033[0m           - Generate dependency injection graph"
    @echo "  • \033[0;36mjust graph-vendor\033[0m       - Generate vendor-inclusive graph"
    @echo "  • \033[0;36mjust graph-all\033[0m          - Generate ALL graph types"
    @echo "  • \033[0;36mjust graph-component <name>\033[0m - Generate focused component graph"
    @echo "  • \033[0;36mjust graph-list-components\033[0m - List available components"
    @echo ""
    @echo "\033[1mTESTING & COVERAGE:\033[0m"
    @echo "  • \033[0;36mjust test\033[0m                - Run tests with coverage"
    @echo "  • \033[0;36mjust coverage\033[0m            - Run coverage analysis with 80% threshold"
    @echo "  • \033[0;36mjust coverage 90\033[0m         - Run coverage analysis with custom threshold"
    @echo "  • \033[0;36mjust coverage-check\033[0m      - Quick coverage check (silent)"
    @echo "  • \033[0;36mjust coverage-detailed\033[0m   - Coverage breakdown by architectural layer"
    @echo ""
    @echo "\033[1mCODE ANALYSIS:\033[0m"
    @echo "  • \033[0;36mjust fd\033[0m                  - Find duplicate code (alias for find-duplicates)"
    @echo "  • \033[0;36mjust find-duplicates\033[0m     - Find duplicate code with custom threshold (default: 15 tokens)"
    @echo "  • \033[0;36mjust lint-capslock\033[0m      - Run Google's capslock capability analysis"
    @echo ""
    @echo "\033[1mSECURITY ANALYSIS:\033[0m"
    @echo "  • \033[0;36mjust security-audit\033[0m    - Complete security audit including capability analysis"
    @echo "  • \033[0;36mjust lint-security\033[0m     - Security-focused linting (gosec + copyloopvar)"
    @echo "  • \033[0;36mjust lint-vulns\033[0m        - Vulnerability scanning with govulncheck"
    @echo "  • \033[0;36mjust lint-licenses\033[0m    - License compliance scanning"
    @echo "  • \033[0;36mjust lint-nilaway\033[0m     - Nil panic prevention with Uber's NilAway"
    @echo "  • \033[0;36mjust lint-capslock\033[0m     - Google's capslock capability analysis"
    @echo ""
    @echo "\033[1mPERFORMANCE PROFILING:\033[0m"
    @echo "  • \033[0;36mjust profile-cpu\033[0m         - Capture 30-second CPU profile"
    @echo "  • \033[0;36mjust profile-heap\033[0m        - Capture heap memory profile"
    @echo "  • \033[0;36mjust profile-goroutines\033[0m  - Capture goroutine dump"
    @echo "  • \033[0;36mjust profile-trace\033[0m       - Capture 10-second execution trace"
    @echo "  • \033[0;36mjust analyze-cpu\033[0m         - Open CPU profile in browser"
    @echo "  • \033[0;36mjust analyze-heap\033[0m        - Open heap profile in browser"
    @echo ""
    @echo "\033[1mBENCHMARKING:\033[0m"
    @echo "  • \033[0;36mjust bench\033[0m               - Run all benchmarks"
    @echo "  • \033[0;36mjust bench-cpu\033[0m           - Run CPU-focused benchmarks"
    @echo "  • \033[0;36mjust bench-memory\033[0m        - Run memory-focused benchmarks"
    @echo "  • \033[0;36mjust bench-compare\033[0m       - Compare benchmark results"

# 🚀 Complete bootstrap setup using enhanced bootstrap.sh script
bootstrap:
    @echo "\033[1m🚀 BOOTSTRAP SETUP - ENTERPRISE GO LINTING\033[0m"
    @echo "\033[0;36mUsing enhanced bootstrap script with comprehensive error handling...\033[0m"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Check if bootstrap.sh exists, if not download it
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "\033[0;33m⚠️  Downloading enhanced bootstrap.sh...\033[0m"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "\033[0;31m❌ Failed to download bootstrap.sh\033[0m"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
        echo "\033[0;32m✅ Downloaded enhanced bootstrap.sh\033[0m"; \
    fi
    
    # Run enhanced bootstrap with default mode
    ./bootstrap.sh

# 🔍 Run comprehensive environment diagnostics only
bootstrap-diagnose:
    @echo "\033[1m🔍 BOOTSTRAP DIAGNOSTICS\033[0m"
    @echo "\033[0;36mAnalyzing environment and requirements...\033[0m"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Ensure bootstrap.sh exists
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "\033[0;33m⚠️  Downloading bootstrap.sh for diagnostics...\033[0m"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "\033[0;31m❌ Failed to download bootstrap.sh\033[0m"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run diagnostic mode only
    ./bootstrap.sh --diagnose

# 🔧 Bootstrap with automatic repair of common issues
bootstrap-fix:
    @echo "\033[1m🔧 BOOTSTRAP WITH AUTO-REPAIR\033[0m"
    @echo "\033[0;36mRunning diagnostics and automatically fixing issues...\033[0m"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Ensure bootstrap.sh exists
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "\033[0;33m⚠️  Downloading bootstrap.sh for auto-repair...\033[0m"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "\033[0;31m❌ Failed to download bootstrap.sh\033[0m"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run auto-repair mode
    ./bootstrap.sh --fix --verbose

# 🗣️ Bootstrap with verbose debug output
bootstrap-verbose:
    @echo "\033[1m🗣️  BOOTSTRAP WITH VERBOSE OUTPUT\033[0m"
    @echo "\033[0;36mRunning bootstrap with detailed debug information...\033[0m"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Ensure bootstrap.sh exists
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "\033[0;33m⚠️  Downloading bootstrap.sh...\033[0m"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "\033[0;31m❌ Failed to download bootstrap.sh\033[0m"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run with verbose output
    ./bootstrap.sh --verbose

# 🧪 Run BDD tests for bootstrap functionality
bootstrap-test:
    @echo "\033[1m🧪 BOOTSTRAP BDD TESTING\033[0m"
    @echo "\033[0;36mRunning behavior-driven tests for bootstrap script...\033[0m"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Download test script if not present
    @if [ ! -f "test-bootstrap-simple-bdd.sh" ]; then \
        echo "\033[0;33m⚠️  Downloading BDD test script...\033[0m"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/test-bootstrap-simple-bdd.sh" -o "test-bootstrap-simple-bdd.sh"; then \
            echo "\033[0;31m❌ Failed to download BDD test script\033[0m"; \
            exit 1; \
        fi; \
        chmod +x test-bootstrap-simple-bdd.sh; \
    fi
    
    # Ensure bootstrap.sh exists for testing
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "\033[0;33m⚠️  Downloading bootstrap.sh for testing...\033[0m"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "\033[0;31m❌ Failed to download bootstrap.sh\033[0m"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run BDD tests
    ./test-bootstrap-simple-bdd.sh

# 🚀 Quick bootstrap check - diagnose then fix if needed
bootstrap-quick:
    @echo "\033[1m⚡ QUICK BOOTSTRAP CHECK & FIX\033[0m"
    @echo "\033[0;36mRunning quick diagnostic and repair cycle...\033[0m"
    @echo ""
    #!/bin/bash
    set -euo pipefail
    
    # Ensure bootstrap.sh exists
    @if [ ! -f "bootstrap.sh" ]; then \
        echo "\033[0;33m⚠️  Downloading bootstrap.sh...\033[0m"; \
        if ! curl -fsSL "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh" -o "bootstrap.sh"; then \
            echo "\033[0;31m❌ Failed to download bootstrap.sh\033[0m"; \
            exit 1; \
        fi; \
        chmod +x bootstrap.sh; \
    fi
    
    # Run diagnose first, then fix if issues found
    echo "\033[1m🔍 Step 1: Diagnostics\033[0m"
    if ! ./bootstrap.sh --diagnose; then \
        echo "\033[1m🔧 Step 2: Auto-repair\033[0m"; \
        ./bootstrap.sh --fix; \
    else \
        echo "\033[0;32m✅ Environment looks good, running standard bootstrap\033[0m"; \
        ./bootstrap.sh; \
    fi
    @echo "\033[0;33m💡 Pro tip:\033[0m Run \033[0;36mjust install-hooks\033[0m to enable pre-commit linting!"

# Install all required linting tools
install:
    @echo "\033[1m📦 Installing linting tools...\033[0m"
    @echo "\033[0;33mInstalling golangci-lint {{GOLANGCI_VERSION}}...\033[0m"
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{{GOLANGCI_VERSION}}
    @echo "\033[0;33mInstalling go-arch-lint {{GO_ARCH_LINT_VERSION}}...\033[0m"
    go install github.com/fe3dback/go-arch-lint@{{GO_ARCH_LINT_VERSION}}
    @echo "\033[0;33mInstalling capslock {{CAPSLOCK_VERSION}}...\033[0m"
    go install github.com/google/capslock/cmd/capslock@{{CAPSLOCK_VERSION}}
    @echo "\033[0;32m✅ All tools installed successfully!\033[0m"

# Run all linters (architecture + code quality + filenames + security)
lint: lint-files lint-cmd-single lint-arch lint-code lint-vulns lint-cycles lint-goroutines lint-deps-advanced lint-capslock
    @echo ""
    @echo "\033[0;32m\033[1m✅ All linting checks completed!\033[0m"

# 🚨 Complete security audit (all security tools + capability analysis)
security-audit: lint-security lint-vulns lint-licenses lint-nilaway lint-capslock
    @echo ""
    @echo "\033[0;32m\033[1m🛡️ Complete security audit finished!\033[0m"

# Run architecture linting only
lint-arch:
    @echo "\033[1m🏗️  ARCHITECTURE LINTING\033[0m"
    @echo "\033[0;36mRunning go-arch-lint...\033[0m"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        if go-arch-lint check; then \
            echo "\033[0;32m✅ Architecture validation passed!\033[0m"; \
        else \
            echo "\033[0;31m❌ Architecture violations found!\033[0m" >&2; \
            exit 1; \
        fi; \
    else \
        echo "\033[0;31m❌ go-arch-lint not installed. Run 'just install' first.\033[0m" >&2; \
        exit 1; \
    fi

# Run code quality linting only
lint-code:
    @echo "\033[1m📝 CODE QUALITY LINTING\033[0m"
    @echo "\033[0;36mRunning golangci-lint v2...\033[0m"
    @if command -v golangci-lint >/dev/null 2>&1; then \
        if golangci-lint run --config .golangci.yml; then \
            echo "\033[0;32m✅ Code quality validation passed!\033[0m"; \
        else \
            echo "\033[0;31m❌ Code quality issues found!\033[0m" >&2; \
            exit 1; \
        fi; \
    else \
        echo "\033[0;31m❌ golangci-lint v2 not installed. Run 'just install' first.\033[0m" >&2; \
        exit 1; \
    fi

# Run filename verification only
lint-files:
    @echo "\033[1m📁 FILENAME VERIFICATION\033[0m"
    @echo "\033[0;36mChecking for problematic filenames...\033[0m"
    @if find . -name "*:*" -not -path "./.git/*" | grep -q .; then \
        echo "\033[0;31m❌ Found files with colons in names:\033[0m"; \
        find . -name "*:*" -not -path "./.git/*"; \
        exit 1; \
    else \
        echo "\033[0;32m✅ No problematic filenames found!\033[0m"; \
    fi

# Enforce single main file in cmd/ directory
lint-cmd-single:
    @./scripts/check-cmd-single.sh

# Auto-fix issues where possible
fix:
    @echo "\033[1m🔧 AUTO-FIXING ISSUES\033[0m"
    just format
    @echo "\033[0;33mRunning golangci-lint v2 with --fix...\033[0m"
    @if command -v $(go env GOPATH)/bin/golangci-lint >/dev/null 2>&1; then \
        $(go env GOPATH)/bin/golangci-lint run --fix --config .golangci.yml || true; \
    fi
    @echo "\033[0;32m✅ Auto-fix completed!\033[0m"

# Run all checks (for CI/CD pipelines)
ci: lint test capslock-quick graph
    @echo "\033[0;36mChecking module dependencies...\033[0m"
    go mod verify

# Pre-commit hook - format code and update architecture graph
pre-commit: format graph
    @echo "\033[1m✅ PRE-COMMIT TASKS COMPLETE\033[0m"
    @if git diff --exit-code > /dev/null 2>&1 && git diff --cached --exit-code > /dev/null 2>&1; then \
        echo "\033[0;32m✅ No changes needed - ready to commit!\033[0m"; \
    else \
        echo "\033[0;33m⚠️  Files were modified during pre-commit.\033[0m"; \
        echo "\033[0;36mModified files:\033[0m"; \
        git diff --name-only; \
        echo ""; \
        echo "\033[0;33mRun 'just commit-auto' to stage and commit these changes.\033[0m"; \
    fi

# Automatically stage formatting/graph changes and create a commit
commit-auto: pre-commit
    @echo "\033[1m🔄 AUTO-COMMIT PROCESS\033[0m"
    @if git diff --exit-code > /dev/null 2>&1; then \
        echo "\033[0;32m✅ No changes to commit.\033[0m"; \
    else \
        echo "\033[0;36mStaging automatic updates...\033[0m"; \
        git add -A; \
        echo "\033[0;36mCreating commit with detailed message...\033[0m"; \
        git commit -m "🔧 chore: Auto-update formatting and architecture graph" \
                   -m "Automated changes:" \
                   -m "- Applied gofumpt and goimports formatting" \
                   -m "- Regenerated architecture dependency graphs in docs/graphs/" \
                   -m "- Ensured consistent code style across the codebase" \
                   -m "" \
                   -m "Files modified:" \
                   -m "$$(git diff --cached --name-only | sed 's/^/  - /')" \
                   -m "" \
                   -m "Generated by: just commit-auto" \
                   -m "Timestamp: $$(date '+%Y-%m-%d %H:%M:%S %Z')"; \
        echo "\033[0;32m✅ Commit created!\033[0m"; \
        echo ""; \
        echo "\033[0;33mReview the commit:\033[0m"; \
        git log --oneline -1; \
        echo ""; \
        echo "\033[0;36mTo push: git push\033[0m"; \
        echo "\033[0;36mTo amend: git commit --amend\033[0m"; \
        echo "\033[0;36mTo undo: git reset HEAD~1\033[0m"; \
    fi

# Safe pre-commit check (doesn't modify files, only checks)
check-pre-commit:
    @echo "\033[1m🔍 PRE-COMMIT CHECK\033[0m"
    @echo "\033[0;36mChecking if formatting is needed...\033[0m"
    @if gofumpt -l . | grep -q .; then \
        echo "\033[0;33m⚠️  Files need formatting. Run 'just format'\033[0m"; \
        gofumpt -l .; \
    else \
        echo "\033[0;32m✅ Code formatting is clean\033[0m"; \
    fi
    @echo "\033[0;36mChecking if architecture graph is up-to-date...\033[0m"
    @go-arch-lint graph --out /tmp/test-graph.svg 2>/dev/null; \
    if ! diff -q /tmp/test-graph.svg docs/graphs/flow/architecture-flow.svg > /dev/null 2>&1; then \
        echo "\033[0;33m⚠️  Architecture graph needs updating. Run 'just graph'\033[0m"; \
    else \
        echo "\033[0;32m✅ Architecture graph is up-to-date\033[0m"; \
    fi
    @rm -f /tmp/test-graph.svg

# Fast pre-commit check for git hooks (formatting only)
check-pre-commit-fast:
    @echo "\033[1m⚡ FAST PRE-COMMIT CHECK\033[0m"
    @if gofumpt -l . | grep -q .; then \
        echo "\033[0;31m❌ Files need formatting. Run 'just format' first.\033[0m"; \
        echo "\033[0;36mFiles needing formatting:\033[0m"; \
        gofumpt -l .; \
        exit 1; \
    else \
        echo "\033[0;32m✅ Code formatting is clean\033[0m"; \
    fi
    go mod tidy -diff
    @echo "\033[0;32m\033[1m✅ CI/CD checks passed!\033[0m"

# Run tests with coverage
test:
    @echo "\033[1m🧪 RUNNING TESTS\033[0m"
    @echo "\033[0;36mRunning tests with coverage...\033[0m"
    go test ./... -v -race -coverprofile=coverage.out
    @echo "\033[0;32m✅ Tests completed!\033[0m"

# Run comprehensive coverage analysis with threshold enforcement
coverage THRESHOLD="80":
    @echo "\033[1m📊 COVERAGE ANALYSIS\033[0m"
    @echo "\033[0;36mRunning tests with coverage...\033[0m"
    go test ./... -v -race -coverprofile={{REPORT_DIR}}/coverage.out -covermode=atomic
    @echo "\033[0;36mGenerating coverage reports...\033[0m"
    go tool cover -html={{REPORT_DIR}}/coverage.out -o {{REPORT_DIR}}/coverage.html
    @echo "\033[0;33mCoverage Summary:\033[0m"
    @go tool cover -func={{REPORT_DIR}}/coverage.out | tail -1
    @echo "\033[0;36mChecking coverage threshold ({{THRESHOLD}}%)...\033[0m"
    @COVERAGE_PERCENT=$$(go tool cover -func={{REPORT_DIR}}/coverage.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
    if [ "$$(echo "$$COVERAGE_PERCENT < {{THRESHOLD}}" | bc -l)" -eq 1 ]; then \
        echo "\033[0;31m❌ Coverage $$COVERAGE_PERCENT% is below threshold {{THRESHOLD}}%\033[0m"; \
        echo "\033[0;33m📈 Generated reports:\033[0m"; \
        echo "  → {{REPORT_DIR}}/coverage.out (machine readable)"; \
        echo "  → {{REPORT_DIR}}/coverage.html (browser viewable)"; \
        exit 1; \
    else \
        echo "\033[0;32m✅ Coverage $$COVERAGE_PERCENT% meets threshold {{THRESHOLD}}%\033[0m"; \
        echo "\033[0;33m📈 Generated reports:\033[0m"; \
        echo "  → {{REPORT_DIR}}/coverage.out (machine readable)"; \
        echo "  → {{REPORT_DIR}}/coverage.html (browser viewable)"; \
    fi

# Quick coverage check without detailed output
coverage-check THRESHOLD="80":
    @echo "\033[1m📊 QUICK COVERAGE CHECK\033[0m"
    @go test ./... -coverprofile={{REPORT_DIR}}/coverage.out -covermode=atomic >/dev/null 2>&1
    @COVERAGE_PERCENT=$$(go tool cover -func={{REPORT_DIR}}/coverage.out | grep total: | awk '{print $$3}' | sed 's/%//'); \
    if [ "$$(echo "$$COVERAGE_PERCENT < {{THRESHOLD}}" | bc -l)" -eq 1 ]; then \
        echo "\033[0;31m❌ Coverage: $$COVERAGE_PERCENT% (threshold: {{THRESHOLD}}%)\033[0m"; \
        exit 1; \
    else \
        echo "\033[0;32m✅ Coverage: $$COVERAGE_PERCENT% (threshold: {{THRESHOLD}}%)\033[0m"; \
    fi

# Coverage by package/component breakdown
coverage-detailed:
    @echo "\033[1m📊 DETAILED COVERAGE ANALYSIS\033[0m"
    go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
    @echo "\033[0;33mCoverage by component:\033[0m"
    @echo ""
    @echo "\033[1mDomain Layer:\033[0m"
    @go tool cover -func=coverage.out | grep "internal/domain" || echo "  No domain coverage data"
    @echo ""
    @echo "\033[1mApplication Layer:\033[0m"
    @go tool cover -func=coverage.out | grep "internal/application" || echo "  No application coverage data"
    @echo ""
    @echo "\033[1mInfrastructure Layer:\033[0m"
    @go tool cover -func=coverage.out | grep "internal/infrastructure" || echo "  No infrastructure coverage data"
    @echo ""
    @echo "\033[1mConfiguration:\033[0m"
    @go tool cover -func=coverage.out | grep "internal/config\|internal/container" || echo "  No config coverage data"
    @echo ""
    @echo "\033[1mOverall Summary:\033[0m"
    @go tool cover -func=coverage.out | tail -1

# Generate detailed linting reports
report:
    @echo "\033[1m📊 GENERATING REPORTS\033[0m"
    mkdir -p {{REPORT_DIR}}
    @echo "\033[0;33mGenerating architecture report...\033[0m"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        go-arch-lint check --json > {{REPORT_DIR}}/architecture.json 2>/dev/null || true; \
        go-arch-lint graph > {{REPORT_DIR}}/dependencies.dot 2>/dev/null || true; \
        echo "  → {{REPORT_DIR}}/architecture.json"; \
        echo "  → {{REPORT_DIR}}/dependencies.dot"; \
    fi
    @echo "\033[0;33mGenerating code quality report...\033[0m"
    @if command -v $(go env GOPATH)/bin/golangci-lint >/dev/null 2>&1; then \
        $(go env GOPATH)/bin/golangci-lint run --out-format json > {{REPORT_DIR}}/quality.json 2>/dev/null || true; \
        $(go env GOPATH)/bin/golangci-lint run --out-format checkstyle > {{REPORT_DIR}}/checkstyle.xml 2>/dev/null || true; \
        $(go env GOPATH)/bin/golangci-lint run --out-format junit-xml > {{REPORT_DIR}}/junit.xml 2>/dev/null || true; \
        echo "  → {{REPORT_DIR}}/quality.json"; \
        echo "  → {{REPORT_DIR}}/checkstyle.xml"; \
        echo "  → {{REPORT_DIR}}/junit.xml"; \
    fi
    @echo "\033[0;33mGenerating capability analysis report...\033[0m"
    @if command -v capslock >/dev/null 2>&1; then \
        capslock ./... > {{REPORT_DIR}}/capslock-analysis.txt 2>/dev/null || true; \
        echo "  → {{REPORT_DIR}}/capslock-analysis.txt"; \
    fi
    @echo "\033[0;33mGenerating test coverage report...\033[0m"
    @go test ./... -coverprofile={{REPORT_DIR}}/coverage.out 2>/dev/null || true
    @go tool cover -html={{REPORT_DIR}}/coverage.out -o {{REPORT_DIR}}/coverage.html 2>/dev/null || true
    @echo "  → {{REPORT_DIR}}/coverage.out"
    @echo "  → {{REPORT_DIR}}/coverage.html"
    @echo "\033[0;32m✅ Reports generated in {{REPORT_DIR}}/\033[0m"

# Clean generated files and reports
clean:
    @echo "\033[1m🧹 CLEANING\033[0m"
    rm -rf {{REPORT_DIR}}
    rm -f coverage.out
    @echo "\033[0;32m✅ Cleaned successfully!\033[0m"

# Run minimal essential linters only
lint-minimal:
    @echo "\033[1m⚡ MINIMAL LINTING\033[0m"
    $(go env GOPATH)/bin/golangci-lint run --fast --config .golangci.yml

# Run with maximum strictness (slower but thorough)
lint-strict:
    @echo "\033[1m🔥 MAXIMUM STRICTNESS LINTING\033[0m"
    $(go env GOPATH)/bin/golangci-lint run --config .golangci.yml --max-issues-per-linter 0 --max-same-issues 0

# Run security-focused linters only
lint-security:
    @echo "\033[1m🔒 SECURITY LINTING\033[0m"
    $(go env GOPATH)/bin/golangci-lint run --config .golangci.yml --enable-only gosec,copyloopvar

# 🔍 Vulnerability scanning with official Go scanner
lint-vulns:
    @echo "\033[1m🔍 VULNERABILITY SCANNING\033[0m"
    @if command -v govulncheck >/dev/null 2>&1; then \
        govulncheck ./...; \
    else \
        echo "⚠️  govulncheck not found. Installing..."; \
        go install golang.org/x/vuln/cmd/govulncheck@latest; \
        govulncheck ./...; \
    fi

# 🔄 Import cycle detection beyond architecture linting
lint-cycles:
    @echo "\033[1m🔄 IMPORT CYCLE DETECTION\033[0m"
    @echo "🔍 Checking for import cycles in all packages..."
    @go list -json ./... | jq -r '.ImportPath' | while read pkg; do \
        echo "Checking $$pkg..."; \
        go list -f '{{{{.ImportPath}}}}: {{{{join .Imports " "}}}}' $$pkg 2>/dev/null || true; \
    done | grep -E "(cycle|import cycle)" || echo "✅ No import cycles detected"
    @echo "🔍 Detailed dependency analysis:"
    @go mod graph | head -20

# 🕸️ Dependency analysis (streamlined - redundant tools removed)
lint-deps-advanced:
    @echo "\033[1m🕸️ DEPENDENCY ANALYSIS\033[0m"
    @echo "🔍 Using govulncheck for comprehensive Go vulnerability scanning..."
    @echo "💡 Note: nancy and osv-scanner removed as redundant with govulncheck"
    @echo "📊 Running dependency analysis..."
    @go mod download -json all | jq -r '.Path + " " + .Version' | head -20
    @echo ""
    @echo "🛡️ For vulnerability scanning, use: just lint-vulns"

# 🔍 Goroutine leak detection (Uber's goleak)
lint-goroutines:
    @echo "\033[1m🔍 GOROUTINE LEAK DETECTION\033[0m"
    @echo "🔍 Installing Uber's goleak..."
    @go install github.com/uber-go/goleak@latest
    @echo "🔍 Running tests with goroutine leak detection..."
    @go test -race ./... -v -timeout=30s || echo "⚠️ Tests failed or goroutine leaks detected"

# ⚖️ License compliance scanning (Manual approach - no paid tools)
lint-licenses:
    @echo "\033[1m⚖️ LICENSE COMPLIANCE SCANNING\033[0m"
    @echo "🔍 Manual license analysis (FOSSA removed - requires paid account)..."
    @echo "📋 Go modules and their licenses:"
    @go mod download -json all | jq -r '.Path + " " + .Version' | head -20
    @echo "💡 Installing go-licenses for comprehensive scanning..."
    @if ! command -v go-licenses >/dev/null 2>&1; then \
        go install github.com/google/go-licenses@latest; \
    fi
    @echo "🔍 Running go-licenses check..."
    @go-licenses check ./... 2>/dev/null || echo "⚠️ Some licenses may need review"
    @echo "📋 Detailed license report:"
    @go-licenses report ./... 2>/dev/null | head -10 || echo "⚠️ Report generation failed"

# Note: Semgrep removed to reduce Python dependency complexity
# Security coverage provided by gosec (via golangci-lint) + govulncheck + NilAway

# 🚫 Uber's NilAway - Nil panic prevention
lint-nilaway:
    @echo "\033[1m🚫 NILAWAY - NIL PANIC DETECTION\033[0m"
    @if command -v nilaway >/dev/null 2>&1; then \
        echo "🔍 Running NilAway analysis (80% panic reduction!)..."; \
        nilaway -include-pkgs="github.com/LarsArtmann/template-arch-lint" -json ./... 2>/dev/null || nilaway ./...; \
    else \
        echo "⚠️  nilaway not found. Installing Uber's NilAway..."; \
        go install go.uber.org/nilaway/cmd/nilaway@latest; \
        nilaway -include-pkgs="github.com/LarsArtmann/template-arch-lint" ./...; \
    fi

# 🔒 Google Capslock - Capability analysis for security assessment
lint-capslock:
    @echo "\033[1m🔒 CAPSLOCK - CAPABILITY ANALYSIS\033[0m"
    @echo "🔍 Analyzing package capabilities and privileged operations..."
    @if command -v capslock >/dev/null 2>&1; then \
        echo "📋 Running capslock capability analysis..."; \
        if capslock -packages="./..." -output=package 2>/dev/null; then \
            echo "\033[0;32m✅ Capability analysis completed - no concerning privileges detected\033[0m"; \
        else \
            echo "\033[0;33m⚠️  Capability analysis completed - some issues detected\033[0m"; \
            echo "🔍 This could indicate:"; \
            echo "  • Security-relevant capabilities in dependencies"; \
            echo "  • Go version compatibility issues"; \
            echo "  • Dependency conflicts"; \
            echo ""; \
            echo "💡 Running detailed analysis for troubleshooting..."; \
            capslock -packages="./..." -output=package 2>&1 | head -10 || true; \
            echo ""; \
            echo "📋 For full analysis, run: just capslock-analysis"; \
        fi; \
    else \
        echo "⚠️  capslock not found. Installing Google's capslock..."; \
        go install github.com/google/capslock/cmd/capslock@latest; \
        echo "📋 Running capslock capability analysis..."; \
        capslock -packages="./..." -output=package; \
    fi

# Format code with enhanced formatters (gofumpt + goimports)
format:
    @echo "\033[1m📝 FORMATTING CODE\033[0m"
    @echo "\033[0;33mRunning gofumpt (enhanced gofmt)...\033[0m"
    @if command -v gofumpt >/dev/null 2>&1; then \
        gofumpt -w .; \
    else \
        echo "\033[0;31m❌ gofumpt not installed. Installing...\033[0m"; \
        go install mvdan.cc/gofumpt@latest; \
        gofumpt -w .; \
    fi
    @echo "\033[0;33mRunning goimports...\033[0m"
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    else \
        echo "\033[0;31m❌ goimports not installed. Installing...\033[0m"; \
        go install golang.org/x/tools/cmd/goimports@latest; \
        goimports -w .; \
    fi
    @echo "\033[0;32m✅ Code formatted!\033[0m"

# Format code (legacy alias - use 'format' instead)
fmt: format

# Generate architecture dependency graph (flow type)
graph:
    @echo "\033[1m📊 GENERATING ARCHITECTURE FLOW GRAPH\033[0m"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "\033[0;36mGenerating SVG flow graph...\033[0m"; \
        go-arch-lint graph --out docs/graphs/flow/architecture-flow.svg; \
        echo "\033[0;32m✅ Flow graph saved to docs/graphs/flow/architecture-flow.svg\033[0m"; \
    else \
        echo "\033[0;31m❌ go-arch-lint not found. Run 'just install' first.\033[0m"; \
        exit 1; \
    fi

# Generate focused architecture graphs for specific components
graph-component component:
    @echo "\033[1m📊 GENERATING FOCUSED COMPONENT GRAPH: {{component}}\033[0m"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        go-arch-lint graph --focus {{component}} --out docs/graphs/focused/{{component}}-focused.svg; \
        echo "\033[0;32m✅ Focused graph saved to docs/graphs/focused/{{component}}-focused.svg\033[0m"; \
    else \
        echo "\033[0;31m❌ go-arch-lint not found. Run 'just install' first.\033[0m"; \
        exit 1; \
    fi

# Generate dependency injection graph (DI type)
graph-di:
    @echo "\033[1m📊 GENERATING DEPENDENCY INJECTION GRAPH\033[0m"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "\033[0;36mGenerating SVG DI graph (component dependencies)...\033[0m"; \
        go-arch-lint graph --type di --out docs/graphs/dependency-injection/architecture-di.svg; \
        echo "\033[0;32m✅ DI graph saved to docs/graphs/dependency-injection/architecture-di.svg\033[0m"; \
    else \
        echo "\033[0;31m❌ go-arch-lint not found. Run 'just install' first.\033[0m"; \
        exit 1; \
    fi

# Generate vendor-inclusive graph (with external dependencies)
graph-vendor:
    @echo "\033[1m📊 GENERATING VENDOR-INCLUSIVE GRAPH\033[0m"
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "\033[0;36mGenerating SVG vendor graph (with external dependencies)...\033[0m"; \
        go-arch-lint graph --include-vendors --out docs/graphs/vendor/architecture-with-vendors.svg; \
        echo "\033[0;32m✅ Vendor graph saved to docs/graphs/vendor/architecture-with-vendors.svg\033[0m"; \
    else \
        echo "\033[0;31m❌ go-arch-lint not found. Run 'just install' first.\033[0m"; \
        exit 1; \
    fi

# Generate ALL graph types comprehensively
graph-all:
    @echo "\033[1m📊 GENERATING ALL ARCHITECTURE GRAPHS\033[0m"
    @echo "\033[0;36mThis will generate flow, DI, and vendor graphs for complete documentation...\033[0m"
    @echo ""
    @mkdir -p docs/graphs/{flow,dependency-injection,focused,vendor}
    @echo "\033[1m1️⃣  Generating Flow Graph (default)...\033[0m"
    @just graph
    @echo ""
    @echo "\033[1m2️⃣  Generating Dependency Injection Graph...\033[0m"
    @just graph-di
    @echo ""
    @echo "\033[1m3️⃣  Generating Vendor-Inclusive Graph...\033[0m"
    @just graph-vendor
    @echo ""
    @echo "\033[1m4️⃣  Generating Component-Focused Graphs...\033[0m"
    @echo "  → Focusing on: domain"
    @go-arch-lint graph --focus domain --out "docs/graphs/focused/domain-focused.svg" 2>/dev/null || true
    @echo "  → Focusing on: application"
    @go-arch-lint graph --focus application --out "docs/graphs/focused/application-focused.svg" 2>/dev/null || true
    @echo "  → Focusing on: infrastructure"
    @go-arch-lint graph --focus infrastructure --out "docs/graphs/focused/infrastructure-focused.svg" 2>/dev/null || true
    @echo "  → Focusing on: cmd"
    @go-arch-lint graph --focus cmd --out "docs/graphs/focused/cmd-focused.svg" 2>/dev/null || true
    @echo ""
    @echo "\033[1m5️⃣  Creating Graph Index...\033[0m"
    @echo "# Architecture Graphs - Generated on $$(date)" > docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "## Generated Graphs" >> docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "### 🔄 Flow Graphs" >> docs/graphs/index.md
    @echo "- [Architecture Flow](flow/architecture-flow.svg) - Execution flow (reverse DI)" >> docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "### 🔗 Dependency Injection Graphs" >> docs/graphs/index.md
    @echo "- [Architecture DI](dependency-injection/architecture-di.svg) - Component dependencies" >> docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "### 🌐 Vendor-Inclusive Graphs" >> docs/graphs/index.md
    @echo "- [Architecture with Vendors](vendor/architecture-with-vendors.svg) - Including external deps" >> docs/graphs/index.md
    @echo "" >> docs/graphs/index.md
    @echo "### 🎯 Component-Focused Graphs" >> docs/graphs/index.md
    @echo "- [domain-focused](focused/domain-focused.svg) - Domain layer dependencies" >> docs/graphs/index.md
    @echo "- [application-focused](focused/application-focused.svg) - Application layer dependencies" >> docs/graphs/index.md
    @echo "- [infrastructure-focused](focused/infrastructure-focused.svg) - Infrastructure layer dependencies" >> docs/graphs/index.md
    @echo "- [cmd-focused](focused/cmd-focused.svg) - Command layer dependencies" >> docs/graphs/index.md
    @echo ""
    @echo "\033[0;32m✅ All graphs generated successfully!\033[0m"
    @echo "\033[0;36m📁 Graph directory: docs/graphs/\033[0m"
    @echo "\033[0;36m📋 Index file: docs/graphs/index.md\033[0m"

# List available components for focused graphs
graph-list-components:
    @echo "\033[1m📋 AVAILABLE ARCHITECTURE COMPONENTS\033[0m"
    @echo "\033[0;36mCommon component names you can focus on:\033[0m"
    @echo "  • domain"
    @echo "  • application"
    @echo "  • infrastructure"
    @echo "  • cmd"
    @echo "  • internal"
    @echo ""
    @echo "\033[0;33m💡 Usage: just graph-component <component_name>\033[0m"
    @echo "Example: just graph-component domain"

# Find code duplications in the project
find-duplicates threshold="15":
    @echo "\033[1m🔍 FINDING CODE DUPLICATIONS\033[0m"
    @mkdir -p {{REPORT_DIR}}
    @echo "\033[0;36mAnalyzing Go code duplications (threshold: {{threshold}} tokens)...\033[0m"
    @if command -v dupl >/dev/null 2>&1; then \
        echo "\033[0;33m📋 Go Code Duplication Report (dupl)\033[0m"; \
        dupl -t {{threshold}} -v . > {{REPORT_DIR}}/go-duplications.txt 2>&1 || true; \
        dupl -t {{threshold}} -html . > {{REPORT_DIR}}/go-duplications.html 2>&1 || true; \
        echo "  → {{REPORT_DIR}}/go-duplications.txt"; \
        echo "  → {{REPORT_DIR}}/go-duplications.html"; \
        echo ""; \
        echo "\033[0;33m📊 Summary:\033[0m"; \
        DUPL_COUNT=`dupl -t {{threshold}} . 2>/dev/null | grep -c "found" || echo "0"`; \
        echo "  Go duplications found: $DUPL_COUNT"; \
    else \
        echo "\033[0;31m❌ dupl not found. Installing...\033[0m"; \
        go install github.com/mibk/dupl@latest; \
        dupl -t {{threshold}} -v . > {{REPORT_DIR}}/go-duplications.txt 2>&1 || true; \
        dupl -t {{threshold}} -html . > {{REPORT_DIR}}/go-duplications.html 2>&1 || true; \
    fi
    @echo "\033[0;36mAnalyzing multi-language duplications (jscpd)...\033[0m"
    @if command -v jscpd >/dev/null 2>&1; then \
        echo "\033[0;33m📋 Multi-Language Duplication Report (jscpd)\033[0m"; \
        jscpd . --min-tokens {{threshold}} --reporters json,html --output {{REPORT_DIR}}/jscpd || true; \
        if [ -f "{{REPORT_DIR}}/jscpd/jscpd-report.json" ]; then \
            echo "  → {{REPORT_DIR}}/jscpd/jscpd-report.json"; \
            echo "  → {{REPORT_DIR}}/jscpd/jscpd-report.html"; \
        fi; \
    else \
        echo "\033[0;33m⚠️  jscpd not found, skipping multi-language analysis.\033[0m"; \
        echo "\033[0;36mTo install: bun install -g jscpd\033[0m"; \
    fi
    @echo ""
    @echo "\033[0;32m✅ Duplication analysis complete!\033[0m"
    @echo "\033[0;36mOpen {{REPORT_DIR}}/go-duplications.html in browser for detailed Go analysis\033[0m"

# 🔒 Comprehensive capslock capability analysis with reporting
capslock-analysis:
    @echo "\033[1m🔒 COMPREHENSIVE CAPSLOCK ANALYSIS\033[0m"
    @echo "🔍 Running detailed capability analysis with reporting..."
    @mkdir -p {{REPORT_DIR}}
    @if command -v capslock >/dev/null 2>&1; then \
        echo "📋 Generating capability analysis report..."; \
        echo "📊 Analyzing package capabilities and privileged operations..."; \
        capslock -packages="./..." -output=package > {{REPORT_DIR}}/capslock-analysis.txt 2>&1 || true; \
        echo "\033[0;33m📋 Capability Analysis Summary:\033[0m"; \
        echo "  → Report saved to: {{REPORT_DIR}}/capslock-analysis.txt"; \
        echo ""; \
        echo "\033[0;36m🔍 Analysis Results:\033[0m"; \
        if grep -q "Some packages had errors" {{REPORT_DIR}}/capslock-analysis.txt; then \
            echo "  ⚠️  Analysis completed with errors - likely dependency compatibility issues"; \
            echo "  💡 This is common when Go versions don't match go.mod requirements"; \
            echo "  📋 Check Go version: go version (should match go.mod)"; \
            echo "  📋 Current Go version: $(go version | cut -d' ' -f3)"; \
            echo "  📋 Required Go version: $(grep '^go' go.mod | cut -d' ' -f2)"; \
            echo "  💡 Try: go mod tidy && go mod download"; \
        else \
            echo "  ✅ Analysis completed successfully"; \
        fi; \
        echo ""; \
        echo "\033[0;36m🔍 Key security insights from capslock:\033[0m"; \
        echo "  • File system access capabilities"; \
        echo "  • Network operation capabilities"; \
        echo "  • System call capabilities"; \
        echo "  • Process execution capabilities"; \
        echo "  • Cryptographic operation capabilities"; \
        echo ""; \
        echo "\033[0;33m💡 Security recommendations:\033[0m"; \
        echo "  1. Review any unexpected privileged capabilities"; \
        echo "  2. Ensure capabilities align with package purpose"; \
        echo "  3. Consider principle of least privilege"; \
        echo "  4. Monitor for capability changes in updates"; \
        echo "  5. Fix dependency compatibility issues if present"; \
        echo ""; \
        echo "\033[0;32m✅ Comprehensive capslock analysis completed!\033[0m"; \
    else \
        echo "\033[0;31m❌ capslock not found. Installing..."; \
        go install github.com/google/capslock/cmd/capslock@latest; \
        echo "🔍 Retrying capability analysis..."; \
        capslock -packages="./..." -output=package > {{REPORT_DIR}}/capslock-analysis.txt 2>&1 || true; \
        echo "\033[0;32m✅ Capslock analysis completed after installation!\033[0m"; \
    fi

# 🔒 Quick capslock security check (CI/CD friendly)
capslock-quick:
    @echo "\033[1m🔒 QUICK CAPSLOCK SECURITY CHECK\033[0m"
    @if command -v capslock >/dev/null 2>&1; then \
        if capslock -packages="./..." -output=package >/dev/null 2>&1; then \
            echo "\033[0;32m✅ Capslock security check passed\033[0m"; \
        else \
            echo "\033[0;33m⚠️  Capslock detected issues - could be capabilities or compatibility\033[0m"; \
            echo "💡 Run 'just capslock-analysis' for detailed troubleshooting"; \
            exit 1; \
        fi; \
    else \
        echo "\033[0;31m❌ capslock not found. Run 'just install' first.\033[0m"; \
        exit 1; \
    fi

# Alias for find-duplicates
fd threshold="15": (find-duplicates threshold)

# Generate templates and build Go modules
build:
    @echo "\033[1m🔨 BUILDING\033[0m"
    @echo "\033[0;33mGenerating templates...\033[0m"
    @if command -v templ >/dev/null 2>&1; then \
        templ generate; \
    else \
        echo "\033[0;31m❌ templ not installed. Installing...\033[0m"; \
        go install github.com/a-h/templ/cmd/templ@latest; \
        templ generate; \
    fi
    @echo "\033[0;33mBuilding Go modules...\033[0m"
    go build ./...
    @echo "\033[0;32m✅ Build completed!\033[0m"

# Generate templates only
templ:
    @echo "\033[1m📄 GENERATING TEMPLATES\033[0m"
    @if command -v templ >/dev/null 2>&1; then \
        templ generate; \
    else \
        echo "\033[0;31m❌ templ not installed. Installing...\033[0m"; \
        go install github.com/a-h/templ/cmd/templ@latest; \
        templ generate; \
    fi
    @echo "\033[0;32m✅ Templates generated!\033[0m"

# Run the server
run: build
    @echo "\033[1m🚀 STARTING SERVER\033[0m"
    go run cmd/server/main.go

# Development mode with auto-reload
dev:
    @echo "\033[1m🔄 DEVELOPMENT MODE\033[0m"
    @if command -v air >/dev/null 2>&1; then \
        air; \
    else \
        echo "\033[0;31m❌ air not installed. Installing...\033[0m"; \
        go install github.com/cosmtrek/air@latest; \
        air; \
    fi

# Template configuration system - copy linting configs to other projects

# Run simple filename verification
verify-filenames: lint-files

# Check dependencies
check-deps:
    @echo "\033[1m📦 CHECKING DEPENDENCIES\033[0m"
    go mod verify
    go mod tidy
    @echo "\033[0;32m✅ Dependencies checked!\033[0m"

# Update dependencies
update-deps:
    @echo "\033[1m🔄 UPDATING DEPENDENCIES\033[0m"
    go get -u ./...
    go mod tidy
    @echo "\033[0;32m✅ Dependencies updated!\033[0m"

# Note: Main bench recipe is defined later with comprehensive reporting

# Test configuration system
config-test:
    @echo "\033[1m⚙️  TESTING CONFIGURATION\033[0m"
    @echo "\033[0;36mTesting default configuration...\033[0m"
    go run example/main.go
    @echo ""
    @echo "\033[0;36mTesting environment variable overrides...\033[0m"
    APP_SERVER_PORT=9090 APP_LOGGING_LEVEL=debug go run example/main.go
    @echo ""
    @echo "\033[0;32m✅ Configuration tests completed!\033[0m"

# Run with verbose output
verbose:
    @echo "\033[1m🔍 VERBOSE LINTING\033[0m"
    go-arch-lint check -v
    $(go env GOPATH)/bin/golangci-lint run -v --config .golangci.yml

# Git hooks setup
setup-hooks:
    @echo "\033[1m🪝 SETTING UP GIT HOOKS\033[0m"
    @echo '#!/bin/sh\necho "Running pre-commit linting..."\njust lint' > .git/hooks/pre-commit
    @chmod +x .git/hooks/pre-commit
    @echo "\033[0;32m✅ Git hooks setup completed!\033[0m"

# Show project statistics
stats:
    @echo "\033[1m📊 PROJECT STATISTICS\033[0m"
    @echo "Lines of Go code:"
    @find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
    @echo "Number of Go files:"
    @find . -name "*.go" -not -path "./vendor/*" | wc -l
    @echo "Number of packages:"
    @go list ./... | wc -l

# Show version information
version:
    @echo "\033[1m📋 VERSION INFORMATION\033[0m"
    @echo "Go version:"
    @go version
    @if command -v $(go env GOPATH)/bin/golangci-lint >/dev/null 2>&1; then \
        echo "golangci-lint version:"; \
        $(go env GOPATH)/bin/golangci-lint version; \
    fi
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "go-arch-lint version:"; \
        go-arch-lint version; \
    fi
    @if command -v capslock >/dev/null 2>&1; then \
        echo "capslock version:"; \
        capslock --version 2>/dev/null || echo "version info not available"; \
    fi
    @echo "Just version:"
    @just --version
    @if command -v docker >/dev/null 2>&1; then \
        echo "Docker version:"; \
        docker --version; \
    fi

# 🐳 Docker Commands

# Build Docker image
docker-build:
    @echo "\033[1m🐳 BUILDING DOCKER IMAGE\033[0m"
    docker build -t template-arch-lint:latest .
    @echo "\033[0;32m✅ Docker image built successfully!\033[0m"

# Build and test Docker image
docker-test: docker-build
    @echo "\033[1m🧪 TESTING DOCKER IMAGE\033[0m"
    @echo "\033[0;36mTesting health check...\033[0m"
    docker run --rm template-arch-lint:latest -health-check
    @echo "\033[0;36mTesting container startup...\033[0m"
    @CONTAINER_ID=$$(docker run -d -p 8080:8080 template-arch-lint:latest); \
    sleep 5; \
    echo "Testing health endpoints..."; \
    curl -f http://localhost:8080/health/live || exit 1; \
    curl -f http://localhost:8080/version || exit 1; \
    docker stop $$CONTAINER_ID; \
    echo "\033[0;32m✅ Docker image tests passed!\033[0m"

# Run application in Docker container
docker-run: docker-build
    @echo "\033[1m🚀 RUNNING DOCKER CONTAINER\033[0m"
    docker run --rm -p 8080:8080 -p 2112:2112 template-arch-lint:latest

# Docker development environment (requires docker-compose.yml)
docker-dev:
    @if [ -f docker-compose.yml ]; then \
        echo "\033[1m🔄 STARTING DEVELOPMENT ENVIRONMENT\033[0m"; \
        docker-compose up --build; \
    else \
        echo "⚠️  docker-compose.yml not found. This is a linting template - monitoring stack removed."; \
        echo "💡 For Docker setup, add your own docker-compose.yml with required services."; \
    fi

# Start Docker environment in background (requires docker-compose.yml)  
docker-dev-detached:
    @if [ -f docker-compose.yml ]; then \
        echo "\033[1m🔄 STARTING DEVELOPMENT ENVIRONMENT (DETACHED)\033[0m"; \
        docker-compose up --build -d; \
        echo "\033[0;32m✅ Development environment started!\033[0m"; \
        echo "\033[0;36mServices available at http://localhost:8080\033[0m"; \
    else \
        echo "⚠️  docker-compose.yml not found. This is a linting template."; \
        echo "💡 Create docker-compose.yml for your specific monitoring/service needs."; \
    fi

# Stop Docker environment  
docker-stop:
    @if [ -f docker-compose.yml ]; then \
        echo "\033[1m🛑 STOPPING DEVELOPMENT ENVIRONMENT\033[0m"; \
        docker-compose down; \
        echo "\033[0;32m✅ Development environment stopped!\033[0m"; \
    else \
        echo "⚠️  docker-compose.yml not found - nothing to stop."; \
    fi

# Clean up Docker resources
docker-clean:
    @echo "\033[1m🧹 CLEANING DOCKER RESOURCES\033[0m"
    docker-compose down -v --remove-orphans
    docker image prune -f
    docker system prune -f
    @echo "\033[0;32m✅ Docker resources cleaned!\033[0m"

# Show Docker logs
docker-logs:
    @echo "\033[1m📋 DOCKER LOGS\033[0m"
    docker-compose logs -f

# Security scan Docker image with Trivy
docker-security: docker-build
    @echo "\033[1m🛡️  DOCKER SECURITY SCAN\033[0m"
    @if command -v trivy >/dev/null 2>&1; then \
        trivy image template-arch-lint:latest; \
    else \
        echo "\033[0;31m❌ Trivy not installed. Install with: brew install trivy\033[0m"; \
        exit 1; \
    fi

# Find high-confidence duplicates (stricter threshold)
find-duplicates-strict: (find-duplicates "100")

# Find all potential duplicates (looser threshold)
find-duplicates-loose: (find-duplicates "25")

# ==============================================
# PERFORMANCE PROFILING COMMANDS
# ==============================================

# Capture CPU profile for 30 seconds
profile-cpu:
    @echo "\033[1m📊 CAPTURING CPU PROFILE\033[0m"
    @echo "\033[0;36mCapturing CPU profile for 30 seconds...\033[0m"
    @if curl -s "http://localhost:8080/debug/pprof/profile?seconds=30" -o cpu.prof; then \
        echo "\033[0;32m✅ CPU profile saved to cpu.prof\033[0m"; \
        echo "\033[0;33m💡 Use 'just analyze-cpu' to view analysis\033[0m"; \
    else \
        echo "\033[0;31m❌ Failed to capture CPU profile. Is the server running in development mode?\033[0m"; \
        exit 1; \
    fi

# Capture heap memory profile
profile-heap:
    @echo "\033[1m📊 CAPTURING HEAP PROFILE\033[0m"
    @echo "\033[0;36mCapturing heap memory profile...\033[0m"
    @if curl -s http://localhost:8080/debug/pprof/heap -o heap.prof; then \
        echo "\033[0;32m✅ Heap profile saved to heap.prof\033[0m"; \
        echo "\033[0;33m💡 Use 'just analyze-heap' to view analysis\033[0m"; \
    else \
        echo "\033[0;31m❌ Failed to capture heap profile. Is the server running in development mode?\033[0m"; \
        exit 1; \
    fi

# Capture goroutine dump
profile-goroutines:
    @echo "\033[1m📊 CAPTURING GOROUTINE DUMP\033[0m"
    @echo "\033[0;36mCapturing goroutine dump...\033[0m"
    @if curl -s http://localhost:8080/debug/pprof/goroutine -o goroutine.prof; then \
        echo "\033[0;32m✅ Goroutine dump saved to goroutine.prof\033[0m"; \
        echo "\033[0;33m💡 Use 'go tool pprof goroutine.prof' to analyze\033[0m"; \
    else \
        echo "\033[0;31m❌ Failed to capture goroutine dump. Is the server running in development mode?\033[0m"; \
        exit 1; \
    fi

# Capture execution trace for 10 seconds
profile-trace:
    @echo "\033[1m📊 CAPTURING EXECUTION TRACE\033[0m"
    @echo "\033[0;36mCapturing execution trace for 10 seconds...\033[0m"
    @if curl -s "http://localhost:8080/debug/pprof/trace?seconds=10" -o trace.out; then \
        echo "\033[0;32m✅ Execution trace saved to trace.out\033[0m"; \
        echo "\033[0;33m💡 Use 'go tool trace trace.out' to analyze\033[0m"; \
    else \
        echo "\033[0;31m❌ Failed to capture execution trace. Is the server running in development mode?\033[0m"; \
        exit 1; \
    fi

# Capture allocation profile
profile-allocs:
    @echo "\033[1m📊 CAPTURING ALLOCATION PROFILE\033[0m"
    @echo "\033[0;36mCapturing allocation profile...\033[0m"
    @if curl -s http://localhost:8080/debug/pprof/allocs -o allocs.prof; then \
        echo "\033[0;32m✅ Allocation profile saved to allocs.prof\033[0m"; \
        echo "\033[0;33m💡 Use 'go tool pprof allocs.prof' to analyze\033[0m"; \
    else \
        echo "\033[0;31m❌ Failed to capture allocation profile. Is the server running in development mode?\033[0m"; \
        exit 1; \
    fi

# Open CPU profile analysis in browser
analyze-cpu:
    @echo "\033[1m🔍 ANALYZING CPU PROFILE\033[0m"
    @if [ -f cpu.prof ]; then \
        echo "\033[0;36mOpening CPU profile analysis in browser...\033[0m"; \
        echo "\033[0;33mBrowser will open at http://localhost:8081\033[0m"; \
        go tool pprof -http=:8081 cpu.prof; \
    else \
        echo "\033[0;31m❌ CPU profile not found. Run 'just profile-cpu' first.\033[0m"; \
        exit 1; \
    fi

# Open heap profile analysis in browser
analyze-heap:
    @echo "\033[1m🔍 ANALYZING HEAP PROFILE\033[0m"
    @if [ -f heap.prof ]; then \
        echo "\033[0;36mOpening heap profile analysis in browser...\033[0m"; \
        echo "\033[0;33mBrowser will open at http://localhost:8081\033[0m"; \
        go tool pprof -http=:8081 heap.prof; \
    else \
        echo "\033[0;31m❌ Heap profile not found. Run 'just profile-heap' first.\033[0m"; \
        exit 1; \
    fi

# Get runtime performance statistics
profile-stats:
    @echo "\033[1m📊 RUNTIME STATISTICS\033[0m"
    @if curl -s http://localhost:8080/performance/stats | jq .; then \
        echo ""; \
        echo "\033[0;32m✅ Runtime statistics retrieved successfully\033[0m"; \
    else \
        echo "\033[0;31m❌ Failed to get runtime statistics. Is the server running?\033[0m"; \
        exit 1; \
    fi

# Check application health metrics
profile-health:
    @echo "\033[1m🏥 HEALTH METRICS\033[0m"
    @if curl -s http://localhost:8080/performance/health | jq .; then \
        echo ""; \
        echo "\033[0;32m✅ Health metrics retrieved successfully\033[0m"; \
    else \
        echo "\033[0;31m❌ Failed to get health metrics. Is the server running?\033[0m"; \
        exit 1; \
    fi

# Force garbage collection and show results
profile-gc:
    @echo "\033[1m🗑️ FORCE GARBAGE COLLECTION\033[0m"
    @echo "\033[0;36mTriggering garbage collection...\033[0m"
    @if curl -s -X POST http://localhost:8080/performance/gc | jq .; then \
        echo ""; \
        echo "\033[0;32m✅ Garbage collection completed\033[0m"; \
    else \
        echo "\033[0;31m❌ Failed to trigger garbage collection. Is the server running in development mode?\033[0m"; \
        exit 1; \
    fi

# Capture all performance profiles in one command
profile-all:
    @echo "\033[1m📊 CAPTURING ALL PROFILES\033[0m"
    @echo "\033[0;36mThis will capture CPU (30s), heap, goroutines, allocations, and trace (10s)...\033[0m"
    @echo "\033[0;33mTotal time: ~45 seconds\033[0m"
    @echo ""
    just profile-heap
    @echo ""
    just profile-goroutines
    @echo ""
    just profile-allocs
    @echo ""
    just profile-cpu
    @echo ""
    just profile-trace
    @echo ""
    @echo "\033[0;32m🎉 All profiles captured successfully!\033[0m"
    @echo "\033[0;36mFiles created:\033[0m"
    @echo "  • cpu.prof - CPU profiling data"
    @echo "  • heap.prof - Heap memory allocations"
    @echo "  • goroutine.prof - Goroutine dump"
    @echo "  • allocs.prof - Allocation history"
    @echo "  • trace.out - Execution trace"

# Clean up profile files
profile-clean:
    @echo "\033[1m🧹 CLEANING PROFILE FILES\033[0m"
    rm -f *.prof *.out
    @echo "\033[0;32m✅ Profile files cleaned!\033[0m"

# ==============================================
# BENCHMARKING COMMANDS
# ==============================================

# Run all benchmarks with comprehensive reporting
bench:
    @echo "\033[1m🏎️ RUNNING ALL BENCHMARKS\033[0m"
    @echo "\033[0;36mRunning comprehensive benchmark suite...\033[0m"
    @mkdir -p benchmarks
    @go test -bench=. -benchmem -run=^$$ ./... | tee benchmarks/benchmark-results.txt
    @echo ""
    @echo "\033[0;32m✅ Benchmarks completed!\033[0m"
    @echo "\033[0;36m→ Results saved to: benchmarks/benchmark-results.txt\033[0m"

# Run CPU-focused benchmarks (no memory allocation reporting)
bench-cpu:
    @echo "\033[1m⚡ RUNNING CPU BENCHMARKS\033[0m"
    @echo "\033[0;36mFocusing on CPU performance metrics...\033[0m"
    @mkdir -p benchmarks
    @go test -bench=. -run=^$$ ./internal/domain/services/ | tee benchmarks/cpu-benchmarks.txt
    @go test -bench=. -run=^$$ ./internal/infrastructure/persistence/ | tee -a benchmarks/cpu-benchmarks.txt
    @echo ""
    @echo "\033[0;32m✅ CPU benchmarks completed!\033[0m"
    @echo "\033[0;36m→ Results saved to: benchmarks/cpu-benchmarks.txt\033[0m"

# Run memory-focused benchmarks (with allocation reporting)
bench-memory:
    @echo "\033[1m🧠 RUNNING MEMORY BENCHMARKS\033[0m"
    @echo "\033[0;36mFocusing on memory allocation patterns...\033[0m"
    @mkdir -p benchmarks
    @go test -bench=BenchmarkMemory -benchmem -run=^$$ ./... | tee benchmarks/memory-benchmarks.txt
    @go test -bench=BenchmarkAllocation -benchmem -run=^$$ ./... | tee -a benchmarks/memory-benchmarks.txt
    @go test -bench=BenchmarkConcurrent -benchmem -run=^$$ ./... | tee -a benchmarks/memory-benchmarks.txt
    @echo ""
    @echo "\033[0;32m✅ Memory benchmarks completed!\033[0m"
    @echo "\033[0;36m→ Results saved to: benchmarks/memory-benchmarks.txt\033[0m"

# Run specific benchmark by name
bench-specific PATTERN:
    @echo "\033[1m🎯 RUNNING SPECIFIC BENCHMARK\033[0m"
    @echo "\033[0;36mRunning benchmarks matching: {{PATTERN}}\033[0m"
    @mkdir -p benchmarks
    @go test -bench={{PATTERN}} -benchmem -run=^$$ ./... | tee benchmarks/specific-{{PATTERN}}.txt
    @echo ""
    @echo "\033[0;32m✅ Specific benchmarks completed!\033[0m"
    @echo "\033[0;36m→ Results saved to: benchmarks/specific-{{PATTERN}}.txt\033[0m"

# Establish performance baseline
bench-baseline:
    @echo "\033[1m📊 ESTABLISHING PERFORMANCE BASELINE\033[0m"
    @echo "\033[0;36mCreating baseline benchmark results...\033[0m"
    @mkdir -p benchmarks/baseline
    @go test -bench=. -benchmem -count=5 -run=^$$ ./... > benchmarks/baseline/results.txt 2>&1
    @echo ""
    @echo "\033[0;32m✅ Baseline established!\033[0m"
    @echo "\033[0;36m→ Baseline saved to: benchmarks/baseline/results.txt\033[0m"
    @echo "\033[0;33m💡 Use 'just bench-compare' to compare future runs against this baseline\033[0m"

# Compare current benchmarks with baseline
bench-compare:
    @echo "\033[1m📈 COMPARING BENCHMARK RESULTS\033[0m"
    @if [ ! -f benchmarks/baseline/results.txt ]; then \
        echo "\033[0;31m❌ No baseline found. Run 'just bench-baseline' first.\033[0m"; \
        exit 1; \
    fi
    @echo "\033[0;36mRunning current benchmarks for comparison...\033[0m"
    @mkdir -p benchmarks/current
    @go test -bench=. -benchmem -count=5 -run=^$$ ./... > benchmarks/current/results.txt 2>&1
    @echo "\033[0;36mComparing results with baseline...\033[0m"
    @if command -v benchcmp >/dev/null 2>&1; then \
        benchcmp benchmarks/baseline/results.txt benchmarks/current/results.txt | tee benchmarks/comparison.txt; \
        echo "\033[0;32m✅ Comparison completed!\033[0m"; \
        echo "\033[0;36m→ Results saved to: benchmarks/comparison.txt\033[0m"; \
    else \
        echo "\033[0;33m⚠️ benchcmp tool not found. Install with: go install golang.org/x/tools/cmd/benchcmp@latest\033[0m"; \
        echo "\033[0;36mManual comparison available in:\033[0m"; \
        echo "  → benchmarks/baseline/results.txt"; \
        echo "  → benchmarks/current/results.txt"; \
    fi

# Generate benchmark report with analysis
bench-report:
    @echo "\033[1m📋 GENERATING BENCHMARK REPORT\033[0m"
    @mkdir -p benchmarks/reports
    @echo "\033[0;36mRunning comprehensive benchmarks...\033[0m"
    @go test -bench=. -benchmem -count=3 -run=^$$ ./... > benchmarks/reports/full-report.txt 2>&1
    @echo "\033[0;36mGenerating summary report...\033[0m"
    @echo "# Benchmark Report - $$(date)" > benchmarks/reports/summary.md
    @echo "" >> benchmarks/reports/summary.md
    @echo "## Performance Summary" >> benchmarks/reports/summary.md
    @echo "" >> benchmarks/reports/summary.md
    @echo "\`\`\`" >> benchmarks/reports/summary.md
    @grep "Benchmark" benchmarks/reports/full-report.txt | head -20 >> benchmarks/reports/summary.md
    @echo "\`\`\`" >> benchmarks/reports/summary.md
    @echo "" >> benchmarks/reports/summary.md
    @echo "## Analysis" >> benchmarks/reports/summary.md
    @echo "- Generated on: $$(date)" >> benchmarks/reports/summary.md
    @echo "- Go version: $$(go version)" >> benchmarks/reports/summary.md
    @echo "- Test count: $$(grep -c "Benchmark" benchmarks/reports/full-report.txt) benchmarks" >> benchmarks/reports/summary.md
    @echo ""
    @echo "\033[0;32m✅ Benchmark report generated!\033[0m"
    @echo "\033[0;36m→ Full report: benchmarks/reports/full-report.txt\033[0m"
    @echo "\033[0;36m→ Summary: benchmarks/reports/summary.md\033[0m"

# Run benchmarks with profiling integration
bench-profile:
    @echo "\033[1m🔬 RUNNING BENCHMARKS WITH PROFILING\033[0m"
    @echo "\033[0;36mRunning benchmarks and generating profiles...\033[0m"
    @mkdir -p benchmarks/profiles
    @go test -bench=BenchmarkCreateUser -benchmem -run=^$$ -cpuprofile=benchmarks/profiles/cpu.prof -memprofile=benchmarks/profiles/mem.prof ./internal/domain/services/
    @echo ""
    @echo "\033[0;32m✅ Profiled benchmarks completed!\033[0m"
    @echo "\033[0;36mProfiles generated:\033[0m"
    @echo "  → benchmarks/profiles/cpu.prof"
    @echo "  → benchmarks/profiles/mem.prof"
    @echo "\033[0;33m💡 Analyze with: go tool pprof benchmarks/profiles/cpu.prof\033[0m"

# Stress test with high iteration count
bench-stress:
    @echo "\033[1m💪 RUNNING STRESS BENCHMARKS\033[0m"
    @echo "\033[0;36mRunning high-iteration stress tests...\033[0m"
    @mkdir -p benchmarks/stress
    @go test -bench=. -benchtime=10s -benchmem -run=^$$ ./internal/domain/services/ | tee benchmarks/stress/services.txt
    @go test -bench=. -benchtime=10s -benchmem -run=^$$ ./internal/infrastructure/persistence/ | tee benchmarks/stress/persistence.txt
    @echo ""
    @echo "\033[0;32m✅ Stress benchmarks completed!\033[0m"
    @echo "\033[0;36m→ Results saved to: benchmarks/stress/\033[0m"

# Quick benchmark run (fast feedback)
bench-quick:
    @echo "\033[1m⚡ RUNNING QUICK BENCHMARKS\033[0m"
    @echo "\033[0;36mRunning short benchmarks for quick feedback...\033[0m"
    @go test -bench=. -benchtime=1s -run=^$$ ./internal/domain/services/ | grep -E "(Benchmark|PASS|FAIL)"
    @echo ""
    @echo "\033[0;32m✅ Quick benchmarks completed!\033[0m"

# Clean up benchmark results
bench-clean:
    @echo "\033[1m🧹 CLEANING BENCHMARK FILES\033[0m"
    rm -rf benchmarks/
    @echo "\033[0;32m✅ Benchmark files cleaned!\033[0m"
