CREATE TABLE organization (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	logo TEXT,
	slug TEXT NOT NULL UNIQUE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TABLE organization_members (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	group_id INTEGER NOT NULL REFERENCES groups(id),
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s','now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s','now')) NOT NULL
) STRICT;

CREATE TABLE organization_invites (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL,
	status TEXT DEFAULT 'PENDING',
	token TEXT NOT NULL UNIQUE,
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	invited_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expired_at INTEGER NOT NULL,
	created_at INTEGER DEFAULT (strftime('%s','now')) NOT NULL,
	CONSTRAINT status_check CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED'))
) STRICT;

-- Trigger Function
CREATE TRIGGER organization_updated_at
AFTER UPDATE ON organization
FOR EACH ROW
BEGIN
	UPDATE organization
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;

CREATE TRIGGER organization_members_updated_at
AFTER UPDATE ON organization_members
FOR EACH ROW
BEGIN
	UPDATE organization_members
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;