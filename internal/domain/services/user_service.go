// User service layer containing business logic and rules
package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/shared"
)

// UserService handles business logic for user operations
type UserService struct {
	userRepo repositories.UserRepository
}

// NewUserService creates a new user service with dependency injection
func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with business validation
func (s *UserService) CreateUser(ctx context.Context, id entities.UserID, email, name string) (*entities.User, error) {
	// Business rule: Validate email format
	if err := s.validateEmail(email); err != nil {
		return nil, errors.NewValidationError("email", err.Error())
	}

	// Business rule: Validate username
	if err := s.validateUserName(name); err != nil {
		return nil, errors.NewValidationError("name", err.Error())
	}

	// Business rule: Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && err != repositories.ErrUserNotFound {
		return nil, errors.NewInternalError("failed to check existing user", err)
	}
	if existingUser != nil {
		return nil, repositories.ErrUserAlreadyExists
	}

	// Create new user entity
	user, err := entities.NewUser(id, email, name)
	if err != nil {
		return nil, err // Already typed error from entity
	}

	// Save to repository
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, errors.NewInternalError("failed to save user", err)
	}

	return user, nil
}

// GetUser retrieves a user by ID with business logic
func (s *UserService) GetUser(ctx context.Context, id entities.UserID) (*entities.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Business logic: Could add user activity tracking, audit logging, etc.
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	if err := s.validateEmail(email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// UpdateUser updates user information with business rules
func (s *UserService) UpdateUser(ctx context.Context, id entities.UserID, email, name string) (*entities.User, error) {
	// Get existing user
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}

	// Business rule: Validate new email if changed
	if email != user.Email {
		if err := s.validateEmail(email); err != nil {
			return nil, fmt.Errorf("invalid email: %w", err)
		}

		// Check if new email is already taken
		existingUser, err := s.userRepo.FindByEmail(ctx, email)
		if err != nil && err != repositories.ErrUserNotFound {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		if existingUser != nil {
			return nil, repositories.ErrUserAlreadyExists
		}
	}

	// Business rule: Validate new name if changed
	if name != user.Name {
		if err := s.validateUserName(name); err != nil {
			return nil, fmt.Errorf("invalid username: %w", err)
		}
	}

	// Update user fields using value objects
	if err := user.SetEmail(email); err != nil {
		return nil, fmt.Errorf("failed to set email: %w", err)
	}

	if err := user.SetName(name); err != nil {
		return nil, fmt.Errorf("failed to set name: %w", err)
	}

	// Save updated user
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save updated user: %w", err)
	}

	return user, nil
}

// DeleteUser removes a user with business rules
func (s *UserService) DeleteUser(ctx context.Context, id entities.UserID) error {
	// Business rule: Check if user exists before deletion
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find user for deletion: %w", err)
	}

	// Business logic: Could add soft delete, cascade operations, etc.
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListUsers retrieves all users with business logic
func (s *UserService) ListUsers(ctx context.Context) ([]*entities.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, errors.NewInternalError("failed to list users", err)
	}

	// Business logic: Could add filtering, sorting, pagination, etc.
	return users, nil
}

// FilterActiveUsers demonstrates functional programming with lo library
func (s *UserService) FilterActiveUsers(ctx context.Context) ([]*entities.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, errors.NewInternalError("failed to list users", err)
	}

	// Functional operations using samber/lo
	activeUsers := lo.Filter(users, func(user *entities.User, _ int) bool {
		// Business rule: Users created in the last 30 days are considered active
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		return user.Created.After(thirtyDaysAgo)
	})

	return activeUsers, nil
}

// GetUserEmailsWithResult demonstrates Result pattern
func (s *UserService) GetUserEmailsWithResult(ctx context.Context) shared.Result[[]string] {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return shared.NewError[[]string](errors.NewInternalError("failed to list users", err))
	}

	// Functional operation: extract emails
	emails := lo.Map(users, func(user *entities.User, _ int) string {
		return user.Email
	})

	return shared.NewResult(emails)
}

// FindUserByEmailOption demonstrates Option pattern
func (s *UserService) FindUserByEmailOption(ctx context.Context, email string) shared.Option[*entities.User] {
	if err := s.validateEmail(email); err != nil {
		return shared.None[*entities.User]()
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return shared.None[*entities.User]()
	}

	return shared.Some(user)
}

// BatchValidateUsers demonstrates functional operations for batch processing
func (s *UserService) BatchValidateUsers(users []*entities.User) map[entities.UserID]error {
	// Use lo to create a map of validation results
	validationResults := lo.SliceToMap(users, func(user *entities.User) (entities.UserID, error) {
		return user.ID, user.Validate()
	})

	// Filter only failed validations
	failedValidations := lo.PickBy(validationResults, func(_ entities.UserID, err error) bool {
		return err != nil
	})

	return failedValidations
}

// GetUserStats demonstrates functional aggregation
func (s *UserService) GetUserStats(ctx context.Context) (map[string]int, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, errors.NewInternalError("failed to list users", err)
	}

	stats := make(map[string]int)

	// Count total users
	stats["total"] = len(users)

	// Count active users (created in last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	activeCount := lo.CountBy(users, func(user *entities.User) bool {
		return user.Created.After(thirtyDaysAgo)
	})
	stats["active"] = activeCount

	// Count users by email domain using functional operations
	domains := lo.Map(users, func(user *entities.User, _ int) string {
		parts := strings.Split(user.Email, "@")
		if len(parts) > 1 {
			return parts[1]
		}
		return "unknown"
	})

	domainCounts := lo.CountValues(domains)
	stats["domains"] = len(domainCounts)

	return stats, nil
}

// validateEmail enforces business rules for email validation
func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return errors.NewRequiredFieldError("email")
	}

	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return errors.NewValidationError("email", "must be a valid email address")
	}

	if len(email) > 254 {
		return errors.NewValidationError("email", "too long (max 254 characters)")
	}

	// Business rule: No spaces in email
	if strings.Contains(email, " ") {
		return errors.NewValidationError("email", "cannot contain spaces")
	}

	return nil
}

// validateUserName enforces business rules for display name validation
func (s *UserService) validateUserName(name string) error {
	if name == "" {
		return errors.NewRequiredFieldError("name")
	}

	// Business rule: Name length constraints
	if len(name) < 2 {
		return errors.NewValidationError("name", "too short (min 2 characters)")
	}

	if len(name) > 100 {
		return errors.NewValidationError("name", "too long (max 100 characters)")
	}

	// Business rule: No leading/trailing whitespace
	if strings.TrimSpace(name) != name {
		return errors.NewValidationError("name", "cannot have leading or trailing spaces")
	}

	// Business rule: Must contain at least one letter
	hasLetter := false
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
			break
		}
	}

	if !hasLetter {
		return errors.NewValidationError("name", "must contain at least one letter")
	}

	return nil
}
