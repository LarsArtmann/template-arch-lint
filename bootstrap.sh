#!/bin/bash
# 🚀 Template Architecture Lint - Ultra-Reliable Bootstrap Script
# One command to rule them all: gets you from zero to enterprise-grade Go linting
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh | bash
#
# Or safer two-step:
#   curl -fsSL https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh -o bootstrap.sh
#   chmod +x bootstrap.sh && ./bootstrap.sh

set -euo pipefail  # Exit on any error, undefined variables, or pipe failures

# Colors and formatting
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly BOLD='\033[1m'
readonly NC='\033[0m' # No Color

# Configuration
readonly REPO_URL="https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master"
readonly REQUIRED_FILES=(".go-arch-lint.yml" ".golangci.yml" "justfile")
readonly MIN_GO_VERSION="1.19"

# Global state
INSTALL_LOG=()
INSTALLED_FILES=()
INSTALLED_TOOLS=()

# Logging functions
log_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
    INSTALL_LOG+=("✅ $1")
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
    INSTALL_LOG+=("⚠️  $1")
}

log_error() {
    echo -e "${RED}❌ $1${NC}" >&2
    INSTALL_LOG+=("❌ $1")
}

log_step() {
    echo -e "${BOLD}${BLUE}🔄 $1${NC}"
}

log_header() {
    echo -e "\n${BOLD}${PURPLE}$1${NC}"
    echo -e "${PURPLE}$(printf '=%.0s' $(seq 1 ${#1}))${NC}"
}

# Error handling and cleanup
cleanup_on_error() {
    log_error "Installation failed! Rolling back changes..."
    
    # Remove any files we installed
    for file in "${INSTALLED_FILES[@]}"; do
        if [[ -f "$file" ]]; then
            rm -f "$file" && log_info "Removed $file"
        fi
    done
    
    echo -e "\n${BOLD}${RED}💥 INSTALLATION FAILED${NC}"
    echo -e "${YELLOW}What happened:${NC}"
    for log_entry in "${INSTALL_LOG[@]}"; do
        echo "  $log_entry"
    done
    
    echo -e "\n${YELLOW}Troubleshooting options:${NC}"
    echo "1. Run with debug: bash -x bootstrap.sh"
    echo "2. Check requirements: Go 1.19+, git, curl"
    echo "3. Manual install: https://github.com/LarsArtmann/template-arch-lint#installation"
    echo "4. Open an issue: https://github.com/LarsArtmann/template-arch-lint/issues"
    
    exit 1
}

# Trap errors and run cleanup
trap cleanup_on_error ERR

# Utility functions
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

version_compare() {
    printf '%s\n%s\n' "$2" "$1" | sort -V -C
}

get_os() {
    case "$(uname -s)" in
        Darwin*) echo "macos" ;;
        Linux*)  echo "linux" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *) echo "unknown" ;;
    esac
}

get_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) echo "unknown" ;;
    esac
}

# Requirement checks
check_requirements() {
    log_header "🔍 CHECKING REQUIREMENTS"
    
    # Check if we're in a git repository
    if [[ ! -d ".git" ]]; then
        log_error "Not in a git repository. Please run from the root of your Go project."
        return 1
    fi
    log_success "Git repository detected"
    
    # Check if this is a Go project
    if [[ ! -f "go.mod" ]]; then
        log_error "No go.mod found. Please run from the root of a Go project."
        return 1
    fi
    log_success "Go project detected ($(basename $(head -1 go.mod | cut -d' ' -f2)))"
    
    # Check for required commands
    local missing_commands=()
    for cmd in curl git go; do
        if ! command_exists "$cmd"; then
            missing_commands+=("$cmd")
        fi
    done
    
    if [[ ${#missing_commands[@]} -gt 0 ]]; then
        log_error "Missing required commands: ${missing_commands[*]}"
        log_info "Please install: ${missing_commands[*]}"
        return 1
    fi
    log_success "Required commands available: curl, git, go"
    
    # Check Go version
    local go_version
    go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    if ! version_compare "$go_version" "$MIN_GO_VERSION"; then
        log_error "Go version $go_version is too old. Minimum required: $MIN_GO_VERSION"
        return 1
    fi
    log_success "Go version $go_version is compatible"
    
    # Platform detection
    local os arch
    os=$(get_os)
    arch=$(get_arch)
    log_success "Platform: $os/$arch"
    
    return 0
}

# Install just command runner if not present
install_just() {
    if command_exists just; then
        log_success "just command runner already installed ($(just --version))"
        return 0
    fi
    
    log_step "Installing just command runner..."
    
    local os arch
    os=$(get_os)
    arch=$(get_arch)
    
    case "$os" in
        macos)
            if command_exists brew; then
                brew install just
                INSTALLED_TOOLS+=("just (via homebrew)")
            else
                log_warning "Homebrew not found, trying alternative installation..."
                curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin
                export PATH="$HOME/.local/bin:$PATH"
                INSTALLED_TOOLS+=("just (to ~/.local/bin)")
            fi
            ;;
        linux)
            # Try package manager first, then fallback to direct install
            if command_exists apt-get; then
                sudo apt-get update && sudo apt-get install -y just
                INSTALLED_TOOLS+=("just (via apt)")
            elif command_exists yum; then
                sudo yum install -y just
                INSTALLED_TOOLS+=("just (via yum)")
            else
                log_info "Using direct installation method..."
                curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin
                export PATH="$HOME/.local/bin:$PATH"
                INSTALLED_TOOLS+=("just (to ~/.local/bin)")
            fi
            ;;
        *)
            log_warning "Unsupported platform for automatic just installation"
            log_info "Please install just manually: https://just.systems/man/en/chapter_4.html"
            return 1
            ;;
    esac
    
    # Verify installation
    if command_exists just; then
        log_success "just installed successfully ($(just --version))"
    else
        log_error "Failed to install just command runner"
        return 1
    fi
    
    return 0
}

# Download configuration files with validation
download_config_files() {
    log_header "📥 DOWNLOADING CONFIGURATION FILES"
    
    for file in "${REQUIRED_FILES[@]}"; do
        log_step "Downloading $file..."
        
        # Download to temporary file first
        local temp_file
        temp_file=$(mktemp)
        
        if ! curl -fsSL "$REPO_URL/$file" -o "$temp_file"; then
            log_error "Failed to download $file"
            rm -f "$temp_file"
            return 1
        fi
        
        # Basic validation
        if [[ ! -s "$temp_file" ]]; then
            log_error "Downloaded $file is empty"
            rm -f "$temp_file"
            return 1
        fi
        
        # Move to final location
        if [[ -f "$file" ]]; then
            log_info "Backing up existing $file to $file.backup"
            mv "$file" "$file.backup"
        fi
        
        mv "$temp_file" "$file"
        INSTALLED_FILES+=("$file")
        log_success "Downloaded and validated $file"
    done
    
    return 0
}

# Install linting tools
install_linting_tools() {
    log_header "🛠️  INSTALLING LINTING TOOLS"
    
    log_step "Running 'just install' to install linting tools..."
    
    # Run just install with output capture
    if just install; then
        log_success "All linting tools installed successfully"
        INSTALLED_TOOLS+=("golangci-lint" "go-arch-lint")
    else
        log_error "Failed to install linting tools via 'just install'"
        log_info "You can run 'just install' manually later"
        return 1
    fi
    
    return 0
}

# Verify installation works
verify_installation() {
    log_header "🧪 VERIFYING INSTALLATION"
    
    # Test basic justfile functionality
    log_step "Testing justfile commands..."
    if just --version >/dev/null 2>&1; then
        log_success "Justfile is working"
    else
        log_error "Justfile verification failed"
        return 1
    fi
    
    # Test if linting tools are accessible
    log_step "Testing linting tools..."
    local tools_working=true
    
    if command_exists golangci-lint; then
        log_success "golangci-lint is available ($(golangci-lint version --format short))"
    else
        log_warning "golangci-lint not found in PATH"
        tools_working=false
    fi
    
    if command_exists go-arch-lint; then
        log_success "go-arch-lint is available"
    else
        log_warning "go-arch-lint not found in PATH" 
        tools_working=false
    fi
    
    # Try running a quick architecture check (if we have Go files)
    if ls *.go >/dev/null 2>&1 || find . -name "*.go" -not -path "./vendor/*" | head -1 | grep -q "."; then
        log_step "Running quick architecture validation..."
        if timeout 30s just lint-arch >/dev/null 2>&1; then
            log_success "Architecture validation passed"
        else
            log_warning "Architecture validation had issues (this may be normal for new projects)"
        fi
    else
        log_info "No Go files found, skipping architecture validation"
    fi
    
    return 0
}

# Main installation function
main() {
    log_header "🚀 TEMPLATE ARCHITECTURE LINT - BOOTSTRAP INSTALLER"
    echo -e "${CYAN}Enterprise-grade Go linting setup in one command${NC}\n"
    
    # Run all installation steps
    check_requirements
    install_just
    download_config_files
    install_linting_tools
    verify_installation
    
    # Success message
    echo -e "\n${BOLD}${GREEN}🎉 INSTALLATION COMPLETE!${NC}"
    echo -e "${GREEN}$(printf '=%.0s' $(seq 1 25))${NC}"
    
    echo -e "\n${BOLD}📁 Files installed:${NC}"
    for file in "${INSTALLED_FILES[@]}"; do
        echo -e "  ${GREEN}✓${NC} $file"
    done
    
    if [[ ${#INSTALLED_TOOLS[@]} -gt 0 ]]; then
        echo -e "\n${BOLD}🛠️  Tools installed:${NC}"
        for tool in "${INSTALLED_TOOLS[@]}"; do
            echo -e "  ${GREEN}✓${NC} $tool"
        done
    fi
    
    echo -e "\n${BOLD}🚀 Ready to use:${NC}"
    echo -e "  ${CYAN}just lint${NC}           # Run ALL quality checks"
    echo -e "  ${CYAN}just lint-arch${NC}      # Architecture boundaries only"
    echo -e "  ${CYAN}just security-audit${NC} # Complete security scan"
    echo -e "  ${CYAN}just format${NC}         # Format code automatically"
    echo -e "  ${CYAN}just help${NC}           # Show all available commands"
    
    echo -e "\n${BOLD}📚 What you got:${NC}"
    echo -e "  ${GREEN}•${NC} Clean Architecture enforcement (domain boundaries)"
    echo -e "  ${GREEN}•${NC} 40+ code quality linters (complexity, naming, etc.)"
    echo -e "  ${GREEN}•${NC} Security scanning (gosec + govulncheck + NilAway)"
    echo -e "  ${GREEN}•${NC} Magic number/string detection"
    echo -e "  ${GREEN}•${NC} Zero tolerance for \`interface{}\`, \`any\`, \`panic()\`"
    
    echo -e "\n${YELLOW}💡 Pro tip:${NC} Run ${CYAN}just install-hooks${NC} to enable pre-commit linting!"
    
    return 0
}

# Execute main function only if script is run directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi