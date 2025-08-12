#!/bin/bash

# Deploy template-arch-lint to local Kubernetes cluster
# Supports both minikube and kind

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
K8S_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$K8S_DIR")"

# Default values
ENVIRONMENT="dev"
CLUSTER_TYPE="kind"
NAMESPACE="template-arch-lint-dev"
IMAGE_NAME="template-arch-lint"
IMAGE_TAG="latest"
SKIP_BUILD=false
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
Deploy template-arch-lint to local Kubernetes cluster

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -e, --environment ENV    Target environment (dev|staging|prod) [default: dev]
    -t, --cluster-type TYPE  Cluster type (kind|minikube) [default: kind]
    -i, --image IMAGE        Docker image name [default: template-arch-lint]
    --tag TAG               Docker image tag [default: latest]
    --skip-build            Skip Docker image build
    -v, --verbose           Enable verbose output
    -h, --help              Show this help message

EXAMPLES:
    # Deploy to development environment on kind
    $0

    # Deploy to staging environment
    $0 --environment staging

    # Deploy using minikube
    $0 --cluster-type minikube

    # Skip Docker build and deploy existing image
    $0 --skip-build --tag v1.0.0

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -t|--cluster-type)
            CLUSTER_TYPE="$2"
            shift 2
            ;;
        -i|--image)
            IMAGE_NAME="$2"
            shift 2
            ;;
        --tag)
            IMAGE_TAG="$2"
            shift 2
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
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

# Validate environment
case $ENVIRONMENT in
    dev|staging|prod)
        case $ENVIRONMENT in
            dev) NAMESPACE="template-arch-lint-dev" ;;
            staging) NAMESPACE="template-arch-lint-staging" ;;
            prod) NAMESPACE="template-arch-lint-prod" ;;
        esac
        ;;
    *)
        log_error "Invalid environment: $ENVIRONMENT. Must be one of: dev, staging, prod"
        exit 1
        ;;
esac

# Enable verbose output if requested
if [[ "$VERBOSE" == "true" ]]; then
    set -x
fi

log_info "Starting deployment to $ENVIRONMENT environment using $CLUSTER_TYPE"

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check cluster-specific tools
    case $CLUSTER_TYPE in
        kind)
            if ! command -v kind &> /dev/null; then
                log_error "kind is not installed or not in PATH"
                exit 1
            fi
            ;;
        minikube)
            if ! command -v minikube &> /dev/null; then
                log_error "minikube is not installed or not in PATH"
                exit 1
            fi
            ;;
        *)
            log_error "Unsupported cluster type: $CLUSTER_TYPE"
            exit 1
            ;;
    esac
    
    # Check if kustomize is available (usually bundled with kubectl)
    if ! kubectl kustomize --help &> /dev/null; then
        log_error "kustomize is not available (should be bundled with kubectl 1.14+)"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Setup local cluster
setup_cluster() {
    log_info "Setting up $CLUSTER_TYPE cluster..."
    
    case $CLUSTER_TYPE in
        kind)
            # Check if kind cluster exists
            if ! kind get clusters | grep -q "^kind$"; then
                log_info "Creating kind cluster..."
                kind create cluster --config - << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: kind
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
- role: worker
- role: worker
EOF
            else
                log_info "Kind cluster already exists"
            fi
            
            # Install nginx ingress controller
            log_info "Installing nginx ingress controller..."
            kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
            kubectl wait --namespace ingress-nginx \
                --for=condition=ready pod \
                --selector=app.kubernetes.io/component=controller \
                --timeout=90s
            ;;
            
        minikube)
            # Check if minikube is running
            if ! minikube status &> /dev/null; then
                log_info "Starting minikube..."
                minikube start --cpus=4 --memory=8192 --disk-size=50GB
            else
                log_info "Minikube is already running"
            fi
            
            # Enable ingress addon
            log_info "Enabling ingress addon..."
            minikube addons enable ingress
            ;;
    esac
    
    log_success "Cluster setup completed"
}

# Build Docker image
build_image() {
    if [[ "$SKIP_BUILD" == "true" ]]; then
        log_info "Skipping Docker image build"
        return
    fi
    
    log_info "Building Docker image: $IMAGE_NAME:$IMAGE_TAG"
    
    cd "$PROJECT_ROOT"
    docker build -t "$IMAGE_NAME:$IMAGE_TAG" .
    
    # Load image into cluster
    case $CLUSTER_TYPE in
        kind)
            log_info "Loading image into kind cluster..."
            kind load docker-image "$IMAGE_NAME:$IMAGE_TAG"
            ;;
        minikube)
            log_info "Loading image into minikube..."
            minikube image load "$IMAGE_NAME:$IMAGE_TAG"
            ;;
    esac
    
    log_success "Docker image built and loaded"
}

# Deploy application
deploy_app() {
    log_info "Deploying application to $ENVIRONMENT environment..."
    
    cd "$K8S_DIR/environments/$ENVIRONMENT"
    
    # Apply manifests using kustomize
    kubectl apply -k .
    
    # Wait for deployment to be ready
    log_info "Waiting for deployment to be ready..."
    kubectl wait --for=condition=available --timeout=300s \
        deployment/$(kubectl get deployment -l app.kubernetes.io/name=template-arch-lint -o name | cut -d/ -f2) \
        -n "$NAMESPACE"
    
    log_success "Application deployed successfully"
}

# Show deployment status
show_status() {
    log_info "Deployment Status:"
    echo
    
    log_info "Namespace: $NAMESPACE"
    kubectl get namespace "$NAMESPACE" -o wide 2>/dev/null || log_warn "Namespace not found"
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
    
    # Get application URL
    case $CLUSTER_TYPE in
        kind)
            log_info "Application should be accessible at:"
            log_info "  http://localhost/ (add host entry for your domain)"
            ;;
        minikube)
            MINIKUBE_IP=$(minikube ip)
            log_info "Application should be accessible at:"
            log_info "  http://$MINIKUBE_IP/ (add host entry for your domain)"
            ;;
    esac
    
    log_info "To access the application, add this to your /etc/hosts file:"
    case $ENVIRONMENT in
        dev)
            log_info "  127.0.0.1 dev.template-arch-lint.example.com"
            log_info "  127.0.0.1 api-dev.template-arch-lint.example.com"
            ;;
        staging)
            log_info "  127.0.0.1 staging.template-arch-lint.example.com"
            log_info "  127.0.0.1 api-staging.template-arch-lint.example.com"
            ;;
        prod)
            log_info "  127.0.0.1 template-arch-lint.example.com"
            log_info "  127.0.0.1 api.template-arch-lint.example.com"
            ;;
    esac
}

# Cleanup function
cleanup_on_error() {
    log_error "Deployment failed. Check the logs above for details."
    log_info "To cleanup, run: kubectl delete namespace $NAMESPACE"
    exit 1
}

# Set trap for error handling
trap cleanup_on_error ERR

# Main execution
main() {
    log_info "Template Arch Lint - Local Kubernetes Deployment"
    log_info "=================================================="
    
    check_prerequisites
    setup_cluster
    build_image
    deploy_app
    show_status
    
    log_success "Deployment completed successfully!"
    log_info "Use 'kubectl logs -f deployment/\$(kubectl get deployment -l app.kubernetes.io/name=template-arch-lint -o name | cut -d/ -f2) -n $NAMESPACE' to view logs"
}

# Run main function
main "$@"