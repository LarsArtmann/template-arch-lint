// Package services provides business logic and domain services.
package services

import (
	"context"
	"strings"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	domainerrors "github.com/LarsArtmann/template-arch-lint/pkg/errors"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// UserQueryService defines the interface for user query operations.
// This follows CQRS principles by separating read operations from write operations.
type UserQueryService interface {
	// GetUser retrieves a user by their unique identifier.
	GetUser(ctx context.Context, id values.UserID) (*entities.User, error)

	// GetUserByEmail retrieves a user by their email address.
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)

	// ListUsers retrieves all users in the system.
	ListUsers(ctx context.Context) ([]*entities.User, error)

	// GetUserEmailsWithResult retrieves all user emails using Result pattern.
	GetUserEmailsWithResult(ctx context.Context) mo.Result[[]string]

	// FindUserByEmailOption finds a user by email using Option pattern.
	FindUserByEmailOption(ctx context.Context, email string) mo.Option[*entities.User]

	// GetUserStats retrieves user statistics and metrics.
	GetUserStats(ctx context.Context) (map[string]int, error)

	// GetUsersWithFilters retrieves users based on provided filters.
	GetUsersWithFilters(ctx context.Context, filters UserFilters) ([]*entities.User, error)

	// GetUsersByEmailDomains retrieves users grouped by their email domains.
	GetUsersByEmailDomains(ctx context.Context, domains []string) (map[string][]*entities.User, error)
}

// userQueryServiceImpl implements UserQueryService interface.
type userQueryServiceImpl struct {
	userRepo repositories.UserRepository
}

// NewUserQueryService creates a new instance of UserQueryService.
func NewUserQueryService(userRepo repositories.UserRepository) UserQueryService {
	return &userQueryServiceImpl{
		userRepo: userRepo,
	}
}

// GetUser retrieves a user by their unique identifier.
func (s *userQueryServiceImpl) GetUser(ctx context.Context, id values.UserID) (*entities.User, error) {
	// TODO: Add caching layer for frequently accessed users
	// TODO: Add metrics tracking for query performance
	// TODO: Add validation for user ID format
	// TODO: Consider adding authorization checks

	if id.IsEmpty() {
		return nil, domainerrors.NewValidationError("userID", "user ID cannot be empty")
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domainerrors.WrapRepoError("get", "user", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by their email address.
func (s *userQueryServiceImpl) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	// TODO: Add email validation using Email value object
	// TODO: Add caching by email for performance
	// TODO: Add rate limiting for email lookups
	// TODO: Consider case-insensitive email matching

	if email == "" {
		return nil, domainerrors.NewValidationError("email", "email cannot be empty")
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, domainerrors.WrapRepoError("get by email", "user", err)
	}

	return user, nil
}

// ListUsers retrieves all users in the system.
func (s *userQueryServiceImpl) ListUsers(ctx context.Context) ([]*entities.User, error) {
	// TODO: Add pagination support for large datasets
	// TODO: Add sorting options (by name, email, created date)
	// TODO: Add filtering capabilities
	// TODO: Add caching for frequently accessed lists
	// TODO: Consider streaming for very large result sets

	return s.userRepo.List(ctx)
}

// GetUserEmailsWithResult retrieves all user emails using Result pattern.
func (s *userQueryServiceImpl) GetUserEmailsWithResult(ctx context.Context) mo.Result[[]string] {
	// TODO: Optimize with direct email query instead of fetching full users
	// TODO: Add email deduplication logic
	// TODO: Add email format validation

	users, err := s.userRepo.List(ctx)
	if err != nil {
		return mo.Err[[]string](domainerrors.WrapRepoError("list for emails", "user", err))
	}

	emails := lo.Map(users, func(user *entities.User, _ int) string {
		return user.GetEmail().String()
	})

	return mo.Ok(emails)
}

// FindUserByEmailOption finds a user by email using Option pattern.
func (s *userQueryServiceImpl) FindUserByEmailOption(ctx context.Context, email string) mo.Option[*entities.User] {
	// TODO: Add email validation using Email value object
	// TODO: Add caching support
	// TODO: Add audit logging for security compliance

	if email == "" {
		return mo.None[*entities.User]()
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Log error but return None for Option pattern
		return mo.None[*entities.User]()
	}

	return mo.Some(user)
}

// GetUserStats retrieves user statistics and metrics.
func (s *userQueryServiceImpl) GetUserStats(ctx context.Context) (map[string]int, error) {
	// TODO: Add comprehensive metrics (active users, domains, creation trends)
	// TODO: Add caching for expensive statistics calculations
	// TODO: Add real-time vs cached statistics options
	// TODO: Add date range filtering for statistics

	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.WrapRepoError("list for stats", "user", err)
	}

	stats := map[string]int{
		"total_users": len(users),
	}

	// Calculate domain distribution
	domainCount := make(map[string]int)
	for _, user := range users {
		email := user.GetEmail().String()
		if atIndex := strings.Index(email, "@"); atIndex != -1 {
			domain := email[atIndex+1:]
			domainCount[domain]++
		}
	}
	stats["unique_domains"] = len(domainCount)

	return stats, nil
}

// GetUsersWithFilters retrieves users based on provided filters.
func (s *userQueryServiceImpl) GetUsersWithFilters(ctx context.Context, filters UserFilters) ([]*entities.User, error) {
	// TODO: Add database-level filtering for better performance
	// TODO: Add validation for filter parameters
	// TODO: Add support for complex filter combinations
	// TODO: Add filter result caching

	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.WrapRepoError("list for filtering", "user", err)
	}

	filtered := lo.Filter(users, func(user *entities.User, _ int) bool {
		if filters.Domain != nil && *filters.Domain != "" {
			email := user.GetEmail().String()
			if atIndex := strings.Index(email, "@"); atIndex != -1 {
				domain := email[atIndex+1:]
				if domain != *filters.Domain {
					return false
				}
			}
		}

		if filters.Active != nil {
			// TODO: Implement user.IsActive() method when UserStatus value object exists
			// For now, assume all users are active
			if !*filters.Active {
				return false
			}
		}

		return true
	})

	return filtered, nil
}

// GetUsersByEmailDomains retrieves users grouped by their email domains.
func (s *userQueryServiceImpl) GetUsersByEmailDomains(ctx context.Context, domains []string) (map[string][]*entities.User, error) {
	// TODO: Add database-level domain filtering for performance
	// TODO: Add domain validation
	// TODO: Add support for wildcard domain matching
	// TODO: Add result caching by domain combinations

	if len(domains) == 0 {
		return map[string][]*entities.User{}, nil
	}

	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, domainerrors.WrapRepoError("list for domain grouping", "user", err)
	}

	domainSet := lo.SliceToMap(domains, func(domain string) (string, bool) {
		return strings.ToLower(domain), true
	})

	result := make(map[string][]*entities.User)
	for domain := range domainSet {
		result[domain] = []*entities.User{}
	}

	for _, user := range users {
		email := user.GetEmail().String()
		if atIndex := strings.Index(email, "@"); atIndex != -1 {
			userDomain := strings.ToLower(email[atIndex+1:])
			if domainSet[userDomain] {
				result[userDomain] = append(result[userDomain], user)
			}
		}
	}

	return result, nil
}
