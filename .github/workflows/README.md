# GitHub Actions Workflows

This directory contains the CI/CD workflows for automated building, testing, and deployment.

## Workflows Overview

### `ci-working.yml` - Main CI/CD Pipeline
**Triggers**: Push to main/master/develop, Pull Requests, Manual dispatch  
**Purpose**: Complete CI/CD pipeline with container registry integration

**Jobs**:
1. **validate** - Basic validation and setup
2. **security** - CodeQL analysis and Trivy filesystem scanning
3. **build** - Go application building with templ generation
4. **quality** - Code linting and quality checks
5. **test** - Unit tests with coverage across Go versions
6. **integration** - Integration testing and configuration validation
7. **prepare-metadata** - Container image metadata and versioning
8. **container-security** - Container vulnerability scanning
9. **build-and-push** - Multi-architecture container builds and GHCR publishing
10. **test-container** - Container functionality testing
11. **validate-complete** - Final validation and reporting

**Container Features**:
- ✅ Multi-architecture builds (AMD64/ARM64)
- ✅ GitHub Container Registry (GHCR) integration
- ✅ Semantic versioning and intelligent tagging
- ✅ Comprehensive security scanning (Trivy + Grype)
- ✅ SBOM generation and build provenance
- ✅ Container signing with GitHub's sigstore

### `registry-cleanup.yml` - Container Registry Maintenance
**Triggers**: Daily schedule (2:00 AM UTC), Manual dispatch  
**Purpose**: Automated cleanup of container images

**Features**:
- **Retention Policy**: Keep latest 10 versions, preserve main/release images
- **PR Image Cleanup**: Remove PR images after 7 days
- **Untagged Cleanup**: Remove untagged images immediately
- **Dry Run Mode**: Preview changes before execution

## Environment Variables

```yaml
GO_VERSION: '1.24.0'
GOLANGCI_LINT_VERSION: 'v2.3.1'
REGISTRY: ghcr.io
IMAGE_NAME: ${{ github.repository }}
```

## Permissions Required

```yaml
permissions:
  contents: read          # Repository access
  security-events: write  # Security scan results
  pull-requests: write    # PR comments and checks
  checks: write          # Check suite results
  packages: write        # Container registry publishing
  attestations: write    # Build provenance attestations
  id-token: write        # OIDC token for signing
```

## Image Tags Generated

| Branch/Event | Tag Examples |
|--------------|--------------|
| Main branch | `latest`, `main`, `v1.0.0`, `main-a1b2c3d-20240101-120000` |
| Feature branch | `feature-name`, `feature-name-a1b2c3d-20240101-120000` |
| Pull request | `pr-123` |
| Release tag | `v1.0.0`, `v1.0`, `v1` |

## Manual Workflow Execution

### Trigger Main CI/CD Pipeline
```bash
# Using GitHub CLI
gh workflow run ci-working.yml

# Via GitHub web interface
# Navigate to Actions > CI/CD Pipeline (Working) > Run workflow
```

### Trigger Registry Cleanup
```bash
# Dry run mode (preview only)
gh workflow run registry-cleanup.yml --field dry_run=true

# Actual cleanup
gh workflow run registry-cleanup.yml --field dry_run=false
```

## Monitoring and Debugging

### View Workflow Runs
```bash
# List recent workflow runs
gh run list

# View specific run details
gh run view RUN_ID

# View logs for failed run
gh run view RUN_ID --log
```

### Check Container Images
```bash
# List published packages
gh api /users/larsartmann/packages/container/template-arch-lint/versions

# Pull and test locally
docker pull ghcr.io/larsartmann/template-arch-lint:latest
docker run --rm -p 8080:8080 ghcr.io/larsartmann/template-arch-lint:latest
```

## Security Integration

### Vulnerability Scanning
- **Trivy**: Scans filesystem and container images
- **Grype**: Additional vulnerability analysis
- **Results**: Uploaded to GitHub Security tab as SARIF reports

### Supply Chain Security
- **SBOM**: Complete software bill of materials generated
- **Provenance**: Build process attestation with GitHub's signing
- **Container Signing**: Images signed with sigstore/cosign

### Security Policies
- Fails build on CRITICAL/HIGH vulnerabilities
- Ignores unfixed vulnerabilities (configurable)
- Scans both source code and container images

## Workflow Customization

### Adding New Jobs
1. Add job definition to `ci-working.yml`
2. Update job dependencies in `needs:` arrays
3. Update final validation job dependencies
4. Test with workflow dispatch before merging

### Modifying Container Build
- Edit `build-and-push` job
- Update platform matrix for different architectures
- Modify tagging strategy in `prepare-metadata` job
- Adjust security scanning parameters

### Registry Configuration
- Modify `REGISTRY` and `IMAGE_NAME` environment variables
- Update authentication in `docker/login-action` steps
- Adjust cleanup policies in `registry-cleanup.yml`

## Best Practices

1. **Feature Development**: Create feature branches to test container builds
2. **Security**: Review security scan results before merging
3. **Testing**: Use PR images for integration testing
4. **Releases**: Use semantic versioning tags for production deployments
5. **Monitoring**: Check workflow status and container registry usage regularly

## Troubleshooting

### Common Issues

#### Authentication Errors
- Verify `GITHUB_TOKEN` has necessary permissions
- Check organization/repository settings for package publishing

#### Build Failures
- Check Docker build context size (use `.dockerignore`)
- Verify multi-architecture support is enabled
- Review build logs for specific error messages

#### Security Scan Failures
- Review vulnerability reports in Security tab
- Update dependencies to address security issues
- Consider ignoring false positives with proper justification

#### Registry Issues
- Monitor storage usage and cleanup policies
- Verify image tags are being generated correctly
- Check container registry permissions and quotas