// Package services provides domain service implementations for business logic.
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	domainerrors "github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/shared"
)

// UserFilters represents the available filters for user queries
type UserFilters struct {
	Domain *string
	Active *bool
}

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
		return nil, domainerrors.NewValidationError("email", err.Error())
	}

	// Business rule: Validate username
	if err := s.validateUserName(name); err != nil {
		return nil, domainerrors.NewValidationError("name", err.Error())
	}

	// Business rule: Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, domainerrors.NewInternalError("failed to check existing user", err)
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
		return nil, domainerrors.NewInternalError("failed to save user", err)
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
	user, err := s.getUserForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.validateUserUpdates(ctx, user, email, name); err != nil {
		return nil, err
	}

	return s.applyUserUpdates(ctx, user, email, name)
}

func (s *UserService) getUserForUpdate(ctx context.Context, id entities.UserID) (*entities.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}
	return user, nil
}

func (s *UserService) validateUserUpdates(ctx context.Context, user *entities.User, email, name string) error {
	if err := s.validateEmailUpdate(ctx, user, email); err != nil {
		return err
	}

	return s.validateNameUpdate(user, name)
}

func (s *UserService) validateEmailUpdate(ctx context.Context, user *entities.User, email string) error {
	if email == user.Email {
		return nil
	}

	if err := s.validateEmail(email); err != nil {
		return fmt.Errorf("invalid email: %w", err)
	}

	return s.checkEmailAvailability(ctx, email)
}

func (s *UserService) checkEmailAvailability(ctx context.Context, email string) error {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return repositories.ErrUserAlreadyExists
	}
	return nil
}

func (s *UserService) validateNameUpdate(user *entities.User, name string) error {
	if name != user.Name {
		if err := s.validateUserName(name); err != nil {
			return fmt.Errorf("invalid username: %w", err)
		}
	}
	return nil
}

func (s *UserService) applyUserUpdates(ctx context.Context, user *entities.User, email, name string) (*entities.User, error) {
	if err := user.SetEmail(email); err != nil {
		return nil, fmt.Errorf("failed to set email: %w", err)
	}

	if err := user.SetName(name); err != nil {
		return nil, fmt.Errorf("failed to set name: %w", err)
	}

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
		return nil, domainerrors.NewInternalError("failed to list users", err)
	}

	// Business logic: Could add filtering, sorting, pagination, etc.
	return users, nil
}

// FilterActiveUsers demonstrates functional programming with lo library
func (s *UserService) FilterActiveUsers(ctx context.Context) ([]*entities.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.NewInternalError("failed to list users", err)
	}

	// Functional operations using samber/lo
	activeUsers := lo.Filter(users, func(user *entities.User, _ int) bool {
		// Business rule: Users created in the last 30 days are considered active
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		return user.Created.After(thirtyDaysAgo)
	})

	return activeUsers, nil
}

// GetUserEmailsWithResult demonstrates Result pattern with Railway Oriented Programming
func (s *UserService) GetUserEmailsWithResult(ctx context.Context) shared.Result[[]string] {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return shared.Err[[]string](domainerrors.NewInternalError("failed to list users", err))
	}

	// Functional operation: extract emails
	emails := lo.Map(users, func(user *entities.User, _ int) string {
		return user.Email
	})

	return shared.Ok(emails)
}

// CreateUserWithResult demonstrates Railway Oriented Programming
func (s *UserService) CreateUserWithResult(ctx context.Context, id entities.UserID, email, name string) shared.Result[*entities.User] {
	// Step 1: Validate inputs
	if validationResult := s.validateUserInputsResult(email, name); validationResult.IsError() {
		return shared.Err[*entities.User](validationResult.Error())
	}

	// Step 2: Check user doesn't exist
	if existsResult := s.checkUserNotExistsResult(ctx, email); existsResult.IsError() {
		return shared.Err[*entities.User](existsResult.Error())
	}

	// Step 3: Create and save user
	return s.createAndSaveUserResult(ctx, id, email, name)
}

// validateUserInputsResult validates user inputs using Result pattern
func (s *UserService) validateUserInputsResult(email, name string) shared.Result[struct{}] {
	if err := s.validateEmail(email); err != nil {
		return shared.Err[struct{}](domainerrors.NewValidationError("email", err.Error()))
	}
	if err := s.validateUserName(name); err != nil {
		return shared.Err[struct{}](domainerrors.NewValidationError("name", err.Error()))
	}
	return shared.Ok(struct{}{})
}

// checkUserNotExistsResult checks if user exists using Result pattern
func (s *UserService) checkUserNotExistsResult(ctx context.Context, email string) shared.Result[*entities.User] {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return shared.Err[*entities.User](domainerrors.NewInternalError("failed to check existing user", err))
	}
	if existingUser != nil {
		return shared.Err[*entities.User](repositories.ErrUserAlreadyExists)
	}
	return shared.Ok[*entities.User](nil)
}

// createAndSaveUserResult creates and saves user using Result pattern
func (s *UserService) createAndSaveUserResult(ctx context.Context, id entities.UserID, email, name string) shared.Result[*entities.User] {
	user, err := entities.NewUser(id, email, name)
	if err != nil {
		return shared.Err[*entities.User](err)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return shared.Err[*entities.User](domainerrors.NewInternalError("failed to save user", err))
	}

	return shared.Ok(user)
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

// GetUserStats demonstrates functional aggregation with lo.Reduce and lo.Ternary
func (s *UserService) GetUserStats(ctx context.Context) (map[string]int, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.NewInternalError("failed to list users", err)
	}

	stats := make(map[string]int)

	// Count total users
	stats["total"] = len(users)

	// Count active users (created in last 30 days) using functional operations
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	activeCount := lo.CountBy(users, func(user *entities.User) bool {
		return user.Created.After(thirtyDaysAgo)
	})
	stats["active"] = activeCount

	// Extract email domains using functional operations with lo.Ternary
	domains := lo.Map(users, func(user *entities.User, _ int) string {
		parts := strings.Split(user.Email, "@")
		return lo.Ternary(len(parts) > 1, parts[1], "unknown")
	})

	// Count unique domains
	domainCounts := lo.CountValues(domains)
	stats["domains"] = len(domainCounts)

	// Calculate average days since registration using lo.Reduce
	now := time.Now()
	totalDays := lo.Reduce(users, func(acc int, user *entities.User, _ int) int {
		days := int(now.Sub(user.Created).Hours() / 24)
		// Ensure non-negative days
		if days < 0 {
			days = 0
		}
		return acc + days
	}, 0)

	// Safe division using lo.Max to prevent divide by zero
	userCount := lo.Max([]int{len(users), 1}) // Ensure at least 1 to prevent division by zero
	avgDays := lo.Ternary(len(users) > 0, totalDays/userCount, 0)
	stats["avg_days_since_registration"] = avgDays

	return stats, nil
}

// GetUsersWithFilters demonstrates advanced functional programming with type-safe filters
func (s *UserService) GetUsersWithFilters(ctx context.Context, filters UserFilters) ([]*entities.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.NewInternalError("failed to list users", err)
	}

	// Start with all users
	filteredUsers := users

	// Filter by domain if specified
	if filters.Domain != nil {
		domainStr := *filters.Domain
		filteredUsers = lo.Filter(filteredUsers, func(user *entities.User, _ int) bool {
			parts := strings.Split(user.Email, "@")
			return len(parts) > 1 && parts[1] == domainStr
		})
	}

	// Filter by active status if specified
	if filters.Active != nil && *filters.Active {
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
		filteredUsers = lo.Filter(filteredUsers, func(user *entities.User, _ int) bool {
			return user.Created.After(thirtyDaysAgo)
		})
	}

	return filteredUsers, nil
}

// ValidateUserBatchWithEither demonstrates Either pattern for batch operations
func (s *UserService) ValidateUserBatchWithEither(users []*entities.User) shared.Either[[]error, []entities.UserID] {
	validUsers := make([]entities.UserID, 0)
	validationErrors := make([]error, 0)

	lo.ForEach(users, func(user *entities.User, _ int) {
		if err := user.Validate(); err != nil {
			validationErrors = append(validationErrors, err)
		} else {
			validUsers = append(validUsers, user.ID)
		}
	})

	// Return either errors (if any) or valid user IDs
	if len(validationErrors) > 0 {
		return shared.Left[[]error, []entities.UserID](validationErrors)
	}
	return shared.Right[[]error, []entities.UserID](validUsers)
}

// GetUsersByEmailDomains demonstrates more complex lo operations
func (s *UserService) GetUsersByEmailDomains(ctx context.Context, domains []string) (map[string][]*entities.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.NewInternalError("failed to list users", err)
	}

	// Group users by email domain using lo.GroupBy
	usersByDomain := lo.GroupBy(users, func(user *entities.User) string {
		parts := strings.Split(user.Email, "@")
		return lo.Ternary(len(parts) > 1, parts[1], "unknown")
	})

	// Filter only requested domains using lo.PickByKeys
	requestedDomainsSet := lo.SliceToMap(domains, func(domain string) (string, bool) {
		return domain, true
	})

	filteredUsers := lo.PickBy(usersByDomain, func(domain string, _ []*entities.User) bool {
		_, exists := requestedDomainsSet[domain]
		return exists
	})

	return filteredUsers, nil
}

// validateEmail enforces business rules for email validation
func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return domainerrors.NewRequiredFieldError("email")
	}

	// Basic email validation
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return domainerrors.NewValidationError("email", "must be a valid email address")
	}

	if len(email) > 254 {
		return domainerrors.NewValidationError("email", "too long (max 254 characters)")
	}

	// Business rule: No spaces in email
	if strings.Contains(email, " ") {
		return domainerrors.NewValidationError("email", "cannot contain spaces")
	}

	return nil
}

// validateUserName enforces business rules for display name validation
func (s *UserService) validateUserName(name string) error {
	if err := s.validateNameNotEmpty(name); err != nil {
		return err
	}

	if err := s.validateNameLength(name); err != nil {
		return err
	}

	if err := s.validateNameWhitespace(name); err != nil {
		return err
	}

	return s.validateNameContainsLetter(name)
}

func (s *UserService) validateNameNotEmpty(name string) error {
	if name == "" {
		return domainerrors.NewRequiredFieldError("name")
	}
	return nil
}

func (s *UserService) validateNameLength(name string) error {
	if len(name) < 2 {
		return domainerrors.NewValidationError("name", "too short (min 2 characters)")
	}
	if len(name) > 100 {
		return domainerrors.NewValidationError("name", "too long (max 100 characters)")
	}
	return nil
}

func (s *UserService) validateNameWhitespace(name string) error {
	if strings.TrimSpace(name) != name {
		return domainerrors.NewValidationError("name", "cannot have leading or trailing spaces")
	}
	return nil
}

func (s *UserService) validateNameContainsLetter(name string) error {
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			return nil
		}
	}
	return domainerrors.NewValidationError("name", "must contain at least one letter")
}
