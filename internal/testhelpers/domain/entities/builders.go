// Package entities provides test data builders for domain entities.
// These builders use the Builder pattern to create complex test scenarios with ease.
package entities

import (
	"fmt"

	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/testhelpers/base"
)

// TestDataBuilder provides a comprehensive builder for creating test scenarios.
type TestDataBuilder struct {
	userCount   int
	userPrefix  string
	emailDomain string
	namePrefix  string
	useSequence bool
	customUsers []*entities.User
}

// NewTestDataBuilder creates a new test data builder.
func NewTestDataBuilder() *TestDataBuilder {
	return &TestDataBuilder{
		userCount:   1,
		userPrefix:  "test-user",
		emailDomain: "example.com",
		namePrefix:  "Test User",
		useSequence: true,
	}
}

// WithUserCount sets the number of users to create.
func (b *TestDataBuilder) WithUserCount(count int) *TestDataBuilder {
	b.userCount = count
	return b
}

// WithUserPrefix sets the prefix for user IDs.
func (b *TestDataBuilder) WithUserPrefix(prefix string) *TestDataBuilder {
	b.userPrefix = prefix
	return b
}

// WithEmailDomain sets the email domain for generated users.
func (b *TestDataBuilder) WithEmailDomain(domain string) *TestDataBuilder {
	b.emailDomain = domain
	return b
}

// WithNamePrefix sets the prefix for user names.
func (b *TestDataBuilder) WithNamePrefix(prefix string) *TestDataBuilder {
	b.namePrefix = prefix
	return b
}

// WithSequentialNaming enables sequential numbering for users.
func (b *TestDataBuilder) WithSequentialNaming() *TestDataBuilder {
	b.useSequence = true
	return b
}

// WithCustomUsers adds specific users to the test data.
func (b *TestDataBuilder) WithCustomUsers(users ...*entities.User) *TestDataBuilder {
	b.customUsers = append(b.customUsers, users...)
	return b
}

// BuildUsers creates the configured test users.
func (b *TestDataBuilder) BuildUsers() []*entities.User {
	users := make([]*entities.User, 0, b.userCount+len(b.customUsers))

	// Add custom users first
	users = append(users, b.customUsers...)

	// Generate standard users
	for i := 0; i < b.userCount; i++ {
		var id, email, name string

		if b.useSequence {
			id = fmt.Sprintf("%s-%d", b.userPrefix, i+1)
			email = fmt.Sprintf("%s%d@%s", b.userPrefix, i+1, b.emailDomain)
			name = fmt.Sprintf("%s %d", b.namePrefix, i+1)
		} else {
			id = b.userPrefix
			email = fmt.Sprintf("%s@%s", b.userPrefix, b.emailDomain)
			name = b.namePrefix
		}

		user := NewUserBuilder().
			WithID(id).
			WithEmail(email).
			WithName(name).
			Build()

		users = append(users, user)
	}

	return users
}

// BuildSingleUser creates one user with the configured settings.
func (b *TestDataBuilder) BuildSingleUser() *entities.User {
	users := b.WithUserCount(1).BuildUsers()
	return users[0]
}

// UserScenarioBuilder builds common user testing scenarios.
type UserScenarioBuilder struct{}

// NewUserScenarioBuilder creates a new scenario builder.
func NewUserScenarioBuilder() *UserScenarioBuilder {
	return &UserScenarioBuilder{}
}

// ValidUser creates a completely valid user for positive testing.
func (b *UserScenarioBuilder) ValidUser() *entities.User {
	return NewUserBuilder().WithValidData().Build()
}

// ValidUserWithCustomID creates a valid user with custom ID.
func (b *UserScenarioBuilder) ValidUserWithCustomID(id string) *entities.User {
	return NewUserBuilder().WithID(id).WithValidData().Build()
}

// InvalidEmailUser creates a user with invalid email for validation testing.
func (b *UserScenarioBuilder) InvalidEmailUser() (*entities.User, error) {
	return NewUserBuilder().WithInvalidData("email").BuildWithError()
}

// InvalidNameUser creates a user with invalid name for validation testing.
func (b *UserScenarioBuilder) InvalidNameUser() (*entities.User, error) {
	return NewUserBuilder().WithInvalidData("name").BuildWithError()
}

// EmptyEmailUser creates a user with empty email for validation testing.
func (b *UserScenarioBuilder) EmptyEmailUser() (*entities.User, error) {
	return NewUserBuilder().WithInvalidData("email_empty").BuildWithError()
}

// ShortNameUser creates a user with too short name for validation testing.
func (b *UserScenarioBuilder) ShortNameUser() (*entities.User, error) {
	return NewUserBuilder().WithInvalidData("name_short").BuildWithError()
}

// NumericNameUser creates a user with numeric name for validation testing.
func (b *UserScenarioBuilder) NumericNameUser() (*entities.User, error) {
	return NewUserBuilder().WithInvalidData("name_numbers").BuildWithError()
}

// EmailWithSpacesUser creates a user with spaces in email for validation testing.
func (b *UserScenarioBuilder) EmailWithSpacesUser() (*entities.User, error) {
	return NewUserBuilder().WithInvalidData("email_spaces").BuildWithError()
}

// BatchUserBuilder creates multiple users for batch operation testing.
type BatchUserBuilder struct {
	users []*entities.User
}

// NewBatchUserBuilder creates a new batch builder.
func NewBatchUserBuilder() *BatchUserBuilder {
	return &BatchUserBuilder{
		users: make([]*entities.User, 0),
	}
}

// AddValidUsers adds multiple valid users to the batch.
func (b *BatchUserBuilder) AddValidUsers(count int) *BatchUserBuilder {
	for i := 0; i < count; i++ {
		user := NewUserBuilder().
			WithID(fmt.Sprintf("batch-user-%d", len(b.users)+1)).
			WithEmail(fmt.Sprintf("batch%d@example.com", len(b.users)+1)).
			WithName(fmt.Sprintf("Batch User %d", len(b.users)+1)).
			Build()
		b.users = append(b.users, user)
	}
	return b
}

// AddCustomUser adds a specific user to the batch.
func (b *BatchUserBuilder) AddCustomUser(user *entities.User) *BatchUserBuilder {
	b.users = append(b.users, user)
	return b
}

// AddUserWithPattern adds a user following a naming pattern.
func (b *BatchUserBuilder) AddUserWithPattern(idPattern, emailPattern, namePattern string, index int) *BatchUserBuilder {
	id := fmt.Sprintf(idPattern, index)
	email := fmt.Sprintf(emailPattern, index)
	name := fmt.Sprintf(namePattern, index)

	user := NewUserBuilder().
		WithID(id).
		WithEmail(email).
		WithName(name).
		Build()

	b.users = append(b.users, user)
	return b
}

// Build returns all users in the batch.
func (b *BatchUserBuilder) Build() []*entities.User {
	return b.users
}

// BuildCount returns the number of users in the batch.
func (b *BatchUserBuilder) BuildCount() int {
	return len(b.users)
}

// ValidationTestSuite provides comprehensive validation testing scenarios.
type ValidationTestSuite struct {
	*base.GinkgoSuite
}

// NewValidationTestSuite creates a validation test suite.
func NewValidationTestSuite() *ValidationTestSuite {
	return &ValidationTestSuite{
		GinkgoSuite: base.NewGinkgoSuite(),
	}
}

// TestAllValidationScenarios runs comprehensive validation tests.
func (s *ValidationTestSuite) TestAllValidationScenarios() {
	scenarios := NewUserScenarioBuilder()

	// Test invalid email scenarios
	user, err := scenarios.InvalidEmailUser()
	base.AssertValidationErrorForField(user, err, "email")

	user, err = scenarios.EmptyEmailUser()
	base.AssertValidationErrorForField(user, err, "email")

	user, err = scenarios.EmailWithSpacesUser()
	base.AssertValidationErrorForField(user, err, "email")

	// Test invalid name scenarios
	user, err = scenarios.InvalidNameUser()
	base.AssertValidationErrorForField(user, err, "name")

	user, err = scenarios.ShortNameUser()
	base.AssertValidationErrorForField(user, err, "name")

	user, err = scenarios.NumericNameUser()
	base.AssertValidationErrorForField(user, err, "name")

	// Test valid user creation
	validUser := scenarios.ValidUser()
	Expect(validUser).ToNot(BeNil())
	Expect(validUser.Email).ToNot(BeEmpty())
	Expect(validUser.Name).ToNot(BeEmpty())
	Expect(validUser.ID.IsEmpty()).To(BeFalse())
}

// Convenience functions for common test data patterns

// CreateTestUserCollection creates a collection of users for testing operations.
func CreateTestUserCollection(count int) []*entities.User {
	return NewTestDataBuilder().
		WithUserCount(count).
		WithSequentialNaming().
		BuildUsers()
}

// CreateTestUserWithDomain creates a user with specific email domain.
func CreateTestUserWithDomain(domain string) *entities.User {
	return NewTestDataBuilder().
		WithEmailDomain(domain).
		BuildSingleUser()
}

// CreateTestUsersWithPrefix creates users with specific ID prefix.
func CreateTestUsersWithPrefix(count int, prefix string) []*entities.User {
	return NewTestDataBuilder().
		WithUserCount(count).
		WithUserPrefix(prefix).
		WithSequentialNaming().
		BuildUsers()
}

// CreateMixedValidationTestData creates users with both valid and invalid data for comprehensive testing.
func CreateMixedValidationTestData() ([]*entities.User, []error) {
	scenarios := NewUserScenarioBuilder()

	validUsers := []*entities.User{
		scenarios.ValidUser(),
		scenarios.ValidUserWithCustomID("custom-123"),
	}

	var errors []error

	// Invalid users (these will fail creation)
	_, err1 := scenarios.InvalidEmailUser()
	_, err2 := scenarios.InvalidNameUser()
	_, err3 := scenarios.EmptyEmailUser()

	errors = append(errors, err1, err2, err3)

	return validUsers, errors
}
