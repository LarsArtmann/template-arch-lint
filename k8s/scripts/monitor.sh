#!/bin/bash

# Monitor template-arch-lint deployment in Kubernetes

set -euo pipefail

# Configuration
ENVIRONMENT="dev"
NAMESPACE=""
WATCH=false
FOLLOW_LOGS=false
SHOW_METRICS=false

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
Monitor template-arch-lint deployment in Kubernetes

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -e, --environment ENV    Target environment (dev|staging|prod) [default: dev]
    -n, --namespace NS       Specific namespace to monitor
    -w, --watch              Watch resources in real-time
    -l, --logs               Follow application logs
    -m, --metrics            Show metrics endpoints
    --health                 Check application health
    --events                 Show recent events
    --describe               Describe all resources
    -h, --help               Show this help message

EXAMPLES:
    # Monitor development environment
    $0

    # Watch resources in real-time
    $0 --watch

    # Follow application logs
    $0 --logs

    # Check application health
    $0 --health

    # Show metrics endpoints
    $0 --metrics

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
        -w|--watch)
            WATCH=true
            shift
            ;;
        -l|--logs)
            FOLLOW_LOGS=true
            shift
            ;;
        -m|--metrics)
            SHOW_METRICS=true
            shift
            ;;
        --health)
            CHECK_HEALTH=true
            shift
            ;;
        --events)
            SHOW_EVENTS=true
            shift
            ;;
        --describe)
            DESCRIBE_RESOURCES=true
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

# Set namespace based on environment
if [[ -z "$NAMESPACE" ]]; then
    case $ENVIRONMENT in
        dev) NAMESPACE="template-arch-lint-dev" ;;
        staging) NAMESPACE="template-arch-lint-staging" ;;
        prod) NAMESPACE="template-arch-lint-prod" ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT. Must be one of: dev, staging, prod"
            exit 1
            ;;
    esac
fi

# Check if namespace exists
check_namespace() {
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_error "Namespace $NAMESPACE does not exist"
        exit 1
    fi
}

# Show basic status
show_status() {
    log_info "Template Arch Lint - Status ($NAMESPACE)"
    log_info "=========================================="
    echo
    
    log_info "Namespace:"
    kubectl get namespace "$NAMESPACE" -o wide
    echo
    
    log_info "Pods:"
    kubectl get pods -n "$NAMESPACE" -o wide
    echo
    
    log_info "Services:"
    kubectl get services -n "$NAMESPACE" -o wide
    echo
    
    log_info "Ingress:"
    kubectl get ingress -n "$NAMESPACE" -o wide
    echo
    
    log_info "Deployments:"
    kubectl get deployments -n "$NAMESPACE" -o wide
    echo
    
    log_info "HPA:"
    kubectl get hpa -n "$NAMESPACE" -o wide 2>/dev/null || log_warn "No HPA found"
    echo
    
    log_info "PVC:"
    kubectl get pvc -n "$NAMESPACE" -o wide 2>/dev/null || log_warn "No PVC found"
    echo
}

# Watch resources
watch_resources() {
    log_info "Watching resources in namespace: $NAMESPACE"
    log_info "Press Ctrl+C to exit"
    echo
    
    kubectl get pods,services,ingress,deployments -n "$NAMESPACE" --watch
}

# Follow logs
follow_logs() {
    log_info "Following logs for template-arch-lint in namespace: $NAMESPACE"
    log_info "Press Ctrl+C to exit"
    echo
    
    # Get the deployment name
    local deployment=$(kubectl get deployment -n "$NAMESPACE" -l app.kubernetes.io/name=template-arch-lint -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    
    if [[ -z "$deployment" ]]; then
        log_error "No template-arch-lint deployment found in namespace $NAMESPACE"
        exit 1
    fi
    
    kubectl logs -f deployment/"$deployment" -n "$NAMESPACE"
}

# Check application health
check_health() {
    log_info "Checking application health in namespace: $NAMESPACE"
    echo
    
    # Get pod names
    local pods=$(kubectl get pods -n "$NAMESPACE" -l app.kubernetes.io/name=template-arch-lint -o jsonpath='{.items[*].metadata.name}')
    
    if [[ -z "$pods" ]]; then
        log_error "No template-arch-lint pods found in namespace $NAMESPACE"
        exit 1
    fi
    
    for pod in $pods; do
        log_info "Health check for pod: $pod"
        
        # Check pod status
        local status=$(kubectl get pod "$pod" -n "$NAMESPACE" -o jsonpath='{.status.phase}')
        log_info "  Pod status: $status"
        
        # Check readiness
        local ready=$(kubectl get pod "$pod" -n "$NAMESPACE" -o jsonpath='{.status.conditions[?(@.type=="Ready")].status}')
        log_info "  Ready: $ready"
        
        # Try health endpoints
        log_info "  Testing health endpoints:"
        
        if kubectl exec -n "$NAMESPACE" "$pod" -- wget -qO- http://localhost:8080/health/live 2>/dev/null; then
            log_success "  ✓ Liveness check passed"
        else
            log_error "  ✗ Liveness check failed"
        fi
        
        if kubectl exec -n "$NAMESPACE" "$pod" -- wget -qO- http://localhost:8080/health/ready 2>/dev/null; then
            log_success "  ✓ Readiness check passed"
        else
            log_error "  ✗ Readiness check failed"
        fi
        
        echo
    done
}

# Show metrics endpoints
show_metrics() {
    log_info "Metrics endpoints for template-arch-lint in namespace: $NAMESPACE"
    echo
    
    # Get service names
    local service=$(kubectl get service -n "$NAMESPACE" -l app.kubernetes.io/name=template-arch-lint -o jsonpath='{.items[0].metadata.name}')
    
    if [[ -z "$service" ]]; then
        log_error "No template-arch-lint service found in namespace $NAMESPACE"
        exit 1
    fi
    
    log_info "Service: $service"
    
    # Port forward to metrics port
    log_info "Setting up port forwarding to metrics endpoint..."
    log_info "Metrics will be available at: http://localhost:2112/metrics"
    log_info "Press Ctrl+C to exit"
    
    kubectl port-forward -n "$NAMESPACE" service/"$service" 2112:2112
}

# Show recent events
show_events() {
    log_info "Recent events in namespace: $NAMESPACE"
    echo
    
    kubectl get events -n "$NAMESPACE" --sort-by='.lastTimestamp'
}

# Describe all resources
describe_resources() {
    log_info "Describing all template-arch-lint resources in namespace: $NAMESPACE"
    log_info "================================================================"
    echo
    
    # Get all resource types
    local resources=$(kubectl api-resources --verbs=describe --namespaced -o name | tr '\n' ',' | sed 's/,$//')
    
    # Describe resources with the app label
    kubectl describe all -n "$NAMESPACE" -l app.kubernetes.io/name=template-arch-lint
    
    echo
    log_info "ConfigMaps:"
    kubectl describe configmap -n "$NAMESPACE" -l app.kubernetes.io/name=template-arch-lint
    
    echo
    log_info "Secrets:"
    kubectl describe secret -n "$NAMESPACE" -l app.kubernetes.io/name=template-arch-lint
    
    echo
    log_info "Ingress:"
    kubectl describe ingress -n "$NAMESPACE" -l app.kubernetes.io/name=template-arch-lint
}

# Main execution
main() {
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
    
    check_namespace
    
    # Handle specific actions
    if [[ "${CHECK_HEALTH:-false}" == "true" ]]; then
        check_health
        exit 0
    fi
    
    if [[ "${SHOW_EVENTS:-false}" == "true" ]]; then
        show_events
        exit 0
    fi
    
    if [[ "${DESCRIBE_RESOURCES:-false}" == "true" ]]; then
        describe_resources
        exit 0
    fi
    
    if [[ "$SHOW_METRICS" == "true" ]]; then
        show_metrics
        exit 0
    fi
    
    if [[ "$FOLLOW_LOGS" == "true" ]]; then
        follow_logs
        exit 0
    fi
    
    if [[ "$WATCH" == "true" ]]; then
        watch_resources
        exit 0
    fi
    
    # Default: show status
    show_status
}

# Run main function
main "$@"