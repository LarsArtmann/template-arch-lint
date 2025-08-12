# 🔧 GitHub Configuration

This directory contains all GitHub-specific configuration files for the template-arch-lint project.

## 📁 Directory Structure

```
.github/
├── workflows/
│   ├── ci.yml           # Complete CI/CD pipeline (architecture linting enabled)
│   ├── ci-working.yml   # Working CI/CD pipeline (architecture linting disabled)
│   └── env.yml          # Shared environment configuration
├── dependabot.yml       # Automated dependency updates
├── CODEOWNERS          # Code ownership and review assignments
└── README.md           # This file
```

## 🔄 Workflows

### ci.yml - Complete CI/CD Pipeline
**Status**: Contains architecture linting issues
**Features**:
- Full security scanning (CodeQL + Trivy)
- Code quality enforcement (golangci-lint)
- Architecture validation (go-arch-lint) - Currently failing
- Comprehensive testing with coverage
- Build verification and artifact generation

### ci-working.yml - Working CI/CD Pipeline
**Status**: ✅ Fully functional
**Features**:
- All features of ci.yml except architecture linting
- Recommended for current use until architecture issues are resolved
- Production-ready quality gates

### env.yml - Environment Configuration
**Purpose**: Shared environment variables and configuration
**Contents**:
- Tool versions and settings
- Timeout configurations
- Cache settings
- Quality gate thresholds

## 🤖 Dependabot Configuration

Automated dependency management:
- **Go modules**: Weekly updates on Mondays
- **GitHub Actions**: Weekly updates on Mondays
- **Grouping**: Minor and patch updates grouped together
- **Reviewers**: Automatically assigned to @LarsArtmann

## 👥 Code Owners

Automatic review assignment for:
- CI/CD configuration files
- Core application code
- Infrastructure and database files
- Configuration and templates

## 🚀 Quick Start

1. **For new development**: Use `ci-working.yml` workflow
2. **For production**: Ensure all quality gates pass
3. **For security**: Review Trivy and CodeQL scan results
4. **For dependencies**: Monitor Dependabot PRs

## 🔧 Maintenance

### Updating Tool Versions
Edit `env.yml` to update:
- Go version
- golangci-lint version
- Other tool versions

### Modifying Quality Gates
Edit the respective workflow files to adjust:
- Coverage thresholds
- Timeout settings
- Security scan sensitivity

### Adding New Workflows
1. Create new `.yml` file in `workflows/`
2. Follow existing naming conventions
3. Include proper permissions and timeouts
4. Test thoroughly before merging

## 📊 Monitoring

- **Workflow Status**: GitHub Actions tab
- **Security Alerts**: Security tab
- **Dependency Updates**: Pull requests from Dependabot
- **Code Coverage**: Codecov integration

## 🐛 Troubleshooting

### Common Issues
1. **Workflow failures**: Check logs in Actions tab
2. **Permission errors**: Verify permissions in workflow files
3. **Tool version conflicts**: Update versions in env.yml
4. **Cache issues**: Clear workflow caches if needed

### Debug Commands
```bash
# Test locally
just ci              # Full CI simulation
just lint            # Code quality only
just test            # Tests only
just build           # Build only
```

---

**Maintained by**: Development Team
**Last Updated**: August 12, 2025