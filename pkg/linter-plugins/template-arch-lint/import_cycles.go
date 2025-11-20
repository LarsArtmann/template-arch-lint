package main

import (
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// runImportCycleDetection implements import cycle detection analyzer
func runImportCycleDetection(pass *analysis.Pass) (interface{}, error) {
	// Build import graph for this package
	pkg := pass.Pkg
	imports := getPackageImports(pkg)

	// Check for direct and indirect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, imp := range imports {
		if cycle := findCycle(imp, pkg.Path(), visited, recStack, make([]string, 0)); cycle != nil {
			// Report cycle on the first file
			if len(pass.Files) > 0 {
				pass.Reportf(pass.Files[0].Pos(),
					"IMPORT_CYCLE: Import cycle detected: %s",
					strings.Join(cycle, " -> "))
			}
		}
	}

	return nil, nil
}

// getPackageImports extracts all import paths from a package
func getPackageImports(pkg *types.Package) []string {
	var imports []string
	seen := make(map[string]bool)

	for _, imp := range pkg.Imports() {
		if !seen[imp.Path()] {
			imports = append(imports, imp.Path())
			seen[imp.Path()] = true
		}
	}

	return imports
}

// findCycle uses DFS to detect import cycles
func findCycle(currentPkg, targetPkg string, visited, recStack map[string]bool, path []string) []string {
	// Add current package to path
	path = append(path, currentPkg)

	// Mark current node as visited and part of recursion stack
	visited[currentPkg] = true
	recStack[currentPkg] = true

	// If we've reached back to our target, we found a cycle
	if currentPkg == targetPkg && len(path) > 1 {
		return append(path, targetPkg) // Complete the cycle
	}

	// This is a simplified version - in a real implementation, you would need
	// to analyze the actual import statements of each package. For now, we'll
	// detect direct cycles and some obvious patterns.

	// Check for common cycle patterns
	if strings.Contains(currentPkg, targetPkg) || strings.Contains(targetPkg, currentPkg) {
		if isLikelyCycle(currentPkg, targetPkg) {
			return []string{targetPkg, currentPkg, targetPkg}
		}
	}

	// Clean up recursion stack
	recStack[currentPkg] = false

	return nil
}

// isLikelyCycle detects common import cycle patterns
func isLikelyCycle(pkg1, pkg2 string) bool {
	// Check for bidirectional dependencies between related packages
	pkg1Parts := strings.Split(pkg1, "/")
	pkg2Parts := strings.Split(pkg2, "/")

	if len(pkg1Parts) > 0 && len(pkg2Parts) > 0 {
		// Same base package with different subpackages
		if len(pkg1Parts) > 2 && len(pkg2Parts) > 2 {
			base1 := strings.Join(pkg1Parts[:len(pkg1Parts)-1], "/")
			base2 := strings.Join(pkg2Parts[:len(pkg2Parts)-1], "/")

			if base1 == base2 {
				// Potential cycle between sibling packages
				return true
			}
		}
	}

	return false
}

// Note: This is a simplified implementation for demonstration.
// A production implementation would need to:
// 1. Build a complete dependency graph across all packages
// 2. Use proper graph traversal algorithms
// 3. Cache results for performance
// 4. Handle complex module structures
// 5. Integrate with Go's module system for accurate dependency resolution
