# Security Best Practices Guide

This guide outlines security best practices for the template-arch-lint application to maintain production-grade security.

## Quick Security Checklist

### Daily Operations
- [ ] Monitor application logs for security events
- [ ] Check system resource usage and anomalies
- [ ] Verify backup integrity

### Weekly Tasks
- [ ] Run vulnerability scan: `govulncheck ./...`
- [ ] Review security logs and alerts
- [ ] Update dependencies if patches available
- [ ] Test security configurations

### Monthly Tasks
- [ ] Comprehensive dependency audit
- [ ] Security configuration review
- [ ] Update Go version if new release available
- [ ] Review and update security documentation

## Essential Security Commands

### Vulnerability Scanning
```bash
# Install vulnerability scanner
go install golang.org/x/vuln/cmd/govulncheck@latest

# Scan for vulnerabilities
govulncheck ./...

# Detailed vulnerability report
govulncheck -show verbose ./...
```

### Dependency Management
```bash
# Check for outdated dependencies
go list -u -m all

# Update dependencies
go get -u ./...
go mod tidy

# Verify module checksums
go mod verify
```

### Security Testing
```bash
# Run all tests
go test ./...

# Run tests with race detection
go test -race ./...

# Generate test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Configuration Security

### Environment Variables
Never hardcode sensitive values. Use environment variables:
```bash
export APP_DATABASE_DSN="your-secure-dsn"
export APP_SERVER_HOST="localhost"
export APP_LOGGING_LEVEL="info"
```

### File Permissions
Secure configuration files:
```bash
chmod 600 config.yaml        # Read/write for owner only
chmod 640 config.production.yaml  # Read for group, write for owner
```

### Database Security
- Use parameterized queries (already implemented via SQLC)
- Implement connection pooling limits
- Regular backup and recovery testing
- Encrypt sensitive data at rest

## Observability Security

### Logging Guidelines
- Never log sensitive information (passwords, tokens, PII)
- Use structured logging for security analysis
- Implement log rotation and archiving
- Monitor log integrity

### Metrics Collection
- Monitor failed authentication attempts
- Track unusual traffic patterns
- Alert on error rate spikes
- Monitor resource exhaustion

## Deployment Security

### Build Process
```bash
# Secure build with vulnerability check
govulncheck ./... && go build -ldflags="-s -w" ./cmd/server
```

### Docker Security (if using containers)
```dockerfile
# Use minimal base images
FROM scratch
# Run as non-root user
USER 65534:65534
# Set security headers
ENV CGO_ENABLED=0
```

### Production Deployment
- Use HTTPS only (TLS 1.3+)
- Implement request rate limiting
- Set security headers
- Configure firewall rules
- Regular security patches

## Incident Response

### Security Event Response
1. **Immediate**: Isolate affected systems
2. **Assessment**: Determine scope and impact
3. **Containment**: Stop active threats
4. **Recovery**: Restore secure operations
5. **Lessons**: Document and improve

### Contact Information
- **Security Team**: security@company.com
- **On-call Engineer**: +1-xxx-xxx-xxxx
- **Incident Commander**: incident-commander@company.com

## Compliance & Auditing

### Audit Trail
- All configuration changes logged
- Code changes require review
- Deployment activities tracked
- Security events recorded

### Regular Reviews
- **Code Reviews**: Security-focused peer review
- **Dependency Audits**: Monthly vulnerability scans
- **Configuration Reviews**: Quarterly security settings review
- **Penetration Testing**: Annual third-party assessment

## Tools & Resources

### Security Tools
- `govulncheck` - Go vulnerability scanner
- `nancy` - OSS Index vulnerability scanner
- `gosec` - Go security checker
- `staticcheck` - Go static analysis

### External Resources
- [Go Security Policy](https://golang.org/security/)
- [OWASP Go Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Go_SCC.html)
- [Go Vulnerability Database](https://vuln.go.dev/)

## Emergency Procedures

### Security Incident
```bash
# 1. Immediate containment
systemctl stop template-arch-lint

# 2. Preserve evidence
cp -r /var/log/template-arch-lint /tmp/incident-$(date +%Y%m%d)

# 3. Notify security team
curl -X POST https://alerts.company.com/security-incident \
  -d "service=template-arch-lint&severity=high&status=active"

# 4. Follow incident response plan
```

### Zero-Day Vulnerability
1. **Monitor**: Watch Go security announcements
2. **Assess**: Evaluate impact on application
3. **Plan**: Prepare update strategy
4. **Execute**: Deploy patches immediately
5. **Verify**: Confirm vulnerability resolution

---

**Remember**: Security is everyone's responsibility. When in doubt, consult the security team.