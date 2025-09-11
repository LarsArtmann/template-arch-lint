package repositories

import (
	"context"
	"sync"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/errors"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/values"
)

// InMemoryUserRepository implements UserRepository interface with in-memory storage.
// TODO: SCALABILITY - Consider implementing LRU cache eviction for production use
// TODO: PERSISTENCE - Add optional backup/restore functionality
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[values.UserID]*entities.User
}

// NewInMemoryUserRepository creates a new in-memory user repository.
func NewInMemoryUserRepository() UserRepository {
	return &InMemoryUserRepository{
		users: make(map[values.UserID]*entities.User),
	}
}

// Save persists a user entity.
func (r *InMemoryUserRepository) Save(_ context.Context, user *entities.User) error {
	if user == nil {
		return errors.NewValidationError("user", "user cannot be nil")
	}

	if err := user.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if user already exists for create operations
	if _, exists := r.users[user.ID]; exists {
		// Update existing user's modified time
		user.Modified = time.Now()
	}

	// Create a copy to avoid external modifications
	userCopy := *user
	r.users[user.ID] = &userCopy

	return nil
}

// FindByID retrieves a user by their unique identifier.
func (r *InMemoryUserRepository) FindByID(_ context.Context, id values.UserID) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Return a copy to prevent external modifications
	userCopy := *user
	return &userCopy, nil
}

// FindByEmail retrieves a user by their email address.
func (r *InMemoryUserRepository) FindByEmail(_ context.Context, email string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.GetEmail().String() == email {
			// Return a copy to prevent external modifications
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, ErrUserNotFound
}

// FindByUsername retrieves a user by their username (name field).
func (r *InMemoryUserRepository) FindByUsername(_ context.Context, username string) (*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.GetUserName().String() == username {
			// Return a copy to prevent external modifications
			userCopy := *user
			return &userCopy, nil
		}
	}

	return nil, ErrUserNotFound
}

// Delete removes a user from the repository.
func (r *InMemoryUserRepository) Delete(_ context.Context, id values.UserID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[id]; !exists {
		return ErrUserNotFound
	}

	delete(r.users, id)
	return nil
}

// List retrieves all users.
func (r *InMemoryUserRepository) List(_ context.Context) ([]*entities.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*entities.User, 0, len(r.users))
	for _, user := range r.users {
		// Return copies to prevent external modifications
		userCopy := *user
		users = append(users, &userCopy)
	}

	return users, nil
}
