-- Users CRUD queries
-- Optimized for performance with proper indexing and query patterns

-- name: SaveUser :exec
INSERT OR REPLACE INTO users (id, email, name, created, modified)
VALUES (?, ?, ?, ?, ?);

-- name: FindUserByID :one
SELECT id, email, name, created, modified
FROM users
WHERE id = ?
LIMIT 1;

-- name: FindUserByEmail :one
SELECT id, email, name, created, modified
FROM users
WHERE email = ?
LIMIT 1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT id, email, name, created, modified
FROM users
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: ListUsersAll :many
SELECT id, email, name, created, modified
FROM users
ORDER BY created DESC;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- Performance-optimized queries

-- name: FindRecentUsers :many
SELECT id, email, name, created, modified
FROM users
WHERE created > datetime('now', '-30 days')
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: SearchUsersByName :many
SELECT id, email, name, created, modified
FROM users
WHERE name LIKE '%' || ? || '%'
ORDER BY name
LIMIT ? OFFSET ?;

-- name: SearchUsersByEmail :many
SELECT id, email, name, created, modified
FROM users
WHERE email LIKE '%' || ? || '%'
ORDER BY email
LIMIT ? OFFSET ?;

-- name: GetUsersCreatedAfter :many
SELECT id, email, name, created, modified
FROM users
WHERE created >= ?
ORDER BY created DESC
LIMIT ? OFFSET ?;

-- name: GetUsersModifiedAfter :many
SELECT id, email, name, created, modified
FROM users
WHERE modified >= ?
ORDER BY modified DESC
LIMIT ? OFFSET ?;

-- name: CountUsersByDateRange :one
SELECT COUNT(*) FROM users
WHERE created BETWEEN ? AND ?;

-- name: GetUserStats :one
SELECT
    COUNT(*) as total_users,
    COUNT(CASE WHEN created > datetime('now', '-1 day') THEN 1 END) as users_today,
    COUNT(CASE WHEN created > datetime('now', '-7 days') THEN 1 END) as users_week,
    COUNT(CASE WHEN created > datetime('now', '-30 days') THEN 1 END) as users_month,
    CAST(COALESCE(MIN(created), '1900-01-01 00:00:00') AS TEXT) as first_user_created,
    CAST(COALESCE(MAX(created), '1900-01-01 00:00:00') AS TEXT) as last_user_created
FROM users;

-- name: GetActiveUserEmails :many
SELECT DISTINCT email
FROM users
WHERE modified > datetime('now', '-7 days')
ORDER BY email;

-- Bulk operations for performance

-- name: BulkInsertUsers :exec
INSERT OR IGNORE INTO users (id, email, name, created, modified)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateUserModified :exec
UPDATE users
SET modified = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: BulkDeleteUsers :exec
DELETE FROM users WHERE id = ?;

-- Maintenance and optimization queries

-- name: VacuumUsers :exec
VACUUM;

-- name: AnalyzeUsers :exec
ANALYZE users;

-- name: GetTableInfo :many
PRAGMA table_info(users);

-- name: GetIndexList :many
PRAGMA index_list(users);

-- name: GetQueryPlan :many
EXPLAIN QUERY PLAN SELECT * FROM users WHERE email = ?;
