-- name: GetUserByID :one
SELECT * FROM users WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: IsOwnerPresent :one
SELECT COUNT(*) FROM users WHERE role = 'OWNER' LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (email, first_name, last_name, avatar, role, password, is_registered, group_id)
VALUES (?, ?, ?, ?, ?, ?, 1, ?)
RETURNING *;
