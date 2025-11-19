package infrastructure

import "database/sql"

// Database represents an infrastructure concern
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dsn string) (*Database, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}
