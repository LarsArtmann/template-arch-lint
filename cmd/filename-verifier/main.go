// filename-verifier: Custom linter to enforce filename conventions
// 🚨 STRICT ENFORCEMENT: No colons allowed in filenames
//
// This tool scans all files in the repository and ensures none contain
// a colon (:) character in their filename, which can cause issues on
// certain filesystems (especially Windows).
//
// Usage:
//   go run cmd/filename-verifier/main.go [directory]
//   filename-verifier [directory]
//
// Exit codes:
//   0 - All filenames are valid
//   1 - Found files with colons in their names
//   2 - Error during execution

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Exit codes
	ExitSuccess           = 0
	ExitViolationFound    = 1
	ExitError             = 2

	// Colors for terminal output
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
)

// ViolationReport represents a filename violation
type ViolationReport struct {
	Path     string
	Filename string
	Issue    string
}

// FileVerifier checks files for naming violations
type FileVerifier struct {
	rootDir           string
	violations        []ViolationReport
	filesScanned      int
	directoriesScanned int
	skipPatterns      []string
}

// NewFileVerifier creates a new verifier instance
func NewFileVerifier(rootDir string) *FileVerifier {
	return &FileVerifier{
		rootDir:      rootDir,
		violations:   []ViolationReport{},
		skipPatterns: getSkipPatterns(),
	}
}

// getSkipPatterns returns patterns to skip during scanning
func getSkipPatterns() []string {
	return []string{
		".git",
		".github",
		"vendor",
		"node_modules",
		".idea",
		".vscode",
		"dist",
		"build",
		"bin",
		".DS_Store",
		"*.pb.go",
		"*_generated.go",
		"*_gen.go",
	}
}

// shouldSkip checks if a path should be skipped
func (v *FileVerifier) shouldSkip(path string) bool {
	for _, pattern := range v.skipPatterns {
		// Check exact directory matches
		if strings.Contains(path, "/"+pattern+"/") || strings.HasPrefix(path, pattern+"/") {
			return true
		}
		// Check file patterns
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

// Verify scans all files and checks for violations
func (v *FileVerifier) Verify() error {
	fmt.Printf("%s🔍 FILENAME VERIFIER%s\n", ColorBold, ColorReset)
	fmt.Printf("Scanning directory: %s\n\n", v.rootDir)

	err := filepath.WalkDir(v.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for cleaner output
		relPath, _ := filepath.Rel(v.rootDir, path)
		if relPath == "" {
			relPath = path
		}

		// Skip certain directories and files
		if v.shouldSkip(relPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			v.directoriesScanned++
			return nil
		}

		v.filesScanned++
		filename := d.Name()

		// Check for colon in filename
		if strings.Contains(filename, ":") {
			v.violations = append(v.violations, ViolationReport{
				Path:     relPath,
				Filename: filename,
				Issue:    "Filename contains colon (:) character",
			})
		}

		// Additional checks can be added here
		v.checkAdditionalViolations(relPath, filename)

		return nil
	})

	return err
}

// checkAdditionalViolations performs additional filename checks
func (v *FileVerifier) checkAdditionalViolations(path, filename string) {
	// Check for spaces in filenames (warning level)
	if strings.Contains(filename, " ") {
		v.violations = append(v.violations, ViolationReport{
			Path:     path,
			Filename: filename,
			Issue:    "Filename contains spaces (consider using hyphens or underscores)",
		})
	}

	// Check for special characters that might cause issues
	problematicChars := []struct {
		char string
		desc string
	}{
		{"<", "less-than sign"},
		{">", "greater-than sign"},
		{"|", "pipe character"},
		{"\"", "quote character"},
		{"*", "asterisk (except in patterns)"},
		{"?", "question mark"},
		{"\t", "tab character"},
		{"\n", "newline character"},
		{"\r", "carriage return"},
	}

	for _, pc := range problematicChars {
		if strings.Contains(filename, pc.char) {
			v.violations = append(v.violations, ViolationReport{
				Path:     path,
				Filename: filename,
				Issue:    fmt.Sprintf("Filename contains %s", pc.desc),
			})
		}
	}

	// Check for files starting with dot (hidden files) - informational
	if strings.HasPrefix(filename, ".") && filename != ".gitignore" && filename != ".gitkeep" {
		// This is often intentional, so we don't add it as a violation
		// Just count it for statistics
	}

	// Check for very long filenames (>255 characters)
	if len(filename) > 255 {
		v.violations = append(v.violations, ViolationReport{
			Path:     path,
			Filename: filename,
			Issue:    fmt.Sprintf("Filename too long (%d characters, max 255)", len(filename)),
		})
	}

	// Check for non-ASCII characters (might cause issues in some systems)
	for _, r := range filename {
		if r > 127 {
			v.violations = append(v.violations, ViolationReport{
				Path:     path,
				Filename: filename,
				Issue:    "Filename contains non-ASCII characters",
			})
			break
		}
	}
}

// PrintReport prints the verification report
func (v *FileVerifier) PrintReport() {
	fmt.Printf("%s📊 SCAN SUMMARY%s\n", ColorBold, ColorReset)
	fmt.Printf("Files scanned:       %d\n", v.filesScanned)
	fmt.Printf("Directories scanned: %d\n", v.directoriesScanned)
	fmt.Printf("Violations found:    %d\n\n", len(v.violations))

	if len(v.violations) == 0 {
		fmt.Printf("%s✅ SUCCESS: All filenames are valid!%s\n", ColorGreen, ColorReset)
		return
	}

	// Group violations by severity
	critical := []ViolationReport{}
	warnings := []ViolationReport{}

	for _, v := range v.violations {
		if strings.Contains(v.Issue, "colon") || 
		   strings.Contains(v.Issue, "less-than") || 
		   strings.Contains(v.Issue, "greater-than") ||
		   strings.Contains(v.Issue, "pipe") ||
		   strings.Contains(v.Issue, "quote") ||
		   strings.Contains(v.Issue, "too long") {
			critical = append(critical, v)
		} else {
			warnings = append(warnings, v)
		}
	}

	// Print critical violations
	if len(critical) > 0 {
		fmt.Printf("%s🚨 CRITICAL VIOLATIONS:%s\n", ColorRed+ColorBold, ColorReset)
		for i, v := range critical {
			fmt.Printf("%s[%d] %s%s\n", ColorRed, i+1, v.Path, ColorReset)
			fmt.Printf("    Issue: %s\n", v.Issue)
			fmt.Printf("    Filename: %s\n\n", v.Filename)
		}
	}

	// Print warnings
	if len(warnings) > 0 {
		fmt.Printf("%s⚠️  WARNINGS:%s\n", ColorYellow+ColorBold, ColorReset)
		for i, v := range warnings {
			fmt.Printf("%s[%d] %s%s\n", ColorYellow, i+1, v.Path, ColorReset)
			fmt.Printf("    Issue: %s\n", v.Issue)
			fmt.Printf("    Filename: %s\n\n", v.Filename)
		}
	}

	// Print fix suggestions
	fmt.Printf("%s💡 FIX SUGGESTIONS:%s\n", ColorBold, ColorReset)
	fmt.Println("1. Rename files to remove colons and special characters")
	fmt.Println("2. Use hyphens (-) or underscores (_) instead of spaces")
	fmt.Println("3. Keep filenames under 255 characters")
	fmt.Println("4. Use only ASCII characters for maximum compatibility")
	fmt.Println("\nExample rename commands:")
	
	count := 0
	for _, v := range critical {
		if count >= 3 {
			fmt.Println("   ... (more violations found)")
			break
		}
		if strings.Contains(v.Issue, "colon") {
			suggested := strings.ReplaceAll(v.Filename, ":", "-")
			fmt.Printf("   mv \"%s\" \"%s\"\n", v.Path, filepath.Join(filepath.Dir(v.Path), suggested))
			count++
		}
	}
}

// HasViolations returns true if any violations were found
func (v *FileVerifier) HasViolations() bool {
	// Only count critical violations (colons and other serious issues)
	for _, violation := range v.violations {
		if strings.Contains(violation.Issue, "colon") || 
		   strings.Contains(violation.Issue, "less-than") || 
		   strings.Contains(violation.Issue, "greater-than") ||
		   strings.Contains(violation.Issue, "pipe") ||
		   strings.Contains(violation.Issue, "quote") {
			return true
		}
	}
	return false
}

func main() {
	// Determine directory to scan
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	// Resolve to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError: Failed to resolve directory path: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(ExitError)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "%sError: Directory does not exist: %s%s\n", ColorRed, absDir, ColorReset)
		os.Exit(ExitError)
	}

	// Create and run verifier
	verifier := NewFileVerifier(absDir)
	
	if err := verifier.Verify(); err != nil {
		fmt.Fprintf(os.Stderr, "%sError during verification: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(ExitError)
	}

	// Print report
	verifier.PrintReport()

	// Exit with appropriate code
	if verifier.HasViolations() {
		os.Exit(ExitViolationFound)
	}
	
	os.Exit(ExitSuccess)
}