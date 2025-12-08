// Package services provides domain service implementations for business logic.
//
// TODO: CRITICAL ARCHITECTURE VIOLATION - This file is 526 lines, violates SRP, needs breaking into smaller services
// TODO: EXTRACT SERVICES - Break into: UserQueryService, UserCommandService, UserValidationService, UserFilterService
// TODO: TYPE SAFETY EMERGENCY - Replace ALL string parameters with value objects (Email, UserName)
// TODO: SPLIT BRAIN RISK - Inconsistent error handling patterns (some use Result[T], others don't)
// TODO: VALIDATION CONSISTENCY - Extract validation logic to dedicated validator following DDD patterns
// TODO: TRANSACTION SAFETY - Add proper transaction boundaries for data consistency
// TODO: PERFORMANCE - Add caching layer, pagination, query optimization
// TODO: DOMAIN MODELING - Create proper domain events for user lifecycle changes
// TODO: FUNCTIONAL PROGRAMMING - Standardize on Result[T] pattern for all operations
// TODO: PRIMITIVE OBSESSION - Remove all string primitives, use value objects everywhere
// TODO: CONCURRENCY SAFETY - Add optimistic locking to prevent concurrent update issues
// TODO: BUSINESS RULES EXTRACTION - Extract business rules to specification pattern
// TODO: OBSERVABILITY - Add comprehensive logging, metrics, and tracing
package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	domainerrors "github.com/LarsArtmann/template-arch-lint/pkg/errors"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

// TODO: TYPE SAFETY - Replace *string with proper value objects (DomainName value object)
// TODO: SPECIFICATION PATTERN - Replace with UserSpecification interface for complex filtering
// TODO: BUILDER PATTERN - Add UserFiltersBuilder for better API design
// UserFilters represents the available filters for user queries.
type UserFilters struct {
	Domain *string // TODO: PRIMITIVE OBSESSION - Should be values.DomainName
	Active *bool   // TODO: DOMAIN MODELING - Could be values.UserStatus enum
}

// TODO: DEPENDENCY INJECTION - Add interfaces for all dependencies (logger, cache, event publisher)
// TODO: SINGLE RESPONSIBILITY - This should be split into multiple focused services
// TODO: CONCURRENCY SAFETY - Add sync.RWMutex for thread-safe operations if needed
// TODO: CACHING - Add cache layer dependency injection
// TODO: OBSERVABILITY - Add logger, metrics, and tracer dependencies
// UserService handles business logic for user operations.
type UserService struct {
	userRepo repositories.UserRepository
	// TODO: MISSING DEPENDENCIES - Should inject: logger, cache, eventPublisher, validator
}

// TODO: INCOMPLETE DEPENDENCY INJECTION - Should accept logger, cache, eventPublisher, validator
// TODO: VALIDATION - Add parameter validation to ensure userRepo is not nil
// TODO: BUILDER PATTERN - Consider using builder pattern for complex service construction
// NewUserService creates a new user service with dependency injection.
func NewUserService(userRepo repositories.UserRepository) *UserService {
	// TODO: NIL SAFETY - Add validation: if userRepo == nil { panic("userRepo cannot be nil") }
	return &UserService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user with business validation.
// TODO: ARCHITECTURAL IMPROVEMENT - Consider splitting this large service (511 lines) into smaller, focused services
// TODO: TYPE SAFETY - Migrate from string parameters to value objects (email values.Email, name values.UserName)
// TODO: BUSINESS RULES - Extract validation logic to dedicated validator service.
func (s *UserService) CreateUser(ctx context.Context, id values.UserID, email, name string) (*entities.User, error) {
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

// GetUser retrieves a user by ID with business logic.
// TODO: ERROR HANDLING - Consider using Result[T] pattern for better functional error handling.
func (s *UserService) GetUser(ctx context.Context, id values.UserID) (*entities.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domainerrors.WrapRepoError("get", "user", err)
	}

	// Business logic: Could add user activity tracking, audit logging, etc.
	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	if err := s.validateEmail(email); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domainerrors.WrapRepoError("get by email", "user", err)
	}

	return user, nil
}

// UpdateUser updates user information with business rules.
// TODO: TRANSACTION SAFETY - Consider implementing optimistic locking to prevent concurrent updates
// TODO: VALUE OBJECTS - Replace string parameters with proper value objects for type safety.
func (s *UserService) UpdateUser(ctx context.Context, id values.UserID, email, name string) (*entities.User, error) {
	user, err := s.getUserForUpdate(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.validateUserUpdates(ctx, user, email, name); err != nil {
		return nil, err
	}

	return s.applyUserUpdates(ctx, user, email, name)
}

// TODO: PERFORMANCE - Consider caching frequently accessed users.
func (s *UserService) getUserForUpdate(ctx context.Context, id values.UserID) (*entities.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domainerrors.WrapRepoError("get for update", "user", err)
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
	if email == user.GetEmail().String() {
		return nil
	}

	if err := s.validateEmail(email); err != nil {
		return err
	}

	return s.checkEmailAvailability(ctx, email)
}

func (s *UserService) checkEmailAvailability(ctx context.Context, email string) error {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return domainerrors.WrapServiceError("check existing email", err)
	}
	if existingUser != nil {
		return repositories.ErrUserAlreadyExists
	}
	return nil
}

func (s *UserService) validateNameUpdate(user *entities.User, name string) error {
	if name != user.GetUserName().String() {
		if err := s.validateUserName(name); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserService) applyUserUpdates(ctx context.Context, user *entities.User, email, name string) (*entities.User, error) {
	if err := user.SetEmail(email); err != nil {
		return nil, domainerrors.WrapServiceError("set email", err)
	}

	if err := user.SetName(name); err != nil {
		return nil, domainerrors.WrapServiceError("set name", err)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, domainerrors.WrapRepoError("save updated", "user", err)
	}

	return user, nil
}

// DeleteUser removes a user with business rules.
// TODO: SOFT DELETE - Consider implementing soft delete for audit trails
// TODO: CASCADE DELETE - Handle dependent entity cleanup (audit logs, user sessions, etc.)
func (s *UserService) DeleteUser(ctx context.Context, id values.UserID) error {
	// Business rule: Check if user exists before deletion
	_, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return domainerrors.WrapRepoError("find for deletion", "user", err)
	}

	// Business logic: Could add soft delete, cascade operations, etc.
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return domainerrors.WrapRepoError("delete", "user", err)
	}

	return nil
}

// ListUsers retrieves all users with business logic.
func (s *UserService) ListUsers(ctx context.Context) ([]*entities.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.NewInternalError("failed to list users", err)
	}

	// Business logic: Could add filtering, sorting, pagination, etc.
	return users, nil
}

// FilterActiveUsers demonstrates functional programming with lo library.
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

// GetUserEmailsWithResult demonstrates Result pattern with Railway Oriented Programming.
func (s *UserService) GetUserEmailsWithResult(ctx context.Context) mo.Result[[]string] {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return mo.Err[[]string](domainerrors.NewInternalError("failed to list users", err))
	}

	// Functional operation: extract emails
	emails := lo.Map(users, func(user *entities.User, _ int) string {
		return user.GetEmail().String()
	})

	return mo.Ok(emails)
}

// CreateUserWithResult demonstrates Railway Oriented Programming.
// TODO: FUNCTIONAL PROGRAMMING - This shows good Result[T] pattern usage - expand this approach.
func (s *UserService) CreateUserWithResult(ctx context.Context, id values.UserID, email, name string) mo.Result[*entities.User] {
	// Step 1: Validate inputs
	if validationResult := s.validateUserInputsResult(email, name); validationResult.IsError() {
		return mo.Err[*entities.User](validationResult.Error())
	}

	// Step 2: Check user doesn't exist
	if existsResult := s.checkUserNotExistsResult(ctx, email); existsResult.IsError() {
		return mo.Err[*entities.User](existsResult.Error())
	}

	// Step 3: Create and save user
	return s.createAndSaveUserResult(ctx, id, email, name)
}

// validateUserInputsResult validates user inputs using Result pattern.
func (s *UserService) validateUserInputsResult(email, name string) mo.Result[struct{}] {
	if err := s.validateEmail(email); err != nil {
		return mo.Err[struct{}](domainerrors.NewValidationError("email", err.Error()))
	}
	if err := s.validateUserName(name); err != nil {
		return mo.Err[struct{}](domainerrors.NewValidationError("name", err.Error()))
	}
	return mo.Ok(struct{}{})
}

// checkUserNotExistsResult checks if user exists using Result pattern.
func (s *UserService) checkUserNotExistsResult(ctx context.Context, email string) mo.Result[*entities.User] {
	existingUser, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return mo.Err[*entities.User](domainerrors.NewInternalError("failed to check existing user", err))
	}
	if existingUser != nil {
		return mo.Err[*entities.User](repositories.ErrUserAlreadyExists)
	}
	return mo.Ok[*entities.User](nil)
}

// createAndSaveUserResult creates and saves user using Result pattern.
// TODO: EXTRACTION - This private method could be part of a UserCreationService.
func (s *UserService) createAndSaveUserResult(ctx context.Context, id values.UserID, email, name string) mo.Result[*entities.User] {
	user, err := entities.NewUser(id, email, name)
	if err != nil {
		return mo.Err[*entities.User](err)
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return mo.Err[*entities.User](domainerrors.NewInternalError("failed to save user", err))
	}

	return mo.Ok(user)
}

// FindUserByEmailOption demonstrates Option pattern.
func (s *UserService) FindUserByEmailOption(ctx context.Context, email string) mo.Option[*entities.User] {
	if err := s.validateEmail(email); err != nil {
		return mo.None[*entities.User]()
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return mo.None[*entities.User]()
	}

	return mo.Some(user)
}

// BatchValidateUsers demonstrates functional operations for batch processing.
// TODO: PERFORMANCE - Consider parallel validation for large batches using goroutines
// TODO: MEMORY OPTIMIZATION - Stream processing for very large user sets.
func (s *UserService) BatchValidateUsers(users []*entities.User) map[values.UserID]error {
	// Use lo to create a map of validation results
	validationResults := lo.SliceToMap(users, func(user *entities.User) (values.UserID, error) {
		return user.ID, user.Validate()
	})

	// Filter only failed validations
	failedValidations := lo.PickBy(validationResults, func(_ values.UserID, err error) bool {
		return err != nil
	})

	return failedValidations
}

// GetUserStats demonstrates functional aggregation with lo.Reduce and lo.Ternary.
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
		parts := strings.Split(user.GetEmail().String(), "@")
		return lo.Ternary(len(parts) > 1, parts[1], "unknown")
	})

	// Count unique domains
	domainCounts := lo.CountValues(domains)
	stats["domains"] = len(domainCounts)

	// Calculate average days since registration using lo.Reduce
	now := time.Now()
	totalDays := lo.Reduce(users, func(acc int, user *entities.User, _ int) int {
		days := max(
			// Ensure non-negative days
			int(now.Sub(user.Created).Hours()/24), 0)
		return acc + days
	}, 0)

	// Safe division using lo.Max to prevent divide by zero
	userCount := lo.Max([]int{len(users), 1}) // Ensure at least 1 to prevent division by zero
	avgDays := lo.Ternary(len(users) > 0, totalDays/userCount, 0)
	stats["avg_days_since_registration"] = avgDays

	return stats, nil
}

// GetUsersWithFilters demonstrates advanced functional programming with type-safe filters.
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
			parts := strings.Split(user.GetEmail().String(), "@")
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

// ValidateUserBatchWithEither demonstrates Either pattern for batch operations.
// TODO: FUNCTIONAL PATTERN - Good Either usage, consider expanding this pattern throughout service layer.
func (s *UserService) ValidateUserBatchWithEither(users []*entities.User) mo.Either[[]error, []values.UserID] {
	validUsers := make([]values.UserID, 0)
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
		return mo.Left[[]error, []values.UserID](validationErrors)
	}
	return mo.Right[[]error, []values.UserID](validUsers)
}

// GetUsersByEmailDomains demonstrates more complex lo operations.
func (s *UserService) GetUsersByEmailDomains(ctx context.Context, domains []string) (map[string][]*entities.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.NewInternalError("failed to list users", err)
	}

	// Group users by email domain using lo.GroupBy
	usersByDomain := lo.GroupBy(users, func(user *entities.User) string {
		parts := strings.Split(user.GetEmail().String(), "@")
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

// validateEmail enforces business rules for email validation.
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

// validateUserName enforces business rules for display name validation.
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
