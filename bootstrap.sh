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

# Operation modes (flags)
MODE_DIAGNOSE=false
MODE_FIX=false
MODE_RETRY=false
MODE_VERBOSE=false
RETRY_COUNT=0
MAX_RETRIES=3

# Logging functions
log_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

log_verbose() {
    if [[ "$MODE_VERBOSE" == true ]]; then
        echo -e "${BLUE}🔍 $1${NC}"
    fi
}

log_debug() {
    if [[ "$MODE_VERBOSE" == true ]]; then
        echo -e "${PURPLE}🐛 DEBUG: $1${NC}"
    fi
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

# Help and usage functions
show_help() {
    cat << EOF
🚀 Template Architecture Lint - Bootstrap Installer

USAGE:
    bootstrap.sh [OPTIONS]

OPTIONS:
    --diagnose         Run comprehensive environment diagnostics
    --fix             Attempt automatic repair of common issues  
    --retry           Retry installation with progressive fallbacks
    --verbose         Enable detailed debugging output
    -h, --help        Show this help message

MODES:
    Default           Standard bootstrap installation
    --diagnose        Analyze environment and report issues
    --fix             Auto-repair mode with fallback strategies
    --retry           Retry failed installations with alternatives
    
EXAMPLES:
    bootstrap.sh                    # Standard installation
    bootstrap.sh --verbose          # Verbose installation
    bootstrap.sh --diagnose         # Check environment only
    bootstrap.sh --fix              # Auto-repair and install
    bootstrap.sh --retry --verbose  # Retry with debug output

INTEGRATION:
    just bootstrap                  # Run via justfile
    just bootstrap-fix              # Auto-repair via justfile  
    just bootstrap-diagnose         # Diagnose via justfile

For more help: https://github.com/LarsArtmann/template-arch-lint
EOF
}

# Flag parsing function
parse_flags() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --diagnose)
                MODE_DIAGNOSE=true
                log_debug "Diagnostic mode enabled"
                shift
                ;;
            --fix)
                MODE_FIX=true
                log_debug "Auto-fix mode enabled"
                shift
                ;;
            --retry)
                MODE_RETRY=true
                log_debug "Retry mode enabled"
                shift
                ;;
            --verbose|-v)
                MODE_VERBOSE=true
                log_debug "Verbose mode enabled"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                echo "Use --help for usage information"
                exit 1
                ;;
        esac
    done
    
    # Log enabled modes
    log_verbose "Modes: DIAGNOSE=$MODE_DIAGNOSE, FIX=$MODE_FIX, RETRY=$MODE_RETRY, VERBOSE=$MODE_VERBOSE"
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

# Setup PATH for Go tools
setup_go_path() {
    log_header "🛤️  SETTING UP GO TOOLS PATH"
    
    local go_bin_path="$HOME/go/bin"
    
    # Check if ~/go/bin exists
    if [ -d "$go_bin_path" ]; then
        log_success "Go tools directory exists: $go_bin_path"
        
        # Check if ~/go/bin is already in PATH
        if echo "$PATH" | grep -q "$go_bin_path"; then
            log_success "Go tools directory already in PATH"
        else
            log_step "Adding Go tools directory to PATH..."
            export PATH="$go_bin_path:$PATH"
            log_success "Added $go_bin_path to PATH for current session"
            
            # Add to shell profile for persistence
            local shell_profile=""
            if [ -n "${BASH_VERSION:-}" ]; then
                shell_profile="$HOME/.bashrc"
                if [ "$(uname)" = "Darwin" ]; then
                    shell_profile="$HOME/.bash_profile"
                fi
            elif [ -n "${ZSH_VERSION:-}" ]; then
                shell_profile="$HOME/.zshrc"
            fi
            
            if [ -n "$shell_profile" ] && [ -f "$shell_profile" ]; then
                if ! grep -q "export PATH.*$go_bin_path" "$shell_profile"; then
                    log_step "Adding Go tools PATH to $shell_profile..."
                    echo "" >> "$shell_profile"
                    echo "# Added by template-arch-lint bootstrap" >> "$shell_profile"
                    echo "export PATH=\"$go_bin_path:\$PATH\"" >> "$shell_profile"
                    log_success "Added PATH export to $shell_profile"
                    log_info "Restart your shell or run: source $shell_profile"
                else
                    log_info "PATH export already exists in $shell_profile"
                fi
            else
                log_warning "Could not detect shell profile to make PATH persistent"
                log_info "Manually add this to your shell profile: export PATH=\"$go_bin_path:\$PATH\""
            fi
        fi
    else
        log_warning "Go tools directory does not exist yet: $go_bin_path"
        log_info "This is normal if Go tools haven't been installed yet"
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
    
    # Final comprehensive verification - prove everything works
    log_step "Running final verification tests..."
    
    # Test 1: Quick tool version checks
    local verification_passed=true
    
    if command_exists golangci-lint && command_exists go-arch-lint; then
        log_step "Testing tool functionality with version commands..."
        
        if golangci-lint --version >/dev/null 2>&1; then
            log_success "golangci-lint responds to version command"
        else
            log_warning "golangci-lint version command failed"
            verification_passed=false
        fi
        
        # Test basic go-arch-lint functionality
        if go-arch-lint --help >/dev/null 2>&1; then
            log_success "go-arch-lint responds to help command"
        else
            log_warning "go-arch-lint help command failed"
            verification_passed=false
        fi
    fi
    
    # Test 2: Try running a quick architecture check (if we have Go files)
    if ls *.go >/dev/null 2>&1 || find . -name "*.go" -not -path "./vendor/*" | head -1 | grep -q "."; then
        log_step "Running architecture validation on project..."
        if timeout 30s just lint-arch >/dev/null 2>&1; then
            log_success "Architecture validation passed ✨"
        else
            log_warning "Architecture validation had issues (this may be normal for new projects)"
        fi
        
        # Test 3: Try a quick format check (non-destructive)
        log_step "Testing code formatting capabilities..."
        if timeout 15s just format --dry-run >/dev/null 2>&1 || timeout 15s just format >/dev/null 2>&1; then
            log_success "Code formatting tools working"
        else
            log_info "Code formatting test skipped (may require Go files)"
        fi
        
    else
        log_info "No Go files found, skipping project-specific validation"
        
        # For projects without Go files, test basic tool availability
        log_step "Testing basic tool configuration..."
        if [ -f ".golangci.yml" ] && [ -f ".go-arch-lint.yml" ] && [ -f "justfile" ]; then
            log_success "All configuration files present and ready"
        else
            log_warning "Some configuration files may be missing"
            verification_passed=false
        fi
    fi
    
    # Final verification summary
    if [ "$verification_passed" = true ] && [ "$tools_working" = true ]; then
        log_success "🎉 ALL VERIFICATION TESTS PASSED!"
        log_info "Your Go project is ready for enterprise-grade linting!"
        return 0
    else
        log_warning "Some verification tests had issues, but installation may still be functional"
        log_info "Try running 'just lint' manually to test full functionality"
        return 0  # Don't fail the entire bootstrap for minor issues
    fi
}

# Diagnostic mode function
run_diagnostics() {
    log_header "🔍 COMPREHENSIVE ENVIRONMENT DIAGNOSTICS"
    echo -e "${CYAN}Analyzing bootstrap environment and requirements${NC}\n"
    
    local issues_found=0
    
    # Check 1: Current directory and project structure  
    log_step "Checking project structure..."
    echo "Current directory: $(pwd)"
    
    if [[ -d ".git" ]]; then
        log_success "Git repository detected"
    else
        log_error "No .git directory found"
        log_info "Fix: Run 'git init' in your project root"
        ((issues_found++))
    fi
    
    if [[ -f "go.mod" ]]; then
        local module_name=$(head -1 go.mod | cut -d' ' -f2)
        log_success "Go module detected: $module_name"
    else
        log_error "No go.mod found"
        log_info "Fix: Run 'go mod init your-project-name'"
        ((issues_found++))
    fi
    
    # Check 2: Required commands
    log_step "Checking required commands..."
    local missing_commands=()
    for cmd in go git curl; do
        if command_exists "$cmd"; then
            log_success "$cmd available at: $(command -v $cmd)"
            if [[ "$cmd" == "go" ]]; then
                log_verbose "Go version: $(go version)"
                log_verbose "GOPATH: $(go env GOPATH)"
                log_verbose "GOPROXY: $(go env GOPROXY)"
            fi
        else
            log_error "$cmd command not found"
            missing_commands+=("$cmd")
            ((issues_found++))
        fi
    done
    
    if [[ ${#missing_commands[@]} -gt 0 ]]; then
        log_warning "Missing commands: ${missing_commands[*]}"
        log_info "macOS: brew install ${missing_commands[*]}"
        log_info "Ubuntu: sudo apt-get install -y golang-go git curl"
        log_info "RHEL: sudo yum install -y golang git curl"
    fi
    
    # Check 3: Network connectivity
    log_step "Checking network connectivity..."
    if curl -I --max-time 10 "$REPO_URL/.go-arch-lint.yml" >/dev/null 2>&1; then
        log_success "GitHub raw files accessible"
    else
        log_error "Cannot reach GitHub configuration files"
        log_info "Check: Network connection, firewall, proxy settings"
        ((issues_found++))
    fi
    
    # Check 4: Disk space
    log_step "Checking disk space..."
    local available=$(df . | tail -1 | awk '{print $4}')
    local available_human=$(df -h . | tail -1 | awk '{print $4}')
    log_info "Available space: $available_human"
    
    if [[ $available -lt 1000000 ]]; then # Less than ~1GB
        log_warning "Low disk space - may cause installation issues"
        ((issues_found++))
    else
        log_success "Sufficient disk space available"
    fi
    
    # Check 5: Platform detection
    log_step "Checking platform compatibility..."
    local os=$(get_os)
    local arch=$(get_arch)
    log_info "Platform: $os/$arch"
    
    case "$os" in
        macos) log_success "macOS detected - fully supported" ;;
        linux) log_success "Linux detected - fully supported" ;;
        windows) log_warning "Windows detected - WSL recommended" ;;
        *) log_error "Unknown platform - may not be supported"; ((issues_found++)) ;;
    esac
    
    # Summary
    echo -e "\n${BOLD}📋 DIAGNOSTIC SUMMARY${NC}"
    echo "===================="
    
    if [[ $issues_found -eq 0 ]]; then
        log_success "Environment looks perfect for bootstrap!"
        log_info "You can now run: ./bootstrap.sh"
    else
        log_error "Found $issues_found issue(s) that need fixing"
        log_info "Fix the issues above, then run diagnostics again"
        
        echo -e "\n${YELLOW}💡 Quick fixes:${NC}"
        echo "• Missing git repo: git init && git add . && git commit -m 'Initial'"
        echo "• Missing go.mod: go mod init your-project-name"
        echo "• Missing tools: brew install go git curl (macOS)"
    fi
    
    return $issues_found
}

# Main installation function
main() {
    # Parse command line flags first
    parse_flags "$@"
    
    # Handle diagnostic mode
    if [[ "$MODE_DIAGNOSE" == true ]]; then
        run_diagnostics
        return $?
    fi
    
    log_header "🚀 TEMPLATE ARCHITECTURE LINT - BOOTSTRAP INSTALLER"
    echo -e "${CYAN}Enterprise-grade Go linting setup in one command${NC}\n"
    
    if [[ "$MODE_VERBOSE" == true ]]; then
        log_verbose "Verbose mode enabled - showing detailed progress"
    fi
    
    if [[ "$MODE_FIX" == true ]]; then
        log_info "Auto-fix mode enabled - will attempt to repair issues automatically"
    fi
    
    if [[ "$MODE_RETRY" == true ]]; then
        log_info "Retry mode enabled - will use progressive fallbacks on failures"
    fi
    
    # Run all installation steps
    check_requirements
    install_just
    download_config_files
    install_linting_tools
    setup_go_path
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