// Integration test to demonstrate DDD patterns working together
package domain_test

import (
	"context"
	"errors"
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
	
	// Setup infrastructure (in-memory repository)
	userRepo := memrepos.NewInMemoryUserRepository()
	
	// Setup domain service
	userService := services.NewUserService(userRepo)
	
	// Test 1: Create user with value objects
	userID, err := values.GenerateUserID()
	if err != nil {
		t.Fatalf("Failed to generate user ID: %v", err)
	}
	
	email := "john.doe@example.com"
	name := "john.doe"
	
	createdUser, err := userService.CreateUser(ctx, userID, email, name)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	
	// Verify user was created with proper value objects
	if createdUser.GetEmail().Value() != email {
		t.Errorf("Expected email %s, got %s", email, createdUser.GetEmail().Value())
	}
	
	if createdUser.GetUserName().Value() != name {
		t.Errorf("Expected name %s, got %s", name, createdUser.GetUserName().Value())
	}
	
	if createdUser.EmailDomain() != "example.com" {
		t.Errorf("Expected domain example.com, got %s", createdUser.EmailDomain())
	}
	
	// Test 2: Retrieve user by ID
	retrievedUser, err := userService.GetUser(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to retrieve user: %v", err)
	}
	
	if !retrievedUser.ID.Equals(userID) {
		t.Errorf("User IDs don't match")
	}
	
	// Test 3: Retrieve user by email
	userByEmail, err := userService.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to retrieve user by email: %v", err)
	}
	
	if !userByEmail.ID.Equals(userID) {
		t.Errorf("User IDs don't match when retrieved by email")
	}
	
	// Test 4: Update user with value object validation
	newEmail := "jane.doe@newdomain.org"
	newName := "jane.doe"
	
	updatedUser, err := userService.UpdateUser(ctx, userID, newEmail, newName)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	
	if updatedUser.GetEmail().Value() != newEmail {
		t.Errorf("Expected updated email %s, got %s", newEmail, updatedUser.GetEmail().Value())
	}
	
	if updatedUser.EmailDomain() != "newdomain.org" {
		t.Errorf("Expected updated domain newdomain.org, got %s", updatedUser.EmailDomain())
	}
	
	// Test 5: Business rule validation - invalid email
	_, err = userService.CreateUser(ctx, userID, "invalid-email", "testuser")
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}
	
	// Test 6: Business rule validation - reserved username
	anotherUserID, _ := values.GenerateUserID()
	_, err = userService.CreateUser(ctx, anotherUserID, "admin@test.com", "admin")
	if err == nil {
		t.Error("Expected validation error for reserved username")
	}
	
	// Test 7: Business rule - duplicate email
	duplicateUserID, _ := values.GenerateUserID()
	_, err = userService.CreateUser(ctx, duplicateUserID, newEmail, "differentuser")
	if err == nil {
		t.Error("Expected conflict error for duplicate email")
	}
	
	// Test 8: List all users
	users, err := userService.ListUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
	
	// Test 9: Delete user
	err = userService.DeleteUser(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	
	// Test 10: Verify user was deleted
	_, err = userService.GetUser(ctx, userID)
	if err == nil {
		t.Error("Expected error when retrieving deleted user")
	}
	
	// Check if the error contains the expected message (wrapped error)
	if !containsError(err, repositories.ErrUserNotFound) {
		t.Errorf("Expected error to contain ErrUserNotFound, got %v", err)
	}
}

// TestValueObjectsIntegration tests value objects in isolation
func TestValueObjectsIntegration(t *testing.T) {
	// Test Email value object
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
	
	// Test invalid email
	_, err = values.NewEmail("invalid-email")
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}
	
	// Test UserName value object
	username, err := values.NewUserName("validuser123")
	if err != nil {
		t.Fatalf("Failed to create username: %v", err)
	}
	
	if username.Value() != "validuser123" {
		t.Errorf("Expected validuser123, got %s", username.Value())
	}
	
	// Test reserved username
	_, err = values.NewUserName("admin")
	if err == nil {
		t.Error("Expected validation error for reserved username")
	}
	
	// Test UserID value object
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
	
	// Test custom user ID
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
	
	// Create user with value objects
	userID, _ := values.GenerateUserID()
	user, err := entities.NewUser(userID, "test@example.com", "testuser")
	if err != nil {
		t.Fatalf("Failed to create user entity: %v", err)
	}
	
	// Test Save
	err = repo.Save(ctx, user)
	if err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}
	
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
	
	// Test List
	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}
	
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}
	
	// Test Delete
	err = repo.Delete(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	
	// Verify deletion
	_, err = repo.FindByID(ctx, userID)
	if err != repositories.ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound after deletion, got %v", err)
	}
}