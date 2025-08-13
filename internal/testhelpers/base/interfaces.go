// Package base provides core interfaces and types for the test helper framework.
// These interfaces establish contracts that all test helpers must implement.
package base

import (
	"context"
)

// TestBuilder provides a common interface for test data builders.
// All entity builders should implement this interface for consistent usage patterns.
type TestBuilder[T any] interface {
	// Build creates the entity with current configuration
	Build() T

	// BuildWithError creates the entity and returns any construction error
	BuildWithError() (T, error)

	// Reset clears all configuration and returns to default state
	Reset() TestBuilder[T]

	// Clone creates a copy of the builder with current configuration
	Clone() TestBuilder[T]
}

// EntityBuilder extends TestBuilder with entity-specific functionality.
type EntityBuilder[T any] interface {
	TestBuilder[T]

	// WithDefaults sets commonly used default values for testing
	WithDefaults() EntityBuilder[T]

	// WithValidData ensures all fields have valid test data
	WithValidData() EntityBuilder[T]

	// WithInvalidData sets invalid data for validation testing
	WithInvalidData(field string) EntityBuilder[T]
}

// RepositoryTestHelper provides standardized repository testing functionality.
type RepositoryTestHelper[T any, ID any] interface {
	// SetupRepository initializes the repository for testing
	SetupRepository()

	// CleanupRepository cleans up repository state after testing
	CleanupRepository()

	// CreateTestEntity creates an entity for testing purposes
	CreateTestEntity() T

	// CreateTestEntityWithID creates an entity with a specific ID
	CreateTestEntityWithID(id ID) T

	// AssertEntityExists verifies an entity exists in the repository
	AssertEntityExists(id ID)

	// AssertEntityNotExists verifies an entity does not exist in the repository
	AssertEntityNotExists(id ID)

	// GetTestContext returns the context for repository operations
	GetTestContext() context.Context
}

// ServiceTestHelper provides standardized service testing functionality.
type ServiceTestHelper[T any] interface {
	// SetupService initializes the service with test dependencies
	SetupService()

	// CleanupService cleans up service state after testing
	CleanupService()

	// GetService returns the service instance for testing
	GetService() T

	// SetupMockDependencies configures mock dependencies for the service
	SetupMockDependencies()

	// ResetMocks resets all mock dependencies to clean state
	ResetMocks()

	// GetTestContext returns the context for service operations
	GetTestContext() context.Context
}

// HandlerTestHelper provides standardized HTTP handler testing functionality.
type HandlerTestHelper interface {
	// SetupRouter initializes the test router with handlers
	SetupRouter()

	// CleanupRouter cleans up router state after testing
	CleanupRouter()

	// SetupMockServices configures mock services for handlers
	SetupMockServices()

	// ResetMockServices resets all mock services to clean state
	ResetMockServices()

	// CreateRequest creates an HTTP request for testing
	CreateRequest(method, path string, body any) any

	// ExecuteRequest executes an HTTP request and returns response
	ExecuteRequest(request any) any

	// AssertStatusCode verifies the HTTP response status code
	AssertStatusCode(response any, expectedStatus int)

	// AssertResponseBody verifies the HTTP response body content
	AssertResponseBody(response any, expectedBody any)
}

// ValidationTestHelper provides standardized validation testing functionality.
type ValidationTestHelper interface {
	// AssertValidationPasses verifies that validation succeeds for valid input
	AssertValidationPasses(entity any)

	// AssertValidationFails verifies that validation fails for invalid input
	AssertValidationFails(entity any, expectedFieldErrors ...string)

	// AssertValidationErrorMessage verifies specific validation error messages
	AssertValidationErrorMessage(err error, expectedMessage string)

	// CreateInvalidEntityForField creates an entity with invalid data for specific field
	CreateInvalidEntityForField(field string) any
}

// TestDataFixture provides standardized test data management.
type TestDataFixture[T any] interface {
	// LoadFixture loads test data from fixture definition
	LoadFixture(name string) T

	// SaveFixture saves test data as a named fixture
	SaveFixture(name string, data T)

	// ListFixtures returns all available fixture names
	ListFixtures() []string

	// ClearFixtures removes all loaded fixtures
	ClearFixtures()

	// CreateDefault creates default test data
	CreateDefault() T

	// CreateMany creates multiple test entities with variations
	CreateMany(count int) []T
}

// MockController provides standardized mock management functionality.
type MockController interface {
	// SetupMocks initializes all mocks for testing
	SetupMocks()

	// ResetMocks resets all mocks to clean state
	ResetMocks()

	// CleanupMocks performs cleanup after testing
	CleanupMocks()

	// VerifyMocks verifies all mock expectations were met
	VerifyMocks()

	// SetExpectation sets up an expectation on a mock
	SetExpectation(mockName string, method string, args []any, result any)

	// GetMock retrieves a named mock instance
	GetMock(name string) any
}

// IntegrationTestHelper provides functionality for integration testing.
type IntegrationTestHelper interface {
	// SetupIntegrationEnvironment initializes the full test environment
	SetupIntegrationEnvironment()

	// CleanupIntegrationEnvironment cleans up the test environment
	CleanupIntegrationEnvironment()

	// SeedTestData populates the environment with test data
	SeedTestData()

	// ClearTestData removes all test data from the environment
	ClearTestData()

	// GetDatabaseConnection returns test database connection
	GetDatabaseConnection() any

	// GetTestConfiguration returns test-specific configuration
	GetTestConfiguration() any
}

// TestMetrics provides testing metrics and reporting functionality.
type TestMetrics interface {
	// StartTimer starts timing an operation
	StartTimer(operationName string)

	// EndTimer ends timing an operation and records the duration
	EndTimer(operationName string)

	// RecordTestExecution records test execution data
	RecordTestExecution(testName string, success bool, duration int64)

	// GetMetrics returns collected metrics
	GetMetrics() map[string]any

	// ResetMetrics clears all collected metrics
	ResetMetrics()
}

// TestConfig provides test-specific configuration management.
type TestConfig interface {
	// GetTestValue retrieves a test configuration value
	GetTestValue(key string) any

	// SetTestValue sets a test configuration value
	SetTestValue(key string, value any)

	// LoadTestConfig loads configuration from test-specific sources
	LoadTestConfig()

	// IsTestMode returns true if running in test mode
	IsTestMode() bool

	// GetTestEnv returns the test environment name
	GetTestEnv() string
}
