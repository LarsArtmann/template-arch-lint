#!/bin/bash

# Validate Kubernetes manifests for template-arch-lint

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$(dirname "$SCRIPT_DIR")"
ENVIRONMENT="dev"
VERBOSE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Help function
show_help() {
    cat << EOF
Validate Kubernetes manifests for template-arch-lint

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -e, --environment ENV    Target environment (dev|staging|prod|all) [default: dev]
    -v, --verbose           Enable verbose output
    --dry-run               Perform dry-run validation against cluster
    --lint                  Run additional linting checks
    -h, --help              Show this help message

EXAMPLES:
    # Validate development environment
    $0

    # Validate all environments
    $0 --environment all

    # Validate with dry-run against cluster
    $0 --dry-run

    # Run with additional linting
    $0 --lint

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --lint)
            LINT=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Enable verbose output if requested
if [[ "$VERBOSE" == "true" ]]; then
    set -x
fi

# Validate YAML syntax
validate_yaml_syntax() {
    local file="$1"
    
    if command -v yamllint &> /dev/null; then
        if yamllint "$file" &> /dev/null; then
            log_success "✓ YAML syntax valid: $(basename "$file")"
            return 0
        else
            log_error "✗ YAML syntax error in: $(basename "$file")"
            yamllint "$file"
            return 1
        fi
    else
        # Basic YAML validation using yq if available
        if command -v yq &> /dev/null; then
            if yq eval '.' "$file" &> /dev/null; then
                log_success "✓ YAML syntax valid: $(basename "$file")"
                return 0
            else
                log_error "✗ YAML syntax error in: $(basename "$file")"
                return 1
            fi
        else
            log_warn "No YAML validator found (yamllint or yq), skipping syntax validation"
            return 0
        fi
    fi
}

# Validate Kubernetes manifests
validate_k8s_manifests() {
    local env_dir="$1"
    local env_name="$(basename "$env_dir")"
    
    log_info "Validating Kubernetes manifests for environment: $env_name"
    
    # Check if environment directory exists
    if [[ ! -d "$env_dir" ]]; then
        log_error "Environment directory not found: $env_dir"
        return 1
    fi
    
    # Validate using kustomize
    if ! kubectl kustomize "$env_dir" &> /dev/null; then
        log_error "✗ Kustomize validation failed for: $env_name"
        kubectl kustomize "$env_dir"
        return 1
    else
        log_success "✓ Kustomize validation passed for: $env_name"
    fi
    
    # Validate individual YAML files in base directory
    local base_files=("$K8S_DIR/base"/*.yaml)
    local errors=0
    
    for file in "${base_files[@]}"; do
        if [[ -f "$file" ]]; then
            if ! validate_yaml_syntax "$file"; then
                ((errors++))
            fi
        fi
    done
    
    return $errors
}

# Validate against cluster (dry-run)
validate_against_cluster() {
    local env_dir="$1"
    local env_name="$(basename "$env_dir")"
    
    log_info "Performing dry-run validation against cluster for: $env_name"
    
    # Check if cluster is accessible
    if ! kubectl cluster-info &> /dev/null; then
        log_warn "Cannot connect to Kubernetes cluster, skipping dry-run validation"
        return 0
    fi
    
    # Generate manifests and validate
    local temp_file=$(mktemp)
    kubectl kustomize "$env_dir" > "$temp_file"
    
    if kubectl apply --dry-run=server -f "$temp_file" &> /dev/null; then
        log_success "✓ Dry-run validation passed for: $env_name"
        rm "$temp_file"
        return 0
    else
        log_error "✗ Dry-run validation failed for: $env_name"
        kubectl apply --dry-run=server -f "$temp_file"
        rm "$temp_file"
        return 1
    fi
}

# Run additional linting checks
run_linting_checks() {
    local env_dir="$1"
    local env_name="$(basename "$env_dir")"
    
    log_info "Running additional linting checks for: $env_name"
    
    # Generate manifests for analysis
    local temp_file=$(mktemp)
    kubectl kustomize "$env_dir" > "$temp_file"
    
    local issues=0
    
    # Check for common issues
    log_info "Checking for common security and best practice issues..."
    
    # Check for latest tags (should be avoided in production)
    if grep -q 'image.*:latest' "$temp_file"; then
        log_warn "⚠ Found 'latest' image tags (consider using specific versions)"
        ((issues++))
    fi
    
    # Check for missing resource limits
    if ! grep -q 'resources:' "$temp_file"; then
        log_warn "⚠ No resource limits/requests found"
        ((issues++))
    fi
    
    # Check for missing liveness/readiness probes
    if ! grep -q 'livenessProbe:' "$temp_file"; then
        log_warn "⚠ No liveness probes found"
        ((issues++))
    fi
    
    if ! grep -q 'readinessProbe:' "$temp_file"; then
        log_warn "⚠ No readiness probes found"
        ((issues++))
    fi
    
    # Check for missing security context
    if ! grep -q 'securityContext:' "$temp_file"; then
        log_warn "⚠ No security context found"
        ((issues++))
    fi
    
    # Check for privileged containers
    if grep -q 'privileged: true' "$temp_file"; then
        log_error "✗ Privileged containers found (security risk)"
        ((issues++))
    fi
    
    # Check for hostNetwork usage
    if grep -q 'hostNetwork: true' "$temp_file"; then
        log_warn "⚠ hostNetwork usage found (potential security risk)"
        ((issues++))
    fi
    
    # Check for proper labels
    if ! grep -q 'app.kubernetes.io/name' "$temp_file"; then
        log_warn "⚠ Missing recommended labels (app.kubernetes.io/name)"
        ((issues++))
    fi
    
    rm "$temp_file"
    
    if [[ $issues -eq 0 ]]; then
        log_success "✓ No linting issues found for: $env_name"
    else
        log_warn "Found $issues linting issues for: $env_name"
    fi
    
    return $issues
}

# Validate environment
validate_environment() {
    local env="$1"
    
    case $env in
        dev|staging|prod)
            local env_dir="$K8S_DIR/environments/$env"
            local total_errors=0
            
            # Basic manifest validation
            if ! validate_k8s_manifests "$env_dir"; then
                ((total_errors++))
            fi
            
            # Dry-run validation if requested
            if [[ "${DRY_RUN:-false}" == "true" ]]; then
                if ! validate_against_cluster "$env_dir"; then
                    ((total_errors++))
                fi
            fi
            
            # Linting checks if requested
            if [[ "${LINT:-false}" == "true" ]]; then
                if ! run_linting_checks "$env_dir"; then
                    ((total_errors++))
                fi
            fi
            
            return $total_errors
            ;;
        all)
            local total_errors=0
            for env in dev staging prod; do
                log_info "Validating environment: $env"
                if ! validate_environment "$env"; then
                    ((total_errors++))
                fi
                echo
            done
            return $total_errors
            ;;
        *)
            log_error "Invalid environment: $env. Must be one of: dev, staging, prod, all"
            return 1
            ;;
    esac
}

# Main execution
main() {
    log_info "Template Arch Lint - Kubernetes Manifest Validation"
    log_info "===================================================="
    echo
    
    # Check prerequisites
    log_info "Checking prerequisites..."
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check if kustomize is available
    if ! kubectl kustomize --help &> /dev/null; then
        log_error "kustomize is not available (should be bundled with kubectl 1.14+)"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
    echo
    
    # Run validation
    if validate_environment "$ENVIRONMENT"; then
        log_success "All validations passed!"
        exit 0
    else
        log_error "Validation failed with errors"
        exit 1
    fi
}

# Run main function
main "$@"