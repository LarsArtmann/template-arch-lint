// Package persistence provides infrastructure layer data persistence implementations.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
)

// SQLUserRepository implements UserRepository using SQL database.
type SQLUserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSQLUserRepository creates a new SQL user repository.
func NewSQLUserRepository(db *sql.DB, logger *slog.Logger) *SQLUserRepository {
	repo := &SQLUserRepository{
		db:     db,
		logger: logger,
	}

	// Initialize database schema if db is not nil
	if db != nil {
		if err := repo.initSchema(); err != nil {
			if logger != nil {
				logger.Error("Failed to initialize database schema", "error", err)
			}
		}
	}

	return repo
}

// initSchema creates the users table if it doesn't exist.
func (r *SQLUserRepository) initSchema() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			created DATETIME NOT NULL,
			modified DATETIME NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_created ON users(created);
	`

	_, err := r.db.ExecContext(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("Database schema initialized successfully")
	}
	return nil
}

// Save persists a user entity.
func (r *SQLUserRepository) Save(ctx context.Context, user *entities.User) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if r.logger != nil {
		r.logger.Debug("Saving user", "user_id", user.ID, "email", user.Email)
	}

	query := `
		INSERT OR REPLACE INTO users (id, email, name, created, modified)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Created,
		user.Modified,
	)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to save user", "user_id", user.ID, "error", err)
		}
		return fmt.Errorf("failed to save user: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("User saved successfully", "user_id", user.ID, "email", user.Email)
	}
	return nil
}

// FindByID retrieves a user by their unique identifier.
func (r *SQLUserRepository) FindByID(ctx context.Context, id entities.UserID) (*entities.User, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if r.logger != nil {
		r.logger.Debug("Finding user by ID", "user_id", id)
	}

	query := `
		SELECT id, email, name, created, modified
		FROM users
		WHERE id = ?
	`

	var user entities.User
	var created, modified time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&created,
		&modified,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if r.logger != nil {
				r.logger.Debug("User not found", "user_id", id)
			}
			return nil, repositories.ErrUserNotFound
		}
		if r.logger != nil {
			r.logger.Error("Failed to find user by ID", "user_id", id, "error", err)
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	user.Created = created
	user.Modified = modified

	if r.logger != nil {
		r.logger.Debug("User found successfully", "user_id", id, "email", user.Email)
	}
	return &user, nil
}

// FindByEmail retrieves a user by their email address.
func (r *SQLUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if r.logger != nil {
		r.logger.Debug("Finding user by email", "email", email)
	}

	query := `
		SELECT id, email, name, created, modified
		FROM users
		WHERE email = ?
	`

	var user entities.User
	var created, modified time.Time

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&created,
		&modified,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if r.logger != nil {
				r.logger.Debug("User not found", "email", email)
			}
			return nil, repositories.ErrUserNotFound
		}
		if r.logger != nil {
			r.logger.Error("Failed to find user by email", "email", email, "error", err)
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	user.Created = created
	user.Modified = modified

	if r.logger != nil {
		r.logger.Debug("User found successfully", "email", email, "user_id", user.ID)
	}
	return &user, nil
}

// Delete removes a user from the repository.
func (r *SQLUserRepository) Delete(ctx context.Context, id entities.UserID) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	if r.logger != nil {
		r.logger.Debug("Deleting user", "user_id", id)
	}

	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to delete user", "user_id", id, "error", err)
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to get rows affected", "user_id", id, "error", err)
		}
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		if r.logger != nil {
			r.logger.Debug("User not found for deletion", "user_id", id)
		}
		return repositories.ErrUserNotFound
	}

	if r.logger != nil {
		r.logger.Info("User deleted successfully", "user_id", id)
	}
	return nil
}

// List retrieves all users (useful for testing and admin operations).
func (r *SQLUserRepository) List(ctx context.Context) ([]*entities.User, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if r.logger != nil {
		r.logger.Debug("Listing all users")
	}

	query := `
		SELECT id, email, name, created, modified
		FROM users
		ORDER BY created ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to list users", "error", err)
		}
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			if r.logger != nil {
				r.logger.Error("Failed to close rows", "error", err)
			}
		}
	}()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		var created, modified time.Time

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&created,
			&modified,
		)
		if err != nil {
			if r.logger != nil {
				r.logger.Error("Failed to scan user row", "error", err)
			}
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		user.Created = created
		user.Modified = modified
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		if r.logger != nil {
			r.logger.Error("Error occurred during row iteration", "error", err)
		}
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	if r.logger != nil {
		r.logger.Info("Users listed successfully", "count", len(users))
	}
	return users, nil
}

// FindByUsername retrieves a user by their username (name field).
func (r *SQLUserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if r.logger != nil {
		r.logger.Debug("Finding user by username", "username", username)
	}

	query := `
		SELECT id, email, name, created, modified
		FROM users
		WHERE name = ?
	`

	var user entities.User
	var created, modified time.Time

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&created,
		&modified,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if r.logger != nil {
				r.logger.Debug("User not found", "username", username)
			}
			return nil, repositories.ErrUserNotFound
		}
		if r.logger != nil {
			r.logger.Error("Failed to find user by username", "username", username, "error", err)
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}

	user.Created = created
	user.Modified = modified

	if r.logger != nil {
		r.logger.Debug("User found successfully", "username", username, "user_id", user.ID)
	}
	return &user, nil
}

// Update updates an existing user.
func (r *SQLUserRepository) Update(ctx context.Context, user *entities.User) error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if r.logger != nil {
		r.logger.Debug("Updating user", "user_id", user.ID, "email", user.Email)
	}

	// First check if user exists
	existingUser, err := r.FindByID(ctx, user.ID)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET email = ?, name = ?, modified = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.Name,
		user.Modified,
		user.ID,
	)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to update user", "user_id", user.ID, "error", err)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to get rows affected", "user_id", user.ID, "error", err)
		}
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		if r.logger != nil {
			r.logger.Debug("No rows updated", "user_id", user.ID)
		}
		return repositories.ErrUserNotFound
	}

	if r.logger != nil {
		r.logger.Info("User updated successfully", "user_id", user.ID, "email", user.Email, "old_email", existingUser.Email)
	}
	return nil
}

// Count returns the total number of users.
func (r *SQLUserRepository) Count(ctx context.Context) (int, error) {
	if r.db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	if r.logger != nil {
		r.logger.Debug("Counting users")
	}

	query := `SELECT COUNT(*) FROM users`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("Failed to count users", "error", err)
		}
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	if r.logger != nil {
		r.logger.Debug("Users counted successfully", "count", count)
	}
	return count, nil
}

// Exists checks if a user with the given ID exists.
func (r *SQLUserRepository) Exists(ctx context.Context, id entities.UserID) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("database connection is nil")
	}

	if r.logger != nil {
		r.logger.Debug("Checking if user exists", "user_id", id)
	}

	query := `SELECT 1 FROM users WHERE id = ? LIMIT 1`

	var exists int
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if r.logger != nil {
				r.logger.Debug("User does not exist", "user_id", id)
			}
			return false, nil
		}
		if r.logger != nil {
			r.logger.Error("Failed to check if user exists", "user_id", id, "error", err)
		}
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}

	if r.logger != nil {
		r.logger.Debug("User existence checked", "user_id", id, "exists", exists == 1)
	}
	return exists == 1, nil
}
