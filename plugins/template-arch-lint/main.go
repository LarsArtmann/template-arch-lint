// Package main implements the unified template-arch-lint plugin for golangci-lint
// This plugin consolidates filename validation, CMD single main enforcement,
// import cycle detection, and code duplication analysis into a single analyzer.
package main

import (
	"golang.org/x/tools/go/analysis"
)

// New returns all analyzers provided by the template-arch-lint plugin.
// This is the required entry point for golangci-lint custom plugins.
func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		FilenameValidatorAnalyzer,
		CmdSingleMainAnalyzer,
		ImportCycleAnalyzer,
		CodeDuplicationAnalyzer,
	}, nil
}

// FilenameValidatorAnalyzer validates Go file naming conventions
var FilenameValidatorAnalyzer = &analysis.Analyzer{
	Name: "filename-validator",
	Doc:  "Validates Go file naming conventions following standard patterns",
	Run:  runFilenameValidation,
}

// CmdSingleMainAnalyzer enforces exactly one main.go file in cmd/ directory
var CmdSingleMainAnalyzer = &analysis.Analyzer{
	Name: "cmd-single-main",
	Doc:  "Enforces exactly one main.go file in cmd/ directory for clean architecture",
	Run:  runCmdSingleMainValidation,
}

// ImportCycleAnalyzer detects import cycles and circular dependencies
var ImportCycleAnalyzer = &analysis.Analyzer{
	Name: "import-cycle-detector",
	Doc:  "Detects import cycles and circular dependencies using AST analysis",
	Run:  runImportCycleDetection,
}

// CodeDuplicationAnalyzer detects code duplications using AST analysis
var CodeDuplicationAnalyzer = &analysis.Analyzer{
	Name: "code-duplication-detector",
	Doc:  "Detects code duplications using AST analysis with configurable thresholds",
	Run:  runCodeDuplicationDetection,
}
