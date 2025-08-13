# 🔥 ENTERPRISE-GRADE GO LINTING JUSTFILE
# Complete architecture and code quality enforcement
#
# Just is a handy way to save and run project-specific commands.
# https://github.com/casey/just

# Tool versions
GOLANGCI_VERSION := "v2.3.1"
GO_ARCH_LINT_VERSION := "v1.12.0"

# Directories
ROOT_DIR := justfile_directory()
REPORT_DIR := ROOT_DIR / "reports"

# Default recipe (runs when just is called without arguments)
default: help

# Show this help message
help:
    @echo "\033[1m🔥 ENTERPRISE GO LINTING JUSTFILE\033[0m"
    @echo ""
    @echo "\033[1mUSAGE:\033[0m"
    @just --list --unsorted
    @echo ""
    @echo "\033[1mQUICK START:\033[0m"
    @echo "  1. \033[0;36mjust install\033[0m  - Install required tools"
    @echo "  2. \033[0;36mjust lint\033[0m     - Run all linters"
    @echo "  3. \033[0;36mjust fix\033[0m      - Auto-fix issues"
    @echo "  4. \033[0;36mjust run\033[0m      - Run the application"
    @echo ""
    @echo "\033[1mDOCKER COMMANDS:\033[0m"
    @echo "  • \033[0;36mjust docker-test\033[0m         - Build and test Docker image"
    @echo "  • \033[0;36mjust docker-dev-detached\033[0m - Start full dev environment"
    @echo "  • \033[0;36mjust docker-stop\033[0m         - Stop development environment"
    @echo ""
    @echo "\033[1mCODE ANALYSIS:\033[0m"
    @echo "  • \033[0;36mjust fd\033[0m                  - Find duplicate code (alias for find-duplicates)"
    @echo "  • \033[0;36mjust find-duplicates\033[0m     - Find duplicate code with custom threshold"

# Install all required linting tools
install:
    @echo "\033[1m📦 Installing linting tools...\033[0m"
    @echo "\033[0;33mInstalling golangci-lint {{GOLANGCI_VERSION}}...\033[0m"
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@{{GOLANGCI_VERSION}}
    @echo "\033[0;33mInstalling go-arch-lint {{GO_ARCH_LINT_VERSION}}...\033[0m"
    go install github.com/fe3dback/go-arch-lint@{{GO_ARCH_LINT_VERSION}}
    @echo "\033[0;32m✅ All tools installed successfully!\033[0m"

# Run all linters (architecture + code quality + filenames)
lint: lint-files lint-arch lint-code
    @echo ""
    @echo "\033[0;32m\033[1m✅ All linting checks completed!\033[0m"

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
    @echo "\033[0;36mRunning golangci-lint...\033[0m"
    @if command -v golangci-lint >/dev/null 2>&1; then \
        if golangci-lint run --config .golangci.yml; then \
            echo "\033[0;32m✅ Code quality validation passed!\033[0m"; \
        else \
            echo "\033[0;31m❌ Code quality issues found!\033[0m" >&2; \
            exit 1; \
        fi; \
    else \
        echo "\033[0;31m❌ golangci-lint not installed. Run 'just install' first.\033[0m" >&2; \
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

# Auto-fix issues where possible
fix:
    @echo "\033[1m🔧 AUTO-FIXING ISSUES\033[0m"
    @echo "\033[0;33mRunning gofmt...\033[0m"
    gofmt -w -s .
    @echo "\033[0;33mRunning goimports...\033[0m"
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    else \
        go install golang.org/x/tools/cmd/goimports@latest; \
        goimports -w .; \
    fi
    @echo "\033[0;33mRunning golangci-lint with --fix...\033[0m"
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run --fix --config .golangci.yml || true; \
    fi
    @echo "\033[0;32m✅ Auto-fix completed!\033[0m"

# Run all checks (for CI/CD pipelines)
ci: lint test
    @echo "\033[0;36mChecking module dependencies...\033[0m"
    go mod verify
    go mod tidy -diff
    @echo "\033[0;32m\033[1m✅ CI/CD checks passed!\033[0m"

# Run tests with coverage
test:
    @echo "\033[1m🧪 RUNNING TESTS\033[0m"
    @echo "\033[0;36mRunning tests with coverage...\033[0m"
    go test ./... -v -race -coverprofile=coverage.out
    @echo "\033[0;32m✅ Tests completed!\033[0m"

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
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run --out-format json > {{REPORT_DIR}}/quality.json 2>/dev/null || true; \
        golangci-lint run --out-format checkstyle > {{REPORT_DIR}}/checkstyle.xml 2>/dev/null || true; \
        golangci-lint run --out-format junit-xml > {{REPORT_DIR}}/junit.xml 2>/dev/null || true; \
        echo "  → {{REPORT_DIR}}/quality.json"; \
        echo "  → {{REPORT_DIR}}/checkstyle.xml"; \
        echo "  → {{REPORT_DIR}}/junit.xml"; \
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
    golangci-lint run --fast --config .golangci.yml

# Run with maximum strictness (slower but thorough)
lint-strict:
    @echo "\033[1m🔥 MAXIMUM STRICTNESS LINTING\033[0m"
    golangci-lint run --config .golangci.yml --max-issues-per-linter 0 --max-same-issues 0

# Run security-focused linters only
lint-security:
    @echo "\033[1m🔒 SECURITY LINTING\033[0m"
    golangci-lint run --config .golangci.yml --enable-only gosec,copyloopvar

# Format code
fmt:
    @echo "\033[1m📝 FORMATTING CODE\033[0m"
    gofmt -w -s .
    @if command -v goimports >/dev/null 2>&1; then \
        goimports -w .; \
    fi
    @echo "\033[0;32m✅ Code formatted!\033[0m"

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

# Run benchmarks
bench:
    @echo "\033[1m⚡ RUNNING BENCHMARKS\033[0m"
    go test -bench=. -benchmem ./...

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
    golangci-lint run -v --config .golangci.yml

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
    @if command -v golangci-lint >/dev/null 2>&1; then \
        echo "golangci-lint version:"; \
        golangci-lint version; \
    fi
    @if command -v go-arch-lint >/dev/null 2>&1; then \
        echo "go-arch-lint version:"; \
        go-arch-lint version; \
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

# Start development environment with Docker Compose
docker-dev:
    @echo "\033[1m🔄 STARTING DEVELOPMENT ENVIRONMENT\033[0m"
    docker-compose up --build

# Start development environment in background
docker-dev-detached:
    @echo "\033[1m🔄 STARTING DEVELOPMENT ENVIRONMENT (DETACHED)\033[0m"
    docker-compose up --build -d
    @echo "\033[0;32m✅ Development environment started!\033[0m"
    @echo "\033[0;36mServices available at:\033[0m"
    @echo "  - Application: http://localhost:8080"
    @echo "  - Grafana: http://localhost:3000 (admin/admin)"
    @echo "  - Prometheus: http://localhost:9090"
    @echo "  - Jaeger UI: http://localhost:16686"

# Stop development environment
docker-stop:
    @echo "\033[1m🛑 STOPPING DEVELOPMENT ENVIRONMENT\033[0m"
    docker-compose down
    @echo "\033[0;32m✅ Development environment stopped!\033[0m"

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

# 🔍 Code Duplication Detection

# Find duplicate code blocks using dupl tool
find-duplicates threshold="50":
    @echo "\033[1m🔍 FINDING DUPLICATE CODE\033[0m"
    @echo "\033[0;36mUsing threshold: {{threshold}} tokens...\033[0m"
    @if command -v dupl >/dev/null 2>&1; then \
        echo "\033[0;33mAnalyzing Go source files...\033[0m"; \
        dupl -t {{threshold}} . > duplication-report.txt; \
        if dupl -t {{threshold}} -html . > duplication-report.html; then \
            echo "\033[0;32m✅ Duplication analysis completed!\033[0m"; \
            echo "\033[0;36m→ Text report: ./duplication-report.txt\033[0m"; \
            echo "\033[0;36m→ HTML report: ./duplication-report.html\033[0m"; \
        else \
            echo "\033[0;31m❌ Duplication analysis failed!\033[0m" >&2; \
            exit 1; \
        fi; \
    else \
        echo "\033[0;31m❌ dupl not installed. Installing...\033[0m"; \
        go install github.com/mibk/dupl@latest; \
        echo "\033[0;33mAnalyzing Go source files...\033[0m"; \
        dupl -t {{threshold}} . > duplication-report.txt; \
        if dupl -t {{threshold}} -html . > duplication-report.html; then \
            echo "\033[0;32m✅ Duplication analysis completed!\033[0m"; \
            echo "\033[0;36m→ Text report: ./duplication-report.txt\033[0m"; \
            echo "\033[0;36m→ HTML report: ./duplication-report.html\033[0m"; \
        else \
            echo "\033[0;31m❌ Duplication analysis failed!\033[0m" >&2; \
            exit 1; \
        fi; \
    fi

# Alias for find-duplicates (fd)
fd threshold="50": (find-duplicates threshold)

# Find high-confidence duplicates (stricter threshold)
find-duplicates-strict: (find-duplicates "100")

# Find all potential duplicates (looser threshold) 
find-duplicates-loose: (find-duplicates "25")