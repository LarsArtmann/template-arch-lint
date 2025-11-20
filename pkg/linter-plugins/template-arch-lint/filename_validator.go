package main

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// runFilenameValidation implements filename validation analyzer
func runFilenameValidation(pass *analysis.Pass) (interface{}, error) {
	// Standard Go filename patterns
	validFilenameRegex := regexp.MustCompile(`^[a-z][a-z0-9_]*(_test)?\.go$`)

	for _, file := range pass.Files {
		filename := filepath.Base(pass.Fset.Position(file.Pos()).Filename)

		// Skip generated files
		if isGeneratedFile(filename) {
			continue
		}

		// Validate filename pattern
		if !validFilenameRegex.MatchString(filename) {
			pass.Reportf(file.Pos(),
				"filename %q does not follow Go naming conventions. Use lowercase with underscores: example_service.go",
				filename)
		}

		// Check for common anti-patterns
		if err := checkFilenameAntiPatterns(pass, filename, file); err != nil {
			pass.Reportf(file.Pos(), "%v", err)
		}
	}

	return nil, nil
}

// isGeneratedFile checks if a file is generated and should be skipped
func isGeneratedFile(filename string) bool {
	generatedPatterns := []string{
		"_gen.go",
		"_generated.go",
		".pb.go",
		"_templ.go",
		"_mock.go",
	}

	for _, pattern := range generatedPatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	return false
}

// checkFilenameAntiPatterns validates against common filename anti-patterns
func checkFilenameAntiPatterns(pass *analysis.Pass, filename string, file *ast.File) error {
	// Check for camelCase filenames
	if strings.ContainsAny(filename[:len(filename)-3], "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return fmt.Errorf("filename %q uses camelCase. Use snake_case: %s",
			filename, strings.ToLower(filename))
	}

	// Check for dashes (should use underscores)
	if strings.Contains(filename, "-") {
		suggested := strings.ReplaceAll(filename, "-", "_")
		return fmt.Errorf("filename %q uses dashes. Use underscores: %s",
			filename, suggested)
	}

	// Check package name alignment
	packageName := file.Name.Name
	expectedPrefix := strings.TrimSuffix(filename, ".go")
	expectedPrefix = strings.TrimSuffix(expectedPrefix, "_test")

	// For single-word packages, filename should match or be descriptive
	if packageName != "main" && !strings.HasPrefix(expectedPrefix, packageName) &&
		!isDescriptiveFilename(expectedPrefix, packageName) {
		return fmt.Errorf("filename %q doesn't align with package %q. Consider %s.go or %s_%s.go",
			filename, packageName, packageName, packageName, expectedPrefix)
	}

	return nil
}

// isDescriptiveFilename checks if a filename is descriptively appropriate for the package
func isDescriptiveFilename(filename, packageName string) bool {
	// Allow common descriptive patterns
	descriptivePatterns := []string{
		"handler", "service", "repository", "model", "entity", "value",
		"error", "config", "client", "server", "middleware", "util", "helper",
		"test", "benchmark", "example", "mock", "stub",
	}

	for _, pattern := range descriptivePatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}

	// Allow if filename contains package name
	return strings.Contains(filename, packageName)
}
