# Database Schema

Goploy is a self-hosted, developer-friendly panel designed to deploy, monitor, and manage Docker Swarm applications, databases, and Docker Compose stacks. The database utilizes SQLite with `STRICT` table definitions to enforce strict type-safety, inline foreign key constraints for cascading deletes, custom index paths, and auto-updating modification triggers.

---

### 1 groups Table

Stores user permission groups for access control.

```sql
CREATE TABLE groups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	-- is_system: 1 = ADMIN, MEMBER (locked, cannot edit/delete)
	-- is_system: 0 = custom groups (owner can edit/delete)
	is_system INTEGER NOT NULL DEFAULT 0,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER groups_updated_at
AFTER UPDATE ON groups
FOR EACH ROW
BEGIN
	UPDATE groups
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 2 policy Table

Defines system access permission policy rules.

```sql
CREATE TABLE policy (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	action TEXT NOT NULL UNIQUE,
	-- Unique action name (e.g. 'read:containers', 'write:users')
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;
```

### 3 group_policy Table

Links policy rules to permission groups (many-to-many relationship).

```sql
CREATE TABLE group_policy (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	group_id INTEGER NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
	policy_id INTEGER NOT NULL REFERENCES policy(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;
```

### 4 users Table

Stores account metadata, password credentials, role permissions, and group associations.

```sql
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

CREATE UNIQUE INDEX idx_single_owner ON users(is_owner) WHERE is_owner = 1;

CREATE TRIGGER users_updated_at
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
	UPDATE users
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 5 user_policy Table

Defines user policy rules (GRANT or DENY overrides) within a specific organization.

```sql
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
```

### 6 resource_access Table

Defines fine-grained user access permissions for specific resources (e.g., projects, servers, environments) within an organization.

```sql
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
```

### 7 two_factor Table

Stores TOTP secret keys and backup recovery codes for user account multi-factor authentication.

```sql
CREATE TABLE two_factor (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	secret TEXT NOT NULL,
	backup_codes TEXT NOT NULL,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
) STRICT;
```

### 8 jwt_tokens Table

Tracks active JWT tokens for user session blacklists and expirations.

```sql
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

CREATE TRIGGER jwt_tokens_updated_at
AFTER UPDATE ON jwt_tokens
FOR EACH ROW
BEGIN
	UPDATE jwt_tokens
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 9 organization Table

Stores tenant organizations hosting resources.

```sql
CREATE TABLE organization (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	logo TEXT,
	slug TEXT NOT NULL UNIQUE,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER organization_updated_at
AFTER UPDATE ON organization
FOR EACH ROW
BEGIN
	UPDATE organization
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 10 organization_members Table

Links users to organizations with specific memberships (many-to-many relationship).

```sql
CREATE TABLE organization_members (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	group_id INTEGER NOT NULL REFERENCES groups(id),
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s','now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s','now')) NOT NULL
) STRICT;

CREATE TRIGGER organization_members_updated_at
AFTER UPDATE ON organization_members
FOR EACH ROW
BEGIN
	UPDATE organization_members
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 11 organization_invites Table

Stores pending/accepted invites sent to new organization members.

```sql
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
```

### 12 projects Table

Stores target projects encapsulating environments and services.

```sql
CREATE TABLE projects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	-- Unique slug used in Docker service names (e.g. 'my-app')
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	env_var TEXT NOT NULL DEFAULT '',
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER projects_updated_at
AFTER UPDATE ON projects
FOR EACH ROW
BEGIN
	UPDATE projects
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 13 tags Table

Stores project label tags.

```sql
CREATE TABLE tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	color TEXT NOT NULL,
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	UNIQUE(name, organization_id)
) STRICT;
```

### 14 project_tags Table

Maps tags to projects (many-to-many relationship).

```sql
CREATE TABLE project_tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
	UNIQUE(project_id, tag_id)
) STRICT;
```

### 15 postgres_dbs Table

Defines managed PostgreSQL database services, resources, and replica settings.

```sql
CREATE TABLE postgres_dbs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	app_name TEXT NOT NULL UNIQUE,
	description TEXT,
	docker_image TEXT NOT NULL DEFAULT 'postgres:18',
	database_name TEXT NOT NULL,
	database_user TEXT NOT NULL,
	database_password TEXT NOT NULL,
	external_port INTEGER,
	command TEXT,
	args TEXT,	-- JSON array
	env_var TEXT,
	memory_reservation TEXT,
	memory_limit TEXT,
	cpu_reservation TEXT,
	cpu_limit TEXT,
	replicas INTEGER NOT NULL DEFAULT 1,
	app_status TEXT NOT NULL DEFAULT 'IDLE',
	-- Swarm configs (JSON)
	health_check_swarm TEXT,
	restart_policy_swarm TEXT,
	placement_swarm TEXT,
	update_config_swarm TEXT,
	rollback_config_swarm TEXT,
	mode_swarm TEXT,
	labels_swarm TEXT,
	network_swarm TEXT,
	endpoint_spec_swarm TEXT,
	ulimits_swarm TEXT,
	stop_grace_period_swarm INTEGER,
	environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT pg_status_check CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;

CREATE TRIGGER pg_updated_at
AFTER UPDATE ON postgres_dbs
FOR EACH ROW
BEGIN
	UPDATE postgres_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 16 mysql_dbs Table

Defines managed MySQL database services, including root password credentials.

```sql
CREATE TABLE mysql_dbs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	app_name TEXT NOT NULL UNIQUE,
	description TEXT,
	docker_image TEXT NOT NULL DEFAULT 'mysql:9',
	database_name TEXT NOT NULL,
	database_user TEXT NOT NULL,
	database_password TEXT NOT NULL,
	database_root_password TEXT NOT NULL,
	external_port INTEGER,
	command TEXT,
	args TEXT,	-- JSON array
	env_var TEXT,
	memory_reservation TEXT,
	memory_limit TEXT,
	cpu_reservation TEXT,
	cpu_limit TEXT,
	replicas INTEGER NOT NULL DEFAULT 1,
	app_status TEXT NOT NULL DEFAULT 'IDLE',
	health_check_swarm TEXT,
	restart_policy_swarm TEXT,
	placement_swarm TEXT,
	update_config_swarm TEXT,
	rollback_config_swarm TEXT,
	mode_swarm TEXT,
	labels_swarm TEXT,
	network_swarm TEXT,
	endpoint_spec_swarm TEXT,
	ulimits_swarm TEXT,
	stop_grace_period_swarm INTEGER,
	environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT mysql_status_check CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;

CREATE TRIGGER mysql_updated_at
AFTER UPDATE ON mysql_dbs
FOR EACH ROW
BEGIN
	UPDATE mysql_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 17 mariadb_dbs Table

Defines managed MariaDB database services.

```sql
CREATE TABLE mariadb_dbs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	app_name TEXT NOT NULL UNIQUE,
	description TEXT,
	docker_image TEXT NOT NULL DEFAULT 'mariadb:13',
	database_name TEXT NOT NULL,
	database_user TEXT NOT NULL,
	database_password TEXT NOT NULL,
	database_root_password TEXT NOT NULL,
	external_port INTEGER,
	command TEXT,
	args TEXT,	-- JSON array
	env_var TEXT,
	memory_reservation TEXT,
	memory_limit TEXT,
	cpu_reservation TEXT,
	cpu_limit TEXT,
	replicas INTEGER NOT NULL DEFAULT 1,
	app_status TEXT NOT NULL DEFAULT 'IDLE',
	health_check_swarm TEXT,
	restart_policy_swarm TEXT,
	placement_swarm TEXT,
	update_config_swarm TEXT,
	rollback_config_swarm TEXT,
	mode_swarm TEXT,
	labels_swarm TEXT,
	network_swarm TEXT,
	endpoint_spec_swarm TEXT,
	ulimits_swarm TEXT,
	stop_grace_period_swarm INTEGER,
	environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT mariadb_status_check CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;

CREATE TRIGGER mariadb_updated_at
AFTER UPDATE ON mariadb_dbs
FOR EACH ROW
BEGIN
	UPDATE mariadb_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 18 mongo_dbs Table

Defines managed MongoDB database services, including replica sets.

```sql
CREATE TABLE mongo_dbs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	app_name TEXT NOT NULL UNIQUE,
	description TEXT,
	docker_image TEXT NOT NULL DEFAULT 'mongo:8',
	database_user TEXT NOT NULL,
	database_password TEXT NOT NULL,
	external_port INTEGER,
	replica_sets INTEGER NOT NULL DEFAULT 0,
	command TEXT,
	args TEXT,	-- JSON array
	env_var TEXT,
	memory_reservation TEXT,
	memory_limit TEXT,
	cpu_reservation TEXT,
	cpu_limit TEXT,
	replicas INTEGER NOT NULL DEFAULT 1,
	app_status TEXT NOT NULL DEFAULT 'IDLE',
	health_check_swarm TEXT,
	restart_policy_swarm TEXT,
	placement_swarm TEXT,
	update_config_swarm TEXT,
	rollback_config_swarm TEXT,
	mode_swarm TEXT,
	labels_swarm TEXT,
	network_swarm TEXT,
	endpoint_spec_swarm TEXT,
	ulimits_swarm TEXT,
	stop_grace_period_warm INTEGER,
	environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT mongo_status_check CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;

CREATE TRIGGER mongo_updated_at
AFTER UPDATE ON mongo_dbs
FOR EACH ROW
BEGIN
	UPDATE mongo_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 19 redis_dbs Table

Defines managed Redis database cache services.

```sql
CREATE TABLE redis_dbs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	app_name TEXT NOT NULL UNIQUE,
	description TEXT,
	docker_image TEXT NOT NULL DEFAULT 'redis:8',
	database_password TEXT NOT NULL,
	external_port INTEGER,
	command TEXT,
	args TEXT,	-- JSON array
	env_var TEXT,
	memory_reservation TEXT,
	memory_limit TEXT,
	cpu_reservation TEXT,
	cpu_limit TEXT,
	replicas INTEGER NOT NULL DEFAULT 1,
	app_status TEXT NOT NULL DEFAULT 'IDLE',
	health_check_swarm TEXT,
	restart_policy_swarm TEXT,
	placement_swarm TEXT,
	update_config_swarm TEXT,
	rollback_config_swarm TEXT,
	mode_swarm TEXT,
	labels_swarm TEXT,
	network_swarm TEXT,
	endpoint_spec_swarm TEXT,
	ulimits_swarm TEXT,
	stop_grace_period_swarm INTEGER,
	environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT redis_status_check CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;

CREATE TRIGGER redis_updated_at
AFTER UPDATE ON redis_dbs
FOR EACH ROW
BEGIN
	UPDATE redis_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 20 libsql_dbs Table

Defines managed LibSQL (Turso core engine) instances, supporting replica nodes.

```sql
CREATE TABLE libsql_dbs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	app_name TEXT NOT NULL UNIQUE,
	description TEXT,
	docker_image TEXT NOT NULL DEFAULT 'ghcr.io/tursodatabase/libsql-server:latest',
	database_user TEXT NOT NULL,
	database_password TEXT NOT NULL,
	sqld_node TEXT NOT NULL DEFAULT 'PRIMARY', -- primary | replica
	sqld_primary_url TEXT,
	enable_namespaces INTEGER NOT NULL DEFAULT 0,
	external_port INTEGER,
	external_grpc_port INTEGER,
	external_admin_port INTEGER,
	command TEXT,
	args TEXT,	-- JSON array
	env_var TEXT,
	memory_reservation TEXT,
	memory_limit TEXT,
	cpu_reservation TEXT,
	cpu_limit TEXT,
	replicas INTEGER NOT NULL DEFAULT 1,
	app_status TEXT NOT NULL DEFAULT 'IDLE',
	-- Swarm configs (JSON)
	health_check_swarm TEXT,
	restart_policy_swarm TEXT,
	placement_swarm TEXT,
	update_config_swarm TEXT,
	rollback_config_swarm TEXT,
	mode_swarm TEXT,
	labels_swarm TEXT,
	network_swarm TEXT,
	endpoint_spec_swarm TEXT,
	ulimits_swarm TEXT,
	stop_grace_period_swarm INTEGER,
	environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT libsql_status_check CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR')),
	CONSTRAINT libsql_node_check CHECK (sqld_node IN ('PRIMARY', 'REPLICA'))
) STRICT;

CREATE TRIGGER libsql_updated_at
AFTER UPDATE ON libsql_dbs
FOR EACH ROW
BEGIN
	UPDATE libsql_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 21 server_metrics Table

Stores time-series CPU, RAM, and Disk metric performance data for the host machine.

```sql
CREATE TABLE server_metrics (
	timestamp INTEGER PRIMARY KEY,
	cpu REAL NOT NULL,
	cpu_model TEXT NOT NULL,
	cpu_cores INTEGER NOT NULL,
	cpu_physical_cores INTEGER NOT NULL,
	cpu_speed REAL NOT NULL,
	os TEXT NOT NULL,
	distro TEXT NOT NULL,
	kernel TEXT NOT NULL,
	arch TEXT NOT NULL,
	mem_used REAL NOT NULL,
	mem_used_gb REAL NOT NULL,
	mem_total REAL NOT NULL,
	uptime INTEGER NOT NULL,
	disk_used REAL NOT NULL,
	total_disk REAL NOT NULL,
	network_in REAL NOT NULL,
	network_out REAL NOT NULL
) STRICT;
```

### 22 container_metrics Table

Stores time-series Docker stats for individual application and database containers.

```sql
CREATE TABLE container_metrics (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	timestamp INTEGER NOT NULL,
	container_id TEXT NOT NULL,
	container_name TEXT NOT NULL,
	metrics_json TEXT NOT NULL
) STRICT;

CREATE INDEX idx_container_metrics_timestamp ON container_metrics(timestamp);

CREATE INDEX idx_container_metrics_name ON container_metrics(container_name);
```

### 23 ssh_keys Table

Stores securely encrypted SSH Private Keys used to provision remote build/deploy servers.

```sql
CREATE TABLE ssh_keys (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	description TEXT,
	private_key TEXT NOT NULL DEFAULT '',
	public_key TEXT NOT NULL,
	last_used_at INTEGER,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER ssh_keys_updated_at
AFTER UPDATE ON ssh_keys
FOR EACH ROW
BEGIN
	UPDATE ssh_keys
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 24 registries Table

Configures container registry access credentials (e.g. Docker Hub, GHCR).

```sql
CREATE TABLE registries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	registry_name TEXT NOT NULL,
	image_prefix TEXT,
	username TEXT NOT NULL,
	password TEXT NOT NULL,
	registry_url TEXT NOT NULL DEFAULT '',
	-- registry_type: cloud | selfHosted
	registry_type TEXT NOT NULL DEFAULT 'CLOUD',
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT registry_type_check CHECK (registry_type IN ('CLOUD', 'SELF_HOSTED'))
) STRICT;

CREATE TRIGGER registries_updated_at
AFTER UPDATE ON registries
FOR EACH ROW
BEGIN
	UPDATE registries
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 25 environments Table

Stores project sub-environments (e.g. 'production', 'staging').

```sql
CREATE TABLE environments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	description TEXT,
	env_var TEXT NOT NULL DEFAULT '',
	is_default INTEGER NOT NULL DEFAULT 0,
	project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER environments_updated_at
AFTER UPDATE ON environments
FOR EACH ROW
BEGIN
	UPDATE environments
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 26 servers Table

Stores configurations for remote Docker servers.

```sql
CREATE TABLE servers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	description TEXT,
	ip_address TEXT NOT NULL,
	port INTEGER NOT NULL DEFAULT 22,
	username TEXT NOT NULL DEFAULT 'root',
	app_name TEXT NOT NULL UNIQUE,
	-- server_status: active | inactive
	server_status TEXT NOT NULL DEFAULT 'ACTIVE',
	-- server_type: deploy | build
	server_type TEXT NOT NULL DEFAULT 'DEPLOY',
	enable_docker_cleanup INTEGER NOT NULL DEFAULT 0,
	log_cleanup_cron TEXT DEFAULT '0 0 * * *',
	command TEXT NOT NULL DEFAULT '',
	-- JSON: metrics config object { server: {...}, containers: {...} }
	metrics_config TEXT NOT NULL DEFAULT '{}',
	ssh_key_id INTEGER REFERENCES ssh_keys(id) ON DELETE SET NULL,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT server_status_check CHECK (server_status IN ('ACTIVE', 'INACTIVE')),
	CONSTRAINT server_type_check CHECK (server_type IN ('DEPLOY', 'BUILD'))
) STRICT;

CREATE TRIGGER servers_updated_at
AFTER UPDATE ON servers
FOR EACH ROW
BEGIN
	UPDATE servers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 27 git_providers Table

Polymorphic umbrella table linking credentials for Git platforms (GitHub, GitLab, etc.).

```sql
CREATE TABLE git_providers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	-- provider_type: github | gitlab | gitea | bitbucket
	provider_type TEXT NOT NULL DEFAULT 'GITHUB',
	-- Share provider across all users (single-tenant: always true)
	shared INTEGER NOT NULL DEFAULT 1,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT git_provider_type_check CHECK (
		provider_type IN ('GITHUB', 'GITLAB', 'GITEA', 'BITBUCKET')
	)
) STRICT;

CREATE TRIGGER git_providers_updated_at
AFTER UPDATE ON git_providers
FOR EACH ROW
BEGIN
	UPDATE git_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 28 github_providers Table

Stores GitHub App Client IDs, Installation IDs, and Private Keys.

```sql
CREATE TABLE github_providers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	github_app_name TEXT,
	github_app_id INTEGER,
	github_client_id TEXT,
	github_client_secret TEXT,
	github_installation_id TEXT,
	github_private_key TEXT,
	github_webhook_secret TEXT,
	git_provider_id INTEGER NOT NULL REFERENCES git_providers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER github_providers_updated_at
AFTER UPDATE ON github_providers
FOR EACH ROW
BEGIN
	UPDATE github_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 29 gitlab_providers Table

Stores GitLab OAuth2 tokens and application IDs.

```sql
CREATE TABLE gitlab_providers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	gitlab_url TEXT NOT NULL DEFAULT 'https://gitlab.com',
	gitlab_internal_url TEXT,
	application_id TEXT,
	redirect_uri TEXT,
	secret TEXT,
	access_token TEXT,
	refresh_token TEXT,
	group_name TEXT,
	expires_at INTEGER,
	git_provider_id INTEGER NOT NULL REFERENCES git_providers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER gitlab_providers_updated_at
AFTER UPDATE ON gitlab_providers
FOR EACH ROW
BEGIN
	UPDATE gitlab_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 30 gitea_providers Table

Stores Gitea integration access details.

```sql
CREATE TABLE gitea_providers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	gitea_url TEXT NOT NULL DEFAULT 'https://gitea.com',
	gitea_internal_url TEXT,
	redirect_uri TEXT,
	client_id TEXT,
	client_secret TEXT,
	access_token TEXT,
	refresh_token TEXT,
	expires_at INTEGER,
	scopes TEXT DEFAULT 'repo,repo:status,read:user,read:org',
	last_authenticated_at INTEGER,
	git_provider_id INTEGER NOT NULL REFERENCES git_providers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER gitea_providers_updated_at
AFTER UPDATE ON gitea_providers
FOR EACH ROW
BEGIN
	UPDATE gitea_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 31 bitbucket_providers Table

Stores Bitbucket App password keys.

```sql
CREATE TABLE bitbucket_providers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	bitbucket_username TEXT,
	bitbucket_email TEXT,
	app_password TEXT,
	api_token TEXT,
	bitbucket_workspace_name TEXT,
	git_provider_id INTEGER NOT NULL REFERENCES git_providers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE TRIGGER bitbucket_providers_updated_at
AFTER UPDATE ON bitbucket_providers
FOR EACH ROW
BEGIN
	UPDATE bitbucket_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 32 applications Table

Configures applications, buildpacks (Nixpacks, Dockerfile), Git sources, environment variables, resource limits, and preview settings.

```sql
CREATE TABLE applications (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	-- Unique slug used as Docker service name (e.g. 'my-app-x7k2')
	app_name TEXT NOT NULL UNIQUE,
	description TEXT,
	-- source_type: DOCKER | GIT | GITHUB | GITLAB | BITBUCKET | GITEA | DROP
	source_type TEXT NOT NULL DEFAULT 'GITHUB',
	-- build_type: DOCKERFILE | HEROKU_BUILDPACKS | PAKETO_BUILDPACKS | NIXPACKS | STATIC | RAILPACK
	build_type TEXT NOT NULL DEFAULT 'NIXPACKS',
	-- app_status: IDLE | RUNNING | DONE | ERROR
	app_status TEXT NOT NULL DEFAULT 'IDLE',
	-- trigger_type: PUSH | TAG
	trigger_type TEXT NOT NULL DEFAULT 'PUSH',
	-- Build config
	build_args TEXT,
	build_secrets TEXT,
	dockerfile TEXT DEFAULT 'Dockerfile',
	docker_context_path TEXT,
	docker_build_stage TEXT,
	publish_directory TEXT,
	is_static_spa INTEGER DEFAULT 0,
	create_env_file INTEGER NOT NULL DEFAULT 1,
	railpack_version TEXT DEFAULT '0.15.4',
	heroku_version TEXT DEFAULT '24',
	command TEXT,
	args TEXT, -- JSON array of args e.g. '["--port","3000"]'
	env_var TEXT,
	build_path TEXT DEFAULT '/',
	clean_cache INTEGER NOT NULL DEFAULT 0,
	drop_build_path TEXT,
	enable_submodules INTEGER NOT NULL DEFAULT 0,
	watch_paths TEXT, -- JSON array of paths to watch for auto-deploy
	refresh_token TEXT,
	icon TEXT,
	-- Resource limits
	memory_reservation TEXT,
	memory_limit TEXT,
	cpu_reservation TEXT,
	cpu_limit TEXT,
	replicas INTEGER NOT NULL DEFAULT 1,
	-- Docker Swarm JSON configs (stored as JSON text)
	health_check_swarm TEXT,
	restart_policy_swarm TEXT,
	placement_swarm TEXT,
	update_config_swarm TEXT,
	rollback_config_swarm TEXT,
	mode_swarm TEXT,
	labels_swarm TEXT,
	network_swarm TEXT,
	endpoint_spec_swarm TEXT,
	ulimits_swarm TEXT,
	stop_grace_period_swarm INTEGER,
	-- GitHub source
	repository TEXT,
	owner TEXT,
	branch TEXT,
	auto_deploy INTEGER DEFAULT 1,
	-- GitLab source
	gitlab_project_id INTEGER,
	gitlab_repository TEXT,
	gitlab_owner TEXT,
	gitlab_branch TEXT,
	gitlab_build_path TEXT DEFAULT '/',
	gitlab_path_namespace TEXT,
	-- Gitea source
	gitea_repository TEXT,
	gitea_owner TEXT,
	gitea_branch TEXT,
	gitea_build_path TEXT DEFAULT '/',
	-- Bitbucket source
	bitbucket_repository TEXT,
	bitbucket_repository_slug TEXT,
	bitbucket_owner TEXT,
	bitbucket_branch TEXT,
	bitbucket_build_path TEXT DEFAULT '/',
	-- Docker image source
	docker_image TEXT,
	docker_username TEXT,
	docker_password TEXT,
	registry_url TEXT,
	-- Custom Git (SSH) source
	custom_git_url TEXT,
	custom_git_branch TEXT,
	custom_git_build_path TEXT,
	custom_git_ssh_key_id INTEGER REFERENCES ssh_keys(id) ON DELETE SET NULL,
	-- Preview deployments
	preview_env TEXT,
	preview_build_args TEXT,
	preview_build_secrets TEXT,
	preview_labels TEXT, -- JSON array of preview labels
	preview_wildcard TEXT,
	preview_port INTEGER DEFAULT 3000,
	preview_https INTEGER NOT NULL DEFAULT 0,
	preview_path TEXT DEFAULT '/',
	-- preview_certificate_type: LETSENCRYPT | NONE | CUSTOM
	preview_certificate_type TEXT NOT NULL DEFAULT 'NONE',
	preview_custom_cert_resolver TEXT,
	preview_limit INTEGER DEFAULT 3,
	is_preview_deployments_active INTEGER NOT NULL DEFAULT 0,
	preview_require_collaborator_permissions INTEGER NOT NULL DEFAULT 1,
	rollback_active INTEGER NOT NULL DEFAULT 0,
	-- Foreign keys (Inline References)
	environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	build_server_id INTEGER REFERENCES servers(id) ON DELETE SET NULL,
	registry_id INTEGER REFERENCES registries(id) ON DELETE SET NULL,
	rollback_registry_id INTEGER REFERENCES registries(id) ON DELETE SET NULL,
	build_registry_id INTEGER REFERENCES registries(id) ON DELETE SET NULL,
	github_provider_id INTEGER REFERENCES github_providers(id) ON DELETE SET NULL,
	gitlab_provider_id INTEGER REFERENCES gitlab_providers(id) ON DELETE SET NULL,
	gitea_provider_id INTEGER REFERENCES gitea_providers(id) ON DELETE SET NULL,
	bitbucket_provider_id INTEGER REFERENCES bitbucket_providers(id) ON DELETE SET NULL,
	-- Timestamps
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	-- Constraints
	CONSTRAINT app_source_type_check CHECK (source_type IN ('DOCKER', 'GIT', 'GITHUB', 'GITLAB', 'BITBUCKET', 'GITEA', 'DROP')),
	CONSTRAINT app_build_type_check CHECK (build_type IN ('DOCKERFILE', 'HEROKU_BUILDPACKS', 'PAKETO_BUILDPACKS', 'NIXPACKS', 'STATIC', 'RAILPACK')),
	CONSTRAINT app_status_check CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR')),
	CONSTRAINT app_trigger_type_check CHECK (trigger_type IN ('PUSH', 'TAG')),
	CONSTRAINT app_preview_cert_check CHECK (preview_certificate_type IN ('LETSENCRYPT', 'NONE', 'CUSTOM'))
) STRICT;

CREATE INDEX idx_applications_environment_id ON applications(environment_id);

CREATE INDEX idx_applications_server_id ON applications(server_id);

CREATE INDEX idx_applications_app_status ON applications(app_status);

CREATE TRIGGER applications_updated_at
AFTER UPDATE ON applications
FOR EACH ROW
BEGIN
	UPDATE applications
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 33 compose_projects Table

Stores configuration settings for multi-container Docker Compose stacks.

```sql
CREATE TABLE compose_projects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	-- Unique slug used as Docker stack name
	app_name TEXT NOT NULL UNIQUE,
	description TEXT,
	env_var TEXT,
	compose_file TEXT NOT NULL DEFAULT '',
	refresh_token TEXT,
	-- source_type: GIT | GITHUB | GITLAB | BITBUCKET | GITEA | RAW
	source_type TEXT NOT NULL DEFAULT 'GITHUB',
	-- compose_type: DOCKER-COMPOSE | STACK
	compose_type TEXT NOT NULL DEFAULT 'DOCKER-COMPOSE',
	-- compose_status: IDLE | RUNNING | DONE | ERROR
	compose_status TEXT NOT NULL DEFAULT 'IDLE',
	-- trigger_type: PUSH | TAG
	trigger_type TEXT NOT NULL DEFAULT 'PUSH',
	-- Github source
	repository TEXT,
	owner TEXT,
	branch TEXT,
	auto_deploy INTEGER NOT NULL DEFAULT 1,
	-- GitLab source
	gitlab_project_id INTEGER,
	gitlab_repository TEXT,
	gitlab_owner TEXT,
	gitlab_branch TEXT,
	gitlab_path_namespace TEXT,
	-- Bitbucket source
	bitbucket_repository TEXT,
	bitbucket_repository_slug TEXT,
	bitbucket_owner TEXT,
	bitbucket_branch TEXT,
	-- Gitea source
	gitea_repository TEXT,
	gitea_owner TEXT,
	gitea_branch TEXT,
	-- Custom Git source
	custom_git_url TEXT,
	custom_git_branch TEXT,
	custom_git_ssh_key_id INTEGER REFERENCES ssh_keys(id) ON DELETE SET NULL,
	-- Build & run config
	command TEXT NOT NULL DEFAULT '',
	enable_submodules INTEGER NOT NULL DEFAULT 0,
	compose_path TEXT NOT NULL DEFAULT './docker-compose.yml',
	suffix TEXT NOT NULL DEFAULT '',
	randomize INTEGER NOT NULL DEFAULT 0,
	isolated_deployment INTEGER NOT NULL DEFAULT 0,
	isolated_deployments_volume INTEGER NOT NULL DEFAULT 0,
	watch_paths TEXT, -- JSON array of strings
	-- Foreign keys
	environment_id INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	github_provider_id INTEGER REFERENCES github_providers(id) ON DELETE SET NULL,
	gitlab_provider_id INTEGER REFERENCES gitlab_providers(id) ON DELETE SET NULL,
	gitea_provider_id INTEGER REFERENCES gitea_providers(id) ON DELETE SET NULL,
	bitbucket_provider_id INTEGER REFERENCES bitbucket_providers(id) ON DELETE SET NULL,
	-- Timestamps
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	-- Constraints
	CONSTRAINT compose_source_type_check CHECK (source_type IN ('GIT', 'GITHUB', 'GITLAB', 'BITBUCKET', 'GITEA', 'RAW')),
	CONSTRAINT compose_type_check CHECK (compose_type IN ('DOCKER-COMPOSE', 'STACK')),
	CONSTRAINT compose_status_check CHECK (compose_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR')),
	CONSTRAINT compose_trigger_type_check CHECK (trigger_type IN ('PUSH', 'TAG'))
) STRICT;

CREATE INDEX idx_compose_projects_environment_id ON compose_projects(environment_id);

CREATE INDEX idx_compose_projects_server_id ON compose_projects(server_id);

CREATE TRIGGER compose_projects_updated_at
AFTER UPDATE ON compose_projects
FOR EACH ROW
BEGIN
	UPDATE compose_projects
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 34 domains Table

Saves host domains mapped to apps or compose projects, with TLS resolver configurations.

```sql
CREATE TABLE domains (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	host TEXT NOT NULL,
	https INTEGER NOT NULL DEFAULT 0,
	port INTEGER DEFAULT 3000,
	path TEXT DEFAULT '/',
	internal_path TEXT DEFAULT '/',
	custom_entrypoint TEXT,
	service_name TEXT,
	custom_cert_resolver TEXT,
	strip_path INTEGER NOT NULL DEFAULT 0,
	-- JSON array of middleware names e.g. '[redirect-to-https]'
	middlewares TEXT NOT NULL DEFAULT '[]',
	-- domain_type: application | compose | preview
	domain_type TEXT NOT NULL DEFAULT 'APPLICATION',
	-- certificate_type: letsencrypt | none | custom
	certificate_type TEXT NOT NULL DEFAULT 'NONE',
	-- One of these will be set
	application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
	compose_id INTEGER REFERENCES compose_projects(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT domain_cert_type_check CHECK (certificate_type IN ('LETSENCRYPT', 'NONE', 'CUSTOM')),
	CONSTRAINT domain_type_check CHECK (domain_type IN ('APPLICATION', 'COMPOSE', 'PREVIEW'))
) STRICT;

CREATE INDEX idx_domains_application_id ON domains(application_id);

CREATE INDEX idx_domains_compose_id ON domains(compose_id);

CREATE INDEX idx_domains_host ON domains(host);

CREATE TRIGGER domains_updated_at
AFTER UPDATE ON domains
FOR EACH ROW
BEGIN
	UPDATE domains
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 35 patches Table

Saves text files/patches applied to files inside the workspace directory during builds.

```sql
CREATE TABLE patches (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	-- patch_type: CREATE | UPDATE | DELETE
	patch_type TEXT NOT NULL DEFAULT 'UPDATE',
	file_path TEXT NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1,
	content TEXT NOT NULL,
	-- Foreign keys
	application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
	compose_id INTEGER REFERENCES compose_projects(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	-- Constraints
	CONSTRAINT patch_type_check CHECK (patch_type IN ('CREATE', 'UPDATE', 'DELETE')),
	UNIQUE(file_path, application_id),
	UNIQUE(file_path, compose_id)
) STRICT;

CREATE INDEX idx_patches_application_id ON patches(application_id);

CREATE INDEX idx_patches_compose_id ON patches(compose_id);

CREATE TRIGGER patches_updated_at
AFTER UPDATE ON patches
FOR EACH ROW
BEGIN
	UPDATE patches
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 36 deployments Table

Contains the execution logs and status metadata for individual container deployment history.

```sql
CREATE TABLE deployments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT,
	-- status: RUNNING | DONE | ERROR | CANCELLED
	status TEXT NOT NULL DEFAULT 'RUNNING',
	log_path TEXT NOT NULL,
	pid TEXT,
	error_message TEXT,
	is_preview_deployment INTEGER NOT NULL DEFAULT 0,
	started_at INTEGER,
	finished_at INTEGER,
	-- Foreign keys
	application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
	compose_id INTEGER REFERENCES compose_projects(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT deployment_status_check CHECK (status IN ('RUNNING', 'DONE', 'ERROR', 'CANCELLED'))
) STRICT;

CREATE INDEX idx_deployments_status ON deployments(status);

CREATE INDEX idx_deployments_created_at ON deployments(created_at);

CREATE INDEX idx_deployments_compose_id ON deployments(compose_id);

CREATE INDEX idx_deployments_application_id ON deployments(application_id);
```

### 37 rollbacks Table

Saves context snapshots and image tags of past stable builds to allow 1-click manual registry rollback.

```sql
CREATE TABLE rollbacks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	deployment_id INTEGER NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
	version INTEGER NOT NULL DEFAULT 1,
	image TEXT,
	full_context TEXT, -- JSON snapshot of application configs
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE INDEX idx_rollbacks_deployment_id ON rollbacks(deployment_id);
```

### 38 mounts Table

Saves bind directory paths, volume paths, or configuration files mounted to services.

```sql
CREATE TABLE mounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	mount_type TEXT NOT NULL DEFAULT 'VOLUME',
	service_type TEXT NOT NULL DEFAULT 'APPLICATION',
	host_path TEXT,
	volume_name TEXT,
	file_path TEXT,
	content TEXT,
	mount_path TEXT NOT NULL,
	-- Foreign keys
	postgres_id INTEGER REFERENCES postgres_dbs(id) ON DELETE CASCADE,
	mysql_id INTEGER REFERENCES mysql_dbs(id) ON DELETE CASCADE,
	mariadb_id INTEGER REFERENCES mariadb_dbs(id) ON DELETE CASCADE,
	mongo_id INTEGER REFERENCES mongo_dbs(id) ON DELETE CASCADE,
	redis_id INTEGER REFERENCES redis_dbs(id) ON DELETE CASCADE,
	libsql_id INTEGER REFERENCES libsql_dbs(id) ON DELETE CASCADE,
	compose_id INTEGER REFERENCES compose_projects(id) ON DELETE CASCADE,
	application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
	-- Timestamp
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	-- Constraints
	CONSTRAINT mount_type_check CHECK (mount_type IN ('BIND', 'VOLUME', 'FILE')),
	CONSTRAINT mount_service_type_check CHECK (
		service_type IN ('APPLICATION', 'COMPOSE', 'POSTGRES', 'MYSQL', 'MARIADB', 'MONGO', 'REDIS', 'LIBSQL')
	)
) STRICT;

CREATE INDEX idx_mounts_application_id ON mounts(application_id);

CREATE INDEX idx_mounts_compose_id ON mounts(compose_id);

CREATE INDEX idx_mounts_postgres_id ON mounts(postgres_id);

CREATE INDEX idx_mounts_mysql_id ON mounts(mysql_id);

CREATE INDEX idx_mounts_mariadb_id ON mounts(mariadb_id);

CREATE INDEX idx_mounts_mongo_id ON mounts(mongo_id);

CREATE INDEX idx_mounts_redis_id ON mounts(redis_id);

CREATE INDEX idx_mounts_libsql_id ON mounts(libsql_id);

CREATE TRIGGER mounts_updated_at
AFTER UPDATE ON mounts
FOR EACH ROW
BEGIN
	UPDATE mounts
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 39 certificates Table

Saves manually uploaded SSL `.crt` and `.key` files.

```sql
CREATE TABLE certificates (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	certificate_data TEXT NOT NULL,
	private_key TEXT NOT NULL,
	certificate_path TEXT NOT NULL UNIQUE,
	auto_renew INTEGER NOT NULL DEFAULT 0,
	-- Foreign keys
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	-- Timestamp
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT auto_renew_check CHECK (auto_renew IN (0, 1))
) STRICT;

CREATE INDEX idx_certificates_server_id ON certificates(server_id);

CREATE INDEX idx_certificates_organization_id ON certificates(organization_id);

CREATE TRIGGER certificates_updated_at
AFTER UPDATE ON certificates
FOR EACH ROW
BEGIN
	UPDATE certificates
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 40 destinations Table

Saves access credentials and configurations for S3-compatible remote storage backends.

```sql
CREATE TABLE destinations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	-- provider: S3 | R2 | BACKBLAZE | GCS | DO_SPACES
	provider TEXT NOT NULL DEFAULT 'S3',
	access_key TEXT NOT NULL,
	secret_access_key TEXT NOT NULL,
	bucket TEXT NOT NULL,
	region TEXT NOT NULL,
	endpoint TEXT NOT NULL,
	additional_flags TEXT, -- JSON array of strings (e.g. ['--max-depth', '1'])
	-- Foreign keys (Inline References)
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	-- Timestamp
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	-- Constraints
	CONSTRAINT destination_provider_check CHECK (provider IN ('S3', 'R2', 'BACKBLAZE', 'GCS', 'DO_SPACES'))
) STRICT;

CREATE INDEX idx_destinations_organization_id ON destinations(organization_id);

CREATE TRIGGER destinations_updated_at
AFTER UPDATE ON destinations
FOR EACH ROW
BEGIN
	UPDATE destinations
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 41 backups Table

Configures cron-scheduled relational database logical dump backup tasks.

```sql
CREATE TABLE backups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	-- Unique slug for Docker service name
	app_name TEXT NOT NULL UNIQUE,
	-- Cron expression e.g. '0 2 * * *'
	schedule TEXT NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1,
	database_name TEXT NOT NULL,
	prefix TEXT NOT NULL,
	service_name TEXT, -- For compose backups
	keep_latest_count INTEGER,
	-- backup_type: DATABASE | COMPOSE
	backup_type TEXT NOT NULL DEFAULT 'DATABASE',
	-- database_type: POSTGRES | MARIADB | MYSQL | MONGO | REDIS | LIBSQL
	database_type TEXT NOT NULL,
	metadata TEXT, -- JSON string for extra config
	-- Foreign keys
	compose_id INTEGER REFERENCES compose_projects(id) ON DELETE CASCADE,
	postgres_id INTEGER REFERENCES postgres_dbs(id) ON DELETE CASCADE,
	mysql_id INTEGER REFERENCES mysql_dbs(id) ON DELETE CASCADE,
	mariadb_id INTEGER REFERENCES mariadb_dbs(id) ON DELETE CASCADE,
	mongo_id INTEGER REFERENCES mongo_dbs(id) ON DELETE CASCADE,
	redis_id INTEGER REFERENCES redis_dbs(id) ON DELETE CASCADE,
	libsql_id INTEGER REFERENCES libsql_dbs(id) ON DELETE CASCADE,
	destination_id INTEGER NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	-- Timestamp
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	-- Constraints
	CONSTRAINT backup_type_check CHECK (backup_type IN ('DATABASE', 'COMPOSE')),
	CONSTRAINT backup_db_type_check CHECK (database_type IN ('POSTGRES', 'MARIADB', 'MYSQL', 'MONGO', 'REDIS', 'LIBSQL'))
) STRICT;

CREATE INDEX idx_backups_destination_id ON backups(destination_id);

CREATE INDEX idx_backups_organization_id ON backups(organization_id);

CREATE INDEX idx_backups_compose_id ON backups(compose_id);

CREATE INDEX idx_backups_postgres_id ON backups(postgres_id);

CREATE INDEX idx_backups_mysql_id ON backups(mysql_id);

CREATE INDEX idx_backups_mariadb_id ON backups(mariadb_id);

CREATE INDEX idx_backups_mongo_id ON backups(mongo_id);

CREATE INDEX idx_backups_redis_id ON backups(redis_id);

CREATE INDEX idx_backups_libsql_id ON backups(libsql_id);

CREATE TRIGGER backups_updated_at
AFTER UPDATE ON backups
FOR EACH ROW
BEGIN
	UPDATE backups
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 42 volume_backups Table

Configures cron-scheduled raw Docker volume compression and S3 upload.

```sql
CREATE TABLE volume_backups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	volume_name TEXT NOT NULL,
	prefix TEXT NOT NULL,
	service_type TEXT NOT NULL DEFAULT 'APPLICATION',
	app_name TEXT NOT NULL UNIQUE,
	service_name TEXT,
	turn_off INTEGER NOT NULL DEFAULT 0,
	cron_expression TEXT NOT NULL,
	keep_latest_count INTEGER,
	enabled INTEGER NOT NULL DEFAULT 1,
	-- Foreign keys
	destination_id INTEGER NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
	postgres_id INTEGER REFERENCES postgres_dbs(id) ON DELETE CASCADE,
	mysql_id INTEGER REFERENCES mysql_dbs(id) ON DELETE CASCADE,
	mariadb_id INTEGER REFERENCES mariadb_dbs(id) ON DELETE CASCADE,
	mongo_id INTEGER REFERENCES mongo_dbs(id) ON DELETE CASCADE,
	redis_id INTEGER REFERENCES redis_dbs(id) ON DELETE CASCADE,
	libsql_id INTEGER REFERENCES libsql_dbs(id) ON DELETE CASCADE,
	compose_id INTEGER REFERENCES compose_projects(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	-- Constraints
	CONSTRAINT volume_backup_service_check CHECK (
		service_type IN ('APPLICATION', 'COMPOSE', 'POSTGRES', 'MYSQL', 'MARIADB', 'MONGO', 'REDIS', 'LIBSQL')
	)
) STRICT;

CREATE INDEX idx_volume_backups_destination_id ON volume_backups(destination_id);

CREATE INDEX idx_volume_backups_organization_id ON volume_backups(organization_id);

CREATE INDEX idx_volume_backups_application_id ON volume_backups(application_id);

CREATE INDEX idx_volume_backups_compose_id ON volume_backups(compose_id);

CREATE TRIGGER volume_backups_updated_at
AFTER UPDATE ON volume_backups
FOR EACH ROW
BEGIN
	UPDATE volume_backups
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 43 notif_slack Table

Stores Slack Webhook configurations.

```sql
CREATE TABLE notif_slack (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	webhook_url TEXT NOT NULL,
	channel TEXT
) STRICT;
```

### 44 notif_telegram Table

Stores Telegram Bot configurations.

```sql
CREATE TABLE notif_telegram (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	bot_token TEXT NOT NULL,
	chat_id TEXT NOT NULL,
	message_thread_id TEXT
) STRICT;
```

### 45 notif_discord Table

Stores Discord Webhook configurations.

```sql
CREATE TABLE notif_discord (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	webhook_url TEXT NOT NULL,
	decoration INTEGER NOT NULL DEFAULT 0
) STRICT;
```

### 46 notif_email Table

Stores SMTP mail server settings.

```sql
CREATE TABLE notif_email (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	smtp_server TEXT NOT NULL,
	smtp_port INTEGER NOT NULL,
	username TEXT NOT NULL,
	password TEXT NOT NULL,
	from_address TEXT NOT NULL,
	to_addresses TEXT NOT NULL DEFAULT '[]' -- JSON array of strings
) STRICT;
```

### 47 notif_resend Table

Stores Resend API service settings.

```sql
CREATE TABLE notif_resend (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	api_key TEXT NOT NULL,
	from_address TEXT NOT NULL,
	to_addresses TEXT NOT NULL DEFAULT '[]' -- JSON array of strings
) STRICT;
```

### 48 notif_gotify Table

Stores Gotify server access settings.

```sql
CREATE TABLE notif_gotify (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	server_url TEXT NOT NULL,
	app_token TEXT NOT NULL,
	priority INTEGER NOT NULL DEFAULT 5,
	decoration INTEGER NOT NULL DEFAULT 0
) STRICT;
```

### 49 notif_ntfy Table

Stores Ntfy server access settings.

```sql
CREATE TABLE notif_ntfy (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	server_url TEXT NOT NULL,
	topic TEXT NOT NULL,
	access_token TEXT,
	priority INTEGER NOT NULL DEFAULT 3
) STRICT;
```

### 50 notif_mattermost Table

Stores Mattermost integration settings.

```sql
CREATE TABLE notif_mattermost (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	webhook_url TEXT NOT NULL,
	channel TEXT,
	username TEXT
) STRICT;
```

### 51 notif_teams Table

Stores Microsoft Teams Webhook settings.

```sql
CREATE TABLE notif_teams (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	webhook_url TEXT NOT NULL
) STRICT;
```

### 52 notif_lark Table

Stores Lark Webhook integration settings.

```sql
CREATE TABLE notif_lark (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	webhook_url TEXT NOT NULL
) STRICT;
```

### 53 notif_pushover Table

Stores Pushover mobile application alerts settings.

```sql
CREATE TABLE notif_pushover (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_key TEXT NOT NULL,
	api_token TEXT NOT NULL,
	priority INTEGER NOT NULL DEFAULT 0,
	retry INTEGER,
	expire INTEGER
) STRICT;
```

### 54 notif_custom Table

Stores Custom HTTP webhook endpoints and custom headers JSON.

```sql
CREATE TABLE notif_custom (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	endpoint TEXT NOT NULL,
	headers TEXT -- JSON object (headers Record<string, string>)
) STRICT;
```

### 55 notifications Table

Tracks active notification triggers for system events (app deploy, build error, db backup) and maps them to channel sub-tables.

```sql
CREATE TABLE notifications (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	-- notification_type: SLACK | TELEGRAM | DISCORD | EMAIL | RESEND | GOTIFY | NTFY | MATTERMOST | PUSHOVER | CUSTOM | LARK | TEAMS
	notification_type TEXT NOT NULL,
	on_app_deploy INTEGER NOT NULL DEFAULT 0,
	on_app_build_error INTEGER NOT NULL DEFAULT 0,
	on_database_backup INTEGER NOT NULL DEFAULT 0,
	on_volume_backup INTEGER NOT NULL DEFAULT 0,
	on_panel_restart INTEGER NOT NULL DEFAULT 0,
	on_docker_cleanup INTEGER NOT NULL DEFAULT 0,
	on_server_threshold INTEGER NOT NULL DEFAULT 0,
	-- Foreign keys
	slack_id INTEGER REFERENCES notif_slack(id) ON DELETE CASCADE,
	telegram_id INTEGER REFERENCES notif_telegram(id) ON DELETE CASCADE,
	discord_id INTEGER REFERENCES notif_discord(id) ON DELETE CASCADE,
	email_id INTEGER REFERENCES notif_email(id) ON DELETE CASCADE,
	resend_id INTEGER REFERENCES notif_resend(id) ON DELETE CASCADE,
	gotify_id INTEGER REFERENCES notif_gotify(id) ON DELETE CASCADE,
	ntfy_id INTEGER REFERENCES notif_ntfy(id) ON DELETE CASCADE,
	mattermost_id INTEGER REFERENCES notif_mattermost(id) ON DELETE CASCADE,
	custom_id INTEGER REFERENCES notif_custom(id) ON DELETE CASCADE,
	lark_id INTEGER REFERENCES notif_lark(id) ON DELETE CASCADE,
	pushover_id INTEGER REFERENCES notif_pushover(id) ON DELETE CASCADE,
	teams_id INTEGER REFERENCES notif_teams(id) ON DELETE CASCADE,
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT notif_type_check CHECK (
		notification_type IN ('SLACK', 'TELEGRAM', 'DISCORD', 'EMAIL', 'RESEND', 'GOTIFY', 'NTFY', 'MATTERMOST', 'PUSHOVER', 'CUSTOM', 'LARK', 'TEAMS')
	)
) STRICT;

CREATE INDEX idx_notifications_organization_id ON notifications(organization_id);

CREATE TRIGGER notifications_updated_at
AFTER UPDATE ON notifications
FOR EACH ROW
BEGIN
	UPDATE notifications
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 56 schedules Table

Schedules automated tasks (bash or sh commands) to run inside containers or on the host.

```sql
CREATE TABLE schedules (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	description TEXT,
	cron_expression TEXT NOT NULL,
	app_name TEXT NOT NULL UNIQUE,
	service_name TEXT,
	-- shell_type: BASH | SH
	shell_type TEXT NOT NULL DEFAULT 'BASH',
	-- schedule_type: APPLICATION | COMPOSE | SERVER | Goploy-SERVER
	schedule_type TEXT NOT NULL DEFAULT 'APPLICATION',
	command TEXT NOT NULL,
	script TEXT,
	timezone TEXT,
	enabled INTEGER NOT NULL DEFAULT 1,
	-- Foreign keys
	application_id INTEGER REFERENCES applications(id) ON DELETE CASCADE,
	compose_id INTEGER REFERENCES compose_projects(id) ON DELETE CASCADE,
	server_id INTEGER REFERENCES servers(id) ON DELETE CASCADE,
	organization_id INTEGER REFERENCES organization(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT schedule_shell_type_check CHECK (shell_type IN ('BASH', 'SH')),
	CONSTRAINT schedule_type_check CHECK (schedule_type IN ('APPLICATION', 'COMPOSE', 'SERVER', 'Goploy-SERVER'))
) STRICT;

CREATE INDEX idx_schedules_application_id ON schedules(application_id);

CREATE INDEX idx_schedules_compose_id ON schedules(compose_id);

CREATE INDEX idx_schedules_server_id ON schedules(server_id);

CREATE INDEX idx_schedules_organization_id ON schedules(organization_id);

CREATE TRIGGER schedules_updated_at
AFTER UPDATE ON schedules
FOR EACH ROW
BEGIN
	UPDATE schedules
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 57 redirects Table

Saves custom Traefik redirection middleware matching rules (regex and replacements).

```sql
CREATE TABLE redirects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	regex TEXT NOT NULL,
	replacement TEXT NOT NULL,
	permanent INTEGER NOT NULL DEFAULT 0,
	unique_config_key INTEGER,
	application_id INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE INDEX idx_redirects_application_id ON redirects(application_id);

CREATE TRIGGER redirects_updated_at
AFTER UPDATE ON redirects
FOR EACH ROW
BEGIN
	UPDATE redirects
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 58 ports Table

Defines host-published-to-container-target port mapping configurations.

```sql
CREATE TABLE ports (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	published_port INTEGER NOT NULL,
	target_port INTEGER NOT NULL,
	-- protocol: TCP | UDP
	protocol TEXT NOT NULL DEFAULT 'TCP',
	-- publish_mode: INGRESS | HOST
	publish_mode TEXT NOT NULL DEFAULT 'HOST',
	application_id INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT port_protocol_check CHECK (protocol IN ('TCP', 'UDP')),
	CONSTRAINT port_publish_mode_check CHECK (publish_mode IN ('INGRESS', 'HOST'))
) STRICT;

CREATE INDEX idx_ports_application_id ON ports(application_id);
```

### 59 security Table

Saves HTTP Basic authentication username and password credentials per application router.

```sql
CREATE TABLE security (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,
	password TEXT NOT NULL,
	application_id INTEGER NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	UNIQUE(username, application_id)
) STRICT;

CREATE INDEX idx_security_application_id ON security(application_id);
```

### 60 audit_logs Table

Tracks user action trails and resource updates for security monitoring.

```sql
CREATE TABLE audit_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_email TEXT NOT NULL,
	user_role TEXT NOT NULL,
	action TEXT NOT NULL,
	resource_type TEXT NOT NULL,
	resource_id TEXT,
	resource_name TEXT,
	metadata TEXT, -- Extra info / JSON string
	-- Foreign keys
	organization_id INTEGER REFERENCES organization(id) ON DELETE SET NULL,
	user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL
) STRICT;

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);

CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

CREATE INDEX idx_audit_logs_organization_id ON audit_logs(organization_id);
```

### 61 settings Table

Configures global dashboard server configurations (Traefik resolvers, docker cleanup tasks, SSL domains).

```sql
CREATE TABLE settings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	server_ip TEXT,
	-- certificate_type: NONE | LETSENCRYPT | CUSTOM
	certificate_type TEXT NOT NULL DEFAULT 'NONE',
	custom_cert_resolver TEXT,
	https INTEGER NOT NULL DEFAULT 0,
	host TEXT, -- Domain Name for server
	lets_encrypt_email TEXT,
	enable_docker_cleanup INTEGER NOT NULL DEFAULT 1,
	log_cleanup_cron TEXT DEFAULT '0 0 * * *',
	-- JSON: metrics config object { server: {...}, containers: {...} }
	metrics_config TEXT NOT NULL DEFAULT '',
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT settings_certificate_check CHECK (certificate_type IN ('NONE', 'LETSENCRYPT', 'CUSTOM'))
) STRICT;

CREATE TRIGGER settings_updated_at
AFTER UPDATE ON settings
FOR EACH ROW
BEGIN
	UPDATE settings
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```

### 62 ai_settings Table

Stores API key configuration keys and OpenAI-compatible endpoints to generate composing stack configs.

```sql
CREATE TABLE ai_settings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	api_url TEXT NOT NULL,
	api_key TEXT NOT NULL,
	model TEXT NOT NULL,
	is_enabled INTEGER NOT NULL DEFAULT 1,
	-- Foreign keys
	organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
	created_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	updated_at INTEGER DEFAULT (strftime('%s', 'now')) NOT NULL,
	CONSTRAINT ai_enabled_check CHECK (is_enabled IN (0, 1))
) STRICT;

CREATE INDEX IF NOT EXISTS idx_ai_settings_organization_id ON ai_settings(organization_id);

CREATE TRIGGER ai_settings_updated_at
AFTER UPDATE ON ai_settings
FOR EACH ROW
BEGIN
	UPDATE ai_settings
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
```
