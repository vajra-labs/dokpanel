-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS groups (
	id text PRIMARY KEY,
  name text NOT NULL UNIQUE,
	created_at integer DEFAULT (strftime('%s','now')) NOT NULL,
	updated_at integer DEFAULT (strftime('%s','now')) NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS policy (
	id text PRIMARY KEY,
	action text NOT NULL UNIQUE,
	created_at integer DEFAULT (strftime('%s','now')) NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS group_policy (
	id text PRIMARY KEY,
	group_id text NOT NULL,
	policy_id text NOT NULL,
	created_at integer DEFAULT (strftime('%s','now')) NOT NULL,
  CONSTRAINT fk_group_policy_group_id_groups_id FOREIGN KEY (group_id) REFERENCES groups(id),
  CONSTRAINT fk_group_policy_policy_id_policy_id FOREIGN KEY (policy_id) REFERENCES policy(id)
) STRICT;

-- Trigger function
CREATE TRIGGER IF NOT EXISTS update_groups_updated_at
AFTER UPDATE ON groups
BEGIN
  UPDATE groups SET updated_at = strftime('%s','now') WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS update_groups_updated_at;
DROP TABLE IF EXISTS group_policy;
DROP TABLE IF EXISTS policy;
DROP TABLE IF EXISTS groups;
