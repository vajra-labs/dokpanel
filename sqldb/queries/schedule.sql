-- name: CreateSchedule :one
INSERT INTO schedules (
	name, description, cron_expression, app_name, service_name,
	shell_type, schedule_type, command, script, timezone, enabled,
	application_id, compose_id, server_id, organization_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetScheduleByID :one
SELECT * FROM schedules WHERE id = ?;

-- name: GetScheduleWithRelations :one
SELECT
	s.*,
	a.app_name AS application_app_name,
	a.server_id AS application_server_id,
	c.app_name AS compose_app_name,
	c.server_id AS compose_server_id,
	srv.name AS server_name
FROM schedules s
LEFT JOIN applications a ON s.application_id = a.id
LEFT JOIN compose_projects c ON s.compose_id = c.id
LEFT JOIN servers srv ON s.server_id = srv.id
WHERE s.id = ?;

-- name: ListEnabledSchedules :many
SELECT
	s.*,
	a.app_name AS application_app_name,
	a.server_id AS application_server_id,
	c.app_name AS compose_app_name,
	c.server_id AS compose_server_id,
	srv.name AS server_name
FROM schedules s
LEFT JOIN applications a ON s.application_id = a.id
LEFT JOIN compose_projects c ON s.compose_id = c.id
LEFT JOIN servers srv ON s.server_id = srv.id
WHERE s.enabled = 1;

-- name: UpdateSchedule :one
UPDATE schedules
SET
	name = ?,
	description = ?,
	cron_expression = ?,
	service_name = ?,
	shell_type = ?,
	schedule_type = ?,
	command = ?,
	script = ?,
	timezone = ?,
	enabled = ?,
	application_id = ?,
	compose_id = ?,
	server_id = ?,
	organization_id = ?,
	updated_at = (strftime('%s', 'now'))
WHERE id = ?
RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM schedules WHERE id = ?;