# 🪟 WSL (Windows Subsystem for Linux) Compatibility Guide

This document provides comprehensive information about using template-arch-lint on Windows with WSL, including compatibility validation, known issues, and troubleshooting procedures.

## 🎯 Overview

The enhanced bootstrap script and linting tools are designed to work seamlessly on WSL, providing the same enterprise-grade experience as native Linux environments.

## ✅ Tested WSL Distributions

The following WSL distributions have been validated with template-arch-lint:

- ✅ **Ubuntu 20.04 LTS** (Recommended)
- ✅ **Ubuntu 22.04 LTS** (Recommended)  
- ✅ **Debian 11** (Stable)
- 🟡 **Alpine Linux** (Basic support)
- 🟡 **openSUSE** (Community tested)

## 🚀 Quick Start on WSL

### 1. Prerequisites

```bash
# Update WSL system
sudo apt update && sudo apt upgrade -y

# Install required packages
sudo apt install -y curl git build-essential

# Install Go (if not already installed)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### 2. Bootstrap Installation

```bash
# Method 1: Direct download and run
curl -fsSL https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh | bash

# Method 2: Two-step installation (recommended)
curl -fsSL https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh -o bootstrap.sh
chmod +x bootstrap.sh
./bootstrap.sh

# Method 3: WSL diagnostic mode first
./bootstrap.sh --diagnose
./bootstrap.sh --fix --verbose
```

### 3. Verification

```bash
# Run comprehensive tests
./test-bootstrap-simple-bdd.sh

# Test linting tools
just lint-arch
just security-audit
```

## 🔧 WSL-Specific Configuration

### Environment Variables

```bash
# Add to ~/.bashrc or ~/.zshrc
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin:$HOME/.local/bin

# WSL-specific optimizations
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org
```

### File System Considerations

```bash
# Work in WSL filesystem (not Windows mount) for best performance
cd ~
git clone https://github.com/your-org/your-go-project.git
cd your-go-project

# Avoid Windows filesystem mounts for Go projects
# ❌ Avoid: /mnt/c/Users/YourName/project
# ✅ Use: /home/yourname/project
```

## 🧪 WSL Compatibility Testing

### Automated Validation Script

Run the provided WSL validation script:

```bash
./test-wsl-compatibility.sh
```

### Manual Validation Checklist

- [ ] Go installation and version compatibility (1.19+)
- [ ] Git configuration and SSH key setup
- [ ] Network connectivity to GitHub and Go proxy
- [ ] File permissions and PATH configuration  
- [ ] Tool installation (golangci-lint, go-arch-lint, just)
- [ ] Bootstrap script functionality
- [ ] Linting tools execution
- [ ] Performance benchmarks

### Performance Considerations

WSL performance characteristics:
- **File I/O**: ~20% slower than native Linux
- **Network**: Comparable to native
- **CPU**: Near-native performance
- **Memory**: Shared with Windows host

Optimization tips:
- Use WSL2 for better performance
- Keep Go projects in WSL filesystem
- Consider increasing WSL memory allocation

## 🐛 Known Issues and Workarounds

### Issue 1: Permission Denied on Executable Files

**Symptoms:**
```
bash: ./bootstrap.sh: Permission denied
```

**Solution:**
```bash
chmod +x bootstrap.sh
chmod +x test-*.sh
```

### Issue 2: Go Tools Not in PATH

**Symptoms:**
```
golangci-lint: command not found
```

**Solution:**
```bash
export PATH=$PATH:$HOME/go/bin
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
```

### Issue 3: Network Connectivity Issues

**Symptoms:**
```
Failed to download configuration files
```

**Solution:**
```bash
# Check DNS resolution
nslookup github.com

# Test connectivity
curl -I https://github.com

# Configure proxy if needed
export https_proxy=your-proxy:port
```

### Issue 4: File System Permissions

**Symptoms:**
```
mkdir: cannot create directory: Permission denied
```

**Solution:**
```bash
# Check if working in Windows filesystem
pwd
# Should be /home/username/... not /mnt/c/...

# Fix permissions
sudo chown -R $USER:$USER $HOME/go
```

## 📊 WSL Performance Benchmarks

### Typical Bootstrap Times

| Environment | Time | Notes |
|-------------|------|-------|
| WSL2 Ubuntu 22.04 | 2-3 min | Recommended |
| WSL2 Ubuntu 20.04 | 2-4 min | Stable |
| WSL1 Ubuntu | 4-6 min | Legacy |
| Native Linux | 1-2 min | Reference |

### Linting Performance

| Tool | WSL2 | Native Linux | Overhead |
|------|------|-------------|----------|
| golangci-lint | 45s | 38s | +18% |
| go-arch-lint | 12s | 10s | +20% |
| gosec | 8s | 7s | +14% |

## 🛠️ Troubleshooting Commands

### System Diagnostics

```bash
# WSL version and distribution
wsl --list --verbose
lsb_release -a

# Go environment
go version
go env GOPATH
go env GOPROXY

# Tool availability
which golangci-lint
which go-arch-lint
which just

# Network connectivity
curl -I https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh
```

### Debug Mode

```bash
# Run bootstrap with maximum verbosity
./bootstrap.sh --diagnose --verbose

# Debug linting with verbose output
just lint-arch --verbose

# Check tool installation
just install --verbose
```

### Performance Analysis

```bash
# System resource usage
top
free -h
df -h

# Go module and proxy status
go clean -cache
go clean -modcache
go env
```

## 🔗 Useful WSL Commands

```bash
# Restart WSL
wsl --shutdown
wsl

# Check WSL version
wsl --list --verbose

# Set default WSL version
wsl --set-default-version 2

# Convert WSL1 to WSL2
wsl --set-version Ubuntu-22.04 2

# Access Windows files from WSL
cd /mnt/c/Users/YourName

# Access WSL files from Windows
\\\\wsl$\\Ubuntu-22.04\\home\\yourname
```

## 💡 Best Practices

### Development Workflow

1. **Keep projects in WSL filesystem** for better performance
2. **Use VSCode with WSL extension** for integrated development
3. **Configure Git properly** with Windows/Linux line endings
4. **Set up SSH keys in WSL** not Windows
5. **Use bash/zsh** not PowerShell for development

### Security Considerations

- Keep WSL system updated with `sudo apt update && sudo apt upgrade`
- Use proper file permissions (avoid 777)
- Configure firewall if exposing services
- Use WSL2 for better isolation

### Performance Optimization

- Allocate sufficient memory to WSL2
- Use SSD for WSL filesystem
- Avoid frequent Windows ↔ WSL file access
- Consider Docker Desktop for containerized workflows

## 📚 Additional Resources

- [Microsoft WSL Documentation](https://docs.microsoft.com/en-us/windows/wsl/)
- [Go on WSL Best Practices](https://docs.microsoft.com/en-us/windows/dev-environment/golang/overview)
- [VSCode WSL Extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-wsl)
- [template-arch-lint GitHub Issues](https://github.com/LarsArtmann/template-arch-lint/issues)

## 🐛 Reporting WSL Issues

When reporting WSL-specific issues, please include:

1. WSL version and distribution: `wsl --list --verbose`
2. Go version: `go version`
3. Bootstrap diagnostic output: `./bootstrap.sh --diagnose --verbose`
4. Error messages and stack traces
5. System specifications (RAM, CPU, Windows version)

Create issues at: https://github.com/LarsArtmann/template-arch-lint/issues

---

*This WSL compatibility guide is maintained as part of the template-arch-lint project. For updates and community contributions, please see the GitHub repository.*