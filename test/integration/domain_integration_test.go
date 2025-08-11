// Integration test to demonstrate DDD patterns working together
package integration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/services"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	memrepos "github.com/LarsArtmann/template-arch-lint/internal/infrastructure/repositories"
)

// containsError checks if err wraps target error
func containsError(err, target error) bool {
	return errors.Is(err, target)
}

// TestDDDIntegration demonstrates the complete DDD architecture working together
func TestDDDIntegration(t *testing.T) {
	ctx := context.Background()
	userService := setupDDDTestServices()

	userID, email, name := setupInitialTestData()
	_ = runUserCreationTest(t, ctx, userService, userID, email, name)
	
	runUserRetrievalTests(t, ctx, userService, userID, email)
	runUserUpdateTest(t, ctx, userService, userID)
	runValidationTests(t, ctx, userService, userID)
	runUserListAndDeletionTests(t, ctx, userService, userID)
}

// setupDDDTestServices sets up the service layer for testing
func setupDDDTestServices() *services.UserService {
	userRepo := memrepos.NewInMemoryUserRepository()
	return services.NewUserService(userRepo)
}

// setupInitialTestData prepares test data constants
func setupInitialTestData() (values.UserID, string, string) {
	userID, err := values.GenerateUserID()
	if err != nil {
		panic(fmt.Sprintf("Failed to generate user ID: %v", err))
	}
	email := "john.doe@example.com"
	name := "john.doe"
	return userID, email, name
}

// runUserCreationTest creates a user and validates the result
func runUserCreationTest(t *testing.T, ctx context.Context, userService *services.UserService, userID values.UserID, email, name string) *entities.User {
	t.Helper()
	
	createdUser, err := userService.CreateUser(ctx, userID, email, name)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	validateCreatedUser(t, createdUser, email, name)
	return createdUser
}

// validateCreatedUser verifies the created user has correct values
func validateCreatedUser(t *testing.T, user *entities.User, expectedEmail, expectedName string) {
	t.Helper()
	
	if user.GetEmail().Value() != expectedEmail {
		t.Errorf("Expected email %s, got %s", expectedEmail, user.GetEmail().Value())
	}
	if user.GetUserName().Value() != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, user.GetUserName().Value())
	}
	if user.EmailDomain() != "example.com" {
		t.Errorf("Expected domain example.com, got %s", user.EmailDomain())
	}
}

// runUserRetrievalTests tests user retrieval by ID and email
func runUserRetrievalTests(t *testing.T, ctx context.Context, userService *services.UserService, userID values.UserID, email string) {
	t.Helper()
	
	// Test retrieve by ID
	retrievedUser, err := userService.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	if !retrievedUser.ID.Equals(userID) {
		t.Errorf("User IDs don't match")
	}

	// Test retrieve by email
	userByEmail, err := userService.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to retrieve user by email: %v", err)
	}
	if !userByEmail.ID.Equals(userID) {
		t.Errorf("User IDs don't match when retrieved by email")
	}
}

// runUserUpdateTest tests user update functionality
func runUserUpdateTest(t *testing.T, ctx context.Context, userService *services.UserService, userID values.UserID) {
	t.Helper()
	
	newEmail := "jane.doe@newdomain.org"
	newName := "jane.doe"

	updatedUser, err := userService.UpdateUser(ctx, userID, newEmail, newName)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	validateUpdatedUser(t, updatedUser, newEmail)
}

// validateUpdatedUser verifies the updated user has correct values
func validateUpdatedUser(t *testing.T, user *entities.User, expectedEmail string) {
	t.Helper()
	
	if user.GetEmail().Value() != expectedEmail {
		t.Errorf("Expected updated email %s, got %s", expectedEmail, user.GetEmail().Value())
	}
	if user.EmailDomain() != "newdomain.org" {
		t.Errorf("Expected updated domain newdomain.org, got %s", user.EmailDomain())
	}
}

// runValidationTests tests business rule validation
func runValidationTests(t *testing.T, ctx context.Context, userService *services.UserService, userID values.UserID) {
	t.Helper()
	
	// Test invalid email validation
	_, err := userService.CreateUser(ctx, userID, "invalid-email", "testuser")
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}

	// Test reserved username validation
	anotherUserID, err := values.GenerateUserID()
	if err != nil {
		t.Fatalf("Failed to generate user ID: %v", err)
	}
	_, err = userService.CreateUser(ctx, anotherUserID, "admin@test.com", "admin")
	if err == nil {
		t.Error("Expected validation error for reserved username")
	}

	// Test duplicate email validation
	duplicateUserID, err := values.GenerateUserID()
	if err != nil {
		t.Fatalf("Failed to generate user ID: %v", err)
	}
	_, err = userService.CreateUser(ctx, duplicateUserID, "jane.doe@newdomain.org", "differentuser")
	if err == nil {
		t.Error("Expected conflict error for duplicate email")
	}
}

// runUserListAndDeletionTests tests listing users and deletion
func runUserListAndDeletionTests(t *testing.T, ctx context.Context, userService *services.UserService, userID values.UserID) {
	t.Helper()
	
	// Test list users
	users, err := userService.ListUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	// Test delete user
	err = userService.DeleteUser(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify user was deleted
	_, err = userService.GetUser(ctx, userID)
	if err == nil {
		t.Error("Expected error when retrieving deleted user")
	}
	if !containsError(err, repositories.ErrUserNotFound) {
		t.Errorf("Expected error to contain ErrUserNotFound, got %v", err)
	}
}

// TestValueObjectsIntegration tests value objects in isolation
func TestValueObjectsIntegration(t *testing.T) {
	testEmailValueObject(t)
	testUserNameValueObject(t)
	testUserIDValueObject(t)
}

// testEmailValueObject tests the Email value object behavior
func testEmailValueObject(t *testing.T) {
	t.Helper()
	
	// Test valid email creation
	email, err := values.NewEmail("test@example.com")
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	if email.Domain() != "example.com" {
		t.Errorf("Expected domain example.com, got %s", email.Domain())
	}

	if email.LocalPart() != "test" {
		t.Errorf("Expected local part test, got %s", email.LocalPart())
	}

	// Test invalid email validation
	_, err = values.NewEmail("invalid-email")
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}
}

// testUserNameValueObject tests the UserName value object behavior
func testUserNameValueObject(t *testing.T) {
	t.Helper()
	
	// Test valid username creation
	username, err := values.NewUserName("validuser123")
	if err != nil {
		t.Fatalf("Failed to create username: %v", err)
	}

	if username.Value() != "validuser123" {
		t.Errorf("Expected validuser123, got %s", username.Value())
	}

	// Test reserved username validation
	_, err = values.NewUserName("admin")
	if err == nil {
		t.Error("Expected validation error for reserved username")
	}
}

// testUserIDValueObject tests the UserID value object behavior
func testUserIDValueObject(t *testing.T) {
	t.Helper()
	
	testGeneratedUserID(t)
	testCustomUserID(t)
}

// testGeneratedUserID tests generated user ID behavior
func testGeneratedUserID(t *testing.T) {
	t.Helper()
	
	userID, err := values.GenerateUserID()
	if err != nil {
		t.Fatalf("Failed to generate user ID: %v", err)
	}

	if userID.IsEmpty() {
		t.Error("Generated user ID should not be empty")
	}

	if !userID.IsGenerated() {
		t.Error("Generated user ID should be marked as generated")
	}
}

// testCustomUserID tests custom user ID behavior
func testCustomUserID(t *testing.T) {
	t.Helper()
	
	customID, err := values.NewUserID("custom_user_123")
	if err != nil {
		t.Fatalf("Failed to create custom user ID: %v", err)
	}

	if customID.IsGenerated() {
		t.Error("Custom user ID should not be marked as generated")
	}
}

// TestRepositoryIntegration tests repository patterns in isolation
func TestRepositoryIntegration(t *testing.T) {
	ctx := context.Background()
	repo := memrepos.NewInMemoryUserRepository()
	userID, user := setupTestUserEntity(t)

	testRepositorySaveOperation(t, ctx, repo, user)
	testRepositoryFindOperations(t, ctx, repo, userID)
	testRepositoryListOperation(t, ctx, repo)
	testRepositoryDeleteOperation(t, ctx, repo, userID)
}

// setupTestUserEntity creates a test user entity for repository testing
func setupTestUserEntity(t *testing.T) (values.UserID, *entities.User) {
	t.Helper()
	
	userID, err := values.GenerateUserID()
	if err != nil {
		t.Fatalf("Failed to generate user ID: %v", err)
	}
	user, err := entities.NewUser(userID, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to create user entity: %v", err)
	}
	return userID, user
}

// testRepositorySaveOperation tests the Save repository operation
func testRepositorySaveOperation(t *testing.T, ctx context.Context, repo repositories.UserRepository, user *entities.User) {
	t.Helper()
	
	err := repo.Save(ctx, user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}
}

// testRepositoryFindOperations tests FindByID and FindByEmail operations
func testRepositoryFindOperations(t *testing.T, ctx context.Context, repo repositories.UserRepository, userID values.UserID) {
	t.Helper()
	
	// Test FindByID
	foundUser, err := repo.FindByID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to find user by ID: %v", err)
	}

	if !foundUser.ID.Equals(userID) {
		t.Error("Found user ID doesn't match")
	}

	// Test FindByEmail
	foundByEmail, err := repo.FindByEmail(ctx, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to find user by email: %v", err)
	}

	if !foundByEmail.ID.Equals(userID) {
		t.Error("Found user ID doesn't match when searched by email")
	}
}

// testRepositoryListOperation tests the List repository operation
func testRepositoryListOperation(t *testing.T, ctx context.Context, repo repositories.UserRepository) {
	t.Helper()
	
	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
}

// testRepositoryDeleteOperation tests the Delete repository operation
func testRepositoryDeleteOperation(t *testing.T, ctx context.Context, repo repositories.UserRepository, userID values.UserID) {
	t.Helper()
	
	err := repo.Delete(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify deletion
	_, err = repo.FindByID(ctx, userID)
	if !errors.Is(err, repositories.ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound after deletion, got %v", err)
	}
}
