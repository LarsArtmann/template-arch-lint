-- Users CRUD queries
-- Matches the existing UserRepository interface methods

-- name: SaveUser :exec
INSERT OR REPLACE INTO users (id, email, name, created, modified)
VALUES (?, ?, ?, ?, ?);

-- name: FindUserByID :one
SELECT id, email, name, created, modified
FROM users
WHERE id = ?;

-- name: FindUserByEmail :one
SELECT id, email, name, created, modified
FROM users
WHERE email = ?;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = ?;

-- name: ListUsers :many
SELECT id, email, name, created, modified
FROM users
ORDER BY created ASC;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;