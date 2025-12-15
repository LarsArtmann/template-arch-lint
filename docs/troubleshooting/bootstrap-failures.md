# Bootstrap Troubleshooting Guide

**Last Updated:** 2025-08-28  
**Version:** 1.0

This guide covers common bootstrap.sh failure scenarios and their solutions.

## 🚨 Quick Emergency Fixes

If bootstrap fails, try these immediate solutions:

```bash
# 1. Check you're in the right place
pwd && ls -la go.mod .git

# 2. Verify basic requirements
go version && git --version && curl --version

# 3. Clean retry with debug output
bash -x bootstrap.sh

# 4. Manual fallback installation
curl -fsSL -o .go-arch-lint.yml https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.go-arch-lint.yml
curl -fsSL -o .golangci.yml https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.golangci.yml
curl -fsSL -o justfile https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/justfile
```

---

## 📋 Complete Failure Scenarios & Solutions

### 1. Environment Validation Failures

#### 1.1 "Not in a git repository"

**Error:** `❌ Not in a git repository. Please run from the root of your Go project.`

**Cause:** Bootstrap must run from the root of a git-initialized Go project.

**Solution:**

```bash
# Check current location
pwd
ls -la

# Initialize git if needed
git init
git add .
git commit -m "Initial commit"

# Then retry bootstrap
./bootstrap.sh
```

**Prevention:** Always run bootstrap from your project root directory.

---

#### 1.2 "No go.mod found"

**Error:** `❌ No go.mod found. Please run from the root of a Go project.`

**Cause:** Bootstrap requires a valid Go module.

**Solution:**

```bash
# Initialize Go module
go mod init your-project-name

# Add basic main.go if needed
cat > main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
EOF

# Then retry bootstrap
./bootstrap.sh
```

---

#### 1.3 "Missing required commands"

**Error:** `❌ Missing required commands: [curl|git|go]`

**Diagnostic Commands:**

```bash
# Check what's missing
command -v go || echo "Go missing"
command -v git || echo "Git missing"
command -v curl || echo "Curl missing"
```

**Solutions by Platform:**

**macOS:**

```bash
# Install missing tools
brew install go git curl
```

**Ubuntu/Debian:**

```bash
sudo apt-get update
sudo apt-get install -y golang-go git curl
```

**RHEL/CentOS/Fedora:**

```bash
sudo yum install -y golang git curl
# or for newer versions:
sudo dnf install -y golang git curl
```

---

#### 1.4 "Go version too old"

**Error:** `❌ Go version X.X is too old. Minimum required: 1.19`

**Diagnostic:**

```bash
go version
```

**Solution:**

```bash
# Remove old Go
sudo rm -rf /usr/local/go

# Install latest Go (Linux)
wget https://golang.org/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Install latest Go (macOS)
brew install go

# Verify
go version
```

---

#### 1.5 "Unsupported platform"

**Error:** `⚠️ Unsupported platform for automatic just installation`

**Diagnostic:**

```bash
uname -s  # Operating system
uname -m  # Architecture
```

**Solution:**

```bash
# Manual just installation
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin
export PATH="$HOME/.local/bin:$PATH"

# Verify
just --version
```

---

### 2. Tool Installation Failures

#### 2.1 "Failed to install just command runner"

**Common Causes:**

- Homebrew not available on macOS
- Package manager failures on Linux
- Network connectivity issues

**Diagnostic Commands:**

```bash
# Check if just is available
command -v just

# Check brew (macOS)
command -v brew

# Check package managers (Linux)
command -v apt-get
command -v yum
command -v dnf
```

**Solutions:**

**macOS without Homebrew:**

```bash
# Install Homebrew first
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Then install just
brew install just
```

**Alternative Installation (any platform):**

```bash
# Direct installation
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin
export PATH="$HOME/.local/bin:$PATH"
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc  # or ~/.zshrc
```

---

#### 2.2 "Failed to install linting tools via 'just install'"

**Error:** `❌ Failed to install linting tools via 'just install'`

**Diagnostic Commands:**

```bash
# Test just functionality
just --version
just --list

# Check if justfile is present and valid
ls -la justfile
just help

# Test Go tool installation manually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/fe3dback/go-arch-lint@latest
```

**Common Solutions:**

**Network/Proxy Issues:**

```bash
# Configure Go proxy
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

# Retry installation
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

**Disk Space Issues:**

```bash
# Check disk space
df -h

# Clean Go cache
go clean -cache -modcache

# Retry installation
just install
```

**Permission Issues:**

```bash
# Check Go bin directory permissions
ls -la ~/go/bin
mkdir -p ~/go/bin
chmod 755 ~/go/bin

# Retry installation
just install
```

---

### 3. Configuration File Download Failures

#### 3.1 "Failed to download [config-file]"

**Common Causes:**

- Network connectivity issues
- GitHub rate limiting
- Corporate firewall blocking raw.githubusercontent.com

**Diagnostic Commands:**

```bash
# Test network connectivity
curl -I https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.go-arch-lint.yml

# Test with verbose output
curl -v https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.golangci.yml

# Check DNS resolution
nslookup raw.githubusercontent.com
```

**Solutions:**

**Network Issues:**

```bash
# Retry with different timeout
curl --max-time 30 -fsSL -o .go-arch-lint.yml https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.go-arch-lint.yml

# Use alternative method via git
git clone --depth 1 https://github.com/LarsArtmann/template-arch-lint.git temp-configs
cp temp-configs/.go-arch-lint.yml .
cp temp-configs/.golangci.yml .
cp temp-configs/justfile .
rm -rf temp-configs
```

**Corporate Firewall:**

```bash
# Use proxy if available
export https_proxy=your-proxy-url
export http_proxy=your-proxy-url

# Download files through proxy
curl -fsSL -o .go-arch-lint.yml https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.go-arch-lint.yml
```

---

#### 3.2 "Downloaded [file] is empty"

**Cause:** Partial download or network interruption

**Diagnostic:**

```bash
# Check file sizes
ls -la .go-arch-lint.yml .golangci.yml justfile

# Expected approximate sizes:
# .go-arch-lint.yml: ~1-3 KB
# .golangci.yml: ~15-25 KB
# justfile: ~8-15 KB
```

**Solution:**

```bash
# Remove corrupted files and re-download
rm -f .go-arch-lint.yml .golangci.yml justfile

# Re-download with retries
for file in .go-arch-lint.yml .golangci.yml justfile; do
  echo "Downloading $file..."
  curl --retry 3 --retry-delay 2 -fsSL -o "$file" "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/$file"
  if [ ! -s "$file" ]; then
    echo "❌ $file is empty, retrying..."
    sleep 2
    curl -fsSL -o "$file" "https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/$file"
  fi
done
```

---

### 4. PATH Setup Issues

#### 4.1 "Go tools directory does not exist yet"

**Warning:** `⚠️ Go tools directory does not exist yet: ~/go/bin`

**This is usually normal** during first-time installation.

**Diagnostic:**

```bash
# Check if directory exists
ls -la ~/go/bin

# Check GOPATH
go env GOPATH
```

**Manual Fix (if needed):**

```bash
# Create directory
mkdir -p ~/go/bin

# Ensure GOPATH is set correctly
export GOPATH=$(go env GOPATH)
export PATH="$GOPATH/bin:$PATH"
```

---

#### 4.2 "Could not detect shell profile to make PATH persistent"

**Warning:** Shell profile detection failed

**Manual Solution:**

```bash
# Add to your shell profile manually
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc   # for Bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc    # for Zsh

# Source the profile
source ~/.bashrc  # or ~/.zshrc

# Verify
echo $PATH | grep -q go/bin && echo "✅ PATH updated" || echo "❌ PATH not updated"
```

---

#### 4.3 "Tools installed but not accessible"

**Issue:** `golangci-lint` or `go-arch-lint` command not found after installation

**Diagnostic Commands:**

```bash
# Check if tools were installed
ls -la ~/go/bin/golangci-lint ~/go/bin/go-arch-lint

# Check current PATH
echo $PATH | grep -q go/bin && echo "✅ ~/go/bin in PATH" || echo "❌ ~/go/bin NOT in PATH"

# Test direct execution
~/go/bin/golangci-lint --version
~/go/bin/go-arch-lint --help
```

**Immediate Fix:**

```bash
# Add to current session
export PATH="$HOME/go/bin:$PATH"

# Test
golangci-lint --version
go-arch-lint --help

# Make permanent (choose your shell)
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc   # Bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc    # Zsh
```

---

### 5. Verification Failures

#### 5.1 "Justfile is not working"

**Error:** `❌ Justfile verification failed`

**Diagnostic Commands:**

```bash
# Test just command
just --version

# Test justfile syntax
just --list

# Check justfile exists
ls -la justfile
```

**Solutions:**

```bash
# Re-download justfile
curl -fsSL -o justfile https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/justfile

# Make executable if needed (usually not required)
chmod +x justfile

# Test again
just --list
```

---

#### 5.2 "Architecture validation had issues"

**Warning:** Architecture validation failed (may be normal for new projects)

**Diagnostic:**

```bash
# Check if you have Go files
find . -name "*.go" -not -path "./vendor/*" | head -5

# Test architecture validation manually
just lint-arch

# Check go-arch-lint config
ls -la .go-arch-lint.yml
```

**Common Solutions:**

**For New Projects (Normal):**

```bash
# Create basic Go structure that passes validation
mkdir -p internal/domain internal/application internal/infrastructure
echo 'package domain' > internal/domain/domain.go
echo 'package application' > internal/application/app.go
echo 'package infrastructure' > internal/infrastructure/infra.go

# Test again
just lint-arch
```

**For Existing Projects:**

```bash
# Check specific violations
go-arch-lint check --config .go-arch-lint.yml

# Common fixes:
# - Move business logic to internal/domain/
# - Move HTTP handlers to internal/application/
# - Move database code to internal/infrastructure/
```

---

#### 5.3 "Some verification tests had issues"

**Warning:** Partial verification failure

**This usually means the installation is functional** but some optional features didn't work.

**Diagnostic Steps:**

```bash
# Test each component individually
echo "Testing just..."
just --version

echo "Testing golangci-lint..."
golangci-lint --version

echo "Testing go-arch-lint..."
go-arch-lint --help

echo "Testing basic functionality..."
just help
```

**Manual Verification:**

```bash
# Test complete workflow
just lint          # Should run without errors
just format        # Should format code
just help          # Should show available commands
```

---

## 🔄 Complete Recovery Procedures

### Full Reset & Retry

```bash
# 1. Clean up any partial installation
rm -f .go-arch-lint.yml .golangci.yml justfile
rm -f .go-arch-lint.yml.backup .golangci.yml.backup justfile.backup

# 2. Ensure clean environment
unset GOPATH GOPROXY GOSUMDB
export PATH=$(echo $PATH | sed 's|:'$HOME'/go/bin||g' | sed 's|'$HOME'/go/bin:||g')

# 3. Verify requirements
go version
git --version
curl --version

# 4. Fresh download and retry
curl -fsSL -o bootstrap.sh https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/bootstrap.sh
chmod +x bootstrap.sh
./bootstrap.sh
```

### Manual Installation Fallback

```bash
# Download configs
curl -fsSL -o .go-arch-lint.yml https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.go-arch-lint.yml
curl -fsSL -o .golangci.yml https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/.golangci.yml
curl -fsSL -o justfile https://raw.githubusercontent.com/LarsArtmann/template-arch-lint/master/justfile

# Install just
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/.local/bin
export PATH="$HOME/.local/bin:$PATH"

# Install linting tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/fe3dback/go-arch-lint@latest

# Add to PATH
export PATH="$HOME/go/bin:$PATH"
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc

# Test
just lint
```

---

## 📞 Getting Help

If none of these solutions work:

1. **Debug Mode:** Run `bash -x bootstrap.sh` to see detailed execution
2. **GitHub Issues:** Report at https://github.com/LarsArtmann/template-arch-lint/issues
3. **Include This Info:**
   - Operating system & version (`uname -a`)
   - Go version (`go version`)
   - Error message (exact text)
   - Output of `bash -x bootstrap.sh`

---

## 🎯 Prevention Checklist

Before running bootstrap.sh:

- [ ] In git repository root (`ls .git`)
- [ ] Go module present (`ls go.mod`)
- [ ] Go 1.19+ installed (`go version`)
- [ ] Network connectivity (`curl -I google.com`)
- [ ] Sufficient disk space (`df -h`)
- [ ] Write permissions in directory (`touch test-file && rm test-file`)

This should eliminate 90%+ of bootstrap failures.
