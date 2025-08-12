#!/bin/bash

# Cleanup template-arch-lint deployment from local Kubernetes cluster

set -euo pipefail

# Configuration
ENVIRONMENT="dev"
NAMESPACE=""
FORCE=false
CLUSTER_TYPE="kind"

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
Cleanup template-arch-lint deployment from local Kubernetes cluster

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -e, --environment ENV    Target environment (dev|staging|prod|all) [default: dev]
    -n, --namespace NS       Specific namespace to cleanup
    -f, --force              Skip confirmation prompts
    -t, --cluster-type TYPE  Cluster type (kind|minikube) [default: kind]
    --destroy-cluster        Destroy the entire cluster
    -h, --help               Show this help message

EXAMPLES:
    # Cleanup development environment
    $0

    # Cleanup all environments
    $0 --environment all

    # Cleanup specific namespace
    $0 --namespace my-namespace

    # Force cleanup without confirmation
    $0 --force

    # Destroy the entire kind cluster
    $0 --destroy-cluster

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -n|--namespace)
            NAMESPACE="$2"
            shift 2
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -t|--cluster-type)
            CLUSTER_TYPE="$2"
            shift
            ;;
        --destroy-cluster)
            DESTROY_CLUSTER=true
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

# Confirm action unless force flag is set
confirm_action() {
    local message="$1"
    
    if [[ "$FORCE" == "true" ]]; then
        return 0
    fi
    
    echo -n -e "${YELLOW}$message${NC} [y/N]: "
    read -r response
    case $response in
        [yY][eE][sS]|[yY])
            return 0
            ;;
        *)
            log_info "Aborted by user"
            exit 0
            ;;
    esac
}

# Cleanup namespace
cleanup_namespace() {
    local ns="$1"
    
    log_info "Checking namespace: $ns"
    
    if ! kubectl get namespace "$ns" &> /dev/null; then
        log_warn "Namespace $ns does not exist"
        return
    fi
    
    confirm_action "Are you sure you want to delete namespace '$ns' and all its resources?"
    
    log_info "Deleting namespace: $ns"
    kubectl delete namespace "$ns" --timeout=300s
    
    log_success "Namespace $ns deleted successfully"
}

# Cleanup by environment
cleanup_environment() {
    local env="$1"
    local namespaces=()
    
    case $env in
        dev)
            namespaces=("template-arch-lint-dev")
            ;;
        staging)
            namespaces=("template-arch-lint-staging")
            ;;
        prod)
            namespaces=("template-arch-lint-prod")
            ;;
        all)
            namespaces=("template-arch-lint-dev" "template-arch-lint-staging" "template-arch-lint-prod")
            ;;
        *)
            log_error "Invalid environment: $env. Must be one of: dev, staging, prod, all"
            exit 1
            ;;
    esac
    
    for ns in "${namespaces[@]}"; do
        cleanup_namespace "$ns"
    done
}

# Destroy entire cluster
destroy_cluster() {
    confirm_action "Are you sure you want to destroy the entire $CLUSTER_TYPE cluster?"
    
    case $CLUSTER_TYPE in
        kind)
            log_info "Destroying kind cluster..."
            kind delete cluster
            ;;
        minikube)
            log_info "Destroying minikube cluster..."
            minikube delete
            ;;
        *)
            log_error "Unsupported cluster type: $CLUSTER_TYPE"
            exit 1
            ;;
    esac
    
    log_success "Cluster destroyed successfully"
}

# Show current status
show_current_status() {
    log_info "Current Kubernetes Resources:"
    echo
    
    log_info "Namespaces related to template-arch-lint:"
    kubectl get namespaces | grep -E "(template-arch-lint|NAME)" || log_warn "No template-arch-lint namespaces found"
    echo
    
    log_info "All namespaces:"
    kubectl get namespaces
    echo
}

# Main execution
main() {
    log_info "Template Arch Lint - Kubernetes Cleanup"
    log_info "========================================"
    
    # Check if kubectl is available
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check if cluster is accessible
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Cannot connect to Kubernetes cluster"
        exit 1
    fi
    
    show_current_status
    
    if [[ "${DESTROY_CLUSTER:-false}" == "true" ]]; then
        destroy_cluster
        exit 0
    fi
    
    if [[ -n "$NAMESPACE" ]]; then
        cleanup_namespace "$NAMESPACE"
    else
        cleanup_environment "$ENVIRONMENT"
    fi
    
    log_success "Cleanup completed!"
}

# Run main function
main "$@"