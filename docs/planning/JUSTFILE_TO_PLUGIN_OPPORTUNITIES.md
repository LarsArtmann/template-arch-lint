# Justfile → golangci-lint Plugin Consolidation Opportunities

**Analysis Date**: 2025-09-06  
**Current State**: Multiple shell-based custom linting rules  
**Target State**: Consolidated golangci-lint plugin architecture

## 🎯 EXECUTIVE SUMMARY

Current justfile contains **6 major custom linting rules** that could benefit from golangci-lint plugin consolidation:

1. **`lint-files`** - Filename validation (HIGH VALUE)
2. **`lint-cmd-single`** - Single main.go enforcement (HIGH VALUE)
3. **Architecture extensions** - Enhanced go-arch-lint integration (MEDIUM VALUE)
4. **Project structure validation** - Directory organization (MEDIUM VALUE)
5. **Template validation** - Templ generation checks (LOW VALUE)
6. **Custom security rules** - Project-specific security patterns (FUTURE VALUE)

## 📊 DETAILED ANALYSIS

### 🚀 HIGH-VALUE PLUGIN CANDIDATES

#### 1. Filename Validation (`lint-files`)

**Current Implementation:**

```bash
lint-files:
    @if find . -name "*:*" -not -path "./.git/*" | grep -q .; then
        echo "❌ Found files with colons in names:"
        find . -name "*:*" -not -path "./.git/*"
        exit 1
    fi
```

**Plugin Potential:**

- **Value**: HIGH - Prevents cross-platform compatibility issues
- **Complexity**: LOW - Simple file system validation
- **Integration**: Native golangci-lint error reporting
- **Extensibility**: Could validate other problematic characters, length limits, case sensitivity

**Plugin Implementation:**

```go
var FilenameAnalyzer = &analysis.Analyzer{
    Name: "filename-validator",
    Doc:  "Validates filename compatibility across platforms",
    Run:  runFilenameValidation,
}

func runFilenameValidation(pass *analysis.Pass) (interface{}, error) {
    // Check for problematic characters: colons, reserved names, length limits
    // Report with precise file locations and suggestions
    // Support configurable rules (Windows compat, length limits, etc.)
}
```

#### 2. CMD Single Main Enforcement (`lint-cmd-single`)

**Current Implementation**: ✅ Already analyzed and implemented
**Plugin Potential**: HIGH - Clean architecture enforcement

### 🔧 MEDIUM-VALUE PLUGIN CANDIDATES

#### 3. Enhanced Architecture Validation

**Current Implementation:**

```bash
lint-arch:
    @if command -v go-arch-lint >/dev/null 2>&1; then
        go-arch-lint check
    fi
```

**Plugin Enhancement Opportunities:**

- **Project-specific rules** beyond standard go-arch-lint
- **Template-specific architecture** (templ + HTMX patterns)
- **Clean architecture boundaries** with custom error messages
- **Dependency injection validation** (samber/do patterns)

**Plugin Implementation:**

```go
var TemplateArchAnalyzer = &analysis.Analyzer{
    Name: "template-arch",
    Doc:  "Template project architecture validation",
    Run:  runTemplateArchValidation,
}

func runTemplateArchValidation(pass *analysis.Pass) (interface{}, error) {
    // Validate Clean Architecture boundaries
    // Check DDD patterns compliance
    // Enforce templ + HTMX best practices
    // Validate dependency injection patterns
}
```

#### 4. Project Structure Validation

**Current Gaps**: No formal project structure validation
**Plugin Potential:**

- **Directory structure** enforcement (cmd/, internal/, pkg/, web/)
- **File organization** rules (naming conventions, package structure)
- **Template project compliance** (follows template-arch-lint patterns)

### 🔮 LOW-VALUE / FUTURE PLUGIN CANDIDATES

#### 5. Template Generation Validation

**Current Implementation**: Basic templ command execution
**Plugin Enhancement**:

- Validate templ templates are properly generated
- Check for template syntax issues
- Ensure template-Go code synchronization

#### 6. Custom Security Patterns

**Future Implementation**:

- Project-specific security rules beyond gosec
- Template-specific XSS prevention patterns
- Clean architecture security boundary validation

## 🚀 PROPOSED PLUGIN ARCHITECTURE

### Single Unified Plugin: `template-arch-lint-plugin`

**Motivation**: Rather than multiple separate plugins, create one comprehensive plugin with multiple analyzers:

```go
package templatearchlint

var Analyzers = []*analysis.Analyzer{
    FilenameAnalyzer,
    CmdSingleMainAnalyzer,
    ProjectStructureAnalyzer,
    TemplateArchAnalyzer,
    // Future: TemplGenAnalyzer, SecurityAnalyzer
}

func New(conf any) ([]*analysis.Analyzer, error) {
    // Return configured analyzers based on settings
    return Analyzers, nil
}
```

### Configuration Integration

```yaml
# .custom-gcl.yml
plugins:
  - module: "github.com/LarsArtmann/template-arch-lint-plugin"
    import: "github.com/LarsArtmann/template-arch-lint-plugin"
    version: v1.0.0

# .golangci.yml
linters:
  enable:
    - filename-validator
    - cmd-single-main
    - project-structure
    - template-arch

linters-settings:
  custom:
    filename-validator:
      type: "module"
      description: "Cross-platform filename validation"
      settings:
        check_colons: true
        check_length: true
        max_length: 255
        windows_compat: true

    cmd-single-main:
      type: "module"
      description: "Enforce single main.go in cmd/"
      settings:
        strict: true
        suggest_consolidation: true

    project-structure:
      type: "module"
      description: "Template project structure validation"
      settings:
        require_internal: true
        require_cmd: true
        allow_pkg: false
```

## 📋 IMPLEMENTATION PHASES

### Phase 1: Core Plugin Development (8-12 hours)

1. **Setup Plugin Module**: Create unified plugin architecture
2. **Filename Validator**: Port lint-files logic to plugin
3. **CMD Single Main**: Port our shell script to plugin analyzer
4. **Basic Testing**: Unit tests for core analyzers
5. **Configuration**: Basic .custom-gcl.yml setup

### Phase 2: Advanced Analysis (6-8 hours)

1. **Project Structure**: Directory organization validation
2. **Enhanced Errors**: Rich error messages with suggestions
3. **Integration Tests**: Test with real projects
4. **Performance**: Optimize for large codebases

### Phase 3: Template-Specific Rules (4-6 hours)

1. **Template Architecture**: Clean Architecture + DDD validation
2. **Security Patterns**: Template-specific security rules
3. **Documentation**: Complete usage guides
4. **CI/CD Integration**: Replace justfile shell commands

## 💡 STRATEGIC ADVANTAGES

### Development Benefits

- **Single Plugin**: Unified development, testing, and maintenance
- **Native Integration**: Deep golangci-lint integration vs external scripts
- **Performance**: Single AST pass vs multiple shell executions
- **Error Quality**: Rich IDE integration with precise locations

### User Experience Benefits

- **Consistent Interface**: All rules in single .golangci.yml
- **IDE Integration**: Native VS Code/GoLand support
- **Error Reporting**: Consistent format across all custom rules
- **Configuration**: Single configuration point vs scattered justfile rules

### Maintenance Benefits

- **Standard Go Tooling**: Use Go testing, modules, releases
- **Reusability**: Plugin can be used across multiple projects
- **Versioning**: Semantic versioning for rule changes
- **Community**: Standard plugin distribution and contribution patterns

## 🔄 MIGRATION STRATEGY

### Phase 1: Dual Operation (1-2 weeks)

- Keep existing justfile commands as fallback
- Develop plugin alongside current implementation
- Test plugin behavior matches shell script behavior

### Phase 2: Plugin Integration (1 week)

- Add plugin to .custom-gcl.yml configuration
- Update justfile to use plugin where possible
- Validate identical behavior in CI/CD

### Phase 3: Shell Script Deprecation (1 week)

- Remove shell-based custom linting commands
- Update documentation and help text
- Archive shell scripts with historical documentation

## 🎯 SUCCESS METRICS

### Technical Metrics

- **Performance**: < 100ms overhead vs shell script approach
- **Accuracy**: 100% parity with existing shell script validation
- **Coverage**: All current custom rules converted to plugin
- **Reliability**: Zero false positives/negatives

### User Experience Metrics

- **IDE Integration**: Error highlighting and quick fixes work
- **Configuration**: Single .golangci.yml configuration point
- **Error Quality**: Rich error messages with actionable suggestions
- **Documentation**: Complete migration and usage guides

---

## 🚀 IMMEDIATE NEXT STEPS

1. **Complete Current Shell Implementation** (finish justfile approach)
2. **Create Plugin MVP** with filename-validator + cmd-single-main
3. **Test Plugin Integration** alongside existing shell scripts
4. **Document Migration Path** for future plugin adoption
5. **Plan Phase 2 Features** (project-structure, template-arch)

**ROI Analysis**: Plugin approach requires ~20-30 hours initial investment but provides:

- **Developer Experience**: 10x improvement (native IDE integration)
- **Maintenance**: 50% reduction (standard Go patterns vs shell scripts)
- **Reusability**: Cross-project plugin distribution
- **Performance**: 3-5x faster execution (single AST pass vs multiple processes)
