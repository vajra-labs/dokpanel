-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS groups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	-- Unique group name (e.g. 'admin', 'devops')
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS policy (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	action TEXT NOT NULL UNIQUE,
	-- Unique action name (e.g. 'read:containers', 'write:users')
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS group_policy (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	group_id INTEGER NOT NULL REFERENCES groups(id),
	policy_id INTEGER NOT NULL REFERENCES policy(id),
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

-- Update trigger function
CREATE TRIGGER IF NOT EXISTS groups_updated_at
	AFTER UPDATE ON groups BEGIN
		UPDATE groups
		SET updated_at = strftime('%s', 'now')
		WHERE id = OLD.id;
	END;

-- +goose StatementEnd
-- +goose Down
DROP TRIGGER IF EXISTS groups_updated_at;

DROP TABLE IF EXISTS group_policy;
DROP TABLE IF EXISTS policy;
DROP TABLE IF EXISTS groups;