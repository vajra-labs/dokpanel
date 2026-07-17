-- name: GetSSHKeyByID :one
SELECT * FROM ssh_keys WHERE id = ? LIMIT 1;

-- name: ListSSHKeys :many
SELECT * FROM ssh_keys ORDER BY created_at DESC;

-- name: CreateSSHKey :one
INSERT INTO ssh_keys (
	name, description, private_key, public_key, last_used_at
)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateSSHKey :one
UPDATE ssh_keys
SET
	name = ?,
	description = ?,
	private_key = ?,
	public_key = ?,
	last_used_at = ?
WHERE id = ?
RETURNING *;

-- name: DeleteSSHKey :one
DELETE FROM ssh_keys WHERE id = ? RETURNING *;
