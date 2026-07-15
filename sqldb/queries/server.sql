-- name: GetServerSSHCredentials :one
SELECT
	sr.ip_address,
	sr.username,
	sr.port,
	sh.public_key,
	sh.private_key
FROM servers sr
LEFT JOIN ssh_keys sh ON sr.ssh_key_id = sh.id
WHERE sr.id = ?;
