// SQLC-based implementation of UserRepository
//go:build sqlite3

package persistence

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/LarsArtmann/template-arch-lint/internal/db"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/entities"
	"github.com/LarsArtmann/template-arch-lint/internal/domain/repositories"
)

// SQLCUserRepository implements UserRepository using SQLC generated code
type SQLCUserRepository struct {
	queries *db.Queries
	db      *sql.DB
	logger  *slog.Logger
}

// NewSQLCUserRepository creates a new SQLC-based user repository
func NewSQLCUserRepository(database *sql.DB, logger *slog.Logger) *SQLCUserRepository {
	repo := &SQLCUserRepository{
		queries: db.New(database),
		db:      database,
		logger:  logger,
	}

	// Initialize database schema
	if err := repo.initSchema(); err != nil {
		logger.Error("Failed to initialize database schema", "error", err)
	}

	return repo
}

// initSchema creates the users table if it doesn't exist
func (r *SQLCUserRepository) initSchema() error {
	// Use the schema from our SQL files
	schema := `
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

	_, err := r.db.Exec(schema)
	if err != nil {
		return err
	}

	r.logger.Info("Database schema initialized successfully")
	return nil
}

// Save persists a user entity using SQLC generated code
func (r *SQLCUserRepository) Save(ctx context.Context, user *entities.User) error {
	r.logger.Debug("Saving user", "user_id", user.ID, "email", user.Email)

	params := &db.SaveUserParams{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Created:  user.Created,
		Modified: user.Modified,
	}

	err := r.queries.SaveUser(ctx, params)
	if err != nil {
		r.logger.Error("Failed to save user", "user_id", user.ID, "error", err)
		return err
	}

	r.logger.Info("User saved successfully", "user_id", user.ID, "email", user.Email)
	return nil
}

// FindByID retrieves a user by their unique identifier using SQLC generated code
func (r *SQLCUserRepository) FindByID(ctx context.Context, id entities.UserID) (*entities.User, error) {
	r.logger.Debug("Finding user by ID", "user_id", id)

	dbUser, err := r.queries.FindUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug("User not found", "user_id", id)
			return nil, repositories.ErrUserNotFound
		}
		r.logger.Error("Failed to find user by ID", "user_id", id, "error", err)
		return nil, err
	}

	// Convert from SQLC generated struct to domain entity
	user := &entities.User{
		ID:       dbUser.ID,
		Email:    dbUser.Email,
		Name:     dbUser.Name,
		Created:  dbUser.Created,
		Modified: dbUser.Modified,
	}

	r.logger.Debug("User found successfully", "user_id", id, "email", user.Email)
	return user, nil
}

// FindByEmail retrieves a user by their email address using SQLC generated code
func (r *SQLCUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	r.logger.Debug("Finding user by email", "email", email)

	dbUser, err := r.queries.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Debug("User not found", "email", email)
			return nil, repositories.ErrUserNotFound
		}
		r.logger.Error("Failed to find user by email", "email", email, "error", err)
		return nil, err
	}

	// Convert from SQLC generated struct to domain entity
	user := &entities.User{
		ID:       dbUser.ID,
		Email:    dbUser.Email,
		Name:     dbUser.Name,
		Created:  dbUser.Created,
		Modified: dbUser.Modified,
	}

	r.logger.Debug("User found successfully", "email", email, "user_id", user.ID)
	return user, nil
}

// Delete removes a user from the repository using SQLC generated code
func (r *SQLCUserRepository) Delete(ctx context.Context, id entities.UserID) error {
	r.logger.Debug("Deleting user", "user_id", id)

	err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		r.logger.Error("Failed to delete user", "user_id", id, "error", err)
		return err
	}

	r.logger.Info("User deleted successfully", "user_id", id)
	return nil
}

// List retrieves all users using SQLC generated code
func (r *SQLCUserRepository) List(ctx context.Context) ([]*entities.User, error) {
	r.logger.Debug("Listing all users")

	dbUsers, err := r.queries.ListUsers(ctx)
	if err != nil {
		r.logger.Error("Failed to list users", "error", err)
		return nil, err
	}

	// Convert from SQLC generated structs to domain entities
	users := make([]*entities.User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		user := &entities.User{
			ID:       dbUser.ID,
			Email:    dbUser.Email,
			Name:     dbUser.Name,
			Created:  dbUser.Created,
			Modified: dbUser.Modified,
		}
		users = append(users, user)
	}

	r.logger.Info("Users listed successfully", "count", len(users))
	return users, nil
}

// Additional helper methods that leverage SQLC

// CountUsers returns the total number of users
func (r *SQLCUserRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}