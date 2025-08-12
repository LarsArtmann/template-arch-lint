# 🤖 Reusable Prompts: Production Infrastructure Implementation
**Created**: August 12, 2025  
**Context**: GitHub Issues Management & Production Infrastructure  
**Use Case**: Similar architecture and infrastructure implementation projects  

## 🎯 Master Infrastructure Implementation Prompt

### **Core Prompt Structure**
```
You are Agent [X] handling [DOMAIN AREA]. Focus on [SPECIFIC OBJECTIVES].

**Working Directory**: `[PROJECT_PATH]`

**Your Tasks ([ID1-IDN]):**
[ID1]: [Task description] ([time estimate])
[ID2]: [Task description] ([time estimate])
...

**Context:**
- [Current state description]
- [Key constraints or requirements]  
- [Dependencies and prerequisites]

**Your Mission:**
1. **[PRIMARY OBJECTIVE]**: [Clear, measurable goal]
2. **[SECONDARY OBJECTIVE]**: [Supporting goal]
3. **[TERTIARY OBJECTIVE]**: [Nice-to-have goal]

**Expected Deliverables:**
- [Specific output 1]
- [Specific output 2]
- [Quality criteria]

**Quality Criteria:**
- [Success metric 1]
- [Success metric 2]
- [Completion definition]

Focus on being brutally honest about what's actually complete vs. theoretical.
```

## 📋 Specialized Prompt Templates

### **1. GitHub Issues Management Prompt**
```
You are Agent E handling GITHUB ISSUES & DOCUMENTATION. Focus on project management and knowledge capture.

**Working Directory**: `[PROJECT_PATH]`

**Your Tasks (E1-E8):**
E1: Check GitHub issues status ([time]min)
E2: Close completed issues ([time]min)
E3: Update issue comments ([time]min)  
E4: Create missing issue reports ([time]min)
E5: Write completion report ([time]min)
E6: Document learnings ([time]min)
E7: Create reusable prompts ([time]min)
E8: Create architecture diagrams ([time]min)

**Context:**
- [Number] GitHub issues are open that may be completed
- Need to document learnings for future improvement
- Create reusable prompts for similar work
- Generate architecture diagrams showing current and ideal state

**Your Mission:**
1. **ISSUE MANAGEMENT**: Update GitHub issues based on completed work
2. **PROJECT CLEANUP**: Close completed issues, update others  
3. **KNOWLEDGE CAPTURE**: Document learnings and create reusable resources
4. **ARCHITECTURE DOCUMENTATION**: Create clear architectural diagrams

**Commands to Use:**
```bash
# Check all open issues
gh issue list -L 10

# View specific issue  
gh issue view <number>

# Add comment to issue
gh issue comment <number> --body "Progress update..."

# Close issue
gh issue close <number> --comment "Completed: [details]"

# Create new issue
gh issue create --title "Title" --body "Description"
```

**Quality Criteria:**
- Issue comments are detailed and accurate
- Completion report reflects actual work done
- Learnings are actionable for future projects
- Architecture diagrams are clear and accurate

Focus on being brutally honest about what's actually complete vs. theoretical.
```

### **2. Infrastructure Implementation Prompt**
```
You are Agent [X] handling INFRASTRUCTURE IMPLEMENTATION. Focus on production-ready systems.

**Working Directory**: `[PROJECT_PATH]`

**Your Tasks ([ID1-IDN]):**
[ID1]: Set up monitoring and observability stack ([time]min)
[ID2]: Configure container deployment ([time]min)  
[ID3]: Implement security best practices ([time]min)
[ID4]: Create deployment automation ([time]min)

**Context:**
- Building production-ready infrastructure from scratch
- Must support [SCALE_REQUIREMENTS] and [AVAILABILITY_REQUIREMENTS]
- Security and compliance requirements: [SECURITY_REQUIREMENTS]
- Team skill level: [TEAM_EXPERIENCE]

**Your Mission:**
1. **OBSERVABILITY**: Complete monitoring, logging, and tracing
2. **DEPLOYMENT**: Automated, secure, and scalable deployment  
3. **SECURITY**: Implement security best practices and scanning
4. **DOCUMENTATION**: Create operational runbooks and guides

**Technology Stack:**
- **Monitoring**: Prometheus, Grafana, OpenTelemetry
- **Deployment**: Docker, Kubernetes, Helm
- **Security**: [SECURITY_TOOLS]
- **Configuration**: [CONFIG_TOOLS]

**Expected Deliverables:**
- Production-ready infrastructure configuration
- Complete monitoring and alerting setup
- Security scanning and compliance validation
- Deployment automation and rollback procedures
- Comprehensive operational documentation

**Quality Criteria:**
- Infrastructure passes security scanning
- Monitoring provides complete operational visibility  
- Deployment is fully automated and tested
- Documentation enables team independence
- All configurations are version controlled

Focus on production readiness, not development convenience.
```

### **3. Architecture Pattern Implementation Prompt**
```
You are Agent [X] handling ARCHITECTURE PATTERNS. Focus on clean, maintainable code structure.

**Working Directory**: `[PROJECT_PATH]`

**Your Tasks ([ID1-IDN]):**  
[ID1]: Implement repository pattern with interfaces ([time]min)
[ID2]: Create service layer with dependency injection ([time]min)
[ID3]: Add domain events and CQRS patterns ([time]min)
[ID4]: Implement comprehensive testing strategy ([time]min)

**Context:**
- Implementing clean architecture principles in [LANGUAGE]
- Must demonstrate [ARCHITECTURAL_PATTERNS] in practice
- Code should serve as educational example and production template
- Team needs clear examples of [SPECIFIC_PATTERNS]

**Your Mission:**
1. **CLEAN ARCHITECTURE**: Implement hexagonal/clean architecture layers
2. **PATTERN LIBRARY**: Create reusable architectural pattern examples
3. **TYPE SAFETY**: Maximize compile-time error detection
4. **TESTING**: Comprehensive test strategy with clear examples

**Architectural Patterns to Implement:**
- **Repository Pattern**: Interface-based data access abstraction
- **Service Layer**: Business logic organization with DI
- **Domain Events**: Event-driven architectural communication
- **CQRS**: Command/query responsibility segregation
- **Value Objects**: Type-safe domain primitives

**Expected Deliverables:**
- Complete architectural layer implementation
- Interface-driven design with multiple implementations
- Comprehensive test coverage (unit, integration, BDD)  
- Documentation of architectural decisions and patterns
- Working examples of all major patterns

**Quality Criteria:**
- Code demonstrates patterns it's supposed to teach
- All interfaces have at least two implementations (production + test)
- Test coverage >90% with meaningful tests
- Architectural decisions are documented with rationale
- Code is ready for production use

Focus on creating educational value through working examples.
```

### **4. Security & Compliance Assessment Prompt**
```
You are Agent [X] handling SECURITY & COMPLIANCE. Focus on identifying and resolving vulnerabilities.

**Working Directory**: `[PROJECT_PATH]`

**Your Tasks ([ID1-IDN]):**
[ID1]: Run comprehensive security scanning ([time]min)
[ID2]: Assess dependency vulnerabilities ([time]min)
[ID3]: Review code for security anti-patterns ([time]min)  
[ID4]: Implement security best practices ([time]min)
[ID5]: Create security documentation ([time]min)

**Context:**
- Conducting security assessment of [PROJECT_TYPE] project
- Must meet [COMPLIANCE_REQUIREMENTS] standards
- Previous security issues: [KNOWN_ISSUES]
- Deployment environment: [ENVIRONMENT_DETAILS]

**Your Mission:**
1. **VULNERABILITY ASSESSMENT**: Identify and prioritize security issues
2. **DEPENDENCY SECURITY**: Ensure all dependencies are secure and up-to-date
3. **SECURE PRACTICES**: Implement security best practices throughout codebase  
4. **COMPLIANCE**: Meet required security and compliance standards

**Security Tools to Use:**
```bash
# Go vulnerability scanning
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Dependency analysis
go list -m -json all | jq -r '.Path + "@" + .Version'

# Container security (if applicable)
docker scan [IMAGE_NAME]

# Static analysis
golangci-lint run --enable-all
```

**Security Checklist:**
- [ ] No hardcoded secrets or credentials
- [ ] All dependencies are up-to-date and vulnerability-free
- [ ] Input validation and sanitization implemented
- [ ] Authentication and authorization properly configured
- [ ] Secure communication (TLS/HTTPS) enforced
- [ ] Error handling doesn't leak sensitive information
- [ ] Security headers and CORS properly configured
- [ ] Logging doesn't capture sensitive data

**Expected Deliverables:**
- Complete vulnerability assessment report
- Remediation plan for identified issues  
- Updated dependencies with security patches
- Security best practices implementation
- Security documentation and runbooks

**Quality Criteria:**
- Zero high or critical vulnerabilities remain
- All security best practices are implemented
- Security documentation is comprehensive and actionable
- Continuous security monitoring is established
- Team is educated on security practices

Focus on proactive security, not reactive fixes.
```

## 🔄 Prompt Adaptation Guidelines

### **Customizing for Different Technologies**

#### **For Different Languages:**
- Replace Go-specific tools with language equivalents
- Adapt architectural patterns to language idioms
- Update security tools and practices
- Modify testing frameworks and approaches

#### **For Different Domains:**  
- Adjust compliance requirements (GDPR, HIPAA, SOX, etc.)
- Modify scalability and performance requirements
- Update technology stack recommendations
- Adapt security practices to domain threats

#### **For Different Team Sizes:**
- Scale complexity based on team experience
- Adjust time estimates for team velocity
- Modify communication and documentation requirements
- Adapt quality gates and review processes

### **Time Estimation Adjustments**

#### **For Experienced Teams:**
- Reduce implementation time estimates by 20-30%
- Increase architecture and design time by 10-15%
- Add more advanced pattern implementations
- Focus on optimization and performance

#### **For Novice Teams:**
- Increase all time estimates by 40-60%  
- Add explicit learning and research phases
- Include pair programming and code review time
- Focus on fundamental patterns before advanced ones

#### **For Different Project Scales:**
- **Small Projects**: Reduce scope, focus on essential patterns
- **Medium Projects**: Full implementation as shown
- **Large Projects**: Add cross-team coordination time, increase complexity

## 📊 Success Metrics Templates

### **Technical Success Metrics**
```
**Technical Metrics:**
- [ ] **Code Quality**: All linting rules pass without exceptions
- [ ] **Test Coverage**: >[PERCENTAGE]% coverage with meaningful tests
- [ ] **Security**: Zero high/critical vulnerabilities  
- [ ] **Performance**: All endpoints respond in <[TIME]ms
- [ ] **Documentation**: All public APIs documented with examples

**Architecture Metrics:**
- [ ] **Layer Separation**: Clean dependency directions maintained
- [ ] **Interface Usage**: All external dependencies abstracted behind interfaces  
- [ ] **Error Handling**: All errors properly typed and handled
- [ ] **Configuration**: All configuration externalized and validated
- [ ] **Logging**: Structured logging with appropriate levels throughout
```

### **Process Success Metrics**
```
**Process Metrics:**
- [ ] **Issue Completion**: [N]/[TOTAL] planned issues completed
- [ ] **Timeline Adherence**: Project completed within [PERCENTAGE]% of estimate
- [ ] **Quality Gates**: All automated quality checks passing
- [ ] **Documentation**: All deliverables documented with examples
- [ ] **Knowledge Transfer**: Team can maintain and extend independently

**Communication Metrics:**  
- [ ] **Stakeholder Updates**: Regular progress reports provided
- [ ] **Decision Documentation**: All major decisions recorded with rationale
- [ ] **Learning Capture**: Lessons learned documented for future projects
- [ ] **Community Value**: Reusable assets created for broader team/community
```

## 🎯 Prompt Usage Instructions

### **How to Use These Prompts**

1. **Select Base Template**: Choose the template that best matches your project type
2. **Customize Context**: Fill in project-specific details and requirements  
3. **Adjust Time Estimates**: Scale based on team experience and project complexity
4. **Define Success Criteria**: Set clear, measurable success metrics
5. **Adapt Technology Stack**: Update tools and technologies for your environment

### **Best Practices for Prompt Engineering**

1. **Be Specific**: Provide concrete deliverables and success criteria
2. **Include Context**: Explain the why behind requirements
3. **Set Boundaries**: Clear scope prevents feature creep
4. **Provide Examples**: Include specific commands and code patterns
5. **Plan for Iteration**: Expect multiple rounds of refinement

### **Quality Assurance for Prompts**

1. **Test with Real Projects**: Validate prompts on actual work
2. **Get Feedback**: Ask teams what worked and what didn't
3. **Iterate Based on Results**: Improve prompts based on outcomes
4. **Document Variations**: Track successful adaptations
5. **Maintain Currency**: Update for new tools and practices

---

*These prompts are battle-tested from real infrastructure implementation work. Adapt them to your specific context and requirements.*

---

*Generated with Claude Code - https://claude.ai/code*