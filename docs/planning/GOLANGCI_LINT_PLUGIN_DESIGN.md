# golangci-lint Custom Plugin Design: CMD Single Main Enforcement

**Plugin Name**: `cmd-single-main`  
**Purpose**: Native golangci-lint integration for enforcing single main.go file in cmd/ directory  
**Priority**: High-value replacement for shell script approach

## 🎯 PLUGIN ARCHITECTURE OVERVIEW

### Core Implementation Strategy

Based on golangci-lint module plugin architecture:

```go
// Package structure
github.com/LarsArtmann/golangci-cmd-single-main/
├── cmd/
│   └── analyzer/
│       └── main.go          // Plugin analyzer
├── pkg/
│   └── cmdsingle/
│       ├── analyzer.go      // Core analysis logic
│       ├── checker.go       // File system checking
│       └── types.go         // Type definitions
├── go.mod
├── go.sum
└── README.md
```

### Go Analysis Framework Implementation

```go
// analyzer.go - Core plugin implementation
package cmdsingle

import (
    "go/analysis"
    "go/ast"
    "path/filepath"
    "os"
)

var Analyzer = &analysis.Analyzer{
    Name:     "cmd-single-main",
    Doc:      "Enforces exactly one main.go file in cmd/ directory for clean architecture",
    Run:      run,
    Requires: []*analysis.Analyzer{},
}

func run(pass *analysis.Pass) (interface{}, error) {
    // 1. Detect if we're analyzing a main package in cmd/
    // 2. Count all main.go files in project's cmd/ directory
    // 3. Report violations with precise locations and suggestions

    if !isMainPackageInCmd(pass) {
        return nil, nil
    }

    cmdMainFiles, err := findCmdMainFiles(pass)
    if err != nil {
        return nil, err
    }

    return validateSingleMain(pass, cmdMainFiles)
}
```

## 🔧 CONFIGURATION INTEGRATION

### .custom-gcl.yml Configuration

```yaml
version: "2"
plugins:
  - module: "github.com/LarsArtmann/golangci-cmd-single-main"
    import: "github.com/LarsArtmann/golangci-cmd-single-main/pkg/cmdsingle"

linters:
  enable:
    - cmd-single-main

linters-settings:
  custom:
    cmd-single-main:
      type: "module"
      description: "Enforce single main.go file in cmd/ directory"
      settings:
        enforce: true
        suggest-consolidation: true
        allow-symlinks: true
```

### .golangci.yml Integration

```yaml
linters:
  enable:
    - cmd-single-main

linters-settings:
  cmd-single-main:
    # Enable strict enforcement
    strict: true
    # Suggest consolidation patterns
    suggest-cobra: true
    # Custom message templates
    message-template: "Found {{.Count}} main.go files in cmd/, expected exactly 1"
```

## 🚀 ADVANTAGES OVER SHELL SCRIPT

### Technical Benefits

- **Native Integration**: Uses golangci-lint's native error reporting and IDE integration
- **AST Analysis**: Sophisticated Go code analysis beyond file system checking
- **Performance**: Runs alongside other linters in single pass
- **Caching**: Benefits from golangci-lint's sophisticated caching system
- **Context**: Provides precise source code locations and context

### User Experience Benefits

- **IDE Integration**: Works natively with VS Code, GoLand, etc.
- **Consistent UI**: Same error reporting format as other linters
- **Configuration**: Centralized in .golangci.yml
- **CI/CD**: Single tool execution instead of multiple scripts

### Enterprise Benefits

- **Reusability**: Plugin can be shared across multiple projects
- **Versioning**: Semantic versioning and dependency management
- **Maintenance**: Standard Go module maintenance and updates
- **Testing**: Standard Go testing practices and frameworks

## 📋 IMPLEMENTATION PHASES

### Phase 1: Core Plugin Development (4-6 hours)

1. **Setup Go Module**: Create plugin module structure
2. **Implement Analyzer**: Core analysis logic using go/analysis
3. **File System Logic**: Robust cmd/ directory scanning
4. **Error Reporting**: Rich error messages with suggestions
5. **Unit Tests**: Comprehensive test coverage

### Phase 2: Integration & Configuration (2-3 hours)

1. **Plugin Configuration**: .custom-gcl.yml setup
2. **Settings Schema**: Configuration options validation
3. **Documentation**: Usage and configuration guides
4. **Integration Tests**: Test with real projects

### Phase 3: Advanced Features (2-4 hours)

1. **AST Analysis**: Validate package main and func main()
2. **Suggestion Engine**: Cobra/CLI framework suggestions
3. **Performance Optimization**: Efficient directory scanning
4. **IDE Integration**: Enhanced error reporting

## 🔬 TECHNICAL IMPLEMENTATION DETAILS

### Core Analysis Logic

```go
func validateSingleMain(pass *analysis.Pass, cmdFiles []string) (interface{}, error) {
    switch len(cmdFiles) {
    case 0:
        pass.Report(analysis.Diagnostic{
            Pos:     pass.Fset.Position(pass.Files[0].Pos()).Pos,
            Message: "No main.go files found in cmd/ directory",
            Category: "architecture",
            SuggestedFixes: []analysis.SuggestedFix{{
                Message: "Create cmd/server/main.go",
                TextEdits: generateMainGoSuggestion(),
            }},
        })
    case 1:
        // Success case - validate it's actually a main package
        return validateMainPackage(pass, cmdFiles[0])
    default:
        // Multiple files - report violation with consolidation suggestions
        return reportMultipleMainFiles(pass, cmdFiles)
    }
    return nil, nil
}
```

### File System Integration

```go
func findCmdMainFiles(pass *analysis.Pass) ([]string, error) {
    // Use pass.Pkg.Path() to determine project root
    // Scan cmd/ directory efficiently
    // Handle symbolic links appropriately
    // Provide detailed error context
}
```

## 📊 COMPARISON: SHELL SCRIPT vs PLUGIN

| Aspect                  | Shell Script ✅             | golangci-lint Plugin 🚀 |
| ----------------------- | --------------------------- | ----------------------- |
| **Implementation Time** | 4 hours                     | 8-12 hours              |
| **Integration**         | External script             | Native linter           |
| **Error Reporting**     | Basic terminal output       | Rich IDE integration    |
| **Performance**         | Separate process            | Integrated pipeline     |
| **Maintainability**     | Shell script complexity     | Standard Go patterns    |
| **Reusability**         | Copy/paste configs          | Go module distribution  |
| **Testing**             | Custom test framework       | Standard Go testing     |
| **CI/CD Integration**   | Additional script execution | Single linter command   |

## 🎯 SUCCESS METRICS

### Technical Metrics

- **Performance**: < 50ms overhead to existing golangci-lint execution
- **Accuracy**: 100% detection of cmd/ main.go violations
- **Reliability**: Zero false positives/negatives in test suite
- **Compatibility**: Works with golangci-lint v2.3.0+

### User Experience Metrics

- **IDE Integration**: Error squiggles and quick fixes in VS Code/GoLand
- **Error Quality**: Actionable error messages with specific suggestions
- **Configuration**: Single .golangci.yml configuration point
- **Documentation**: Complete usage and configuration guides

## 🚧 MIGRATION STRATEGY

### Phase 1: Dual Operation

- Keep shell script as fallback during plugin development
- Test plugin alongside existing implementation
- Validate identical behavior and error detection

### Phase 2: Plugin Integration

- Add plugin configuration to .custom-gcl.yml
- Update justfile to include plugin validation
- Update documentation with plugin instructions

### Phase 3: Shell Script Deprecation

- Remove shell script after plugin validation period
- Update pre-commit hooks to use plugin
- Archive shell script approach with historical documentation

## 🔮 FUTURE ENHANCEMENTS

### Advanced Analysis Features

- **Package Structure Validation**: Ensure proper package organization
- **Import Cycle Detection**: Identify architectural violations
- **Dependency Analysis**: Validate clean architecture boundaries
- **Performance Profiling**: Built-in performance impact measurement

### Enterprise Features

- **Custom Rules Engine**: User-defined architectural constraints
- **Reporting Dashboard**: Architectural compliance reporting
- **Integration Hooks**: CI/CD pipeline integration endpoints
- **Multi-Project Analysis**: Cross-repository architectural consistency

---

**Next Steps**: Begin Phase 1 implementation with core plugin development while maintaining current shell script functionality.
