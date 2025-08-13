// Package repositories provides in-memory repository implementations and helpers for testing.
// These helpers provide consistent test data management and repository behavior.
package repositories

import (
	"context"
	"fmt"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/LarsArtmann/template-arch-lint/internal/testhelpers/base"
	entityHelpers "github.com/LarsArtmann/template-arch-lint/internal/testhelpers/domain/entities"
)

// InMemoryTestRepository provides an enhanced in-memory repository for testing.
// This extends the basic InMemoryUserRepository with additional testing features.
type InMemoryTestRepository struct {
	*repositories.InMemoryUserRepository
	testData     *RepositoryTestData
	operationLog []string
	beforeSave   func(*entities.User) error
	afterSave    func(*entities.User) error
	beforeDelete func(values.UserID) error
	afterDelete  func(values.UserID) error
}

// NewInMemoryTestRepository creates an enhanced in-memory repository for testing.
func NewInMemoryTestRepository() *InMemoryTestRepository {
	return &InMemoryTestRepository{
		InMemoryUserRepository: repositories.NewInMemoryUserRepository().(*repositories.InMemoryUserRepository),
		testData:               NewRepositoryTestData(),
		operationLog:           make([]string, 0),
	}
}

// Save wraps the base Save method with test enhancements.
func (r *InMemoryTestRepository) Save(ctx context.Context, user *entities.User) error {
	r.operationLog = append(r.operationLog, "Save")

	if r.beforeSave != nil {
		if err := r.beforeSave(user); err != nil {
			return err
		}
	}

	err := r.InMemoryUserRepository.Save(ctx, user)

	if err == nil && r.afterSave != nil {
		if afterErr := r.afterSave(user); afterErr != nil {
			return afterErr
		}
	}

	return err
}

// Delete wraps the base Delete method with test enhancements.
func (r *InMemoryTestRepository) Delete(ctx context.Context, id values.UserID) error {
	r.operationLog = append(r.operationLog, "Delete")

	if r.beforeDelete != nil {
		if err := r.beforeDelete(id); err != nil {
			return err
		}
	}

	err := r.InMemoryUserRepository.Delete(ctx, id)

	if err == nil && r.afterDelete != nil {
		if afterErr := r.afterDelete(id); afterErr != nil {
			return afterErr
		}
	}

	return err
}

// SetBeforeSaveHook sets a callback to execute before saving users.
func (r *InMemoryTestRepository) SetBeforeSaveHook(hook func(*entities.User) error) {
	r.beforeSave = hook
}

// SetAfterSaveHook sets a callback to execute after saving users.
func (r *InMemoryTestRepository) SetAfterSaveHook(hook func(*entities.User) error) {
	r.afterSave = hook
}

// SetBeforeDeleteHook sets a callback to execute before deleting users.
func (r *InMemoryTestRepository) SetBeforeDeleteHook(hook func(values.UserID) error) {
	r.beforeDelete = hook
}

// SetAfterDeleteHook sets a callback to execute after deleting users.
func (r *InMemoryTestRepository) SetAfterDeleteHook(hook func(values.UserID) error) {
	r.afterDelete = hook
}

// GetOperationLog returns the log of operations performed on the repository.
func (r *InMemoryTestRepository) GetOperationLog() []string {
	return r.operationLog
}

// ClearOperationLog clears the operation log.
func (r *InMemoryTestRepository) ClearOperationLog() {
	r.operationLog = make([]string, 0)
}

// LoadTestData loads predefined test data into the repository.
func (r *InMemoryTestRepository) LoadTestData(ctx context.Context, scenario string) error {
	users := r.testData.GetScenarioUsers(scenario)
	for _, user := range users {
		if err := r.Save(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

// RepositoryTestData manages test data scenarios for repositories.
type RepositoryTestData struct {
	scenarios map[string][]*entities.User
}

// NewRepositoryTestData creates a new test data manager.
func NewRepositoryTestData() *RepositoryTestData {
	data := &RepositoryTestData{
		scenarios: make(map[string][]*entities.User),
	}
	data.initializeDefaultScenarios()
	return data
}

// initializeDefaultScenarios sets up common test data scenarios.
func (d *RepositoryTestData) initializeDefaultScenarios() {
	// Empty scenario
	d.scenarios["empty"] = []*entities.User{}

	// Single user scenario
	d.scenarios["single"] = []*entities.User{
		entityHelpers.TestUserWithID("single-user"),
	}

	// Multiple users scenario
	d.scenarios["multiple"] = entityHelpers.CreateTestUserCollection(5)

	// Users with same domain scenario
	d.scenarios["same_domain"] = []*entities.User{
		entityHelpers.NewUserBuilder().WithID("user1").WithEmail("user1@company.com").WithName("User One").Build(),
		entityHelpers.NewUserBuilder().WithID("user2").WithEmail("user2@company.com").WithName("User Two").Build(),
		entityHelpers.NewUserBuilder().WithID("user3").WithEmail("user3@company.com").WithName("User Three").Build(),
	}

	// Users with different domains scenario
	d.scenarios["mixed_domains"] = []*entities.User{
		entityHelpers.NewUserBuilder().WithID("gmail-user").WithEmail("user@gmail.com").WithName("Gmail User").Build(),
		entityHelpers.NewUserBuilder().WithID("yahoo-user").WithEmail("user@yahoo.com").WithName("Yahoo User").Build(),
		entityHelpers.NewUserBuilder().WithID("company-user").WithEmail("user@company.com").WithName("Company User").Build(),
	}

	// Large dataset scenario
	d.scenarios["large"] = entityHelpers.CreateTestUserCollection(100)
}

// GetScenarioUsers returns users for a specific test scenario.
func (d *RepositoryTestData) GetScenarioUsers(scenario string) []*entities.User {
	users, exists := d.scenarios[scenario]
	if !exists {
		return []*entities.User{}
	}
	return users
}

// AddScenario adds a custom test scenario.
func (d *RepositoryTestData) AddScenario(name string, users []*entities.User) {
	d.scenarios[name] = users
}

// GetAvailableScenarios returns all available scenario names.
func (d *RepositoryTestData) GetAvailableScenarios() []string {
	scenarios := make([]string, 0, len(d.scenarios))
	for name := range d.scenarios {
		scenarios = append(scenarios, name)
	}
	return scenarios
}

// RepositoryTestHelper provides high-level repository testing utilities.
type RepositoryTestHelper struct {
	*base.GinkgoSuite
	repository repositories.UserRepository
	testData   *RepositoryTestData
}

// NewRepositoryTestHelper creates a new repository test helper.
func NewRepositoryTestHelper(repo repositories.UserRepository) *RepositoryTestHelper {
	return &RepositoryTestHelper{
		GinkgoSuite: base.NewGinkgoSuite(),
		repository:  repo,
		testData:    NewRepositoryTestData(),
	}
}

// SetupWithScenario initializes the repository with a specific test scenario.
func (h *RepositoryTestHelper) SetupWithScenario(scenario string) {
	ctx := h.GetContext()

	// Clear repository if it's a test repository
	if testRepo, ok := h.repository.(*InMemoryTestRepository); ok {
		testRepo.ClearOperationLog()
	}

	// Load test data
	users := h.testData.GetScenarioUsers(scenario)
	for _, user := range users {
		err := h.repository.Save(ctx, user)
		if err != nil {
			panic(err) // Fail fast during test setup
		}
	}
}

// AssertScenarioState verifies the repository contains expected scenario data.
func (h *RepositoryTestHelper) AssertScenarioState(scenario string) {
	ctx := h.GetContext()
	expectedUsers := h.testData.GetScenarioUsers(scenario)

	allUsers, err := h.repository.List(ctx)
	base.AssertSuccess(allUsers, err)

	// Check user count matches
	if len(allUsers) != len(expectedUsers) {
		panic(fmt.Sprintf("Expected %d users, got %d", len(expectedUsers), len(allUsers)))
	}

	// Verify all expected users exist
	for _, expectedUser := range expectedUsers {
		found, err := h.repository.FindByID(ctx, expectedUser.ID)
		base.AssertSuccess(found, err)
	}
}

// Repository test convenience functions

// SetupInMemoryRepository creates a fresh in-memory repository for testing.
func SetupInMemoryRepository() repositories.UserRepository {
	return repositories.NewInMemoryUserRepository()
}

// SetupEnhancedInMemoryRepository creates an enhanced in-memory repository with test features.
func SetupEnhancedInMemoryRepository() *InMemoryTestRepository {
	return NewInMemoryTestRepository()
}

// SetupMockRepository creates a mock repository for service testing.
func SetupMockRepository() *MockUserRepository {
	return NewMockUserRepository()
}

// SetupRepositoryWithData creates a repository pre-loaded with test data.
func SetupRepositoryWithData(scenario string) repositories.UserRepository {
	repo := SetupEnhancedInMemoryRepository()
	ctx := context.Background()

	if err := repo.LoadTestData(ctx, scenario); err != nil {
		panic(err)
	}

	return repo
}

// PreloadRepositoryWithUsers adds specific users to a repository.
func PreloadRepositoryWithUsers(repo repositories.UserRepository, users ...*entities.User) error {
	ctx := context.Background()
	for _, user := range users {
		if err := repo.Save(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

// ClearRepository removes all users from an in-memory repository.
func ClearRepository(repo repositories.UserRepository) error {
	ctx := context.Background()
	users, err := repo.List(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		if err := repo.Delete(ctx, user.ID); err != nil {
			return err
		}
	}

	return nil
}
