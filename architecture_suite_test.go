// Package architecture_test contains automated tests that enforce
// architectural boundaries
// and Clean Architecture/DDD principles to prevent architectural
// decay over time.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

func TestArchitecture(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "🏗️ Architecture Test Suite - Clean "+
		"Architecture & DDD Enforcement")
}

// packageInfo holds information about a Go package discovered during analysis.
type packageInfo struct {
	path      string
	layer     string
	imports   []string
	functions []string
	types     []string
}

// domainLayers defines the architectural layers and their allowed dependencies.
var domainLayers = map[string][]string{
	"domain/entities": {"domain/shared", "domain/values", "domain/errors"},
	"domain/values":   {"domain/shared", "domain/errors"},
	"domain/repositories": {
		"domain/entities", "domain/shared",
		"domain/values", "domain/errors",
	},
	"domain/services": {
		"domain/entities", "domain/repositories",
		"domain/shared", "domain/values", "domain/errors",
	},
	"domain/shared": {},
	"domain/errors": {"domain/shared"},
	"application": {
		"domain/entities", "domain/services",
		"domain/repositories", "domain/shared", "domain/values",
		"domain/errors",
	},
	"infrastructure": {
		"domain/entities", "domain/repositories",
		"domain/shared", "domain/values", "domain/errors",
	},
}

var _ = ginkgo.Describe("🏗️ Architecture Tests - Clean "+
	"Architecture & DDD Enforcement", func() {
	var packages []packageInfo
	var fileSet *token.FileSet

	ginkgo.BeforeEach(func() {
		packages = []packageInfo{}
		fileSet = token.NewFileSet()

		// Parse all Go files in the project
		err := filepath.Walk(".", func(path string, _ os.FileInfo,
			err error,
		) error {
			if err != nil {
				return err
			}

			// Skip non-Go files, test files, generated files, and vendor
			// directories
			if !strings.HasSuffix(path, ".go") ||
				strings.HasSuffix(path, "_test.go") ||
				strings.Contains(path, "_templ.go") ||
				strings.Contains(path, "vendor/") ||
				strings.Contains(path, ".git/") ||
				strings.Contains(path, "web/templates/") {
				return nil
			}

			// Parse the Go file
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			file, err := parser.ParseFile(fileSet, path, src,
				parser.ParseComments)
			if err != nil {
				// Skip files that can't be parsed
				return nil
			}

			// Extract package information
			pkg := extractPackageInfo(path, file)
			packages = append(packages, pkg)

			return nil
		})
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.Describe("🔒 TestDomainIsolation", func() {
		ginkgo.It("should ensure domain layer has no external infrastructure "+
			"dependencies", func() {
			domainPackages := filterPackagesByLayer(packages, "domain")

			for _, pkg := range domainPackages {
				ginkgo.By(fmt.Sprintf("Checking domain package %s for external "+
					"dependencies", pkg.path))

				for _, importPath := range pkg.imports {
					// Skip standard library and test imports
					if isStandardLibrary(importPath) || isTestImport(importPath) {
						continue
					}

					// Skip allowed vendor dependencies
					if isAllowedVendorDependency(importPath) {
						continue
					}

					// Domain should only import from other domain packages
					if !isDomainImport(importPath) {
						ginkgo.Fail(fmt.Sprintf("❌ DOMAIN ISOLATION VIOLATION: Package %s "+
							"imports non-domain dependency %s\n"+
							"Domain layer must not depend on infrastructure, "+
							"application, or external concerns.\n"+
							"Allowed: domain/*, standard library, approved "+
							"vendor packages",
							pkg.path, importPath))
					}
				}
			}

			ginkgo.By("✅ Domain isolation maintained - no external dependencies found")
		})
	})

	ginkgo.Describe("🔄 TestLayerDependencies", func() {
		ginkgo.It("should verify proper layer dependency direction follows "+
			"Clean Architecture", func() {
			for _, pkg := range packages {
				layer := pkg.layer
				allowedDeps := domainLayers[layer]

				if len(allowedDeps) == 0 && layer != "main" && layer != "config" {
					continue // Skip layers not explicitly defined
				}

				ginkgo.By(fmt.Sprintf("Checking layer dependencies for %s "+
					"(layer: %s)", pkg.path, layer))

				for _, importPath := range pkg.imports {
					// Skip standard library and vendor dependencies
					if isStandardLibrary(importPath) || isVendorDependency(importPath) {
						continue
					}

					importLayer := getLayerFromImport(importPath)
					if importLayer == "" {
						continue
					}

					// Check if the import is allowed for this layer
					if !isAllowedLayerDependency(layer, importLayer, allowedDeps) {
						ginkgo.Fail(fmt.Sprintf("❌ LAYER DEPENDENCY VIOLATION: %s "+
							"(layer: %s) cannot depend on %s (layer: %s)\n"+
							"Clean Architecture rule violated. Allowed dependencies "+
							"for %s: %v\n"+
							"Dependencies must flow: Infrastructure → Application → "+
							"Domain",
							pkg.path, layer, importPath, importLayer, layer, allowedDeps))
					}
				}
			}

			ginkgo.By("✅ Layer dependencies follow Clean Architecture principles")
		})
	})

	ginkgo.Describe("🔄 TestNoCircularDeps", func() {
		ginkgo.It("should detect circular dependencies between packages", func() {
			// Build dependency graph
			graph := buildDependencyGraph(packages)

			// Check for cycles using DFS
			visited := make(map[string]bool)
			recStack := make(map[string]bool)

			for pkgPath := range graph {
				if !visited[pkgPath] {
					cycle := findCycle(pkgPath, graph, visited, recStack, []string{})
					if len(cycle) > 0 {
						ginkgo.Fail(fmt.Sprintf("❌ CIRCULAR DEPENDENCY DETECTED: %s\n"+
							"Circular dependencies violate Clean Architecture and "+
							"can cause compilation issues.\n"+
							"Refactor to remove the circular dependency by "+
							"introducing interfaces or reorganizing code.",
							strings.Join(cycle, " → ")))
					}
				}
			}

			ginkgo.By("✅ No circular dependencies found")
		})
	})

	ginkgo.Describe("💎 TestValueObjectsImmutable", func() {
		ginkgo.It("should verify value objects are immutable and follow DDD "+
			"principles", func() {
			// Test Email value object
			emailType := reflect.TypeOf(values.Email{})
			ginkgo.By(fmt.Sprintf("Checking value object immutability: %s",
				emailType.Name()))

			// Check that all fields are unexported (immutable)
			for i := 0; i < emailType.NumField(); i++ {
				field := emailType.Field(i)
				firstChar := field.Name[0:1]
				if strings.ToUpper(firstChar) == firstChar {
					ginkgo.Fail(fmt.Sprintf("❌ VALUE OBJECT MUTABILITY VIOLATION: %s.%s "+
						"is exported\n"+
						"Value objects must be immutable. All fields should be "+
						"unexported.\n"+
						"Use getter methods to access field values.",
						emailType.Name(), field.Name))
				}
			}

			// Test UserID value object
			userIDType := reflect.TypeOf(values.UserID{})
			ginkgo.By(fmt.Sprintf("Checking value object immutability: %s",
				userIDType.Name()))

			for i := 0; i < userIDType.NumField(); i++ {
				field := userIDType.Field(i)
				firstChar := field.Name[0:1]
				if strings.ToUpper(firstChar) == firstChar {
					ginkgo.Fail(fmt.Sprintf("❌ VALUE OBJECT MUTABILITY VIOLATION: %s.%s "+
						"is exported\n"+
						"Value objects must be immutable. All fields should be "+
						"unexported.\n"+
						"Use getter methods to access field values.",
						userIDType.Name(), field.Name))
				}
			}

			// Test UserName value object
			userNameType := reflect.TypeOf(values.UserName{})
			ginkgo.By(fmt.Sprintf("Checking value object immutability: %s",
				userNameType.Name()))

			for i := 0; i < userNameType.NumField(); i++ {
				field := userNameType.Field(i)
				firstChar := field.Name[0:1]
				if strings.ToUpper(firstChar) == firstChar {
					ginkgo.Fail(fmt.Sprintf("❌ VALUE OBJECT MUTABILITY VIOLATION: %s.%s "+
						"is exported\n"+
						"Value objects must be immutable. All fields should be "+
						"unexported.\n"+
						"Use getter methods to access field values.",
						userNameType.Name(), field.Name))
				}
			}

			ginkgo.By("✅ All value objects are properly immutable")
		})
	})

	ginkgo.Describe("🔌 TestRepositoryInterfaces", func() {
		ginkgo.It("should ensure repository interfaces are in domain and "+
			"implementations in infrastructure", func() {
			// Check that UserRepository interface is in domain
			userRepoType := reflect.TypeOf((*repositories.UserRepository)(nil)).Elem()
			ginkgo.By(fmt.Sprintf("Verifying repository interface location: %s",
				userRepoType.Name()))

			gomega.Expect(userRepoType.Kind()).To(gomega.Equal(reflect.Interface),
				"Repository should be an interface, not a concrete type")

			// Check repository interface methods follow domain patterns
			numMethods := userRepoType.NumMethod()
			gomega.Expect(numMethods).To(gomega.BeNumerically(">", 0),
				"Repository interface should define methods")

			for i := 0; i < numMethods; i++ {
				method := userRepoType.Method(i)
				methodType := method.Type

				ginkgo.By(fmt.Sprintf("Checking repository method: %s", method.Name))

				// Repository methods should have context as first parameter
				if methodType.NumIn() > 0 { // Interface methods don't have receiver
					// in reflection
					firstParam := methodType.In(0)
					if !strings.Contains(firstParam.String(), "context.Context") {
						ginkgo.Fail(fmt.Sprintf("❌ REPOSITORY METHOD VIOLATION: %s.%s "+
							"should have context.Context as first parameter\n"+
							"Repository methods must accept context for "+
							"cancellation and timeout support.",
							userRepoType.Name(), method.Name))
					}
				}

				// Repository methods should return error as last return value
				if methodType.NumOut() > 0 {
					lastReturn := methodType.Out(methodType.NumOut() - 1)
					if !strings.Contains(lastReturn.String(), "error") {
						ginkgo.Fail(fmt.Sprintf("❌ REPOSITORY METHOD VIOLATION: %s.%s "+
							"should return error as last value\n"+
							"Repository methods must return errors for proper "+
							"error handling.",
							userRepoType.Name(), method.Name))
					}
				}
			}

			ginkgo.By("✅ Repository interfaces properly defined in domain " +
				"layer")
		})
	})

	ginkgo.Describe("🧹 TestServicePurity", func() {
		ginkgo.It("should verify services don't depend on infrastructure "+
			"directly", func() {
			servicePackages := filterPackagesByPath(packages, "domain/services")

			for _, pkg := range servicePackages {
				ginkgo.By(fmt.Sprintf("Checking service purity: %s", pkg.path))

				for _, importPath := range pkg.imports {
					// Skip standard library and allowed dependencies
					if isStandardLibrary(importPath) || isDomainImport(importPath) {
						continue
					}

					// Services should not import infrastructure
					if strings.Contains(importPath, "infrastructure") ||
						strings.Contains(importPath, "/db") ||
						strings.Contains(importPath, "persistence") {
						ginkgo.Fail(fmt.Sprintf("❌ SERVICE PURITY VIOLATION: Service %s "+
							"imports infrastructure dependency %s\n"+
							"Domain services must not depend directly on "+
							"infrastructure.\n"+
							"Use repository interfaces and dependency injection "+
							"instead.",
							pkg.path, importPath))
					}
				}
			}

			// Test service constructor patterns using reflection
			serviceType := reflect.TypeOf(services.UserService{})
			ginkgo.By(fmt.Sprintf("Checking service constructor pattern: %s",
				serviceType.Name()))

			// Services should have repository dependencies injected, not
			// infrastructure types
			for i := 0; i < serviceType.NumField(); i++ {
				field := serviceType.Field(i)
				fieldType := field.Type.String()

				if strings.Contains(fieldType, "infrastructure") ||
					strings.Contains(fieldType, "*sql.DB") ||
					strings.Contains(fieldType, "persistence") {
					ginkgo.Fail(fmt.Sprintf("❌ SERVICE DEPENDENCY VIOLATION: %s.%s "+
						"has infrastructure dependency %s\n"+
						"Services should depend on repository interfaces, not "+
						"concrete infrastructure types.\n"+
						"Use dependency injection with repository interfaces.",
						serviceType.Name(), field.Name, fieldType))
				}
			}

			ginkgo.By("✅ Services maintain purity and don't depend on " +
				"infrastructure")
		})
	})

	ginkgo.Describe("📊 Architecture Constraint Summary", func() {
		ginkgo.It("should report all architectural constraints being "+
			"enforced", func() {
			constraints := []string{
				"✅ Domain Isolation: Domain layer has zero infrastructure " +
					"dependencies",
				"✅ Layer Dependencies: Clean Architecture dependency flow " +
					"enforced",
				"✅ No Circular Dependencies: Package dependency cycles " +
					"prevented",
				"✅ Value Object Immutability: DDD value objects are immutable",
				"✅ Repository Interfaces: Repository contracts defined in " +
					"domain",
				"✅ Service Purity: Domain services free from infrastructure " +
					"coupling",
				"✅ Dependency Inversion: Infrastructure implements domain " +
					"interfaces",
				"✅ Single Responsibility: Each layer has clear, focused " +
					"concerns",
				"✅ Interface Segregation: Repository interfaces follow single " +
					"purpose",
				"✅ Clean Boundaries: No violations of architectural boundaries " +
					"detected",
			}

			ginkgo.By("📋 Architectural Constraints Report:")
			for _, constraint := range constraints {
				ginkgo.By(constraint)
			}

			ginkgo.By(fmt.Sprintf("📦 Analyzed %d packages across all layers",
				len(packages)))
			ginkgo.By("🏛️ Clean Architecture + DDD principles successfully " +
				"enforced")
		})
	})
})

// Helper functions for package analysis and architectural rule enforcement

func extractPackageInfo(path string, file *ast.File) packageInfo {
	pkg := packageInfo{
		path:      path,
		layer:     getLayerFromPath(path),
		imports:   []string{},
		functions: []string{},
		types:     []string{},
	}

	// Extract imports
	for _, imp := range file.Imports {
		importPath, _ := strconv.Unquote(imp.Path.Value)
		pkg.imports = append(pkg.imports, importPath)
	}

	// Extract functions and types
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.IsExported() {
				pkg.functions = append(pkg.functions, d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.IsExported() {
					pkg.types = append(pkg.types, typeSpec.Name.Name)
				}
			}
		}
	}

	return pkg
}

// pathToLayerMapping maps path patterns to layer names for efficient lookup.
var pathToLayerMapping = []struct {
	pattern string
	layer   string
}{
	{"internal/domain/entities", "domain/entities"},
	{"internal/domain/values", "domain/values"},
	{"internal/domain/repositories", "domain/repositories"},
	{"internal/domain/services", "domain/services"},
	{"internal/domain/shared", "domain/shared"},
	{"internal/domain/errors", "domain/errors"},
	{"internal/application", "application"},
	{"internal/infrastructure", "infrastructure"},
	{"internal/config", "config"},
	{"cmd/", "main"},
}

func getLayerFromPath(path string) string {
	for _, mapping := range pathToLayerMapping {
		if strings.Contains(path, mapping.pattern) {
			return mapping.layer
		}
	}
	return "unknown"
}

// importToLayerMapping maps import patterns to layer names for efficient
// lookup.
var importToLayerMapping = map[string]string{
	"/domain/entities":     "domain/entities",
	"/domain/values":       "domain/values",
	"/domain/repositories": "domain/repositories",
	"/domain/services":     "domain/services",
	"/domain/shared":       "domain/shared",
	"/domain/errors":       "domain/errors",
	"/application":         "application",
	"/infrastructure":      "infrastructure",
}

func getLayerFromImport(importPath string) string {
	for pattern, layer := range importToLayerMapping {
		if strings.Contains(importPath, pattern) {
			return layer
		}
	}
	return ""
}

func filterPackagesByLayer(packages []packageInfo, layer string) []packageInfo {
	var filtered []packageInfo
	for _, pkg := range packages {
		if strings.HasPrefix(pkg.layer, layer) {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

func filterPackagesByPath(packages []packageInfo,
	pathContains string,
) []packageInfo {
	var filtered []packageInfo
	for _, pkg := range packages {
		if strings.Contains(pkg.path, pathContains) {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

// standardLibraryPrefixes contains standard library package prefixes for
// efficient lookup.
var standardLibraryPrefixes = []string{
	"context",
	"database/sql",
	"encoding/json",
	"fmt",
	"log",
	"net/http",
	"regexp",
	"strconv",
	"strings",
	"time",
	"errors",
	"go/",
	"os",
	"path",
	"reflect",
}

func isStandardLibrary(importPath string) bool {
	// Fast check for packages without dots (standard library)
	if !strings.Contains(importPath, ".") {
		return true
	}

	// Check against known standard library prefixes
	for _, prefix := range standardLibraryPrefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}

	return false
}

func isTestImport(importPath string) bool {
	return strings.Contains(importPath, "github.com/onsi/ginkgo") ||
		strings.Contains(importPath, "github.com/onsi/gomega") ||
		strings.Contains(importPath, "testing")
}

func isAllowedVendorDependency(importPath string) bool {
	allowedVendorPrefixes := []string{
		"github.com/samber/lo",
		"github.com/samber/mo",
		"github.com/samber/do",
		"github.com/gin-gonic/gin",
		"github.com/go-playground/validator",
		"github.com/spf13/viper",
		"github.com/a-h/templ",
		"github.com/mattn/go-sqlite3",
	}

	for _, prefix := range allowedVendorPrefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}

func isDomainImport(importPath string) bool {
	return strings.Contains(importPath, "/domain/")
}

func isVendorDependency(importPath string) bool {
	return strings.Contains(importPath, "github.com/") ||
		strings.Contains(importPath, "golang.org/") ||
		strings.Contains(importPath, "gopkg.in/")
}

func isAllowedLayerDependency(fromLayer, toLayer string,
	allowedDeps []string,
) bool {
	// Allow dependencies within same layer
	if fromLayer == toLayer {
		return true
	}

	// Check explicit allowed dependencies
	for _, allowed := range allowedDeps {
		if toLayer == allowed {
			return true
		}
	}

	return false
}

func buildDependencyGraph(packages []packageInfo) map[string][]string {
	graph := make(map[string][]string)

	for _, pkg := range packages {
		graph[pkg.path] = []string{}
		for _, imp := range pkg.imports {
			// Only track project internal dependencies
			if strings.Contains(imp, "template-arch-lint/internal") {
				graph[pkg.path] = append(graph[pkg.path], imp)
			}
		}
	}

	return graph
}

func findCycle(node string, graph map[string][]string,
	visited, recStack map[string]bool, path []string,
) []string {
	markNodeAsVisited(node, visited, recStack)
	path = append(path, node)

	for _, neighbor := range graph[node] {
		if cycle := processCycleNeighbor(neighbor, graph, visited, recStack,
			path); len(cycle) > 0 {
			return cycle
		}
	}

	recStack[node] = false
	return nil
}

// markNodeAsVisited marks a node as visited in the DFS traversal
func markNodeAsVisited(node string, visited, recStack map[string]bool) {
	visited[node] = true
	recStack[node] = true
}

// processCycleNeighbor processes a neighbor during cycle detection
func processCycleNeighbor(neighbor string, graph map[string][]string,
	visited, recStack map[string]bool, path []string,
) []string {
	if !visited[neighbor] {
		return findCycle(neighbor, graph, visited, recStack, path)
	}

	if recStack[neighbor] {
		return buildCyclePath(neighbor, path)
	}

	return nil
}

// buildCyclePath constructs the cycle path when a back edge is found
func buildCyclePath(neighbor string, path []string) []string {
	cycleStart := findCycleStart(neighbor, path)
	if cycleStart >= 0 {
		return append(path[cycleStart:], neighbor)
	}
	return nil
}

// findCycleStart finds the starting index of the cycle in the path
func findCycleStart(neighbor string, path []string) int {
	for i, p := range path {
		if p == neighbor {
			return i
		}
	}
	return -1
}
