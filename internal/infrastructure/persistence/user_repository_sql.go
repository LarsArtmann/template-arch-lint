// SQL implementation of UserRepository
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
)

// SQLUserRepository implements UserRepository using SQL database
type SQLUserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewSQLUserRepository creates a new SQL user repository
func NewSQLUserRepository(db *sql.DB, logger *slog.Logger) *SQLUserRepository {
	repo := &SQLUserRepository{
		db:     db,
		logger: logger,
	}

	// Initialize database schema
	if err := repo.initSchema(); err != nil {
		logger.Error("Failed to initialize database schema", "error", err)
	}

	return repo
}

// initSchema creates the users table if it doesn't exist
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

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	r.logger.Info("Database schema initialized successfully")
	return nil
}

// Save persists a user entity
func (r *SQLUserRepository) Save(ctx context.Context, user *entities.User) error {
	r.logger.Debug("Saving user", "user_id", user.ID, "email", user.Email)

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
		r.logger.Error("Failed to save user", "user_id", user.ID, "error", err)
		return fmt.Errorf("failed to save user: %w", err)
	}

	r.logger.Info("User saved successfully", "user_id", user.ID, "email", user.Email)
	return nil
}

// FindByID retrieves a user by their unique identifier
func (r *SQLUserRepository) FindByID(ctx context.Context, id entities.UserID) (*entities.User, error) {
	r.logger.Debug("Finding user by ID", "user_id", id)

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
		if err == sql.ErrNoRows {
			r.logger.Debug("User not found", "user_id", id)
			return nil, repositories.ErrUserNotFound
		}
		r.logger.Error("Failed to find user by ID", "user_id", id, "error", err)
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	user.Created = created
	user.Modified = modified

	r.logger.Debug("User found successfully", "user_id", id, "email", user.Email)
	return &user, nil
}

// FindByEmail retrieves a user by their email address
func (r *SQLUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.logger.Debug("Finding user by email", "email", email)

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
		if err == sql.ErrNoRows {
			r.logger.Debug("User not found", "email", email)
			return nil, repositories.ErrUserNotFound
		}
		r.logger.Error("Failed to find user by email", "email", email, "error", err)
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	user.Created = created
	user.Modified = modified

	r.logger.Debug("User found successfully", "email", email, "user_id", user.ID)
	return &user, nil
}

// Delete removes a user from the repository
func (r *SQLUserRepository) Delete(ctx context.Context, id entities.UserID) error {
	r.logger.Debug("Deleting user", "user_id", id)

	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete user", "user_id", id, "error", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "user_id", id, "error", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Debug("User not found for deletion", "user_id", id)
		return repositories.ErrUserNotFound
	}

	r.logger.Info("User deleted successfully", "user_id", id)
	return nil
}

// List retrieves all users (useful for testing and admin operations)
func (r *SQLUserRepository) List(ctx context.Context) ([]*entities.User, error) {
	r.logger.Debug("Listing all users")

	query := `
		SELECT id, email, name, created, modified
		FROM users
		ORDER BY created ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		r.logger.Error("Failed to list users", "error", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

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
			r.logger.Error("Failed to scan user row", "error", err)
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		user.Created = created
		user.Modified = modified
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Error occurred during row iteration", "error", err)
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	r.logger.Info("Users listed successfully", "count", len(users))
	return users, nil
}
