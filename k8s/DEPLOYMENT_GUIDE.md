# Kubernetes Deployment Guide

A step-by-step guide for deploying template-arch-lint to Kubernetes environments.

## 🚀 Quick Deployment (Local)

### Prerequisites Check

```bash
# Verify prerequisites
kubectl version --short
docker --version
kind --version  # or minikube version

# Clone the repository
git clone <repository-url>
cd template-arch-lint
```

### 1. Build and Deploy Locally

```bash
# Deploy to development environment using kind
./k8s/scripts/deploy-local.sh

# Or deploy to staging
./k8s/scripts/deploy-local.sh --environment staging

# Monitor the deployment
./k8s/scripts/monitor.sh --logs
```

### 2. Access the Application

```bash
# Add to /etc/hosts (or use your preferred method)
echo "127.0.0.1 dev.template-arch-lint.example.com" | sudo tee -a /etc/hosts

# Access the application
curl http://dev.template-arch-lint.example.com/health
```

### 3. Cleanup

```bash
# Remove the deployment
./k8s/scripts/cleanup.sh

# Or destroy the entire cluster
./k8s/scripts/cleanup.sh --destroy-cluster
```

## 🏢 Production Deployment

### Prerequisites

1. **Kubernetes Cluster** (v1.19+)
   - EKS, GKE, AKS, or self-managed cluster
   - Minimum 3 nodes with 2 CPU / 4GB RAM each

2. **Required Add-ons**
   - NGINX Ingress Controller
   - cert-manager (for SSL certificates)
   - Metrics Server (for HPA)

3. **External Dependencies**
   - Container registry (Docker Hub, ECR, GCR, etc.)
   - DNS management (Route53, CloudDNS, etc.)
   - Certificate authority (Let's Encrypt, etc.)

### Step-by-Step Production Deployment

#### 1. Prepare Container Image

```bash
# Build production image
docker build -t your-registry.com/template-arch-lint:v1.0.0 .

# Push to registry
docker push your-registry.com/template-arch-lint:v1.0.0
```

#### 2. Configure Environment

```bash
# Copy and customize production configuration
cp k8s/environments/prod/kustomization.yaml k8s/environments/prod/kustomization.yaml.backup

# Update image reference in kustomization.yaml
sed -i 's|newTag: v1.0.0|newTag: v1.0.0|g' k8s/environments/prod/kustomization.yaml

# Update domain names
sed -i 's|template-arch-lint.example.com|your-domain.com|g' k8s/environments/prod/kustomization.yaml
```

#### 3. Create Secrets

```bash
# Create namespace first
kubectl create namespace template-arch-lint-prod

# Create database secrets (if using external database)
kubectl create secret generic template-arch-lint-secrets \
  --from-literal=APP_DATABASE_PASSWORD='your-secure-password' \
  --from-literal=APP_JWT_SECRET='your-jwt-secret' \
  --namespace=template-arch-lint-prod

# Create container registry secret (if using private registry)
kubectl create secret docker-registry regcred \
  --docker-server=your-registry.com \
  --docker-username=your-username \
  --docker-password=your-password \
  --namespace=template-arch-lint-prod
```

#### 4. Install Prerequisites

```bash
# Install NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/cloud/deploy.yaml

# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer for Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

#### 5. Deploy Application

```bash
# Validate manifests
./k8s/scripts/validate.sh --environment prod --lint

# Apply production configuration
kubectl apply -k k8s/environments/prod/

# Wait for deployment to be ready
kubectl wait --for=condition=available --timeout=300s \
  deployment/template-arch-lint -n template-arch-lint-prod
```

#### 6. Configure DNS

```bash
# Get ingress IP address
kubectl get ingress template-arch-lint -n template-arch-lint-prod -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# Configure DNS A record
# your-domain.com -> INGRESS_IP
# api.your-domain.com -> INGRESS_IP
```

#### 7. Verify Deployment

```bash
# Check all resources
kubectl get all -n template-arch-lint-prod

# Test health endpoints
curl https://your-domain.com/health
curl https://your-domain.com/version

# Check SSL certificate
curl -I https://your-domain.com

# Monitor logs
kubectl logs -f deployment/template-arch-lint -n template-arch-lint-prod
```

## 🔧 Configuration Management

### Environment-Specific Configurations

#### Development
```bash
# Deploy development environment
kubectl apply -k k8s/environments/dev/

# Access locally
kubectl port-forward service/dev-template-arch-lint 8080:80 -n template-arch-lint-dev
```

#### Staging
```bash
# Deploy staging environment  
kubectl apply -k k8s/environments/staging/

# Run smoke tests
curl https://staging.your-domain.com/health
```

#### Production
```bash
# Production deployment with zero downtime
kubectl apply -k k8s/environments/prod/

# Monitor rollout
kubectl rollout status deployment/template-arch-lint -n template-arch-lint-prod
```

### Secret Management

#### Using External Secret Management

```bash
# Install External Secrets Operator
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets -n external-secrets-system --create-namespace

# Configure secret store (example with AWS Secrets Manager)
cat <<EOF | kubectl apply -f -
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: template-arch-lint-prod
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-west-2
      auth:
        secretRef:
          accessKeyId:
            name: aws-credentials
            key: access-key-id
          secretAccessKey:
            name: aws-credentials
            key: secret-access-key
EOF

# Create ExternalSecret
cat <<EOF | kubectl apply -f -
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: template-arch-lint-secrets
  namespace: template-arch-lint-prod
spec:
  refreshInterval: 15s
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: template-arch-lint-secrets
    creationPolicy: Owner
  data:
  - secretKey: APP_DATABASE_PASSWORD
    remoteRef:
      key: prod/template-arch-lint/database
      property: password
EOF
```

## 📊 Monitoring and Alerting

### Prometheus Setup

```bash
# Install Prometheus Stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace

# Access Prometheus UI
kubectl port-forward service/kube-prometheus-stack-prometheus 9090:9090 -n monitoring

# Access Grafana
kubectl get secret kube-prometheus-stack-grafana -n monitoring -o jsonpath="{.data.admin-password}" | base64 --decode
kubectl port-forward service/kube-prometheus-stack-grafana 3000:80 -n monitoring
```

### Application Metrics

```bash
# Create ServiceMonitor for application metrics
cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: template-arch-lint
  namespace: template-arch-lint-prod
  labels:
    app.kubernetes.io/name: template-arch-lint
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: template-arch-lint
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
EOF
```

### Alerting Rules

```bash
# Create PrometheusRule for alerts
cat <<EOF | kubectl apply -f -
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: template-arch-lint-alerts
  namespace: template-arch-lint-prod
spec:
  groups:
  - name: template-arch-lint
    rules:
    - alert: TemplateArchLintDown
      expr: up{job="template-arch-lint"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Template Arch Lint is down"
        description: "Template Arch Lint has been down for more than 1 minute."
    
    - alert: HighErrorRate
      expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
      for: 2m
      labels:
        severity: warning
      annotations:
        summary: "High error rate detected"
        description: "Error rate is above 5% for more than 2 minutes."
EOF
```

## 🔄 CI/CD Integration

### GitHub Actions Example

```yaml
# .github/workflows/deploy.yml
name: Deploy to Kubernetes

on:
  push:
    branches: [main]
    tags: ['v*']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Configure AWS credentials
      uses: aws-actions/configure-aws-credentials@v2
      with:
        aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
        aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
        aws-region: us-west-2
    
    - name: Login to Amazon ECR
      uses: aws-actions/amazon-ecr-login@v1
    
    - name: Build and push Docker image
      run: |
        docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$GITHUB_SHA .
        docker push $ECR_REGISTRY/$ECR_REPOSITORY:$GITHUB_SHA
    
    - name: Configure kubectl
      run: |
        aws eks get-token --cluster-name prod-cluster | kubectl apply -f -
    
    - name: Deploy to Kubernetes
      run: |
        cd k8s/environments/prod
        kustomize edit set image template-arch-lint=$ECR_REGISTRY/$ECR_REPOSITORY:$GITHUB_SHA
        kubectl apply -k .
        kubectl rollout status deployment/template-arch-lint -n template-arch-lint-prod
```

## 🚨 Troubleshooting Common Issues

### Image Pull Errors

```bash
# Check if image exists
docker pull your-registry.com/template-arch-lint:v1.0.0

# Verify registry credentials
kubectl get secret regcred -n template-arch-lint-prod -o yaml

# Update deployment with correct image
kubectl patch deployment template-arch-lint -n template-arch-lint-prod \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"template-arch-lint","image":"your-registry.com/template-arch-lint:v1.0.0"}]}}}}'
```

### Certificate Issues

```bash
# Check certificate status
kubectl get certificate -n template-arch-lint-prod

# Debug certificate request
kubectl describe certificaterequest -n template-arch-lint-prod

# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager

# Force certificate renewal
kubectl delete certificate template-arch-lint-tls -n template-arch-lint-prod
```

### Database Connection Issues

```bash
# Check database connectivity
kubectl exec -it deployment/template-arch-lint -n template-arch-lint-prod -- \
  nc -zv database-host 5432

# Verify database credentials
kubectl get secret template-arch-lint-secrets -n template-arch-lint-prod -o yaml

# Check database logs
kubectl logs -f deployment/template-arch-lint -n template-arch-lint-prod | grep -i database
```

### Performance Issues

```bash
# Check resource usage
kubectl top pods -n template-arch-lint-prod

# Check HPA status
kubectl get hpa -n template-arch-lint-prod

# Scale manually if needed
kubectl scale deployment template-arch-lint --replicas=10 -n template-arch-lint-prod

# Check node resources
kubectl describe nodes
```

## 🔒 Security Hardening

### Runtime Security

```bash
# Install Falco for runtime security monitoring
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm install falco falcosecurity/falco --namespace falco-system --create-namespace

# Install OPA Gatekeeper for policy enforcement
kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/release-3.14/deploy/gatekeeper.yaml
```

### Network Policies

```bash
# Test network connectivity
kubectl run netshoot --image=nicolaka/netshoot -it --rm -- /bin/bash

# From within netshoot pod
nslookup template-arch-lint.template-arch-lint-prod.svc.cluster.local
nc -zv template-arch-lint.template-arch-lint-prod.svc.cluster.local 8080
```

### Security Scanning

```bash
# Scan container image for vulnerabilities
trivy image your-registry.com/template-arch-lint:v1.0.0

# Scan Kubernetes manifests
trivy config k8s/environments/prod/

# Run CIS Kubernetes Benchmark
kube-bench run --targets master,node
```

This deployment guide provides comprehensive instructions for deploying template-arch-lint to various Kubernetes environments with proper security, monitoring, and operational practices.