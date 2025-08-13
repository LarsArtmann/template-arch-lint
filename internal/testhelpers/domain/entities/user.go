// Package entities provides test helpers for domain entities.
// These helpers eliminate repetitive entity creation and validation testing patterns.
package entities

import (
	"fmt"

	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/internal/testhelpers/base"
)

// UserBuilder provides a fluent API for creating test User entities.
// This eliminates the repetitive user creation patterns found in existing tests.
type UserBuilder struct {
	id    values.UserID
	email string
	name  string

	// Configuration options
	useDefaults   bool
	shouldBeValid bool
	invalidField  string
}

// NewUserBuilder creates a new UserBuilder with no defaults.
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{}
}

// WithID sets the user ID for the builder.
func (b *UserBuilder) WithID(id string) *UserBuilder {
	userID, err := values.NewUserID(id)
	Expect(err).ToNot(HaveOccurred()) // Fail fast on invalid test data
	b.id = userID
	return b
}

// WithUserID sets the user ID value object for the builder.
func (b *UserBuilder) WithUserID(id values.UserID) *UserBuilder {
	b.id = id
	return b
}

// WithEmail sets the email for the builder.
func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}

// WithName sets the name for the builder.
func (b *UserBuilder) WithName(name string) *UserBuilder {
	b.name = name
	return b
}

// WithDefaults sets commonly used default values for testing.
func (b *UserBuilder) WithDefaults() *UserBuilder {
	b.useDefaults = true
	return b
}

// WithValidData ensures all fields have valid test data.
func (b *UserBuilder) WithValidData() *UserBuilder {
	b.shouldBeValid = true
	b.useDefaults = true
	return b
}

// WithInvalidData sets invalid data for a specific field.
func (b *UserBuilder) WithInvalidData(field string) *UserBuilder {
	b.invalidField = field
	b.shouldBeValid = false
	return b
}

// Reset clears all configuration and returns to default state.
func (b *UserBuilder) Reset() *UserBuilder {
	return &UserBuilder{}
}

// Clone creates a copy of the builder with current configuration.
func (b *UserBuilder) Clone() *UserBuilder {
	return &UserBuilder{
		id:            b.id,
		email:         b.email,
		name:          b.name,
		useDefaults:   b.useDefaults,
		shouldBeValid: b.shouldBeValid,
		invalidField:  b.invalidField,
	}
}

// Build creates the User entity with current configuration.
func (b *UserBuilder) Build() *entities.User {
	user, err := b.BuildWithError()
	if b.shouldBeValid {
		Expect(err).ToNot(HaveOccurred())
	}
	return user
}

// BuildWithError creates the User entity and returns any construction error.
func (b *UserBuilder) BuildWithError() (*entities.User, error) {
	b.applyDefaults()
	b.applyInvalidData()

	return entities.NewUser(b.id, b.email, b.name)
}

// applyDefaults sets default values if requested and not already set.
func (b *UserBuilder) applyDefaults() {
	if !b.useDefaults {
		return
	}

	if b.id.IsEmpty() {
		id, err := values.NewUserID("test-user-123")
		Expect(err).ToNot(HaveOccurred())
		b.id = id
	}

	if b.email == "" {
		b.email = "test@example.com"
	}

	if b.name == "" {
		b.name = "Test User"
	}
}

// applyInvalidData sets invalid data for the specified field.
func (b *UserBuilder) applyInvalidData() {
	if b.invalidField == "" {
		return
	}

	switch b.invalidField {
	case "id":
		b.id = values.UserID{} // Empty ID
	case "email":
		b.email = "invalid-email-format"
	case "name":
		b.name = ""
	case "email_empty":
		b.email = ""
	case "name_short":
		b.name = "A"
	case "name_numbers":
		b.name = "123"
	case "email_spaces":
		b.email = "test @example.com"
	}
}

// UserTestSuite provides comprehensive User entity testing functionality.
type UserTestSuite struct {
	*base.GinkgoSuite
	builder *UserBuilder
}

// NewUserTestSuite creates a new User entity test suite.
func NewUserTestSuite() *UserTestSuite {
	return &UserTestSuite{
		GinkgoSuite: base.NewGinkgoSuite(),
		builder:     NewUserBuilder(),
	}
}

// CreateValidUser creates a valid user for testing.
func (s *UserTestSuite) CreateValidUser() *entities.User {
	return s.builder.WithValidData().Build()
}

// CreateValidUserWithID creates a valid user with specific ID.
func (s *UserTestSuite) CreateValidUserWithID(id string) *entities.User {
	return s.builder.Reset().WithID(id).WithValidData().Build()
}

// CreateUserWithInvalidField creates a user with invalid data for testing validation.
func (s *UserTestSuite) CreateUserWithInvalidField(field string) (*entities.User, error) {
	return s.builder.Reset().WithInvalidData(field).BuildWithError()
}

// AssertUserValidationError verifies that user creation fails with validation error.
func (s *UserTestSuite) AssertUserValidationError(field string) {
	user, err := s.CreateUserWithInvalidField(field)
	base.AssertValidationErrorForField(user, err, field)
}

// AssertValidUserCreation verifies successful user creation with expected values.
func (s *UserTestSuite) AssertValidUserCreation(user *entities.User, expectedID values.UserID, expectedEmail, expectedName string) {
	base.AssertSuccess(user, nil)
	Expect(user.ID).To(Equal(expectedID))
	Expect(user.Email).To(Equal(expectedEmail))
	Expect(user.Name).To(Equal(expectedName))
}

// Convenience functions for global use

// DefaultTestUser creates a user with standard test values.
func DefaultTestUser() *entities.User {
	return NewUserBuilder().WithDefaults().Build()
}

// TestUserWithID creates a user with specific ID and default other values.
func TestUserWithID(id string) *entities.User {
	return NewUserBuilder().WithID(id).WithDefaults().Build()
}

// TestUserWithEmail creates a user with specific email and default other values.
func TestUserWithEmail(email string) *entities.User {
	return NewUserBuilder().WithEmail(email).WithDefaults().Build()
}

// TestUserWithName creates a user with specific name and default other values.
func TestUserWithName(name string) *entities.User {
	return NewUserBuilder().WithName(name).WithDefaults().Build()
}

// TestUsersMany creates multiple test users with incremented IDs.
func TestUsersMany(count int) []*entities.User {
	users := make([]*entities.User, count)
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("test-user-%d", i+1)
		email := fmt.Sprintf("user%d@example.com", i+1)
		name := fmt.Sprintf("User %d", i+1)

		users[i] = NewUserBuilder().
			WithID(id).
			WithEmail(email).
			WithName(name).
			Build()
	}
	return users
}

// ValidateUserCreationSuccess is a test helper that verifies successful user creation.
func ValidateUserCreationSuccess(user *entities.User, err error, expectedID values.UserID, expectedEmail, expectedName string) {
	base.AssertSuccess(user, err)
	Expect(user.ID).To(Equal(expectedID))
	Expect(user.Email).To(Equal(expectedEmail))
	Expect(user.Name).To(Equal(expectedName))
	Expect(user.Created).ToNot(BeZero())
	Expect(user.Modified).ToNot(BeZero())
	Expect(user.Created).To(Equal(user.Modified))
}

// ValidateUserValidationError is a test helper that verifies validation error for specific field.
func ValidateUserValidationError(user *entities.User, err error, expectedFieldError string) {
	base.AssertValidationErrorForField(user, err, expectedFieldError)
}
