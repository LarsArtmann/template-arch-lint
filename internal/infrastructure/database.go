package infrastructure

import (
	"database/sql"
	"fmt"
)

// Database represents an infrastructure concern.
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection.
func NewDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("dsn=%s: %w", dsn, err)
	}

	return &Database{db: db}, nil
}
