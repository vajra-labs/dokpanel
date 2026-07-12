-- name: CreateJwtToken :one
INSERT INTO jwt_tokens (jti, role, user_id, is_blacklist, expired_at)
VALUES (?, ?, ?, 0, ?)
RETURNING *;

-- name: GetJwtTokenByJti :one
SELECT * FROM jwt_tokens WHERE jti = ? LIMIT 1;

-- name: GetJwtTokenByJtiAndBlacklist :one
SELECT * FROM jwt_tokens WHERE jti = ? AND is_blacklist = ? LIMIT 1;

-- name: UpdateJwtTokenByJti :exec
UPDATE jwt_tokens
SET is_blacklist = ?, blacklist_at = ?
WHERE jti = ?;

-- name: UpdateJwtTokensByUserID :exec
UPDATE jwt_tokens
SET is_blacklist = ?, blacklist_at = ?
WHERE user_id = ?;
