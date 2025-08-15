// Package constants provides application-wide constants to eliminate magic numbers
// and improve code readability and maintainability.
package constants

// ArchitectureAnalysis contains constants for architecture testing and cycle detection
const (
	// NoCycleFound indicates that no dependency cycle was found during graph traversal.
	// Used as a sentinel value to mark that the cycle start position is undefined.
	NoCycleFound = -1

	// EmptyCycle represents the length of an empty cycle slice.
	// Used to check if a dependency cycle was detected during graph analysis.
	EmptyCycle = 0
)

// Testing contains constants for testing and validation
const (
	// MinimumMethodCount represents the minimum number of methods expected
	// in repository interfaces to ensure they are properly defined.
	MinimumMethodCount = 0

	// FirstParameterIndex represents the index of the first parameter
	// in a method signature for validation purposes.
	FirstParameterIndex = 0

	// FirstCharacterIndex represents the index of the first character
	// in a string for case checking operations.
	FirstCharacterIndex = 0

	// SingleCharacterLength represents the length of a single character substring.
	SingleCharacterLength = 1
)

// PackageAnalysis contains constants for package and dependency analysis
const (
	// NoPackagesFound indicates that no packages were found during analysis.
	NoPackagesFound = 0

	// FirstField represents the index of the first field in a struct
	// when iterating through struct fields using reflection.
	FirstField = 0
)
