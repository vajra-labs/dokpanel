-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
	id text PRIMARY KEY,
	email text UNIQUE,
	last_name text,
	first_name text,
	avatar text NOT NULL,
	role text DEFAULT 'OWNER',
	about_me text,
	password text NOT NULL,
	is_email_verify integer DEFAULT 0,
	email_verify_at integer, -- Timestamp in Number format
	two_factor_enable integer DEFAULT 0,
	is_registered integer DEFAULT 0 NOT NULL,
	added_by text DEFAULT NULL,
	group_id text NOT NULL,
	created_at integer DEFAULT (strftime('%s','now')) NOT NULL,
	updated_at integer DEFAULT (strftime('%s','now')) NOT NULL,
	CONSTRAINT fk_users_added_by_users_id FOREIGN KEY (added_by) REFERENCES users(id),
	CONSTRAINT fk_users_group_id_groups_id FOREIGN KEY (group_id) REFERENCES groups(id),
	CONSTRAINT role_check CHECK(role IN ('OWNER','ADMIN','MEMBER'))
) STRICT;

CREATE TABLE IF NOT EXISTS two_factor (
	id text PRIMARY KEY,
	secret text NOT NULL,
	backup_codes text NOT NULL,
	user_id text NOT NULL,
	CONSTRAINT fk_two_factor_user_id_users_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) STRICT;

CREATE TABLE IF NOT EXISTS jwt_tokens (
	id text PRIMARY KEY,
	jti text NOT NULL,
	role text NOT NULL,
	user_id text NOT NULL,
	is_blacklist integer DEFAULT 0,
	blacklist_at integer, -- Timestamp in Number format
	expired_at integer, -- Timestamp in Number format
	created_at integer DEFAULT (strftime('%s','now')) NOT NULL,
	updated_at integer DEFAULT (strftime('%s','now')) NOT NULL,
	CONSTRAINT fk_jwt_tokens_user_id_users_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	CONSTRAINT role_check CHECK(role IN ('OWNER','ADMIN','MEMBER'))
) STRICT;

CREATE TABLE IF NOT EXISTS activity_logs (
  id text PRIMARY KEY,
  user_id text NOT NULL,
  --'LOGIN','LOGOUT', ...etc
  activity text NOT NULL,
  source text NOT NULL,
  client_ip text NOT NULL,
  created_at integer DEFAULT (strftime('%s','now')) NOT NULL,
  CONSTRAINT fk_activity_logs_user_id_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT activity_check CHECK(activity IN ('LOGIN','LOGOUT','REGISTER'))
) STRICT;

-- Trigger function
CREATE TRIGGER IF NOT EXISTS update_users_updated_at
AFTER UPDATE ON users
BEGIN
  UPDATE users SET updated_at = strftime('%s','now') WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_jwt_tokens_updated_at
AFTER UPDATE ON jwt_tokens
BEGIN
  UPDATE jwt_tokens SET updated_at = strftime('%s','now') WHERE id = NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS update_jwt_tokens_updated_at;
DROP TRIGGER IF EXISTS update_users_updated_at;
DROP TABLE IF EXISTS jwt_tokens;
DROP TABLE IF EXISTS two_factor;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS activity_logs;
