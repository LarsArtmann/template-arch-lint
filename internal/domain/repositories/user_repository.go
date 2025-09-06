// Package repositories provides domain repository interfaces and implementations.
package repositories

import (
	"context"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

// ErrUserNotFound is returned when a user is not found.
var ErrUserNotFound = errors.NewNotFoundError("user", "")

// ErrUserAlreadyExists is returned when a user already exists.
var ErrUserAlreadyExists = errors.NewConflictError("user already exists", errors.ErrorDetails{
	Resource: "user",
})

// UserRepository defines the contract for user data persistence.
type UserRepository interface {
	// Save persists a user entity
	Save(ctx context.Context, user *entities.User) error

	// FindByID retrieves a user by their unique identifier
	// TODO: QUERY OPTIMIZATION - Consider implementing query hints for performance
	FindByID(ctx context.Context, id values.UserID) (*entities.User, error)

	// FindByEmail retrieves a user by their email address
	// TODO: TYPE SAFETY - Replace string with values.Email for validation
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// FindByUsername retrieves a user by their username
	// TODO: TYPE SAFETY - Replace string with values.UserName for validation
	FindByUsername(ctx context.Context, username string) (*entities.User, error)

	// Delete removes a user from the repository
	// TODO: SOFT DELETE - Consider adding soft delete support with deleted_at timestamp
	Delete(ctx context.Context, id values.UserID) error

	// List retrieves all users (useful for testing and admin operations)
	// TODO: PAGINATION - Add pagination support for large datasets
	// TODO: FILTERING - Add filtering capabilities (active/inactive, by domain, etc.)
	List(ctx context.Context) ([]*entities.User, error)
}
