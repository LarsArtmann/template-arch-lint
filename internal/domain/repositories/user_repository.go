// Package repositories provides domain repository interfaces and implementations.
package repositories

import (
	"context"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
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
	FindByID(ctx context.Context, id entities.UserID) (*entities.User, error)

	// FindByEmail retrieves a user by their email address
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// FindByUsername retrieves a user by their username
	FindByUsername(ctx context.Context, username string) (*entities.User, error)

	// Delete removes a user from the repository
	Delete(ctx context.Context, id entities.UserID) error

	// List retrieves all users (useful for testing and admin operations)
	List(ctx context.Context) ([]*entities.User, error)
}
