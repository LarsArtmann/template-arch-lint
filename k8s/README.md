# Kubernetes Deployment for Template Arch Lint

This directory contains production-ready Kubernetes manifests for deploying the template-arch-lint application. The configuration supports multiple environments (development, staging, production) with appropriate resource allocation and security configurations.

## 📁 Directory Structure

```
k8s/
├── base/                          # Base Kubernetes manifests
│   ├── configmap.yaml             # Application configuration
│   ├── deployment.yaml            # Main application deployment
│   ├── hpa.yaml                   # Horizontal Pod Autoscaler
│   ├── ingress.yaml               # Ingress configuration
│   ├── kustomization.yaml         # Base kustomization
│   ├── namespace.yaml             # Kubernetes namespace
│   ├── rbac.yaml                  # Role-based access control
│   ├── secret.yaml                # Sensitive configuration
│   ├── security-policies.yaml     # Security policies
│   ├── service.yaml               # Kubernetes services
│   └── storage.yaml               # Persistent volumes
├── environments/                  # Environment-specific configurations
│   ├── dev/
│   │   └── kustomization.yaml     # Development overrides
│   ├── staging/
│   │   └── kustomization.yaml     # Staging overrides
│   └── prod/
│       └── kustomization.yaml     # Production overrides
├── scripts/                       # Deployment and management scripts
│   ├── cleanup.sh                 # Environment cleanup
│   ├── deploy-local.sh            # Local deployment (kind/minikube)
│   ├── monitor.sh                 # Monitoring and logging
│   └── validate.sh                # Manifest validation
└── README.md                      # This file
```

## 🚀 Quick Start

### Prerequisites

- **Kubernetes Cluster**: kind, minikube, or cloud provider cluster
- **kubectl**: v1.19+ with kustomize support
- **Docker**: For building application images

### Local Deployment

1. **Deploy to development environment**:
   ```bash
   ./k8s/scripts/deploy-local.sh
   ```

2. **Deploy to staging environment**:
   ```bash
   ./k8s/scripts/deploy-local.sh --environment staging
   ```

3. **Monitor deployment**:
   ```bash
   ./k8s/scripts/monitor.sh --logs
   ```

4. **Clean up**:
   ```bash
   ./k8s/scripts/cleanup.sh
   ```

## 🏗️ Architecture Overview

The deployment consists of the following components:

### Core Application
- **Deployment**: Stateless Go web application with 3 replicas (configurable per environment)
- **Service**: ClusterIP service exposing HTTP (8080) and metrics (2112) ports
- **Ingress**: NGINX ingress with SSL termination and rate limiting
- **HPA**: Horizontal Pod Autoscaler based on CPU/memory metrics

### Configuration Management
- **ConfigMap**: Non-sensitive application configuration
- **Secrets**: Sensitive data (API keys, certificates, passwords)
- **Environment Variables**: Runtime configuration injection

### Storage
- **EmptyDir**: Temporary storage for logs and cache
- **PersistentVolume**: Optional storage for SQLite database
- **StorageClass**: SSD-backed storage with encryption

### Security
- **RBAC**: Minimal service account permissions
- **SecurityContext**: Non-root user, read-only filesystem
- **NetworkPolicy**: Ingress/egress traffic restrictions
- **PodSecurityPolicy**: Pod security constraints

### Observability
- **Health Checks**: Liveness, readiness, and startup probes
- **Metrics**: Prometheus metrics endpoint on port 2112
- **Logging**: Structured JSON logging to stdout
- **Tracing**: OpenTelemetry integration

## 🌍 Environment Configurations

### Development Environment
- **Namespace**: `template-arch-lint-dev`
- **Replicas**: 1
- **Resources**: 50m CPU, 64Mi memory
- **Domain**: `dev.template-arch-lint.example.com`
- **Logging**: Debug level
- **Sampling**: 100% tracing

### Staging Environment
- **Namespace**: `template-arch-lint-staging`
- **Replicas**: 2
- **Resources**: 75m CPU, 96Mi memory
- **Domain**: `staging.template-arch-lint.example.com`
- **Logging**: Info level
- **Sampling**: 50% tracing

### Production Environment
- **Namespace**: `template-arch-lint-prod`
- **Replicas**: 5 (auto-scaling 3-20)
- **Resources**: 200m CPU, 256Mi memory
- **Domain**: `template-arch-lint.example.com`
- **Logging**: Warn level
- **Sampling**: 1% tracing

## 📊 Resource Requirements

### Minimum System Requirements
- **CPU**: 2 cores
- **Memory**: 4GB RAM
- **Storage**: 50GB available disk space
- **Kubernetes**: v1.19+

### Per-Pod Resources

| Environment | CPU Request | CPU Limit | Memory Request | Memory Limit |
|-------------|-------------|-----------|----------------|--------------|
| Development | 50m         | 200m      | 64Mi           | 128Mi        |
| Staging     | 75m         | 300m      | 96Mi           | 192Mi        |
| Production  | 200m        | 1000m     | 256Mi          | 512Mi        |

## 🔧 Configuration

### Environment Variables

The application accepts configuration through environment variables prefixed with `APP_`:

```bash
# Server Configuration
APP_SERVER_HOST=0.0.0.0
APP_SERVER_PORT=8080
APP_SERVER_READ_TIMEOUT=5s
APP_SERVER_WRITE_TIMEOUT=10s

# Database Configuration
APP_DATABASE_DRIVER=sqlite3
APP_DATABASE_DSN=/data/app.db

# Logging Configuration
APP_LOGGING_LEVEL=info
APP_LOGGING_FORMAT=json

# Observability Configuration
APP_OBSERVABILITY_ENABLED=true
APP_OBSERVABILITY_TRACING_ENABLED=true
APP_OBSERVABILITY_METRICS_ENABLED=true
```

### Secrets Management

Sensitive configuration should be stored in Kubernetes Secrets:

```bash
# Create secret with sensitive data
kubectl create secret generic template-arch-lint-secrets \
  --from-literal=APP_DATABASE_PASSWORD=your-password \
  --from-literal=APP_JWT_SECRET=your-jwt-secret \
  -n template-arch-lint-prod
```

### SSL/TLS Configuration

The ingress uses cert-manager for automatic SSL certificate management:

```yaml
annotations:
  cert-manager.io/cluster-issuer: "letsencrypt-prod"
```

## 📈 Monitoring and Observability

### Health Endpoints

The application exposes several health endpoints:

- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /health` - Comprehensive health check
- `GET /version` - Application version info

### Metrics Collection

Prometheus metrics are available at:
- `http://localhost:2112/metrics` (container port)
- Scraped automatically via service annotations

### Logging

All logs are structured JSON format sent to stdout:

```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "Server started",
  "port": 8080,
  "environment": "production"
}
```

### Tracing

Distributed tracing via OpenTelemetry:
- **Jaeger**: UI for trace visualization
- **OTLP**: Standard telemetry protocol
- **Automatic instrumentation**: HTTP requests, database queries

## 🔐 Security Considerations

### Container Security
- **Non-root user**: Runs as UID 65532
- **Read-only filesystem**: Prevents runtime modifications
- **No privileged escalation**: Security context restrictions
- **Minimal capabilities**: All capabilities dropped

### Network Security
- **Network policies**: Restrict ingress/egress traffic
- **Ingress authentication**: Basic auth for metrics endpoint
- **Rate limiting**: Prevent abuse and DoS attacks
- **Security headers**: HSTS, CSP, X-Frame-Options

### Secret Management
- **External secrets**: Use external secret management systems
- **Encryption at rest**: Kubernetes secrets encryption
- **RBAC**: Minimal service account permissions
- **Certificate management**: Automated SSL/TLS certificates

## 🚨 Troubleshooting

### Common Issues

1. **Pod Not Starting**
   ```bash
   # Check pod events
   kubectl describe pod <pod-name> -n <namespace>
   
   # Check logs
   kubectl logs <pod-name> -n <namespace>
   ```

2. **Service Unreachable**
   ```bash
   # Check service endpoints
   kubectl get endpoints <service-name> -n <namespace>
   
   # Test service connectivity
   kubectl run -it --rm debug --image=busybox -- sh
   wget -qO- http://service-name:8080/health
   ```

3. **Ingress Issues**
   ```bash
   # Check ingress status
   kubectl describe ingress <ingress-name> -n <namespace>
   
   # Check ingress controller logs
   kubectl logs -n ingress-nginx deployment/nginx-ingress-controller
   ```

4. **Resource Issues**
   ```bash
   # Check resource usage
   kubectl top pods -n <namespace>
   
   # Check HPA status
   kubectl get hpa -n <namespace>
   ```

### Debug Commands

```bash
# Get all resources in namespace
kubectl get all -n template-arch-lint-prod

# Describe deployment
kubectl describe deployment template-arch-lint -n template-arch-lint-prod

# Port forward for local testing
kubectl port-forward service/template-arch-lint 8080:80 -n template-arch-lint-prod

# Execute commands in pod
kubectl exec -it deployment/template-arch-lint -n template-arch-lint-prod -- /bin/sh

# View recent events
kubectl get events --sort-by='.lastTimestamp' -n template-arch-lint-prod
```

## 🔄 Deployment Strategies

### Rolling Updates
Default deployment strategy with zero downtime:
```yaml
strategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: 1
    maxSurge: 1
```

### Blue-Green Deployment
For critical production updates:
1. Deploy new version to separate namespace
2. Test new version thoroughly
3. Switch ingress to new version
4. Cleanup old version

### Canary Deployment
Gradual rollout with traffic splitting:
1. Deploy new version with reduced replicas
2. Route small percentage of traffic
3. Monitor metrics and errors
4. Gradually increase traffic
5. Complete rollout or rollback

## 📝 Maintenance

### Regular Tasks

1. **Update Dependencies**
   ```bash
   # Update Docker base images
   docker pull gcr.io/distroless/static-debian12:nonroot
   
   # Rebuild application image
   docker build -t template-arch-lint:v1.1.0 .
   ```

2. **Certificate Renewal**
   ```bash
   # Check certificate expiry
   kubectl get certificate -n template-arch-lint-prod
   
   # Force renewal if needed
   kubectl delete certificate template-arch-lint-tls -n template-arch-lint-prod
   ```

3. **Resource Cleanup**
   ```bash
   # Cleanup old ReplicaSets
   kubectl get rs -n template-arch-lint-prod
   kubectl delete rs <old-replicaset> -n template-arch-lint-prod
   ```

4. **Backup Data**
   ```bash
   # Backup persistent volume data
   kubectl exec deployment/template-arch-lint -n template-arch-lint-prod -- \
     tar czf - /data | kubectl exec -i backup-pod -- tar xzf - -C /backup
   ```

### Monitoring Alerts

Set up alerts for:
- Pod crash loops
- High CPU/memory usage  
- Failed health checks
- Certificate expiry
- High error rates
- Response time degradation

## 🆘 Disaster Recovery

### Backup Strategy
- **Database**: Regular SQLite backups to object storage
- **Configuration**: Version-controlled manifests
- **Secrets**: Encrypted backup in secure storage
- **Persistent volumes**: Snapshot-based backups

### Recovery Procedures
1. **Complete cluster failure**: Redeploy from manifests
2. **Data corruption**: Restore from latest backup
3. **Certificate issues**: Force certificate renewal
4. **Network issues**: Check ingress and network policies

## 📚 References

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kustomize Documentation](https://kustomize.io/)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [cert-manager Documentation](https://cert-manager.io/docs/)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)

## 🤝 Contributing

When contributing to the Kubernetes configuration:

1. **Validate manifests**: Run `./scripts/validate.sh`
2. **Test locally**: Deploy to kind/minikube
3. **Security review**: Check for security best practices
4. **Documentation**: Update this README for changes
5. **Version control**: Use semantic versioning for images