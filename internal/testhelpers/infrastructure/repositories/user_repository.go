// Package repositories provides test helpers for repository layer testing.
// These helpers eliminate repetitive repository setup and testing patterns.
package repositories

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/internal/testhelpers/base"
	entityHelpers "github.com/LarsArtmann/template-arch-lint/internal/testhelpers/domain/entities"
	valueHelpers "github.com/LarsArtmann/template-arch-lint/internal/testhelpers/domain/values"
)

// UserRepositoryTestSuite provides comprehensive UserRepository testing functionality.
type UserRepositoryTestSuite struct {
	*base.GinkgoSuite
	repository repositories.UserRepository
	ctx        context.Context
}

// NewUserRepositoryTestSuite creates a new UserRepository test suite.
func NewUserRepositoryTestSuite(repo repositories.UserRepository) *UserRepositoryTestSuite {
	return &UserRepositoryTestSuite{
		GinkgoSuite: base.NewGinkgoSuite(),
		repository:  repo,
	}
}

// Setup initializes the test suite with fresh context and repository.
func (s *UserRepositoryTestSuite) Setup() {
	s.GinkgoSuite.Setup()
	s.ctx = s.GetContext()
}

// CreateTestUser creates a user in the repository for testing.
func (s *UserRepositoryTestSuite) CreateTestUser() *entities.User {
	user := entityHelpers.DefaultTestUser()
	err := s.repository.Save(s.ctx, user)
	Expect(err).ToNot(HaveOccurred())
	return user
}

// CreateTestUserWithID creates a user with specific ID in the repository.
func (s *UserRepositoryTestSuite) CreateTestUserWithID(id string) *entities.User {
	user := entityHelpers.TestUserWithID(id)
	err := s.repository.Save(s.ctx, user)
	Expect(err).ToNot(HaveOccurred())
	return user
}

// CreateTestUsers creates multiple users in the repository.
func (s *UserRepositoryTestSuite) CreateTestUsers(count int) []*entities.User {
	users := entityHelpers.CreateTestUserCollection(count)
	for _, user := range users {
		err := s.repository.Save(s.ctx, user)
		Expect(err).ToNot(HaveOccurred())
	}
	return users
}

// AssertUserExists verifies a user exists in the repository.
func (s *UserRepositoryTestSuite) AssertUserExists(id values.UserID) {
	user, err := s.repository.FindByID(s.ctx, id)
	base.AssertSuccess(user, err)
}

// AssertUserNotExists verifies a user does not exist in the repository.
func (s *UserRepositoryTestSuite) AssertUserNotExists(id values.UserID) {
	user, err := s.repository.FindByID(s.ctx, id)
	Expect(user).To(BeNil())
	Expect(err).To(Equal(repositories.ErrUserNotFound))
}

// AssertUserByEmailExists verifies a user exists by email.
func (s *UserRepositoryTestSuite) AssertUserByEmailExists(email string) {
	user, err := s.repository.FindByEmail(s.ctx, email)
	base.AssertSuccess(user, err)
}

// AssertUserByEmailNotExists verifies a user does not exist by email.
func (s *UserRepositoryTestSuite) AssertUserByEmailNotExists(email string) {
	user, err := s.repository.FindByEmail(s.ctx, email)
	Expect(user).To(BeNil())
	Expect(err).To(Equal(repositories.ErrUserNotFound))
}

// AssertUserCount verifies the total number of users in the repository.
func (s *UserRepositoryTestSuite) AssertUserCount(expectedCount int) {
	users, err := s.repository.List(s.ctx)
	base.AssertSuccess(users, err)
	Expect(users).To(HaveLen(expectedCount))
}

// AssertRepositoryEmpty verifies the repository contains no users.
func (s *UserRepositoryTestSuite) AssertRepositoryEmpty() {
	s.AssertUserCount(0)
}

// TestRepositoryCRUD runs comprehensive CRUD tests on the repository.
func (s *UserRepositoryTestSuite) TestRepositoryCRUD() {
	// Test Create (Save)
	user := entityHelpers.DefaultTestUser()
	err := s.repository.Save(s.ctx, user)
	base.AssertSuccess(user, err)

	// Test Read (FindByID)
	foundUser, err := s.repository.FindByID(s.ctx, user.ID)
	base.AssertSuccess(foundUser, err)
	Expect(foundUser.ID).To(Equal(user.ID))
	Expect(foundUser.Email).To(Equal(user.Email))
	Expect(foundUser.Name).To(Equal(user.Name))

	// Test Update (Save existing)
	foundUser.Name = "Updated Name"
	err = s.repository.Save(s.ctx, foundUser)
	base.AssertSuccess(foundUser, err)

	updatedUser, err := s.repository.FindByID(s.ctx, user.ID)
	base.AssertSuccess(updatedUser, err)
	Expect(updatedUser.Name).To(Equal("Updated Name"))

	// Test Delete
	err = s.repository.Delete(s.ctx, user.ID)
	Expect(err).ToNot(HaveOccurred())

	s.AssertUserNotExists(user.ID)
}

// MockUserRepository provides a mock implementation for testing services.
type MockUserRepository struct {
	users       map[string]*entities.User
	shouldError bool
	errorType   string
	callLog     []string
}

// NewMockUserRepository creates a new mock user repository.
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[string]*entities.User),
		callLog: make([]string, 0),
	}
}

// Save implements the UserRepository interface for testing.
func (m *MockUserRepository) Save(_ context.Context, user *entities.User) error {
	m.callLog = append(m.callLog, fmt.Sprintf("Save(%s)", user.ID.String()))

	if m.shouldError {
		switch m.errorType {
		case "conflict":
			return repositories.ErrUserAlreadyExists
		default:
			return fmt.Errorf("mock repository error")
		}
	}

	m.users[user.ID.String()] = user
	return nil
}

// FindByID implements the UserRepository interface for testing.
func (m *MockUserRepository) FindByID(_ context.Context, id values.UserID) (*entities.User, error) {
	m.callLog = append(m.callLog, fmt.Sprintf("FindByID(%s)", id.String()))

	if m.shouldError {
		return nil, repositories.ErrUserNotFound
	}

	user, exists := m.users[id.String()]
	if !exists {
		return nil, repositories.ErrUserNotFound
	}

	return user, nil
}

// FindByEmail implements the UserRepository interface for testing.
func (m *MockUserRepository) FindByEmail(_ context.Context, email string) (*entities.User, error) {
	m.callLog = append(m.callLog, fmt.Sprintf("FindByEmail(%s)", email))

	if m.shouldError {
		return nil, repositories.ErrUserNotFound
	}

	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, repositories.ErrUserNotFound
}

// List implements the UserRepository interface for testing.
func (m *MockUserRepository) List(_ context.Context) ([]*entities.User, error) {
	m.callLog = append(m.callLog, "List()")

	if m.shouldError {
		return nil, fmt.Errorf("mock repository error")
	}

	users := make([]*entities.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}

	return users, nil
}

// Delete implements the UserRepository interface for testing.
func (m *MockUserRepository) Delete(_ context.Context, id values.UserID) error {
	m.callLog = append(m.callLog, fmt.Sprintf("Delete(%s)", id.String()))

	if m.shouldError {
		return repositories.ErrUserNotFound
	}

	_, exists := m.users[id.String()]
	if !exists {
		return repositories.ErrUserNotFound
	}

	delete(m.users, id.String())
	return nil
}

// SetShouldError configures the mock to return errors.
func (m *MockUserRepository) SetShouldError(shouldError bool, errorType string) {
	m.shouldError = shouldError
	m.errorType = errorType
}

// GetCallLog returns the log of method calls made to the mock.
func (m *MockUserRepository) GetCallLog() []string {
	return m.callLog
}

// ResetCallLog clears the method call log.
func (m *MockUserRepository) ResetCallLog() {
	m.callLog = make([]string, 0)
}

// GetUserCount returns the number of users in the mock repository.
func (m *MockUserRepository) GetUserCount() int {
	return len(m.users)
}

// HasUser checks if a user exists in the mock repository.
func (m *MockUserRepository) HasUser(id values.UserID) bool {
	_, exists := m.users[id.String()]
	return exists
}

// Clear removes all users from the mock repository.
func (m *MockUserRepository) Clear() {
	m.users = make(map[string]*entities.User)
	m.callLog = make([]string, 0)
}

// PreloadUsers adds users to the mock repository for testing.
func (m *MockUserRepository) PreloadUsers(users ...*entities.User) {
	for _, user := range users {
		m.users[user.ID.String()] = user
	}
}

// RepositoryBehaviorTester provides comprehensive repository behavior testing.
type RepositoryBehaviorTester struct {
	suite *UserRepositoryTestSuite
}

// NewRepositoryBehaviorTester creates a behavior tester for repositories.
func NewRepositoryBehaviorTester(repo repositories.UserRepository) *RepositoryBehaviorTester {
	return &RepositoryBehaviorTester{
		suite: NewUserRepositoryTestSuite(repo),
	}
}

// TestSaveAndFindByID tests basic save and retrieval functionality.
func (t *RepositoryBehaviorTester) TestSaveAndFindByID() {
	user := t.suite.CreateTestUser()
	foundUser, err := t.suite.repository.FindByID(t.suite.ctx, user.ID)
	base.AssertSuccess(foundUser, err)
	Expect(foundUser.ID).To(Equal(user.ID))
}

// TestFindByEmail tests email-based user lookup.
func (t *RepositoryBehaviorTester) TestFindByEmail() {
	user := t.suite.CreateTestUser()
	foundUser, err := t.suite.repository.FindByEmail(t.suite.ctx, user.Email)
	base.AssertSuccess(foundUser, err)
	Expect(foundUser.Email).To(Equal(user.Email))
}

// TestUserNotFound tests not-found error scenarios.
func (t *RepositoryBehaviorTester) TestUserNotFound() {
	nonExistentID := valueHelpers.TestUserIDFromString("nonexistent-user")
	t.suite.AssertUserNotExists(nonExistentID)

	t.suite.AssertUserByEmailNotExists("nonexistent@example.com")
}

// TestListUsers tests listing all users.
func (t *RepositoryBehaviorTester) TestListUsers() {
	users := t.suite.CreateTestUsers(3)

	allUsers, err := t.suite.repository.List(t.suite.ctx)
	base.AssertSuccess(allUsers, err)
	Expect(allUsers).To(HaveLen(3))

	// Verify all created users are in the list
	for _, createdUser := range users {
		found := false
		for _, listedUser := range allUsers {
			if listedUser.ID.Equals(createdUser.ID) {
				found = true
				break
			}
		}
		Expect(found).To(BeTrue(), fmt.Sprintf("User %s should be in the list", createdUser.ID.String()))
	}
}

// TestDeleteUser tests user deletion.
func (t *RepositoryBehaviorTester) TestDeleteUser() {
	user := t.suite.CreateTestUser()

	// Verify user exists
	t.suite.AssertUserExists(user.ID)

	// Delete user
	err := t.suite.repository.Delete(t.suite.ctx, user.ID)
	Expect(err).ToNot(HaveOccurred())

	// Verify user no longer exists
	t.suite.AssertUserNotExists(user.ID)
}

// TestUpdateUser tests user updates through save.
func (t *RepositoryBehaviorTester) TestUpdateUser() {
	user := t.suite.CreateTestUser()

	// Update user data
	originalEmail := user.Email
	user.Email = "updated@example.com"
	user.Name = "Updated Name"

	// Save updated user
	err := t.suite.repository.Save(t.suite.ctx, user)
	Expect(err).ToNot(HaveOccurred())

	// Retrieve and verify updates
	updatedUser, err := t.suite.repository.FindByID(t.suite.ctx, user.ID)
	base.AssertSuccess(updatedUser, err)
	Expect(updatedUser.Email).To(Equal("updated@example.com"))
	Expect(updatedUser.Name).To(Equal("Updated Name"))

	// Verify old email no longer works
	_, err = t.suite.repository.FindByEmail(t.suite.ctx, originalEmail)
	Expect(err).To(Equal(repositories.ErrUserNotFound))
}

// RunAllRepositoryTests executes a comprehensive repository test suite.
func (t *RepositoryBehaviorTester) RunAllRepositoryTests() {
	t.TestSaveAndFindByID()
	t.TestFindByEmail()
	t.TestUserNotFound()
	t.TestListUsers()
	t.TestDeleteUser()
	t.TestUpdateUser()
}
