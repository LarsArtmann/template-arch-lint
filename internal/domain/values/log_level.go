package values

import (
	"fmt"
	"slices"
	"strings"

	domainerrors "github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

// LogLevel represents a valid logging level with business rules validation.
type LogLevel string

// Valid log levels following standard logging practices.
const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic" // Only for critical system failures
)

// getLogLevelHierarchy returns the log level hierarchy for comparison.
func getLogLevelHierarchy() map[LogLevel]int {
	return map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
		LogLevelPanic: 5,
	}
}

// getValidLogLevels returns all valid log levels for validation.
func getValidLogLevels() []LogLevel {
	return []LogLevel{
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelFatal,
		LogLevelPanic,
	}
}

// NewLogLevel creates a new LogLevel with validation.
func NewLogLevel(value string) (LogLevel, error) {
	// Normalize input - convert to lowercase and trim whitespace
	normalized := LogLevel(strings.ToLower(strings.TrimSpace(value)))

	if err := normalized.Validate(); err != nil {
		return "", err
	}

	return normalized, nil
}

// Validate checks if the log level is valid.
func (l LogLevel) Validate() error {
	if l == "" {
		return domainerrors.NewValidationError("log_level", "log level cannot be empty")
	}

	if slices.Contains(getValidLogLevels(), l) {
		return nil
	}

	return domainerrors.NewValidationError("log_level",
		fmt.Sprintf("invalid log level '%s', must be one of: %s",
			l, strings.Join(l.ValidLevels(), ", ")))
}

// IsValid returns true if the log level is valid.
func (l LogLevel) IsValid() bool {
	return l.Validate() == nil
}

// String returns the log level as a string.
func (l LogLevel) String() string {
	return string(l)
}

// ValidLevels returns a slice of all valid log level strings.
func (l LogLevel) ValidLevels() []string {
	validLevels := getValidLogLevels()
	levels := make([]string, len(validLevels))
	for i, level := range validLevels {
		levels[i] = string(level)
	}

	return levels
}

// IsProduction returns true if this log level is appropriate for production.
func (l LogLevel) IsProduction() bool {
	return l == LogLevelInfo || l == LogLevelWarn || l == LogLevelError
}

// IsDevelopment returns true if this log level is useful for development.
func (l LogLevel) IsDevelopment() bool {
	return l == LogLevelDebug || l == LogLevelInfo
}

// IsError returns true if this represents an error condition.
func (l LogLevel) IsError() bool {
	return l == LogLevelError || l == LogLevelFatal || l == LogLevelPanic
}

// Priority returns the priority level (higher number = more severe).
func (l LogLevel) Priority() int {
	if priority, exists := getLogLevelHierarchy()[l]; exists {
		return priority
	}

	return -1 // Invalid level
}

// IsMoreSevereThan returns true if this level is more severe than the other.
func (l LogLevel) IsMoreSevereThan(other LogLevel) bool {
	return l.Priority() > other.Priority()
}

// IsLessOrEqualSevereThan returns true if this level is less or equal severity.
func (l LogLevel) IsLessOrEqualSevereThan(other LogLevel) bool {
	return l.Priority() <= other.Priority()
}

// ShouldLog returns true if a message at this level should be logged
// given the configured minimum log level.
func (l LogLevel) ShouldLog(minimumLevel LogLevel) bool {
	return l.Priority() >= minimumLevel.Priority()
}

// MarshalText implements encoding.TextMarshaler.
func (l LogLevel) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (l *LogLevel) UnmarshalText(text []byte) error {
	level, err := NewLogLevel(string(text))
	if err != nil {
		return err
	}
	*l = level

	return nil
}

// DefaultLogLevel returns the recommended default log level for production.
func DefaultLogLevel() LogLevel {
	return LogLevelInfo
}

// DevelopmentLogLevel returns the recommended log level for development.
func DevelopmentLogLevel() LogLevel {
	return LogLevelDebug
}
