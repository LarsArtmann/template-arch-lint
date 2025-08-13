// Package base provides fundamental test utilities for all test suites.
// These utilities establish consistent patterns for test setup, execution, and cleanup.
package base

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestSuite provides a standard interface for test suite management.
// Implementations should provide setup, teardown, and common utilities
// for specific domains or layers.
type TestSuite interface {
	// Setup initializes the test suite with required dependencies
	Setup()

	// Teardown cleans up resources after test execution
	Teardown()

	// GetContext returns the test context for the suite
	GetContext() context.Context

	// Reset clears any state between test cases
	Reset()
}

// GinkgoSuite provides base functionality for Ginkgo-based test suites.
// It implements common patterns used across all test suites in the project.
type GinkgoSuite struct {
	ctx context.Context
}

// NewGinkgoSuite creates a new base Ginkgo test suite.
func NewGinkgoSuite() *GinkgoSuite {
	return &GinkgoSuite{}
}

// Setup initializes the base test suite with a fresh context.
func (s *GinkgoSuite) Setup() {
	s.ctx = context.Background()
}

// Teardown performs cleanup for the base suite.
func (s *GinkgoSuite) Teardown() {
	s.ctx = nil
}

// GetContext returns the current test context.
func (s *GinkgoSuite) GetContext() context.Context {
	return s.ctx
}

// Reset clears the context and recreates it for the next test case.
func (s *GinkgoSuite) Reset() {
	s.ctx = context.Background()
}

// RegisterGinkgoSuite registers a Ginkgo test suite with standard configuration.
// This eliminates the repetitive RegisterFailHandler and RunSpecs boilerplate.
func RegisterGinkgoSuite(t *testing.T, suiteName string) {
	RegisterFailHandler(Fail)
	RunSpecs(t, suiteName)
}

// SetupBeforeEach creates a standard BeforeEach block that calls Setup and Reset.
// This ensures consistent initialization across all test suites.
func SetupBeforeEach(suite TestSuite) {
	BeforeEach(func() {
		suite.Setup()
		suite.Reset()
	})
}

// TeardownAfterEach creates a standard AfterEach block that calls Teardown.
// This ensures consistent cleanup across all test suites.
func TeardownAfterEach(suite TestSuite) {
	AfterEach(func() {
		suite.Teardown()
	})
}

// SetupStandardSuite combines RegisterGinkgoSuite with BeforeEach/AfterEach setup.
// This provides a one-line solution for standard test suite configuration.
func SetupStandardSuite(t *testing.T, suiteName string, suite TestSuite) {
	RegisterGinkgoSuite(t, suiteName)

	BeforeEach(func() {
		suite.Setup()
		suite.Reset()
	})

	AfterEach(func() {
		suite.Teardown()
	})
}

// DescribeStandardCRUD creates a standard CRUD test structure.
// This eliminates repetitive Describe/Context blocks for entity operations.
func DescribeStandardCRUD(entityName string, createTests, readTests, updateTests, deleteTests func()) {
	Describe(entityName, func() {
		Describe("Create", func() {
			Context("with valid input", func() {
				createTests()
			})
		})

		Describe("Read", func() {
			Context("when entity exists", func() {
				readTests()
			})
		})

		Describe("Update", func() {
			Context("with valid changes", func() {
				updateTests()
			})
		})

		Describe("Delete", func() {
			Context("when entity exists", func() {
				deleteTests()
			})
		})
	})
}
