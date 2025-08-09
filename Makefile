# 🔥 ENTERPRISE-GRADE GO LINTING MAKEFILE
# Complete architecture and code quality enforcement
#
# Usage:
#   make help           - Show this help message
#   make lint           - Run all linters (architecture + code quality + filenames)
#   make lint-arch      - Run architecture linting only
#   make lint-code      - Run code quality linting only
#   make lint-files     - Run filename verification only
#   make install        - Install all required tools
#   make fix            - Auto-fix issues where possible
#   make ci             - Run all checks (for CI/CD)
#   make report         - Generate detailed reports

.PHONY: help lint lint-arch lint-code lint-files install fix ci report clean

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
CYAN := \033[0;36m
WHITE := \033[0;37m
BOLD := \033[1m
RESET := \033[0m

# Tool versions
GOLANGCI_VERSION := v1.61.0
GO_ARCH_LINT_VERSION := v1.8.0

# Directories
ROOT_DIR := $(shell pwd)
REPORT_DIR := reports

help: ## Show this help message
	@echo "$(BOLD)🔥 ENTERPRISE GO LINTING MAKEFILE$(RESET)"
	@echo ""
	@echo "$(BOLD)USAGE:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)QUICK START:$(RESET)"
	@echo "  1. $(CYAN)make install$(RESET)  - Install required tools"
	@echo "  2. $(CYAN)make lint$(RESET)     - Run all linters"
	@echo "  3. $(CYAN)make fix$(RESET)      - Auto-fix issues"

install: ## Install all required linting tools
	@echo "$(BOLD)📦 Installing linting tools...$(RESET)"
	@echo "$(YELLOW)Installing golangci-lint $(GOLANGCI_VERSION)...$(RESET)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION)
	@echo "$(YELLOW)Installing go-arch-lint $(GO_ARCH_LINT_VERSION)...$(RESET)"
	@go install github.com/fe3dback/go-arch-lint@$(GO_ARCH_LINT_VERSION)
	@echo "$(YELLOW)Building filename-verifier...$(RESET)"
	@go build -o bin/filename-verifier cmd/filename-verifier/main.go
	@echo "$(GREEN)✅ All tools installed successfully!$(RESET)"

lint: ## Run all linters (architecture + code quality + filenames)
	@echo "$(BOLD)🔍 RUNNING COMPLETE LINTING SUITE$(RESET)"
	@echo ""
	@$(MAKE) lint-files
	@echo ""
	@$(MAKE) lint-arch
	@echo ""
	@$(MAKE) lint-code
	@echo ""
	@echo "$(GREEN)$(BOLD)✅ All linting checks completed!$(RESET)"

lint-arch: ## Run architecture linting only
	@echo "$(BOLD)🏗️  ARCHITECTURE LINTING$(RESET)"
	@echo "$(CYAN)Running go-arch-lint...$(RESET)"
	@if command -v go-arch-lint >/dev/null 2>&1; then \
		go-arch-lint check || (echo "$(RED)❌ Architecture violations found!$(RESET)" && exit 1); \
		echo "$(GREEN)✅ Architecture validation passed!$(RESET)"; \
	else \
		echo "$(RED)❌ go-arch-lint not installed. Run 'make install' first.$(RESET)"; \
		exit 1; \
	fi

lint-code: ## Run code quality linting only
	@echo "$(BOLD)📝 CODE QUALITY LINTING$(RESET)"
	@echo "$(CYAN)Running golangci-lint...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci.yml || (echo "$(RED)❌ Code quality issues found!$(RESET)" && exit 1); \
		echo "$(GREEN)✅ Code quality validation passed!$(RESET)"; \
	else \
		echo "$(RED)❌ golangci-lint not installed. Run 'make install' first.$(RESET)"; \
		exit 1; \
	fi

lint-files: ## Run filename verification only
	@echo "$(BOLD)📁 FILENAME VERIFICATION$(RESET)"
	@echo "$(CYAN)Running filename-verifier...$(RESET)"
	@if [ -f bin/filename-verifier ]; then \
		./bin/filename-verifier . || (echo "$(RED)❌ Filename violations found!$(RESET)" && exit 1); \
	else \
		go run cmd/filename-verifier/main.go . || (echo "$(RED)❌ Filename violations found!$(RESET)" && exit 1); \
	fi

fix: ## Auto-fix issues where possible
	@echo "$(BOLD)🔧 AUTO-FIXING ISSUES$(RESET)"
	@echo "$(YELLOW)Running gofmt...$(RESET)"
	@gofmt -w -s .
	@echo "$(YELLOW)Running goimports...$(RESET)"
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		go install golang.org/x/tools/cmd/goimports@latest && goimports -w .; \
	fi
	@echo "$(YELLOW)Running golangci-lint with --fix...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix --config .golangci.yml; \
	fi
	@echo "$(GREEN)✅ Auto-fix completed!$(RESET)"

ci: ## Run all checks (for CI/CD pipelines)
	@echo "$(BOLD)🚀 CI/CD PIPELINE CHECKS$(RESET)"
	@$(MAKE) lint
	@echo "$(CYAN)Running tests...$(RESET)"
	@go test ./... -v -race -coverprofile=coverage.out
	@echo "$(CYAN)Checking module dependencies...$(RESET)"
	@go mod verify
	@go mod tidy -diff
	@echo "$(GREEN)$(BOLD)✅ CI/CD checks passed!$(RESET)"

report: ## Generate detailed linting reports
	@echo "$(BOLD)📊 GENERATING REPORTS$(RESET)"
	@mkdir -p $(REPORT_DIR)
	@echo "$(YELLOW)Generating architecture report...$(RESET)"
	@if command -v go-arch-lint >/dev/null 2>&1; then \
		go-arch-lint check --json > $(REPORT_DIR)/architecture.json 2>/dev/null || true; \
		go-arch-lint graph > $(REPORT_DIR)/dependencies.dot 2>/dev/null || true; \
		echo "  → $(REPORT_DIR)/architecture.json"; \
		echo "  → $(REPORT_DIR)/dependencies.dot"; \
	fi
	@echo "$(YELLOW)Generating code quality report...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --out-format json > $(REPORT_DIR)/quality.json 2>/dev/null || true; \
		golangci-lint run --out-format checkstyle > $(REPORT_DIR)/checkstyle.xml 2>/dev/null || true; \
		golangci-lint run --out-format junit-xml > $(REPORT_DIR)/junit.xml 2>/dev/null || true; \
		echo "  → $(REPORT_DIR)/quality.json"; \
		echo "  → $(REPORT_DIR)/checkstyle.xml"; \
		echo "  → $(REPORT_DIR)/junit.xml"; \
	fi
	@echo "$(YELLOW)Generating test coverage report...$(RESET)"
	@go test ./... -coverprofile=$(REPORT_DIR)/coverage.out 2>/dev/null || true
	@go tool cover -html=$(REPORT_DIR)/coverage.out -o $(REPORT_DIR)/coverage.html 2>/dev/null || true
	@echo "  → $(REPORT_DIR)/coverage.out"
	@echo "  → $(REPORT_DIR)/coverage.html"
	@echo "$(GREEN)✅ Reports generated in $(REPORT_DIR)/$(RESET)"

clean: ## Clean generated files and reports
	@echo "$(BOLD)🧹 CLEANING$(RESET)"
	@rm -rf $(REPORT_DIR)
	@rm -rf bin/
	@rm -f coverage.out
	@echo "$(GREEN)✅ Cleaned successfully!$(RESET)"

# Additional targets for specific use cases

.PHONY: lint-minimal lint-strict lint-security

lint-minimal: ## Run minimal essential linters only
	@echo "$(BOLD)⚡ MINIMAL LINTING$(RESET)"
	@golangci-lint run --fast --config .golangci.yml

lint-strict: ## Run with maximum strictness (slower but thorough)
	@echo "$(BOLD)🔥 MAXIMUM STRICTNESS LINTING$(RESET)"
	@golangci-lint run --config .golangci.yml --max-issues-per-linter 0 --max-same-issues 0

lint-security: ## Run security-focused linters only
	@echo "$(BOLD)🔒 SECURITY LINTING$(RESET)"
	@golangci-lint run --config .golangci.yml --enable-only gosec,exportloopref

# Watch mode for development
.PHONY: watch

watch: ## Watch for changes and auto-lint
	@echo "$(BOLD)👁️  WATCH MODE$(RESET)"
	@echo "$(YELLOW)Watching for changes...$(RESET)"
	@while true; do \
		find . -name "*.go" -not -path "./vendor/*" | \
		entr -d -c sh -c 'clear && make lint'; \
	done

# Default target
.DEFAULT_GOAL := help