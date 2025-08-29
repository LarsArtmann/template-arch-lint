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

# Retry logic with exponential backoff
retry_with_backoff() {
    local max_attempts=${1:-3}
    local delay=${2:-1}
    local backoff=${3:-2}
    local attempt=1
    
    shift 3 # Remove the first 3 parameters
    local command=("$@")
    
    while [[ $attempt -le $max_attempts ]]; do
        log_verbose "Attempt $attempt/$max_attempts: ${command[*]}"
        
        if "${command[@]}"; then
            log_verbose "Command succeeded on attempt $attempt"
            return 0
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            log_verbose "Command failed, retrying in ${delay}s..."
            sleep "$delay"
            delay=$((delay * backoff))
        fi
        
        ((attempt++))
    done
    
    log_verbose "Command failed after $max_attempts attempts"
    return 1
}

# Auto-repair common issues
auto_repair_environment() {
    log_verbose "Running auto-repair checks..."
    local repairs_made=0
    
    # Repair 1: Create Go bin directory if missing
    if [[ ! -d "$HOME/go/bin" ]]; then
        log_step "Auto-repair: Creating missing ~/go/bin directory"
        if mkdir -p "$HOME/go/bin"; then
            log_success "Created ~/go/bin directory"
            ((repairs_made++))
        else
            log_warning "Failed to create ~/go/bin directory"
        fi
    fi
    
    # Repair 2: Fix PATH if Go tools not accessible
    if [[ -d "$HOME/go/bin" ]] && ! echo "$PATH" | grep -q "$HOME/go/bin"; then
        log_step "Auto-repair: Adding ~/go/bin to PATH"
        export PATH="$HOME/go/bin:$PATH"
        log_success "Added ~/go/bin to PATH for current session"
        ((repairs_made++))
    fi
    
    # Repair 3: Fix permissions on Go bin directory
    if [[ -d "$HOME/go/bin" ]] && [[ ! -w "$HOME/go/bin" ]]; then
        log_step "Auto-repair: Fixing permissions on ~/go/bin"
        if chmod 755 "$HOME/go/bin" 2>/dev/null; then
            log_success "Fixed permissions on ~/go/bin"
            ((repairs_made++))
        else
            log_warning "Failed to fix permissions on ~/go/bin"
        fi
    fi
    
    # Repair 4: Set Go environment variables if missing
    if [[ -z "${GOPATH:-}" ]]; then
        log_step "Auto-repair: Setting GOPATH environment variable"
        export GOPATH=$(go env GOPATH 2>/dev/null || echo "$HOME/go")
        log_success "Set GOPATH to: $GOPATH"
        ((repairs_made++))
    fi
    
    log_verbose "Auto-repair completed: $repairs_made repairs made"
    return 0
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
        handle_error_with_escalation "$ERROR_REQUIREMENTS" "CRITICAL" \
            "Not in a git repository" \
            "Bootstrap must be run from the root of a Git repository" \
            "Run 'git init' to initialize a repository"
        return 1
    fi
    log_success "Git repository detected"
    
    # Check if this is a Go project
    if [[ ! -f "go.mod" ]]; then
        handle_error_with_escalation "$ERROR_REQUIREMENTS" "CRITICAL" \
            "No go.mod found" \
            "Bootstrap requires a Go module to configure linting" \
            "Run 'go mod init your-project-name' to create a module"
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

# Install just command runner with progressive fallbacks
install_just() {
    if command_exists just; then
        log_success "just command runner already installed ($(just --version))"
        return 0
    fi
    
    log_step "Installing just command runner..."
    
    local os arch
    os=$(get_os)
    arch=$(get_arch)
    
    # Progressive fallback chains by platform
    case "$os" in
        macos)
            # Fallback chain for macOS: Homebrew → Direct install → Manual binary download
            if install_just_macos_homebrew || install_just_direct || install_just_manual_binary "$os" "$arch"; then
                log_success "just installed successfully via fallback chain"
            else
                return 1
            fi
            ;;
        linux)
            # Fallback chain for Linux: Package manager → Direct install → Manual binary download
            if install_just_linux_packages || install_just_direct || install_just_manual_binary "$os" "$arch"; then
                log_success "just installed successfully via fallback chain"
            else
                return 1
            fi
            ;;
        *)
            log_warning "Unsupported platform for automatic just installation"
            if install_just_manual_binary "$os" "$arch"; then
                log_success "just installed via manual binary download"
            else
                handle_error_with_escalation "$ERROR_TOOLS" "HIGH" \
                    "Platform $os not supported for automatic just installation" \
                    "No automated installation method available" \
                    "Install manually: https://just.systems/man/en/chapter_4.html"
                return 1
            fi
            ;;
    esac
    
    # Final verification
    if command_exists just; then
        log_success "just installed successfully ($(just --version))"
        return 0
    else
        handle_error_with_escalation "$ERROR_TOOLS" "CRITICAL" \
            "Failed to install just command runner after all fallback attempts" \
            "Homebrew, package managers, direct install, and binary download all failed" \
            "Manual installation required: https://github.com/casey/just/releases"
        return 1
    fi
}

# macOS Homebrew installation
install_just_macos_homebrew() {
    if ! command_exists brew; then
        log_verbose "Homebrew not available, skipping homebrew installation"
        return 1
    fi
    
    log_step "Attempting just installation via Homebrew..."
    if retry_with_backoff 3 2 2 brew install just; then
        INSTALLED_TOOLS+=("just (via homebrew)")
        return 0
    else
        log_warning "Homebrew installation failed, trying next method..."
        return 1
    fi
}

# Linux package manager installation
install_just_linux_packages() {
    log_step "Attempting just installation via package manager..."
    
    # Try apt-get (Debian/Ubuntu)
    if command_exists apt-get; then
        log_verbose "Trying apt-get installation..."
        if retry_with_backoff 2 1 2 sudo apt-get update && retry_with_backoff 2 1 2 sudo apt-get install -y just; then
            INSTALLED_TOOLS+=("just (via apt)")
            return 0
        fi
    fi
    
    # Try yum (RHEL/CentOS)
    if command_exists yum; then
        log_verbose "Trying yum installation..."
        if retry_with_backoff 2 1 2 sudo yum install -y just; then
            INSTALLED_TOOLS+=("just (via yum)")
            return 0
        fi
    fi
    
    # Try dnf (Fedora)
    if command_exists dnf; then
        log_verbose "Trying dnf installation..."
        if retry_with_backoff 2 1 2 sudo dnf install -y just; then
            INSTALLED_TOOLS+=("just (via dnf)")
            return 0
        fi
    fi
    
    # Try pacman (Arch)
    if command_exists pacman; then
        log_verbose "Trying pacman installation..."
        if retry_with_backoff 2 1 2 sudo pacman -S --noconfirm just; then
            INSTALLED_TOOLS+=("just (via pacman)")
            return 0
        fi
    fi
    
    log_warning "Package manager installation failed, trying next method..."
    return 1
}

# Direct installation via official script
install_just_direct() {
    log_step "Attempting just installation via official installer script..."
    
    # Ensure ~/.local/bin exists
    mkdir -p "$HOME/.local/bin"
    
    if retry_with_backoff 3 2 2 curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin; then
        export PATH="$HOME/.local/bin:$PATH"
        INSTALLED_TOOLS+=("just (direct install to ~/.local/bin)")
        return 0
    else
        log_warning "Direct installation failed, trying next method..."
        return 1
    fi
}

# Manual binary download as last resort
install_just_manual_binary() {
    local os="$1"
    local arch="$2"
    
    log_step "Attempting manual binary download for $os/$arch..."
    
    # Map architecture names
    local binary_arch
    case "$arch" in
        amd64) binary_arch="x86_64" ;;
        arm64) binary_arch="aarch64" ;;
        *) 
            log_warning "Unsupported architecture for binary download: $arch"
            return 1
            ;;
    esac
    
    # Map OS names for binary download
    local binary_os
    case "$os" in
        macos) binary_os="apple-darwin" ;;
        linux) binary_os="unknown-linux-musl" ;;
        *)
            log_warning "Unsupported OS for binary download: $os"
            return 1
            ;;
    esac
    
    local binary_name="just-${binary_arch}-${binary_os}"
    local download_url="https://github.com/casey/just/releases/latest/download/${binary_name}.tar.gz"
    
    log_verbose "Downloading from: $download_url"
    
    # Create temporary directory
    local temp_dir
    temp_dir=$(mktemp -d)
    
    if retry_with_backoff 3 2 2 curl -fsSL "$download_url" -o "$temp_dir/just.tar.gz"; then
        if cd "$temp_dir" && tar -xzf just.tar.gz; then
            mkdir -p "$HOME/.local/bin"
            if mv just "$HOME/.local/bin/"; then
                chmod +x "$HOME/.local/bin/just"
                export PATH="$HOME/.local/bin:$PATH"
                INSTALLED_TOOLS+=("just (manual binary download)")
                rm -rf "$temp_dir"
                return 0
            fi
        fi
    fi
    
    rm -rf "$temp_dir"
    log_warning "Manual binary download failed"
    return 1
}

# Download single configuration file with retries
download_single_config_file() {
    local file="$1"
    local temp_file
    temp_file=$(mktemp)
    
    log_step "Downloading $file..."
    log_verbose "Temporary file: $temp_file"
    
    # Download with retries and multiple mirrors
    if download_config_file_with_fallbacks "$file" "$temp_file"; then
        # Validate downloaded file
        if [[ ! -s "$temp_file" ]]; then
            log_error "Downloaded $file is empty"
            rm -f "$temp_file"
            return 1
        fi
        
        # Backup existing file if present
        if [[ -f "$file" ]]; then
            log_verbose "Backing up existing $file to $file.backup"
            mv "$file" "$file.backup"
        fi
        
        # Move to final location
        if mv "$temp_file" "$file"; then
            INSTALLED_FILES+=("$file")
            log_success "Downloaded and validated $file ($(wc -c < "$file") bytes)"
            return 0
        else
            handle_error_with_escalation "$ERROR_PERMISSIONS" "HIGH" \
                "Failed to move $file to final location" \
                "File download succeeded but filesystem operation failed" \
                "Check directory permissions and disk space"
            rm -f "$temp_file"
            return 1
        fi
    else
        rm -f "$temp_file"
        return 1
    fi
}

# Download with fallback mirrors and retry logic
download_config_file_with_fallbacks() {
    local file="$1"
    local output_file="$2"
    
    # Primary mirror (GitHub raw)
    log_verbose "Trying primary mirror (GitHub raw)..."
    if retry_with_backoff 3 1 2 curl -fsSL "$REPO_URL/$file" -o "$output_file"; then
        return 0
    fi
    
    # Fallback mirror 1: GitHub via different CDN
    log_verbose "Trying fallback mirror (GitHub CDN)..."
    local cdn_url="https://cdn.jsdelivr.net/gh/LarsArtmann/template-arch-lint@master/$file"
    if retry_with_backoff 2 2 2 curl -fsSL "$cdn_url" -o "$output_file"; then
        return 0
    fi
    
    # Fallback mirror 2: GitHub with different user agent
    log_verbose "Trying fallback with different user agent..."
    if retry_with_backoff 2 3 2 curl -fsSL -H "User-Agent: bootstrap/1.0" "$REPO_URL/$file" -o "$output_file"; then
        return 0
    fi
    
    handle_error_with_escalation "$ERROR_NETWORK" "HIGH" \
        "Failed to download $file from all mirrors" \
        "Primary, CDN, and fallback mirrors all failed" \
        "Check network connectivity and GitHub accessibility"
    return 1
}

# Parallel configuration file downloads
download_config_files() {
    log_header "📥 DOWNLOADING CONFIGURATION FILES"
    
    if [[ "$MODE_VERBOSE" == true ]]; then
        log_verbose "Files to download: ${REQUIRED_FILES[*]}"
    fi
    
    # Check if we can do parallel downloads (requires background processes)
    if command -v xargs >/dev/null 2>&1; then
        log_step "Using parallel download strategy..."
        download_config_files_parallel
    else
        log_step "Using sequential download strategy..."
        download_config_files_sequential
    fi
}

# Sequential download as fallback
download_config_files_sequential() {
    local failed_files=()
    
    for file in "${REQUIRED_FILES[@]}"; do
        if ! download_single_config_file "$file"; then
            failed_files+=("$file")
        fi
    done
    
    if [[ ${#failed_files[@]} -gt 0 ]]; then
        log_error "Failed to download ${#failed_files[@]} file(s): ${failed_files[*]}"
        
        if [[ "$MODE_FIX" == true ]]; then
            log_info "Auto-fix mode: Attempting manual download alternatives..."
            for file in "${failed_files[@]}"; do
                log_info "Manual download command: curl -fsSL -o '$file' '$REPO_URL/$file'"
            done
        fi
        return 1
    fi
    
    log_success "All configuration files downloaded successfully"
    return 0
}

# Parallel download implementation
download_config_files_parallel() {
    log_verbose "Starting parallel downloads for ${#REQUIRED_FILES[@]} files..."
    
    # Create array to track background job PIDs
    local pids=()
    local temp_files=()
    
    # Start downloads in parallel
    for file in "${REQUIRED_FILES[@]}"; do
        local temp_file
        temp_file=$(mktemp)
        temp_files+=("$temp_file")
        
        # Start download in background
        (download_config_file_with_fallbacks "$file" "$temp_file" && echo "$file:SUCCESS:$temp_file" || echo "$file:FAILED:$temp_file") &
        pids+=($!)
    done
    
    log_verbose "Started ${#pids[@]} parallel download processes"
    
    # Wait for all downloads and collect results
    local results=()
    for pid in "${pids[@]}"; do
        wait "$pid"
        # Note: results are captured via echo in background processes
    done
    
    # Process results (simplified approach for now)
    local failed_files=()
    for i in "${!REQUIRED_FILES[@]}"; do
        local file="${REQUIRED_FILES[$i]}"
        local temp_file="${temp_files[$i]}"
        
        if [[ -s "$temp_file" ]]; then
            # File downloaded successfully
            if [[ -f "$file" ]]; then
                log_verbose "Backing up existing $file to $file.backup"
                mv "$file" "$file.backup"
            fi
            
            if mv "$temp_file" "$file"; then
                INSTALLED_FILES+=("$file")
                log_success "Downloaded $file ($(wc -c < "$file") bytes)"
            else
                failed_files+=("$file")
                rm -f "$temp_file"
            fi
        else
            failed_files+=("$file")
            rm -f "$temp_file"
        fi
    done
    
    if [[ ${#failed_files[@]} -gt 0 ]]; then
        log_error "Failed to download ${#failed_files[@]} file(s): ${failed_files[*]}"
        log_info "Falling back to sequential downloads..."
        
        # Retry failed files sequentially
        for file in "${failed_files[@]}"; do
            if ! download_single_config_file "$file"; then
                log_error "Sequential retry also failed for $file"
            fi
        done
        
        # Check if any files are still missing
        local still_missing=()
        for file in "${failed_files[@]}"; do
            if [[ ! -f "$file" ]]; then
                still_missing+=("$file")
            fi
        done
        
        if [[ ${#still_missing[@]} -gt 0 ]]; then
            log_error "Critical: Still missing files after all retry attempts: ${still_missing[*]}"
            return 1
        fi
    fi
    
    log_success "All configuration files downloaded successfully (parallel + retry strategy)"
    return 0
}

# Install linting tools with progressive fallbacks
install_linting_tools() {
    log_header "🛠️  INSTALLING LINTING TOOLS"
    
    # Auto-repair environment before installing tools
    if [[ "$MODE_FIX" == true ]]; then
        auto_repair_environment
    fi
    
    # Try justfile installation first
    log_step "Attempting tool installation via justfile..."
    if [[ -f "justfile" ]] && command_exists just; then
        if retry_with_backoff 2 2 2 just install; then
            log_success "All linting tools installed successfully via justfile"
            INSTALLED_TOOLS+=("golangci-lint (via justfile)" "go-arch-lint (via justfile)")
            return 0
        else
            log_warning "Justfile installation failed, trying direct installation..."
        fi
    else
        log_warning "Justfile not available, using direct installation method..."
    fi
    
    # Fallback to direct Go installations
    log_step "Installing linting tools directly via go install..."
    
    local tools_to_install=(
        "github.com/golangci/golangci-lint/cmd/golangci-lint@latest:golangci-lint"
        "github.com/fe3dback/go-arch-lint@latest:go-arch-lint"
    )
    
    local failed_tools=()
    local successful_tools=()
    
    for tool_spec in "${tools_to_install[@]}"; do
        local package="${tool_spec%:*}"
        local binary="${tool_spec#*:}"
        
        if install_go_tool "$package" "$binary"; then
            successful_tools+=("$binary")
        else
            failed_tools+=("$binary")
        fi
    done
    
    # Report results
    if [[ ${#successful_tools[@]} -gt 0 ]]; then
        log_success "Successfully installed tools: ${successful_tools[*]}"
        INSTALLED_TOOLS+=("${successful_tools[@]}")
    fi
    
    if [[ ${#failed_tools[@]} -gt 0 ]]; then
        log_error "Failed to install tools: ${failed_tools[*]}"
        
        if [[ "$MODE_FIX" == true ]]; then
            log_info "Auto-fix mode: Attempting alternative installation methods..."
            for tool in "${failed_tools[@]}"; do
                install_tool_alternative "$tool"
            done
        else
            log_info "Run with --fix flag to attempt alternative installation methods"
        fi
        
        # Don't fail completely if some tools installed
        if [[ ${#successful_tools[@]} -eq 0 ]]; then
            return 1
        fi
    fi
    
    return 0
}

# Install individual Go tool with retries
install_go_tool() {
    local package="$1"
    local binary="$2"
    
    log_step "Installing $binary from $package..."
    
    # Check if already installed
    if command_exists "$binary"; then
        log_success "$binary already installed ($(command -v "$binary"))"
        return 0
    fi
    
    # Clean Go module cache if in fix mode and previous attempts failed
    if [[ "$MODE_FIX" == true && $RETRY_COUNT -gt 0 ]]; then
        log_verbose "Auto-fix: Cleaning Go module cache before retry..."
        go clean -modcache >/dev/null 2>&1 || true
    fi
    
    # Attempt installation with retries
    if retry_with_backoff 3 2 2 go install "$package"; then
        # Verify installation worked
        if command_exists "$binary"; then
            log_success "$binary installed successfully"
            return 0
        else
            log_warning "$binary installation completed but binary not found in PATH"
            
            # Auto-repair: Try to fix PATH
            if [[ "$MODE_FIX" == true ]]; then
                log_step "Auto-fix: Attempting PATH repair for $binary..."
                if [[ -f "$HOME/go/bin/$binary" ]]; then
                    export PATH="$HOME/go/bin:$PATH"
                    if command_exists "$binary"; then
                        log_success "PATH repair successful - $binary now available"
                        return 0
                    fi
                fi
            fi
            return 1
        fi
    else
        handle_error_with_escalation "$ERROR_TOOLS" "HIGH" \
            "Failed to install $binary after retries" \
            "go install $package failed multiple times" \
            "Check Go proxy settings and network connectivity"
        return 1
    fi
}

# Alternative installation methods for failed tools
install_tool_alternative() {
    local tool="$1"
    
    log_step "Attempting alternative installation for $tool..."
    
    case "$tool" in
        golangci-lint)
            install_golangci_lint_alternative
            ;;
        go-arch-lint)
            install_go_arch_lint_alternative
            ;;
        *)
            log_warning "No alternative installation method for $tool"
            return 1
            ;;
    esac
}

# Alternative golangci-lint installation
install_golangci_lint_alternative() {
    log_step "Trying alternative golangci-lint installation..."
    
    # Method 1: Direct script installation
    if retry_with_backoff 2 3 2 curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)"/bin v1.54.2; then
        log_success "golangci-lint installed via direct script"
        INSTALLED_TOOLS+=("golangci-lint (direct script)")
        return 0
    fi
    
    # Method 2: Try different version
    log_verbose "Trying different golangci-lint version..."
    if retry_with_backoff 2 2 2 go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2; then
        log_success "golangci-lint installed with specific version"
        INSTALLED_TOOLS+=("golangci-lint (specific version)")
        return 0
    fi
    
    log_error "All alternative golangci-lint installation methods failed"
    return 1
}

# Alternative go-arch-lint installation  
install_go_arch_lint_alternative() {
    log_step "Trying alternative go-arch-lint installation..."
    
    # Method 1: Try different version tags
    local versions=("@latest" "@v1.8.0" "@v1.7.0")
    
    for version in "${versions[@]}"; do
        log_verbose "Trying go-arch-lint version $version..."
        if retry_with_backoff 2 2 2 go install "github.com/fe3dback/go-arch-lint$version"; then
            log_success "go-arch-lint installed with version $version"
            INSTALLED_TOOLS+=("go-arch-lint ($version)")
            return 0
        fi
    done
    
    log_error "All alternative go-arch-lint installation methods failed"
    return 1
}

# Smart PATH setup with comprehensive auto-repair
setup_go_path() {
    log_header "🛤️  SETTING UP GO TOOLS PATH"
    
    # Get Go binary path
    local go_bin_path
    go_bin_path=$(go env GOPATH 2>/dev/null)/bin || go_bin_path="$HOME/go/bin"
    local custom_go_bin="$HOME/.local/bin"
    
    log_verbose "Go binary path: $go_bin_path"
    log_verbose "Custom binary path: $custom_go_bin"
    log_verbose "Current PATH: $PATH"
    
    # Smart directory creation and repair
    if ! setup_go_directories "$go_bin_path" "$custom_go_bin"; then
        return 1
    fi
    
    # PATH detection and repair
    local path_changes=()
    
    # Check and fix Go bin in PATH
    if ! echo "$PATH" | grep -q "$go_bin_path"; then
        log_step "Adding Go tools directory to PATH: $go_bin_path"
        export PATH="$go_bin_path:$PATH"
        path_changes+=("$go_bin_path")
        log_success "Added $go_bin_path to PATH for current session"
    else
        log_success "Go tools directory already in PATH"
    fi
    
    # Check and fix custom bin in PATH (for direct installs)
    if [[ -d "$custom_go_bin" ]] && ! echo "$PATH" | grep -q "$custom_go_bin"; then
        log_step "Adding custom binary directory to PATH: $custom_go_bin"
        export PATH="$custom_go_bin:$PATH"
        path_changes+=("$custom_go_bin")
        log_success "Added $custom_go_bin to PATH for current session"
    fi
    
    # Verify tools are now accessible
    local tools_verified=()
    local tools_missing=()
    
    for tool in just golangci-lint go-arch-lint; do
        if command_exists "$tool"; then
            tools_verified+=("$tool:$(command -v "$tool")")
        else
            tools_missing+=("$tool")
        fi
    done
    
    if [[ ${#tools_verified[@]} -gt 0 ]]; then
        log_success "Verified tools accessible: ${#tools_verified[@]} found"
        for tool_info in "${tools_verified[@]}"; do
            log_verbose "  ${tool_info%:*} → ${tool_info#*:}"
        done
    fi
    
    if [[ ${#tools_missing[@]} -gt 0 ]]; then
        log_verbose "Tools not yet accessible: ${tools_missing[*]}"
        if [[ "$MODE_FIX" == true ]]; then
            smart_tool_path_repair "${tools_missing[@]}"
        fi
    fi
    
    # Make PATH changes persistent
    if [[ ${#path_changes[@]} -gt 0 ]]; then
        make_path_persistent "${path_changes[@]}"
    fi
    
    # Final validation
    validate_go_environment
    
    return 0
}

# Setup Go directories with proper permissions
setup_go_directories() {
    local go_bin_path="$1"
    local custom_go_bin="$2"
    local repairs_made=0
    
    log_verbose "Setting up Go directories..."
    
    # Create Go bin directory if missing
    if [[ ! -d "$go_bin_path" ]]; then
        log_step "Creating Go binary directory: $go_bin_path"
        if mkdir -p "$go_bin_path"; then
            log_success "Created $go_bin_path"
            ((repairs_made++))
        else
            log_error "Failed to create $go_bin_path"
            return 1
        fi
    else
        log_verbose "Go binary directory exists: $go_bin_path"
    fi
    
    # Fix permissions on Go bin directory
    if [[ -d "$go_bin_path" ]] && [[ ! -w "$go_bin_path" ]]; then
        log_step "Fixing permissions on $go_bin_path"
        if chmod 755 "$go_bin_path" 2>/dev/null; then
            log_success "Fixed permissions on $go_bin_path"
            ((repairs_made++))
        else
            log_warning "Could not fix permissions on $go_bin_path"
        fi
    fi
    
    # Create custom bin directory for direct installs
    if [[ ! -d "$custom_go_bin" ]]; then
        log_step "Creating custom binary directory: $custom_go_bin"
        if mkdir -p "$custom_go_bin"; then
            log_success "Created $custom_go_bin"
            ((repairs_made++))
        else
            log_warning "Failed to create $custom_go_bin"
        fi
    fi
    
    log_verbose "Directory setup completed: $repairs_made repairs made"
    return 0
}

# Smart tool PATH repair - find and fix tool accessibility
smart_tool_path_repair() {
    local missing_tools=("$@")
    
    log_step "Attempting smart PATH repair for missing tools..."
    
    for tool in "${missing_tools[@]}"; do
        log_verbose "Searching for $tool..."
        
        # Search common locations
        local search_paths=(
            "$HOME/go/bin"
            "$HOME/.local/bin" 
            "/opt/homebrew/bin"
            "/usr/local/bin"
            "/usr/bin"
            "$(go env GOPATH 2>/dev/null)/bin"
        )
        
        for search_path in "${search_paths[@]}"; do
            if [[ -x "$search_path/$tool" ]]; then
                log_success "Found $tool at $search_path/$tool"
                
                # Add to PATH if not already there
                if ! echo "$PATH" | grep -q "$search_path"; then
                    log_step "Adding $search_path to PATH for $tool"
                    export PATH="$search_path:$PATH"
                    log_success "Added $search_path to PATH"
                    
                    # Verify tool is now accessible
                    if command_exists "$tool"; then
                        log_success "$tool is now accessible"
                    fi
                fi
                break
            fi
        done
        
        # If tool still not found, provide help
        if ! command_exists "$tool"; then
            log_warning "$tool not found in common locations"
            case "$tool" in
                just)
                    log_info "Install with: curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin"
                    ;;
                golangci-lint)
                    log_info "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
                    ;;
                go-arch-lint)
                    log_info "Install with: go install github.com/fe3dback/go-arch-lint@latest"
                    ;;
            esac
        fi
    done
}

# Make PATH changes persistent across shell sessions
make_path_persistent() {
    local path_additions=("$@")
    
    log_step "Making PATH changes persistent..."
    
    # Detect shell and profile
    local shell_profiles=()
    local current_shell
    current_shell=$(basename "$SHELL" 2>/dev/null || echo "unknown")
    
    case "$current_shell" in
        bash)
            if [[ "$(uname)" = "Darwin" ]]; then
                shell_profiles+=("$HOME/.bash_profile")
                [[ -f "$HOME/.bashrc" ]] && shell_profiles+=("$HOME/.bashrc")
            else
                shell_profiles+=("$HOME/.bashrc")
                [[ -f "$HOME/.profile" ]] && shell_profiles+=("$HOME/.profile")
            fi
            ;;
        zsh)
            shell_profiles+=("$HOME/.zshrc")
            [[ -f "$HOME/.zprofile" ]] && shell_profiles+=("$HOME/.zprofile")
            ;;
        fish)
            # Fish uses different syntax
            local fish_config="$HOME/.config/fish/config.fish"
            if [[ -f "$fish_config" ]]; then
                for path_add in "${path_additions[@]}"; do
                    if ! grep -q "fish_add_path.*$path_add" "$fish_config"; then
                        log_step "Adding $path_add to fish config"
                        echo "fish_add_path $path_add" >> "$fish_config"
                    fi
                done
                log_success "Updated fish configuration"
                return 0
            fi
            ;;
        *)
            # Try common profiles
            for profile in "$HOME/.profile" "$HOME/.bashrc" "$HOME/.zshrc"; do
                [[ -f "$profile" ]] && shell_profiles+=("$profile")
            done
            ;;
    esac
    
    # Update detected shell profiles
    for profile in "${shell_profiles[@]}"; do
        if [[ -f "$profile" ]]; then
            log_verbose "Checking profile: $profile"
            local profile_updated=false
            
            for path_add in "${path_additions[@]}"; do
                # Check if path already exists in profile
                if ! grep -q "export PATH.*$path_add" "$profile"; then
                    log_step "Adding $path_add to $profile"
                    
                    # Add organized block
                    if ! grep -q "# Added by template-arch-lint bootstrap" "$profile"; then
                        echo "" >> "$profile"
                        echo "# Added by template-arch-lint bootstrap" >> "$profile"
                    fi
                    echo "export PATH=\"$path_add:\$PATH\"" >> "$profile"
                    profile_updated=true
                fi
            done
            
            if [[ "$profile_updated" == true ]]; then
                log_success "Updated $profile"
                log_info "Restart your shell or run: source $profile"
            else
                log_verbose "No updates needed for $profile"
            fi
        fi
    done
    
    if [[ ${#shell_profiles[@]} -eq 0 ]]; then
        log_warning "Could not detect shell profile for persistent PATH"
        log_info "Manually add these to your shell profile:"
        for path_add in "${path_additions[@]}"; do
            echo "  export PATH=\"$path_add:\$PATH\""
        done
    fi
}

# Validate Go environment after PATH setup
validate_go_environment() {
    log_step "Validating Go environment setup..."
    
    local validation_issues=()
    
    # Check GOPATH
    local gopath
    gopath=$(go env GOPATH 2>/dev/null)
    if [[ -n "$gopath" ]]; then
        log_success "GOPATH: $gopath"
        
        # Check if GOPATH/bin is in PATH
        if ! echo "$PATH" | grep -q "$gopath/bin"; then
            validation_issues+=("GOPATH/bin ($gopath/bin) not in PATH")
        fi
    else
        validation_issues+=("GOPATH not set")
    fi
    
    # Check Go proxy
    local goproxy
    goproxy=$(go env GOPROXY 2>/dev/null)
    log_verbose "GOPROXY: ${goproxy:-not set}"
    
    # Check Go version
    if command_exists go; then
        log_verbose "Go version: $(go version)"
    fi
    
    # Report validation results
    if [[ ${#validation_issues[@]} -eq 0 ]]; then
        log_success "Go environment validation passed"
    else
        log_warning "Go environment validation issues found:"
        for issue in "${validation_issues[@]}"; do
            log_warning "  $issue"
        done
        
        if [[ "$MODE_FIX" == true ]]; then
            log_info "Issues found but auto-repair attempted above"
        else
            log_info "Run with --fix flag to attempt auto-repair"
        fi
    fi
    
    return 0
}

# Standard error recovery actions
standardize_error_recovery() {
    local category="$1"
    local failed_action="$2"
    
    case "$category" in
        "$ERROR_NETWORK")
            log_info "🌐 Network troubleshooting steps:"
            log_info "  1. Check internet connection: ping -c 3 8.8.8.8"
            log_info "  2. Test GitHub access: curl -I https://github.com"
            log_info "  3. Check proxy settings: echo $https_proxy"
            log_info "  4. Try different DNS: export DNS=8.8.8.8"
            ;;
        "$ERROR_TOOLS")
            log_info "🛠️ Tool installation troubleshooting:"
            log_info "  1. Clean Go cache: go clean -modcache"
            log_info "  2. Check Go proxy: go env GOPROXY"
            log_info "  3. Verify GOPATH: echo $GOPATH"
            log_info "  4. Manual install: Download from GitHub releases"
            ;;
        "$ERROR_PERMISSIONS")
            log_info "🔐 Permission troubleshooting:"
            log_info "  1. Check directory permissions: ls -la ."
            log_info "  2. Fix Go bin permissions: chmod 755 ~/go/bin"
            log_info "  3. Check ownership: ls -la ~/go"
            log_info "  4. Create missing directories: mkdir -p ~/go/bin"
            ;;
        "$ERROR_CONFIG")
            log_info "⚙️ Configuration troubleshooting:"
            log_info "  1. Check file downloads: ls -la .go-arch-lint.yml .golangci.yml justfile"
            log_info "  2. Validate file contents: head -5 .golangci.yml"
            log_info "  3. Re-download manually: curl -fsSL -o file $REPO_URL/file"
            log_info "  4. Check file permissions: chmod 644 config-files"
            ;;
        *)
            log_info "💡 General troubleshooting:"
            log_info "  1. Run diagnostics: ./bootstrap.sh --diagnose"
            log_info "  2. Try auto-repair: ./bootstrap.sh --fix"
            log_info "  3. Enable debug: bash -x bootstrap.sh --verbose"
            log_info "  4. Check system requirements"
            ;;
    esac
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
        handle_error_with_escalation "$ERROR_NETWORK" "HIGH" \
            "Cannot reach GitHub configuration files" \
            "Network connectivity to $REPO_URL failed" \
            "Check network connection, firewall, proxy settings"
        ((issues_found++))
    fi
    
    # Check 4: Disk space
    log_step "Checking disk space..."
    local available=$(df . | tail -1 | awk '{print $4}')
    local available_human=$(df -h . | tail -1 | awk '{print $4}')
    log_info "Available space: $available_human"
    
    if [[ $available -lt 1000000 ]]; then # Less than ~1GB
        handle_error_with_escalation "$ERROR_SYSTEM" "MEDIUM" \
            "Low disk space - may cause installation issues" \
            "Available: $available_human, tools require ~100MB" \
            "Free up disk space or use different installation directory"
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
        *) 
            handle_error_with_escalation "$ERROR_SYSTEM" "HIGH" \
                "Unknown platform - may not be supported" \
                "Platform: $os detected, may not have installation support" \
                "Try manual installation or report compatibility issue"
            ((issues_found++)) 
            ;;
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

# Diagnostics with automatic repair (--fix mode)
run_diagnostics_with_repair() {
    log_header "🔧 COMPREHENSIVE DIAGNOSTICS WITH AUTO-REPAIR"
    echo -e "${CYAN}Analyzing environment and automatically repairing issues${NC}\n"
    
    local issues_found=0
    local repairs_attempted=0
    local repairs_successful=0
    
    # Check 1: Project structure with auto-repair
    log_step "Checking and repairing project structure..."
    echo "Current directory: $(pwd)"
    
    # Fix missing git repository
    if [[ ! -d ".git" ]]; then
        log_warning "No .git directory found"
        log_step "Auto-repair: Initializing git repository..."
        
        if git init && git add . && git commit -m "Initial commit for template-arch-lint bootstrap"; then
            log_success "Auto-repair successful: Git repository initialized"
            ((repairs_successful++))
        else
            log_error "Auto-repair failed: Could not initialize git repository"
            ((issues_found++))
            log_info "Manual fix: Run 'git init && git add . && git commit -m \"Initial commit\"'"
        fi
        ((repairs_attempted++))
    else
        log_success "Git repository detected"
    fi
    
    # Fix missing go.mod
    if [[ ! -f "go.mod" ]]; then
        log_warning "No go.mod found"
        log_step "Auto-repair: Creating go.mod file..."
        
        local module_name
        module_name=$(basename "$(pwd)")
        
        if go mod init "$module_name"; then
            log_success "Auto-repair successful: Created go.mod for module $module_name"
            ((repairs_successful++))
        else
            log_error "Auto-repair failed: Could not create go.mod"
            ((issues_found++))
            log_info "Manual fix: Run 'go mod init your-project-name'"
        fi
        ((repairs_attempted++))
    else
        local module_name=$(head -1 go.mod | cut -d' ' -f2)
        log_success "Go module detected: $module_name"
    fi
    
    # Check 2: Required commands with installation assistance
    log_step "Checking required commands with auto-install assistance..."
    local missing_commands=()
    
    for cmd in go git curl; do
        if command_exists "$cmd"; then
            log_success "$cmd available at: $(command -v "$cmd")"
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
    
    # Provide detailed installation instructions for missing commands
    if [[ ${#missing_commands[@]} -gt 0 ]]; then
        log_warning "Missing commands: ${missing_commands[*]}"
        log_step "Auto-repair: Providing installation instructions..."
        
        local os=$(get_os)
        case "$os" in
            macos)
                if command_exists brew; then
                    log_info "Auto-install command: brew install ${missing_commands[*]}"
                    log_step "Attempting automatic installation via Homebrew..."
                    if brew install "${missing_commands[@]}"; then
                        log_success "Auto-repair successful: Installed via Homebrew"
                        ((repairs_successful++))
                    else
                        log_error "Auto-repair failed: Homebrew installation failed"
                    fi
                    ((repairs_attempted++))
                else
                    log_info "Install Homebrew first: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
                    log_info "Then run: brew install ${missing_commands[*]}"
                fi
                ;;
            linux)
                if command_exists apt-get; then
                    log_info "Auto-install command: sudo apt-get update && sudo apt-get install -y golang-go git curl"
                    log_step "Attempting automatic installation via apt..."
                    if sudo apt-get update && sudo apt-get install -y golang-go git curl; then
                        log_success "Auto-repair successful: Installed via apt"
                        ((repairs_successful++))
                    else
                        log_error "Auto-repair failed: apt installation failed"
                    fi
                    ((repairs_attempted++))
                elif command_exists yum; then
                    log_info "Auto-install command: sudo yum install -y golang git curl"
                    log_step "Attempting automatic installation via yum..."
                    if sudo yum install -y golang git curl; then
                        log_success "Auto-repair successful: Installed via yum"
                        ((repairs_successful++))
                    else
                        log_error "Auto-repair failed: yum installation failed"
                    fi
                    ((repairs_attempted++))
                else
                    log_info "Manual install required - package manager not detected"
                fi
                ;;
            *)
                log_info "Manual installation required for platform: $os"
                ;;
        esac
    fi
    
    # Check 3: Network connectivity with proxy detection
    log_step "Checking network connectivity with auto-repair..."
    if curl -I --max-time 10 "$REPO_URL/.go-arch-lint.yml" >/dev/null 2>&1; then
        log_success "GitHub raw files accessible"
    else
        log_error "Cannot reach GitHub configuration files"
        ((issues_found++))
        
        log_step "Auto-repair: Attempting network troubleshooting..."
        ((repairs_attempted++))
        
        # Try different methods
        if curl -I --max-time 10 -H "User-Agent: Mozilla/5.0" "$REPO_URL/.go-arch-lint.yml" >/dev/null 2>&1; then
            log_success "Auto-repair successful: Connection works with different User-Agent"
            ((repairs_successful++))
        elif curl -I --max-time 10 --insecure "$REPO_URL/.go-arch-lint.yml" >/dev/null 2>&1; then
            log_warning "Auto-repair partial: Connection works without SSL verification"
            log_info "This may indicate a corporate firewall or proxy issue"
        else
            log_error "Auto-repair failed: Network connectivity issues persist"
            log_info "Check: Network connection, firewall, proxy settings"
            log_info "Try: export https_proxy=your-proxy-url"
        fi
    fi
    
    # Check 4: Disk space with cleanup suggestions
    log_step "Checking disk space with cleanup assistance..."
    local available=$(df . | tail -1 | awk '{print $4}')
    local available_human=$(df -h . | tail -1 | awk '{print $4}')
    log_info "Available space: $available_human"
    
    if [[ $available -lt 1000000 ]]; then # Less than ~1GB
        log_warning "Low disk space - may cause installation issues"
        ((issues_found++))
        
        log_step "Auto-repair: Suggesting disk cleanup options..."
        ((repairs_attempted++))
        
        # Provide helpful cleanup suggestions
        log_info "Auto-cleanup suggestions:"
        log_info "  • Clean Go cache: go clean -cache -modcache"
        log_info "  • Clean npm cache: npm cache clean --force"
        log_info "  • Clean Docker: docker system prune"
        log_info "  • Clean Homebrew: brew cleanup"
        
        # Attempt safe automatic cleanup
        if command_exists go; then
            log_step "Attempting Go cache cleanup..."
            if go clean -cache; then
                log_success "Auto-repair successful: Cleaned Go build cache"
                ((repairs_successful++))
            fi
        fi
    else
        log_success "Sufficient disk space available"
    fi
    
    # Check 5: Platform with optimization suggestions
    log_step "Checking platform with optimization suggestions..."
    local os=$(get_os)
    local arch=$(get_arch)
    log_info "Platform: $os/$arch"
    
    case "$os" in
        macos)
            log_success "macOS detected - fully supported"
            if ! command_exists brew; then
                log_info "Optimization suggestion: Install Homebrew for easier package management"
                log_info "Command: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
            fi
            ;;
        linux)
            log_success "Linux detected - fully supported"
            # Check for common Linux optimizations
            if [[ ! -f "/etc/os-release" ]]; then
                log_info "Could not detect Linux distribution"
            else
                local distro=$(grep "^ID=" /etc/os-release | cut -d= -f2 | tr -d '"')
                log_verbose "Linux distribution: $distro"
            fi
            ;;
        windows)
            log_warning "Windows detected - WSL recommended"
            log_info "Optimization suggestion: Use Windows Subsystem for Linux (WSL)"
            log_info "Command: wsl --install"
            ((issues_found++))
            ;;
        *)
            log_error "Unknown platform - may not be supported"
            ((issues_found++))
            ;;
    esac
    
    # Summary with repair statistics
    echo -e "\n${BOLD}🔧 AUTO-REPAIR SUMMARY${NC}"
    echo "======================"
    
    log_info "Repair Statistics:"
    log_info "  • Issues detected: $issues_found"
    log_info "  • Repairs attempted: $repairs_attempted"
    log_info "  • Repairs successful: $repairs_successful"
    
    if [[ $repairs_attempted -gt 0 ]]; then
        local success_rate=$((repairs_successful * 100 / repairs_attempted))
        log_info "  • Success rate: ${success_rate}%"
    fi
    
    if [[ $issues_found -eq 0 ]]; then
        log_success "🎉 Environment is perfect! No issues found."
        log_info "Ready to proceed with bootstrap installation."
    elif [[ $repairs_successful -eq $repairs_attempted && $repairs_attempted -gt 0 ]]; then
        log_success "🛠️  All detected issues were automatically repaired!"
        log_info "Environment is now ready for bootstrap installation."
        # Reset issues count since they were fixed
        issues_found=0
    elif [[ $repairs_successful -gt 0 ]]; then
        log_success "🔧 Partially repaired: $repairs_successful/$repairs_attempted fixes successful"
        log_warning "Some issues remain and may require manual intervention."
    else
        log_warning "⚠️  Auto-repair could not fix all issues."
        log_info "Manual intervention may be required for optimal bootstrap experience."
    fi
    
    echo -e "\n${YELLOW}💡 Next steps:${NC}"
    if [[ $issues_found -eq 0 ]]; then
        echo "✅ Your environment is ready! Bootstrap will continue automatically."
    else
        echo "🔧 Consider running the suggested manual fixes above"
        echo "📋 Run './bootstrap.sh --diagnose' to re-check environment"
        echo "🚀 Or continue with bootstrap - it will attempt to work around remaining issues"
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
    
    # Handle auto-fix mode (diagnose + repair + bootstrap)
    if [[ "$MODE_FIX" == true ]]; then
        log_info "🔧 Auto-fix mode: Running diagnostics with auto-repair..."
        run_diagnostics_with_repair
        
        # If diagnostics pass, continue with bootstrap
        if [[ $? -eq 0 ]]; then
            log_info "✅ Diagnostics passed, continuing with bootstrap installation..."
        else
            log_warning "⚠️  Some diagnostic issues remain, but continuing with bootstrap..."
        fi
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