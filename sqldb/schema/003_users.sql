CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT UNIQUE,
	last_name TEXT,
	first_name TEXT,
	avatar TEXT NOT NULL,
	is_owner INTEGER DEFAULT 0,
	about_me TEXT,
	password TEXT NOT NULL,
	is_email_verify INTEGER DEFAULT 0,
	email_verify_at INTEGER,
	two_factor_enable INTEGER DEFAULT 0,
	is_registered INTEGER DEFAULT 0 NOT NULL,
	added_by INTEGER DEFAULT NULL REFERENCES users(id),
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TABLE user_policy (
	id         INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	org_id     INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	policy_id  INTEGER NOT NULL REFERENCES policy(id) ON DELETE CASCADE,
	-- grant = extra permission add, deny = permission remove
	effect     TEXT NOT NULL DEFAULT 'GRANT',
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT effect_check CHECK (effect IN ('GRANT', 'DENY')),
	UNIQUE(user_id, org_id, policy_id)
) STRICT;

CREATE TABLE resource_access (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
	org_id  INTEGER REFERENCES organization(id) ON DELETE CASCADE,
	resource_type TEXT,  --  "project", "server", "environment"
	resource_id   INTEGER,
	created_at INTEGER DEFAULT (strftime('%s', 'now')),
	CONSTRAINT resource_type_check CHECK (
		resource_type IN ('PROJECT', 'SERVER', 'ENVIRONMENT', 'SERVICE', 'GIT_PROVIDER')
	)
) STRICT;

CREATE TABLE two_factor (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	secret TEXT NOT NULL,
	backup_codes TEXT NOT NULL,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
) STRICT;

CREATE TABLE jwt_tokens (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	jti TEXT NOT NULL,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	is_blacklist INTEGER DEFAULT 0,
	blacklist_at INTEGER,
	expired_at INTEGER,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

-- Only one owner allowed across all users
CREATE UNIQUE INDEX idx_single_owner ON users(is_owner) WHERE is_owner = 1;

-- Trigger Function
CREATE TRIGGER users_updated_at
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
	UPDATE users
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;

CREATE TRIGGER jwt_tokens_updated_at
AFTER UPDATE ON jwt_tokens
FOR EACH ROW
BEGIN
	UPDATE jwt_tokens
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;