-- Users table schema
-- Optimized for performance with proper indexing and constraints

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created ON users(created);
CREATE INDEX IF NOT EXISTS idx_users_modified ON users(modified);
CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_users_name_created ON users(name, created);
CREATE INDEX IF NOT EXISTS idx_users_email_created ON users(email, created);

-- Partial indexes for active/recent users (performance optimization)
CREATE INDEX IF NOT EXISTS idx_users_recent ON users(created, modified)
WHERE created > datetime('now', '-30 days');

-- Full-text search index for name (if needed for search functionality)
-- CREATE INDEX IF NOT EXISTS idx_users_name_fts ON users(name) WHERE name IS NOT NULL;

-- Statistics collection for query optimizer
ANALYZE users;
