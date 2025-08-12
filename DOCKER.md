# Docker Containerization Guide

This document provides comprehensive information about the Docker containerization implementation for the template-arch-lint application.

## Overview

The application is containerized using a multi-stage Docker approach optimized for security, size, and production deployment. The container foundation enables Kubernetes deployment and provides a complete local development environment.

## Quick Start

### Using Pre-built Images from GitHub Container Registry

```bash
# Pull and run the latest image
docker pull ghcr.io/larsartmann/template-arch-lint:latest
docker run -p 8080:8080 ghcr.io/larsartmann/template-arch-lint:latest

# Run with health check
docker run -p 8080:8080 --health-cmd="/server -health-check" \
  ghcr.io/larsartmann/template-arch-lint:latest

# Access the application at http://localhost:8080
```

### Local Development with Docker Compose

```bash
# Start the complete development environment
just docker-dev-detached

# Access services:
# - Application: http://localhost:8080
# - Grafana: http://localhost:3000 (admin/admin)
# - Prometheus: http://localhost:9090
# - Jaeger UI: http://localhost:16686
# - SQLite Web UI: http://localhost:8081

# Stop the environment
just docker-stop
```

### Container Registry Images

The application is automatically built and published to GitHub Container Registry with comprehensive security scanning and multi-architecture support.

**Available Tags:**
- `latest` - Latest main branch build
- `v1.0.0` - Semantic version releases  
- `main` - Main branch builds
- `pr-123` - Pull request builds

**Supported Architectures:**
- `linux/amd64` - Intel/AMD 64-bit
- `linux/arm64` - ARM 64-bit (Apple Silicon, AWS Graviton)

**Security Features:**
- Vulnerability scanning with Trivy and Grype
- SBOM (Software Bill of Materials) generation
- Build provenance attestation
- Container image signing with sigstore

For detailed registry documentation, see [CONTAINER_REGISTRY.md](CONTAINER_REGISTRY.md).

### Building and Testing Docker Images

```bash
# Build and test the Docker image
just docker-test

# Build image only
just docker-build

# Run security scan
just docker-security
```

## Architecture

### Multi-Stage Dockerfile

The `Dockerfile` uses a multi-stage build approach:

1. **Build Stage** (`golang:1.24-alpine`)
   - Installs build dependencies (git, ca-certificates, gcc, musl-dev)
   - Downloads Go dependencies with layer caching optimization
   - Installs and runs `templ` for template generation
   - Compiles the Go application as a static binary

2. **Runtime Stage** (`gcr.io/distroless/static-debian12:nonroot`)
   - Uses distroless base image for minimal attack surface
   - Runs as non-root user (65532:65532)
   - Includes only the compiled binary and essential files
   - Implements health checks for container orchestration

### Key Optimizations

- **Size Optimization**: Uses distroless runtime image (~20MB vs ~300MB+ for full OS)
- **Security**: Non-root user, minimal attack surface, no shell access
- **Performance**: Multi-stage build with layer caching
- **Static Binary**: No external dependencies in runtime
- **Health Checks**: Built-in application health monitoring

## Development Environment

### Docker Compose Services

The `docker-compose.yml` provides a complete observability stack:

#### Core Application
- **app**: Main Go application with health checks
- **db-ui**: SQLite web interface for database inspection

#### Observability Stack
- **otel-collector**: OpenTelemetry collector for traces and metrics
- **jaeger**: Distributed tracing UI and storage
- **prometheus**: Metrics collection and storage
- **grafana**: Metrics visualization and dashboards

### Configuration Files

```
docker/
├── otel-collector-config.yaml  # OpenTelemetry collector configuration
├── prometheus.yml              # Prometheus scraping configuration
└── grafana/
    ├── datasources/            # Grafana datasource configuration
    │   └── datasources.yml
    └── dashboards/             # Dashboard provisioning
        └── dashboards.yml
```

## Production Deployment

### Container Registry

Images are built and pushed to GitHub Container Registry (ghcr.io) via CI/CD:

```bash
# Images are tagged with:
ghcr.io/larsartmann/template-arch-lint:latest      # Main branch
ghcr.io/larsartmann/template-arch-lint:sha-abc123  # Commit SHA
ghcr.io/larsartmann/template-arch-lint:pr-123      # Pull requests
```

### Health Checks

The container includes comprehensive health checking:

```bash
# Docker health check (built-in)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/server", "-health-check"] || exit 1

# Application endpoints
curl http://localhost:8080/health/live   # Liveness probe
curl http://localhost:8080/health/ready  # Readiness probe  
curl http://localhost:8080/health        # Detailed health status
curl http://localhost:8080/version       # Version information
```

### Environment Variables

```bash
# Core application
APP_ENVIRONMENT=production
APP_SERVER_HOST=0.0.0.0
APP_SERVER_PORT=8080

# Observability (optional)
APP_OBSERVABILITY_TRACING_ENDPOINT=http://jaeger:4318/v1/traces
APP_OBSERVABILITY_METRICS_ENDPOINT=http://otel-collector:4318/v1/metrics

# Logging
APP_LOGGING_LEVEL=info
APP_LOGGING_FORMAT=json
```

## Security

### Security Scanning

Automatic security scanning is integrated into CI/CD using Trivy:

```bash
# Local security scan
just docker-security

# CI automatically scans:
# - Source code vulnerabilities
# - Container image vulnerabilities  
# - Dependencies and licenses
```

### Security Features

- **Distroless Base**: Minimal attack surface, no shell, no package manager
- **Non-Root User**: Runs as user 65532 (nonroot)
- **Static Binary**: No dynamic dependencies
- **Read-Only Root**: Container filesystem is read-only where possible
- **Minimal Privileges**: Only necessary capabilities

### Security Best Practices

1. **Regular Updates**: Base images updated automatically via Dependabot
2. **Vulnerability Scanning**: Automated scanning on every build
3. **Secrets Management**: No hardcoded secrets in images
4. **Network Security**: Only expose necessary ports
5. **Resource Limits**: Set appropriate CPU/memory limits in production

## CI/CD Integration

### GitHub Actions Workflow

The `.github/workflows/ci.yml` includes a dedicated Docker job:

1. **Multi-Platform Builds**: linux/amd64, linux/arm64
2. **Layer Caching**: GitHub Actions cache for faster builds
3. **Security Scanning**: Trivy vulnerability assessment
4. **Image Testing**: Functional tests on built images
5. **Registry Push**: Automatic push to GitHub Container Registry

### Build Optimization

```yaml
# Build cache configuration
cache-from: type=gha
cache-to: type=gha,mode=max
```

## Troubleshooting

### Common Issues

1. **Build Failures**
   ```bash
   # Check build context
   docker build --no-cache .
   
   # Verify .dockerignore
   docker build --progress=plain .
   ```

2. **Runtime Issues**
   ```bash
   # Check logs
   docker logs <container-id>
   
   # Debug container
   docker run -it --entrypoint="" template-arch-lint:latest /bin/sh
   # Note: Won't work with distroless - use builder stage for debugging
   ```

3. **Health Check Failures**
   ```bash
   # Test health check manually
   docker run --rm template-arch-lint:latest -health-check
   
   # Check application logs
   docker-compose logs app
   ```

### Performance Monitoring

```bash
# Container resource usage
docker stats

# Application metrics
curl http://localhost:2112/metrics

# Distributed tracing
# Visit http://localhost:16686 (Jaeger UI)
```

## Best Practices

### Development Workflow

1. **Use Docker Compose**: Complete environment with one command
2. **Test Locally**: Run `just docker-test` before pushing
3. **Security First**: Scan images regularly with `just docker-security`
4. **Monitor Resources**: Check container metrics and logs

### Production Deployment

1. **Resource Limits**: Set appropriate CPU/memory limits
2. **Health Checks**: Configure liveness and readiness probes
3. **Logging**: Use structured logging with log aggregation
4. **Monitoring**: Set up metrics collection and alerting
5. **Scaling**: Use horizontal pod autoscaling based on metrics

### Image Management

1. **Tagging Strategy**: Use semantic versioning and commit SHAs
2. **Registry Security**: Enable vulnerability scanning in registry
3. **Image Cleanup**: Regularly clean up old images
4. **Multi-Architecture**: Support both AMD64 and ARM64

## File Structure

```
├── Dockerfile                       # Multi-stage container definition
├── .dockerignore                    # Build context optimization
├── docker-compose.yml               # Development environment
├── docker-compose.override.yml      # Development overrides
└── docker/                         # Configuration files
    ├── otel-collector-config.yaml
    ├── prometheus.yml
    └── grafana/
        ├── datasources/
        └── dashboards/
```

## Integration with Kubernetes

The containerized application is ready for Kubernetes deployment:

```yaml
# Example Kubernetes deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: template-arch-lint
spec:
  replicas: 3
  selector:
    matchLabels:
      app: template-arch-lint
  template:
    spec:
      containers:
      - name: app
        image: ghcr.io/larsartmann/template-arch-lint:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

This Docker containerization foundation provides enterprise-grade deployment capabilities with security, observability, and operational best practices built-in.