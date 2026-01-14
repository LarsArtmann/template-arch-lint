package values

import (
	"fmt"
	"slices"
	"strconv"

	domainerrors "github.com/LarsArtmann/template-arch-lint/pkg/errors"
)

// Port represents a valid network port number with business rules validation.
type Port int

// PortRange defines the valid range for network ports.
const (
	MinPort Port = 1
	MaxPort Port = 65535
)

// Well-known ports for reference.
const (
	DefaultHTTPPort  Port = 8080
	DefaultHTTPSPort Port = 8443
	DefaultDBPort    Port = 5432
)

// NewPort creates a new Port with validation.
func NewPort(value int) (Port, error) {
	port := Port(value)
	if err := port.Validate(); err != nil {
		return 0, err
	}

	return port, nil
}

// NewPortFromString creates a Port from string representation.
func NewPortFromString(value string) (Port, error) {
	if value == "" {
		return 0, domainerrors.NewValidationError("port", "port cannot be empty")
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, domainerrors.NewValidationError("port", "invalid port format: "+value)
	}

	return NewPort(intValue)
}

// Validate checks if the port is within valid range.
func (p Port) Validate() error {
	if p < MinPort {
		return domainerrors.NewValidationError("port", fmt.Sprintf("port %d is too low, minimum is %d", p, MinPort))
	}
	if p > MaxPort {
		return domainerrors.NewValidationError("port", fmt.Sprintf("port %d is too high, maximum is %d", p, MaxPort))
	}

	return nil
}

// IsValid returns true if the port is within valid range.
func (p Port) IsValid() bool {
	return p.Validate() == nil
}

// Int returns the port as an integer.
func (p Port) Int() int {
	return int(p)
}

// String returns the port as a string.
func (p Port) String() string {
	return fmt.Sprintf("%d", p)
}

// IsWellKnown returns true if this is a well-known port (1-1023).
func (p Port) IsWellKnown() bool {
	return p >= 1 && p <= 1023
}

// IsEphemeral returns true if this is an ephemeral port (32768-65535 on Linux).
func (p Port) IsEphemeral() bool {
	return p >= 32768 && p <= 65535
}

// IsDevelopment returns true if this is a common development port.
func (p Port) IsDevelopment() bool {
	developmentPorts := []Port{3000, 3001, 4200, 5000, 8000, 8080, 8443, 9000}

	return slices.Contains(developmentPorts, p)
}

// MarshalText implements encoding.TextMarshaler.
func (p Port) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *Port) UnmarshalText(text []byte) error {
	port, err := NewPortFromString(string(text))
	if err != nil {
		return err
	}
	*p = port

	return nil
}
