package main

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// runCmdSingleMainValidation implements CMD single main enforcement
func runCmdSingleMainValidation(pass *analysis.Pass) (interface{}, error) {
	var mainFiles []string

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename

		// Check if this is a main.go file in cmd/ directory
		if filepath.Base(filename) == "main.go" &&
			strings.Contains(filepath.Dir(filename), "cmd") {
			mainFiles = append(mainFiles, filename)
		}
	}

	// Validate constraint: exactly one main.go in cmd/
	if len(mainFiles) == 0 {
		// Report error on the first file analyzed (this is a project-wide constraint)
		if len(pass.Files) > 0 {
			pass.Reportf(pass.Files[0].Pos(),
				"CMD_SINGLE_MAIN: No main.go files found in cmd/ directory. "+
					"Create exactly one main.go file: mkdir -p cmd/server && touch cmd/server/main.go")
		}
	} else if len(mainFiles) > 1 {
		// Report error on each extra main file
		for i, mainFile := range mainFiles {
			if i > 0 { // Skip the first one (that's allowed)
				pass.Reportf(getFileNodeByPath(pass, mainFile).Pos(),
					"CMD_SINGLE_MAIN: Found %d main.go files in cmd/, expected exactly 1. "+
						"Consolidate using CLI frameworks like cobra: %s",
					len(mainFiles), strings.Join(mainFiles, ", "))
			}
		}
	}

	// Additional validation: ensure main.go contains proper package main and func main()
	for _, mainFile := range mainFiles {
		file := getFileNodeByPath(pass, mainFile)
		if file != nil {
			if err := validateMainFile(pass, file); err != nil {
				pass.Reportf(file.Pos(), "CMD_SINGLE_MAIN: %v", err)
			}
		}
	}

	return nil, nil
}

// getFileNodeByPath finds the AST file node for a given file path
func getFileNodeByPath(pass *analysis.Pass, filePath string) *ast.File {
	for _, file := range pass.Files {
		if pass.Fset.Position(file.Pos()).Filename == filePath {
			return file
		}
	}
	return nil
}

// validateMainFile ensures main.go has proper structure
func validateMainFile(pass *analysis.Pass, file *ast.File) error {
	// Check package name
	if file.Name.Name != "main" {
		return fmt.Errorf("main.go must have 'package main', found 'package %s'", file.Name.Name)
	}

	// Check for main function
	hasMainFunc := false
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == "main" && fn.Recv == nil {
				hasMainFunc = true

				// Validate main function signature
				if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
					pass.Reportf(fn.Pos(),
						"CMD_SINGLE_MAIN: main() function should not have parameters")
				}
				if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
					pass.Reportf(fn.Pos(),
						"CMD_SINGLE_MAIN: main() function should not return values")
				}
			}
		}
		return true
	})

	if !hasMainFunc {
		return fmt.Errorf("main.go must contain a main() function")
	}

	return nil
}
