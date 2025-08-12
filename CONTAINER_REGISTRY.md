# Container Registry Integration & Usage Guide

## Overview

This project implements comprehensive container registry integration with GitHub Container Registry (GHCR), featuring automated image building, publishing, security scanning, and lifecycle management.

## Registry Configuration

### GitHub Container Registry (GHCR)
- **Registry URL**: `ghcr.io`
- **Repository**: `ghcr.io/larsartmann/template-arch-lint`
- **Authentication**: GitHub Actions token (automatic)

### Supported Architectures
- `linux/amd64` (Intel/AMD 64-bit)
- `linux/arm64` (ARM 64-bit, Apple Silicon, AWS Graviton)

## Image Tagging Strategy

### Semantic Versioning
```
ghcr.io/larsartmann/template-arch-lint:v1.0.0
ghcr.io/larsartmann/template-arch-lint:v1.0
ghcr.io/larsartmann/template-arch-lint:v1
ghcr.io/larsartmann/template-arch-lint:latest
```

### Branch-based Tags
```
ghcr.io/larsartmann/template-arch-lint:main
ghcr.io/larsartmann/template-arch-lint:develop
ghcr.io/larsartmann/template-arch-lint:feature-branch-name
```

### Pull Request Tags
```
ghcr.io/larsartmann/template-arch-lint:pr-123
```

### Commit-based Tags
```
ghcr.io/larsartmann/template-arch-lint:main-a1b2c3d-20240101-120000
ghcr.io/larsartmann/template-arch-lint:0.0.0-a1b2c3d
```

## CI/CD Pipeline Integration

### Workflow Triggers

#### Automatic Building & Publishing
- **Main/Master branches**: Publishes tagged images with `latest` tag
- **Feature branches**: Publishes branch-tagged images
- **Pull requests**: Publishes PR-tagged images (for testing)
- **Release tags**: Publishes semantic version tags

#### Manual Workflow Dispatch
```yaml
workflow_dispatch: # Available for manual triggering
```

### Pipeline Stages

1. **Metadata Preparation** (`prepare-metadata`)
   - Determines version and tagging strategy
   - Extracts Docker metadata
   - Sets up image labels and annotations

2. **Security Scanning** (`container-security`)
   - **Trivy**: Vulnerability scanning for OS packages and dependencies
   - **Grype**: Additional vulnerability analysis
   - SARIF reports uploaded to GitHub Security tab
   - Fails on CRITICAL and HIGH vulnerabilities

3. **Build & Push** (`build-and-push`)
   - Multi-architecture builds (AMD64/ARM64)
   - Pushes to GitHub Container Registry
   - Generates SBOM (Software Bill of Materials)
   - Creates build provenance attestations
   - Signs images with GitHub's sigstore

4. **Container Testing** (`test-container`)
   - Tests image startup on both architectures
   - Validates health checks
   - Ensures container functionality

## Security Features

### Vulnerability Scanning
- **Pre-push scanning**: Images are scanned before publishing
- **Continuous monitoring**: Automated security updates via Dependabot
- **SARIF reporting**: Results integrated with GitHub Security tab

### Supply Chain Security
- **SBOM Generation**: Complete software bill of materials
- **Provenance Attestation**: Build process verification
- **Sigstore Signing**: Container image signing with GitHub's keys
- **Base Image**: Uses Google's distroless images for minimal attack surface

### Security Policies
```yaml
# Scan severity levels that fail the build
severity: 'CRITICAL,HIGH,MEDIUM'
exit-code: '1'
ignore-unfixed: true
```

## Registry Lifecycle Management

### Retention Policies

#### Automated Cleanup (Daily at 2:00 AM UTC)
- **Keep**: Latest 10 tagged versions
- **Keep**: All main/master branch images
- **Keep**: All semantic release versions
- **Remove**: Untagged images immediately
- **Remove**: PR images after 7 days
- **Remove**: Feature branch images after 3 days

#### Manual Cleanup
```bash
# Trigger manual cleanup with dry run
gh workflow run registry-cleanup.yml --field dry_run=true

# Trigger actual cleanup
gh workflow run registry-cleanup.yml --field dry_run=false
```

### Storage Optimization
- Multi-stage builds for minimal image size (~20MB)
- Layer caching with GitHub Actions cache
- Efficient `.dockerignore` to reduce build context

## Usage Examples

### Pull and Run Latest Image
```bash
# Pull latest image
docker pull ghcr.io/larsartmann/template-arch-lint:latest

# Run container
docker run -p 8080:8080 ghcr.io/larsartmann/template-arch-lint:latest

# Run with health check
docker run -p 8080:8080 --health-cmd="/server -health-check" \
  ghcr.io/larsartmann/template-arch-lint:latest
```

### Pull Specific Architecture
```bash
# Pull ARM64 version (Apple Silicon)
docker pull --platform=linux/arm64 ghcr.io/larsartmann/template-arch-lint:latest

# Pull AMD64 version
docker pull --platform=linux/amd64 ghcr.io/larsartmann/template-arch-lint:latest
```

### Use in Docker Compose
```yaml
services:
  app:
    image: ghcr.io/larsartmann/template-arch-lint:latest
    ports:
      - "8080:8080"
    environment:
      - APP_ENVIRONMENT=production
    healthcheck:
      test: ["/server", "-health-check"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Use in Kubernetes
```yaml
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
    metadata:
      labels:
        app: template-arch-lint
    spec:
      containers:
      - name: app
        image: ghcr.io/larsartmann/template-arch-lint:v1.0.0
        ports:
        - containerPort: 8080
        livenessProbe:
          exec:
            command: ["/server", "-health-check"]
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command: ["/server", "-health-check"]
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Image Information

### Labels and Annotations
```yaml
org.opencontainers.image.title: "Template Architecture Lint"
org.opencontainers.image.description: "Go-based architecture linting tool with web interface"
org.opencontainers.image.vendor: "LarsArtmann"
org.opencontainers.image.licenses: "MIT"
org.opencontainers.image.source: "https://github.com/LarsArtmann/template-arch-lint"
org.opencontainers.image.documentation: "https://github.com/LarsArtmann/template-arch-lint#readme"
```

### Image Size and Layers
- **Base Image**: `gcr.io/distroless/static-debian12:nonroot`
- **Final Size**: ~20MB (optimized multi-stage build)
- **User**: `nonroot` (UID 65532) for security

### Included Files
```
/server                    # Main application binary
/config.yaml              # Default configuration
/config.production.yaml   # Production configuration
/web/                      # Web templates and static files
/usr/share/zoneinfo        # Timezone data
/etc/ssl/certs/           # CA certificates for HTTPS
```

## Monitoring and Observability

### Registry Metrics
- Image pull counts (via GitHub Insights)
- Storage usage tracking
- Build success/failure rates
- Vulnerability scan results

### Container Metrics
- Built-in Prometheus metrics endpoint (`:2112/metrics`)
- Health check endpoint (`/health`)
- Ready check endpoint (`/ready`)

## Troubleshooting

### Common Issues

#### Image Pull Authentication
```bash
# Login to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Or use GitHub CLI
gh auth token | docker login ghcr.io -u USERNAME --password-stdin
```

#### Multi-architecture Issues
```bash
# Enable Docker BuildKit for multi-arch support
export DOCKER_BUILDKIT=1

# Use buildx for multi-platform builds
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap
```

#### Build Cache Issues
```bash
# Clear GitHub Actions cache
gh workflow run ci-working.yml --field clear_cache=true

# Local cache cleanup
docker builder prune --all
```

### Security Scan Failures
- Check the Security tab in GitHub repository
- Review SARIF reports in Actions logs
- Update base images if vulnerabilities are found
- Use ignore files for false positives

### Performance Issues
- Monitor image layer sizes
- Use `.dockerignore` to reduce build context
- Leverage multi-stage builds
- Enable BuildKit caching

## Best Practices

### Development Workflow
1. **Feature Development**: Use feature branch images for testing
2. **Pull Request Review**: Use PR images for integration testing  
3. **Main Branch**: Automatic latest tag deployment
4. **Releases**: Use semantic versioning tags for production

### Security Guidelines
1. **Regular Updates**: Keep base images updated
2. **Scan Results**: Address security findings promptly
3. **Secrets**: Never include secrets in images
4. **User Context**: Always run as non-root user

### Performance Optimization
1. **Layer Caching**: Optimize Dockerfile for layer reuse
2. **Multi-stage**: Use multi-stage builds to minimize size
3. **Build Context**: Keep build context minimal with `.dockerignore`
4. **Architecture**: Build for target architectures only when needed

## API and Automation

### GitHub CLI Integration
```bash
# List package versions
gh api -X GET /users/larsartmann/packages/container/template-arch-lint/versions

# Delete specific version
gh api -X DELETE /users/larsartmann/packages/container/template-arch-lint/versions/VERSION_ID
```

### Container API
```bash
# Health check
curl -f http://localhost:8080/health || exit 1

# Metrics
curl http://localhost:8080:2112/metrics

# Application endpoints
curl http://localhost:8080/
```

## Compliance and Governance

### Supply Chain Requirements
- ✅ SBOM (Software Bill of Materials) generated
- ✅ Build provenance attestation signed
- ✅ Container images signed with sigstore
- ✅ Vulnerability scanning integrated
- ✅ Base image tracking and updates

### Audit Trail
- All builds logged in GitHub Actions
- Container registry access logged
- Security scan results preserved
- Provenance data available for compliance

---

For additional support, please refer to:
- [Docker Documentation](DOCKER.md)
- [CI/CD Pipeline Documentation](CI_CD_DOCUMENTATION.md)
- [Security Best Practices](SECURITY_BEST_PRACTICES.md)
- [Project Issues](https://github.com/LarsArtmann/template-arch-lint/issues)