// Package persistence provides infrastructure layer data persistence implementations.
package persistence

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
	"github.com/samber/lo"
)

// UserRepositoryMemory provides an in-memory implementation of the user repository.
// This is perfect for a linting template - demonstrates patterns without database complexity.
type UserRepositoryMemory struct {
	users map[string]*entities.User
	mutex sync.RWMutex
}

// NewUserRepositoryMemory creates a new in-memory user repository.
func NewUserRepositoryMemory() repositories.UserRepository {
	return &UserRepositoryMemory{
		users: make(map[string]*entities.User),
	}
}

// Save persists a user entity.
func (r *UserRepositoryMemory) Save(ctx context.Context, user *entities.User) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Set timestamps for new users
	if user.Created.IsZero() {
		user.Created = time.Now()
	}
	user.Modified = time.Now()

	// Store user by ID
	r.users[user.ID.String()] = user

	return nil
}

// FindByID retrieves a user by their unique identifier.
func (r *UserRepositoryMemory) FindByID(ctx context.Context, id entities.UserID) (*entities.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id.String()]
	if !exists {
		return nil, repositories.ErrUserNotFound
	}

	return user, nil
}

// FindByEmail retrieves a user by their email address.
func (r *UserRepositoryMemory) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Use functional programming patterns (samber/lo) as demonstrated in linting template
	user, found := lo.Find(lo.Values(r.users), func(u *entities.User) bool {
		return u.Email == email
	})

	if !found {
		return nil, repositories.ErrUserNotFound
	}

	return user, nil
}

// FindByUsername retrieves a user by their username (name field).
func (r *UserRepositoryMemory) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Use functional programming patterns (samber/lo) as demonstrated in linting template
	user, found := lo.Find(lo.Values(r.users), func(u *entities.User) bool {
		return u.Name == username
	})

	if !found {
		return nil, repositories.ErrUserNotFound
	}

	return user, nil
}

// Delete removes a user from the repository.
func (r *UserRepositoryMemory) Delete(ctx context.Context, id entities.UserID) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	_, exists := r.users[id.String()]
	if !exists {
		return repositories.ErrUserNotFound
	}

	delete(r.users, id.String())
	return nil
}

// List returns all users.
func (r *UserRepositoryMemory) List(ctx context.Context) ([]*entities.User, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Convert to slice using functional patterns
	allUsers := lo.Values(r.users)

	return allUsers, nil
}

// SeedTestData adds sample users for development and testing.
// This demonstrates the repository pattern without database complexity.
func (r *UserRepositoryMemory) SeedTestData() error {
	ctx := context.Background()

	// Create sample users using domain value objects
	users := []struct {
		id    string
		email string
		name  string
	}{
		{"user-001", "admin@example.com", "System Administrator"},
		{"user-002", "user@example.com", "Regular User"},
		{"user-003", "test@example.com", "Test User"},
	}

	for _, u := range users {
		userID, err := values.NewUserID(u.id)
		if err != nil {
			return fmt.Errorf("failed to create user ID: %w", err)
		}

		user := &entities.User{
			ID:      userID,
			Email:   u.email,
			Name:    u.name,
			Created: time.Now().Add(-time.Duration(len(users)-1) * time.Hour),
		}

		if err := r.Save(ctx, user); err != nil {
			return fmt.Errorf("failed to seed user %s: %w", u.id, err)
		}
	}

	return nil
}
