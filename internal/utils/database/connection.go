// Package database provides reusable database connection and configuration utilities.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import sqlite3 driver
)

// ConnectionConfig holds database connection configuration.
type ConnectionConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultSQLiteConfig returns a default configuration for SQLite databases.
func DefaultSQLiteConfig(dsnPath string) *ConnectionConfig {
	return &ConnectionConfig{
		Driver:          "sqlite3",
		DSN:             dsnPath,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// InMemoryTestConfig returns a configuration for in-memory testing databases.
func InMemoryTestConfig() *ConnectionConfig {
	return &ConnectionConfig{
		Driver:          "sqlite3",
		DSN:             ":memory:",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// ConnectionManager provides utilities for database connection management.
type ConnectionManager struct {
	logger *slog.Logger
}

// NewConnectionManager creates a new ConnectionManager instance.
func NewConnectionManager(logger *slog.Logger) *ConnectionManager {
	return &ConnectionManager{
		logger: logger,
	}
}

// Connect establishes a database connection with the provided configuration.
func (cm *ConnectionManager) Connect(ctx context.Context, config *ConnectionConfig) (*sql.DB, error) {
	if config == nil {
		return nil, fmt.Errorf("database configuration cannot be nil")
	}

	// Open database connection
	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	cm.configureConnectionPool(db, config)

	// Test connection
	if err := cm.testConnection(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	cm.logConnectionSuccess(config)
	return db, nil
}

// ConnectWithDefaults establishes a database connection using sensible defaults.
func (cm *ConnectionManager) ConnectWithDefaults(ctx context.Context, driver, dsn string) (*sql.DB, error) {
	config := &ConnectionConfig{
		Driver:          driver,
		DSN:             dsn,
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
	return cm.Connect(ctx, config)
}

// ConnectInMemory establishes an in-memory SQLite connection for testing.
func (cm *ConnectionManager) ConnectInMemory(ctx context.Context) (*sql.DB, error) {
	config := InMemoryTestConfig()
	return cm.Connect(ctx, config)
}

// configureConnectionPool applies connection pool settings to the database.
func (cm *ConnectionManager) configureConnectionPool(db *sql.DB, config *ConnectionConfig) {
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
}

// testConnection verifies the database connection is working.
func (cm *ConnectionManager) testConnection(ctx context.Context, db *sql.DB) error {
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

// logConnectionSuccess logs successful database connection.
func (cm *ConnectionManager) logConnectionSuccess(config *ConnectionConfig) {
	if cm.logger != nil {
		cm.logger.Info("Database connected successfully",
			"driver", config.Driver,
			"max_open_conns", config.MaxOpenConns,
			"max_idle_conns", config.MaxIdleConns,
		)
	}
}

// Disconnect safely closes a database connection.
func (cm *ConnectionManager) Disconnect(db *sql.DB) error {
	if db == nil {
		return nil
	}

	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	if cm.logger != nil {
		cm.logger.Info("Database disconnected successfully")
	}

	return nil
}

// HealthCheck performs a basic health check on the database connection.
func (cm *ConnectionManager) HealthCheck(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Set a timeout for the health check
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(healthCtx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// Transaction utilities

// TransactionFunc represents a function that runs within a database transaction.
type TransactionFunc func(tx *sql.Tx) error

// RunInTransaction executes a function within a database transaction.
func (cm *ConnectionManager) RunInTransaction(ctx context.Context, db *sql.DB, fn TransactionFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is always closed
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Execute the function
	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			cm.logger.Error("Failed to rollback transaction", "error", rollbackErr)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Convenience functions for common patterns

// QuickConnect provides a simple way to connect to a SQLite database.
func QuickConnect(ctx context.Context, dsnPath string, logger *slog.Logger) (*sql.DB, error) {
	manager := NewConnectionManager(logger)
	config := DefaultSQLiteConfig(dsnPath)
	return manager.Connect(ctx, config)
}

// QuickConnectInMemory provides a simple way to connect to an in-memory database for testing.
func QuickConnectInMemory(ctx context.Context, logger *slog.Logger) (*sql.DB, error) {
	manager := NewConnectionManager(logger)
	return manager.ConnectInMemory(ctx)
}

// SafeClose safely closes a database connection without panicking.
func SafeClose(db *sql.DB, logger *slog.Logger) {
	if db == nil {
		return
	}

	if err := db.Close(); err != nil && logger != nil {
		logger.Error("Failed to close database connection", "error", err)
	}
}
