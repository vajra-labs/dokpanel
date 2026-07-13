-- +goose Up
-- create "groups" table
CREATE TABLE `groups` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now'))
) STRICT;
-- create index "groups_name" to table: "groups"
CREATE UNIQUE INDEX `groups_name` ON `groups` (`name`);
-- create trigger "groups_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `groups_updated_at` AFTER UPDATE ON `groups` FOR EACH ROW BEGIN
	UPDATE groups
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "policy" table
CREATE TABLE `policy` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `action` text NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now'))
) STRICT;
-- create index "policy_action" to table: "policy"
CREATE UNIQUE INDEX `policy_action` ON `policy` (`action`);
-- create "group_policy" table
CREATE TABLE `group_policy` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `group_id` integer NOT NULL,
  `policy_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`policy_id`) REFERENCES `policy` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION
) STRICT;
-- create "users" table
CREATE TABLE `users` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `email` text NULL,
  `last_name` text NULL,
  `first_name` text NULL,
  `avatar` text NOT NULL,
  `role` text NULL DEFAULT 'OWNER',
  `about_me` text NULL,
  `password` text NOT NULL,
  `is_email_verify` integer NULL DEFAULT 0,
  `email_verify_at` integer NULL,
  `two_factor_enable` integer NULL DEFAULT 0,
  `is_registered` integer NOT NULL DEFAULT 0,
  `added_by` integer NULL DEFAULT (NULL),
  `group_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `1` FOREIGN KEY (`added_by`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT `role_check` CHECK (role IN ('OWNER', 'ADMIN', 'MEMBER'))
) STRICT;
-- create index "users_email" to table: "users"
CREATE UNIQUE INDEX `users_email` ON `users` (`email`);
-- create trigger "users_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `users_updated_at` AFTER UPDATE ON `users` FOR EACH ROW BEGIN
	UPDATE users
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "two_factor" table
CREATE TABLE `two_factor` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `secret` text NOT NULL,
  `backup_codes` text NOT NULL,
  `user_id` integer NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create "jwt_tokens" table
CREATE TABLE `jwt_tokens` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `jti` text NOT NULL,
  `role` text NOT NULL,
  `user_id` integer NOT NULL,
  `is_blacklist` integer NULL DEFAULT 0,
  `blacklist_at` integer NULL,
  `expired_at` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `role_check` CHECK (role IN ('OWNER', 'ADMIN', 'MEMBER'))
) STRICT;
-- create trigger "jwt_tokens_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `jwt_tokens_updated_at` AFTER UPDATE ON `jwt_tokens` FOR EACH ROW BEGIN
	UPDATE jwt_tokens
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "organization" table
CREATE TABLE `organization` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `logo` text NULL,
  `slug` text NOT NULL,
  `owner_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create index "organization_name" to table: "organization"
CREATE UNIQUE INDEX `organization_name` ON `organization` (`name`);
-- create index "organization_slug" to table: "organization"
CREATE UNIQUE INDEX `organization_slug` ON `organization` (`slug`);
-- create trigger "organization_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `organization_updated_at` AFTER UPDATE ON `organization` FOR EACH ROW BEGIN
	UPDATE organization
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "organization_members" table
CREATE TABLE `organization_members` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `role` text NULL DEFAULT 'MEMBER',
  `user_id` integer NOT NULL,
  `organization_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s','now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s','now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `role_check` CHECK (role IN ('ADMIN', 'MEMBER'))
) STRICT;
-- create trigger "organization_members_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `organization_members_updated_at` AFTER UPDATE ON `organization_members` FOR EACH ROW BEGIN
	UPDATE organization_members
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "organization_invites" table
CREATE TABLE `organization_invites` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `email` text NOT NULL,
  `role` text NULL DEFAULT 'MEMBER',
  `status` text NULL DEFAULT 'PENDING',
  `token` text NOT NULL,
  `group_id` integer NOT NULL,
  `organization_id` integer NOT NULL,
  `invited_by` integer NOT NULL,
  `expired_at` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s','now')),
  CONSTRAINT `0` FOREIGN KEY (`invited_by`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `2` FOREIGN KEY (`group_id`) REFERENCES `groups` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `role_check` CHECK (role IN ('ADMIN', 'MEMBER')),
  CONSTRAINT `status_check` CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED'))
) STRICT;
-- create index "organization_invites_token" to table: "organization_invites"
CREATE UNIQUE INDEX `organization_invites_token` ON `organization_invites` (`token`);
-- create "projects" table
CREATE TABLE `projects` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `description` text NULL,
  `env_var` text NOT NULL DEFAULT '',
  `organization_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create index "projects_name" to table: "projects"
CREATE UNIQUE INDEX `projects_name` ON `projects` (`name`);
-- create trigger "projects_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `projects_updated_at` AFTER UPDATE ON `projects` FOR EACH ROW BEGIN
	UPDATE projects
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "tags" table
CREATE TABLE `tags` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `color` text NOT NULL,
  `organization_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create index "tags_name_organization_id" to table: "tags"
CREATE UNIQUE INDEX `tags_name_organization_id` ON `tags` (`name`, `organization_id`);
-- create "project_tags" table
CREATE TABLE `project_tags` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `project_id` integer NOT NULL,
  `tag_id` integer NOT NULL,
  CONSTRAINT `0` FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create index "project_tags_project_id_tag_id" to table: "project_tags"
CREATE UNIQUE INDEX `project_tags_project_id_tag_id` ON `project_tags` (`project_id`, `tag_id`);
-- create "postgres_dbs" table
CREATE TABLE `postgres_dbs` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `app_name` text NOT NULL,
  `description` text NULL,
  `docker_image` text NOT NULL DEFAULT 'postgres:18',
  `database_name` text NOT NULL,
  `database_user` text NOT NULL,
  `database_password` text NOT NULL,
  `external_port` integer NULL,
  `command` text NULL,
  `args` text NULL,
  `env_var` text NULL,
  `memory_reservation` text NULL,
  `memory_limit` text NULL,
  `cpu_reservation` text NULL,
  `cpu_limit` text NULL,
  `replicas` integer NOT NULL DEFAULT 1,
  `app_status` text NOT NULL DEFAULT 'IDLE',
  `health_check_swarm` text NULL,
  `restart_policy_swarm` text NULL,
  `placement_swarm` text NULL,
  `update_config_swarm` text NULL,
  `rollback_config_swarm` text NULL,
  `mode_swarm` text NULL,
  `labels_swarm` text NULL,
  `network_swarm` text NULL,
  `endpoint_spec_swarm` text NULL,
  `ulimits_swarm` text NULL,
  `stop_grace_period_swarm` integer NULL,
  `environment_id` integer NOT NULL,
  `server_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`environment_id`) REFERENCES `environments` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `pg_status_check` CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;
-- create index "postgres_dbs_app_name" to table: "postgres_dbs"
CREATE UNIQUE INDEX `postgres_dbs_app_name` ON `postgres_dbs` (`app_name`);
-- create trigger "pg_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `pg_updated_at` AFTER UPDATE ON `postgres_dbs` FOR EACH ROW BEGIN
	UPDATE postgres_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "mysql_dbs" table
CREATE TABLE `mysql_dbs` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `app_name` text NOT NULL,
  `description` text NULL,
  `docker_image` text NOT NULL DEFAULT 'mysql:9',
  `database_name` text NOT NULL,
  `database_user` text NOT NULL,
  `database_password` text NOT NULL,
  `database_root_password` text NOT NULL,
  `external_port` integer NULL,
  `command` text NULL,
  `args` text NULL,
  `env_var` text NULL,
  `memory_reservation` text NULL,
  `memory_limit` text NULL,
  `cpu_reservation` text NULL,
  `cpu_limit` text NULL,
  `replicas` integer NOT NULL DEFAULT 1,
  `app_status` text NOT NULL DEFAULT 'IDLE',
  `health_check_swarm` text NULL,
  `restart_policy_swarm` text NULL,
  `placement_swarm` text NULL,
  `update_config_swarm` text NULL,
  `rollback_config_swarm` text NULL,
  `mode_swarm` text NULL,
  `labels_swarm` text NULL,
  `network_swarm` text NULL,
  `endpoint_spec_swarm` text NULL,
  `ulimits_swarm` text NULL,
  `stop_grace_period_swarm` integer NULL,
  `environment_id` integer NOT NULL,
  `server_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`environment_id`) REFERENCES `environments` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `mysql_status_check` CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;
-- create index "mysql_dbs_app_name" to table: "mysql_dbs"
CREATE UNIQUE INDEX `mysql_dbs_app_name` ON `mysql_dbs` (`app_name`);
-- create trigger "mysql_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `mysql_updated_at` AFTER UPDATE ON `mysql_dbs` FOR EACH ROW BEGIN
	UPDATE mysql_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "mariadb_dbs" table
CREATE TABLE `mariadb_dbs` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `app_name` text NOT NULL,
  `description` text NULL,
  `docker_image` text NOT NULL DEFAULT 'mariadb:13',
  `database_name` text NOT NULL,
  `database_user` text NOT NULL,
  `database_password` text NOT NULL,
  `database_root_password` text NOT NULL,
  `external_port` integer NULL,
  `command` text NULL,
  `args` text NULL,
  `env_var` text NULL,
  `memory_reservation` text NULL,
  `memory_limit` text NULL,
  `cpu_reservation` text NULL,
  `cpu_limit` text NULL,
  `replicas` integer NOT NULL DEFAULT 1,
  `app_status` text NOT NULL DEFAULT 'IDLE',
  `health_check_swarm` text NULL,
  `restart_policy_swarm` text NULL,
  `placement_swarm` text NULL,
  `update_config_swarm` text NULL,
  `rollback_config_swarm` text NULL,
  `mode_swarm` text NULL,
  `labels_swarm` text NULL,
  `network_swarm` text NULL,
  `endpoint_spec_swarm` text NULL,
  `ulimits_swarm` text NULL,
  `stop_grace_period_swarm` integer NULL,
  `environment_id` integer NOT NULL,
  `server_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`environment_id`) REFERENCES `environments` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `mariadb_status_check` CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;
-- create index "mariadb_dbs_app_name" to table: "mariadb_dbs"
CREATE UNIQUE INDEX `mariadb_dbs_app_name` ON `mariadb_dbs` (`app_name`);
-- create trigger "mariadb_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `mariadb_updated_at` AFTER UPDATE ON `mariadb_dbs` FOR EACH ROW BEGIN
	UPDATE mariadb_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "mongo_dbs" table
CREATE TABLE `mongo_dbs` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `app_name` text NOT NULL,
  `description` text NULL,
  `docker_image` text NOT NULL DEFAULT 'mongo:8',
  `database_user` text NOT NULL,
  `database_password` text NOT NULL,
  `external_port` integer NULL,
  `replica_sets` integer NOT NULL DEFAULT 0,
  `command` text NULL,
  `args` text NULL,
  `env_var` text NULL,
  `memory_reservation` text NULL,
  `memory_limit` text NULL,
  `cpu_reservation` text NULL,
  `cpu_limit` text NULL,
  `replicas` integer NOT NULL DEFAULT 1,
  `app_status` text NOT NULL DEFAULT 'IDLE',
  `health_check_swarm` text NULL,
  `restart_policy_swarm` text NULL,
  `placement_swarm` text NULL,
  `update_config_swarm` text NULL,
  `rollback_config_swarm` text NULL,
  `mode_swarm` text NULL,
  `labels_swarm` text NULL,
  `network_swarm` text NULL,
  `endpoint_spec_swarm` text NULL,
  `ulimits_swarm` text NULL,
  `stop_grace_period_warm` integer NULL,
  `environment_id` integer NOT NULL,
  `server_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`environment_id`) REFERENCES `environments` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `mongo_status_check` CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;
-- create index "mongo_dbs_app_name" to table: "mongo_dbs"
CREATE UNIQUE INDEX `mongo_dbs_app_name` ON `mongo_dbs` (`app_name`);
-- create trigger "mongo_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `mongo_updated_at` AFTER UPDATE ON `mongo_dbs` FOR EACH ROW BEGIN
	UPDATE mongo_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "redis_dbs" table
CREATE TABLE `redis_dbs` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `app_name` text NOT NULL,
  `description` text NULL,
  `docker_image` text NOT NULL DEFAULT 'redis:8',
  `database_password` text NOT NULL,
  `external_port` integer NULL,
  `command` text NULL,
  `args` text NULL,
  `env_var` text NULL,
  `memory_reservation` text NULL,
  `memory_limit` text NULL,
  `cpu_reservation` text NULL,
  `cpu_limit` text NULL,
  `replicas` integer NOT NULL DEFAULT 1,
  `app_status` text NOT NULL DEFAULT 'IDLE',
  `health_check_swarm` text NULL,
  `restart_policy_swarm` text NULL,
  `placement_swarm` text NULL,
  `update_config_swarm` text NULL,
  `rollback_config_swarm` text NULL,
  `mode_swarm` text NULL,
  `labels_swarm` text NULL,
  `network_swarm` text NULL,
  `endpoint_spec_swarm` text NULL,
  `ulimits_swarm` text NULL,
  `stop_grace_period_swarm` integer NULL,
  `environment_id` integer NOT NULL,
  `server_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`environment_id`) REFERENCES `environments` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `redis_status_check` CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR'))
) STRICT;
-- create index "redis_dbs_app_name" to table: "redis_dbs"
CREATE UNIQUE INDEX `redis_dbs_app_name` ON `redis_dbs` (`app_name`);
-- create trigger "redis_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `redis_updated_at` AFTER UPDATE ON `redis_dbs` FOR EACH ROW BEGIN
	UPDATE redis_dbs
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "server_metrics" table
CREATE TABLE `server_metrics` (
  `timestamp` integer NULL,
  `cpu` real NOT NULL,
  `cpu_model` text NOT NULL,
  `cpu_cores` integer NOT NULL,
  `cpu_physical_cores` integer NOT NULL,
  `cpu_speed` real NOT NULL,
  `os` text NOT NULL,
  `distro` text NOT NULL,
  `kernel` text NOT NULL,
  `arch` text NOT NULL,
  `mem_used` real NOT NULL,
  `mem_used_gb` real NOT NULL,
  `mem_total` real NOT NULL,
  `uptime` integer NOT NULL,
  `disk_used` real NOT NULL,
  `total_disk` real NOT NULL,
  `network_in` real NOT NULL,
  `network_out` real NOT NULL,
  PRIMARY KEY (`timestamp`)
) STRICT;
-- create "container_metrics" table
CREATE TABLE `container_metrics` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `timestamp` integer NOT NULL,
  `container_id` text NOT NULL,
  `container_name` text NOT NULL,
  `metrics_json` text NOT NULL
) STRICT;
-- create index "idx_container_metrics_timestamp" to table: "container_metrics"
CREATE INDEX `idx_container_metrics_timestamp` ON `container_metrics` (`timestamp`);
-- create index "idx_container_metrics_name" to table: "container_metrics"
CREATE INDEX `idx_container_metrics_name` ON `container_metrics` (`container_name`);
-- create "ssh_keys" table
CREATE TABLE `ssh_keys` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `description` text NULL,
  `private_key` text NOT NULL DEFAULT '',
  `public_key` text NOT NULL,
  `last_used_at` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now'))
) STRICT;
-- create trigger "ssh_keys_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `ssh_keys_updated_at` AFTER UPDATE ON `ssh_keys` FOR EACH ROW BEGIN
	UPDATE ssh_keys
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "registries" table
CREATE TABLE `registries` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `registry_name` text NOT NULL,
  `image_prefix` text NULL,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `registry_url` text NOT NULL DEFAULT '',
  `registry_type` text NOT NULL DEFAULT 'CLOUD',
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `registry_type_check` CHECK (registry_type IN ('CLOUD', 'SELF_HOSTED'))
) STRICT;
-- create trigger "registries_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `registries_updated_at` AFTER UPDATE ON `registries` FOR EACH ROW BEGIN
	UPDATE registries
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "environments" table
CREATE TABLE `environments` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `description` text NULL,
  `env_var` text NOT NULL DEFAULT '',
  `is_default` integer NOT NULL DEFAULT 0,
  `project_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create trigger "environments_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `environments_updated_at` AFTER UPDATE ON `environments` FOR EACH ROW BEGIN
	UPDATE environments
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "servers" table
CREATE TABLE `servers` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `description` text NULL,
  `ip_address` text NOT NULL,
  `port` integer NOT NULL DEFAULT 22,
  `username` text NOT NULL DEFAULT 'root',
  `app_name` text NOT NULL,
  `server_status` text NOT NULL DEFAULT 'ACTIVE',
  `server_type` text NOT NULL DEFAULT 'DEPLOY',
  `enable_docker_cleanup` integer NOT NULL DEFAULT 0,
  `log_cleanup_cron` text NULL DEFAULT '0 0 * * *',
  `command` text NOT NULL DEFAULT '',
  `metrics_config` text NOT NULL DEFAULT '{}',
  `ssh_key_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`ssh_key_id`) REFERENCES `ssh_keys` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `server_status_check` CHECK (server_status IN ('ACTIVE', 'INACTIVE')),
  CONSTRAINT `server_type_check` CHECK (server_type IN ('DEPLOY', 'BUILD'))
) STRICT;
-- create index "servers_app_name" to table: "servers"
CREATE UNIQUE INDEX `servers_app_name` ON `servers` (`app_name`);
-- create trigger "servers_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `servers_updated_at` AFTER UPDATE ON `servers` FOR EACH ROW BEGIN
	UPDATE servers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "git_providers" table
CREATE TABLE `git_providers` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `provider_type` text NOT NULL DEFAULT 'GITHUB',
  `shared` integer NOT NULL DEFAULT 1,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `git_provider_type_check` CHECK (
		provider_type IN ('GITHUB', 'GITLAB', 'GITEA', 'BITBUCKET')
	)
) STRICT;
-- create trigger "git_providers_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `git_providers_updated_at` AFTER UPDATE ON `git_providers` FOR EACH ROW BEGIN
	UPDATE git_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "github_providers" table
CREATE TABLE `github_providers` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `github_app_name` text NULL,
  `github_app_id` integer NULL,
  `github_client_id` text NULL,
  `github_client_secret` text NULL,
  `github_installation_id` text NULL,
  `github_private_key` text NULL,
  `github_webhook_secret` text NULL,
  `git_provider_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`git_provider_id`) REFERENCES `git_providers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create trigger "github_providers_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `github_providers_updated_at` AFTER UPDATE ON `github_providers` FOR EACH ROW BEGIN
	UPDATE github_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "gitlab_providers" table
CREATE TABLE `gitlab_providers` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `gitlab_url` text NOT NULL DEFAULT 'https://gitlab.com',
  `gitlab_internal_url` text NULL,
  `application_id` text NULL,
  `redirect_uri` text NULL,
  `secret` text NULL,
  `access_token` text NULL,
  `refresh_token` text NULL,
  `group_name` text NULL,
  `expires_at` integer NULL,
  `git_provider_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`git_provider_id`) REFERENCES `git_providers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create trigger "gitlab_providers_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `gitlab_providers_updated_at` AFTER UPDATE ON `gitlab_providers` FOR EACH ROW BEGIN
	UPDATE gitlab_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "gitea_providers" table
CREATE TABLE `gitea_providers` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `gitea_url` text NOT NULL DEFAULT 'https://gitea.com',
  `gitea_internal_url` text NULL,
  `redirect_uri` text NULL,
  `client_id` text NULL,
  `client_secret` text NULL,
  `access_token` text NULL,
  `refresh_token` text NULL,
  `expires_at` integer NULL,
  `scopes` text NULL DEFAULT 'repo,repo:status,read:user,read:org',
  `last_authenticated_at` integer NULL,
  `git_provider_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`git_provider_id`) REFERENCES `git_providers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create trigger "gitea_providers_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `gitea_providers_updated_at` AFTER UPDATE ON `gitea_providers` FOR EACH ROW BEGIN
	UPDATE gitea_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "bitbucket_providers" table
CREATE TABLE `bitbucket_providers` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `bitbucket_username` text NULL,
  `bitbucket_email` text NULL,
  `app_password` text NULL,
  `api_token` text NULL,
  `bitbucket_workspace_name` text NULL,
  `git_provider_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`git_provider_id`) REFERENCES `git_providers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create trigger "bitbucket_providers_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `bitbucket_providers_updated_at` AFTER UPDATE ON `bitbucket_providers` FOR EACH ROW BEGIN
	UPDATE bitbucket_providers
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "applications" table
CREATE TABLE `applications` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `app_name` text NOT NULL,
  `description` text NULL,
  `source_type` text NOT NULL DEFAULT 'GITHUB',
  `build_type` text NOT NULL DEFAULT 'NIXPACKS',
  `app_status` text NOT NULL DEFAULT 'IDLE',
  `trigger_type` text NOT NULL DEFAULT 'PUSH',
  `build_args` text NULL,
  `build_secrets` text NULL,
  `dockerfile` text NULL DEFAULT 'Dockerfile',
  `docker_context_path` text NULL,
  `docker_build_stage` text NULL,
  `publish_directory` text NULL,
  `is_static_spa` integer NULL DEFAULT 0,
  `create_env_file` integer NOT NULL DEFAULT 1,
  `railpack_version` text NULL DEFAULT '0.15.4',
  `heroku_version` text NULL DEFAULT '24',
  `command` text NULL,
  `args` text NULL,
  `env_var` text NULL,
  `build_path` text NULL DEFAULT '/',
  `clean_cache` integer NOT NULL DEFAULT 0,
  `drop_build_path` text NULL,
  `enable_submodules` integer NOT NULL DEFAULT 0,
  `watch_paths` text NULL,
  `refresh_token` text NULL,
  `icon` text NULL,
  `memory_reservation` text NULL,
  `memory_limit` text NULL,
  `cpu_reservation` text NULL,
  `cpu_limit` text NULL,
  `replicas` integer NOT NULL DEFAULT 1,
  `health_check_swarm` text NULL,
  `restart_policy_swarm` text NULL,
  `placement_swarm` text NULL,
  `update_config_swarm` text NULL,
  `rollback_config_swarm` text NULL,
  `mode_swarm` text NULL,
  `labels_swarm` text NULL,
  `network_swarm` text NULL,
  `endpoint_spec_swarm` text NULL,
  `ulimits_swarm` text NULL,
  `stop_grace_period_swarm` integer NULL,
  `repository` text NULL,
  `owner` text NULL,
  `branch` text NULL,
  `auto_deploy` integer NULL DEFAULT 1,
  `gitlab_project_id` integer NULL,
  `gitlab_repository` text NULL,
  `gitlab_owner` text NULL,
  `gitlab_branch` text NULL,
  `gitlab_build_path` text NULL DEFAULT '/',
  `gitlab_path_namespace` text NULL,
  `gitea_repository` text NULL,
  `gitea_owner` text NULL,
  `gitea_branch` text NULL,
  `gitea_build_path` text NULL DEFAULT '/',
  `bitbucket_repository` text NULL,
  `bitbucket_repository_slug` text NULL,
  `bitbucket_owner` text NULL,
  `bitbucket_branch` text NULL,
  `bitbucket_build_path` text NULL DEFAULT '/',
  `docker_image` text NULL,
  `docker_username` text NULL,
  `docker_password` text NULL,
  `registry_url` text NULL,
  `custom_git_url` text NULL,
  `custom_git_branch` text NULL,
  `custom_git_build_path` text NULL,
  `custom_git_ssh_key_id` integer NULL,
  `preview_env` text NULL,
  `preview_build_args` text NULL,
  `preview_build_secrets` text NULL,
  `preview_labels` text NULL,
  `preview_wildcard` text NULL,
  `preview_port` integer NULL DEFAULT 3000,
  `preview_https` integer NOT NULL DEFAULT 0,
  `preview_path` text NULL DEFAULT '/',
  `preview_certificate_type` text NOT NULL DEFAULT 'NONE',
  `preview_custom_cert_resolver` text NULL,
  `preview_limit` integer NULL DEFAULT 3,
  `is_preview_deployments_active` integer NOT NULL DEFAULT 0,
  `preview_require_collaborator_permissions` integer NOT NULL DEFAULT 1,
  `rollback_active` integer NOT NULL DEFAULT 0,
  `environment_id` integer NOT NULL,
  `server_id` integer NULL,
  `build_server_id` integer NULL,
  `registry_id` integer NULL,
  `rollback_registry_id` integer NULL,
  `build_registry_id` integer NULL,
  `github_provider_id` integer NULL,
  `gitlab_provider_id` integer NULL,
  `gitea_provider_id` integer NULL,
  `bitbucket_provider_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`bitbucket_provider_id`) REFERENCES `bitbucket_providers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`gitea_provider_id`) REFERENCES `gitea_providers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `2` FOREIGN KEY (`gitlab_provider_id`) REFERENCES `gitlab_providers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `3` FOREIGN KEY (`github_provider_id`) REFERENCES `github_providers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `4` FOREIGN KEY (`build_registry_id`) REFERENCES `registries` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `5` FOREIGN KEY (`rollback_registry_id`) REFERENCES `registries` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `6` FOREIGN KEY (`registry_id`) REFERENCES `registries` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `7` FOREIGN KEY (`build_server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `8` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `9` FOREIGN KEY (`environment_id`) REFERENCES `environments` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `10` FOREIGN KEY (`custom_git_ssh_key_id`) REFERENCES `ssh_keys` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `app_source_type_check` CHECK (source_type IN ('DOCKER', 'GIT', 'GITHUB', 'GITLAB', 'BITBUCKET', 'GITEA', 'DROP')),
  CONSTRAINT `app_build_type_check` CHECK (build_type IN ('DOCKERFILE', 'HEROKU_BUILDPACKS', 'PAKETO_BUILDPACKS', 'NIXPACKS', 'STATIC', 'RAILPACK')),
  CONSTRAINT `app_status_check` CHECK (app_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR')),
  CONSTRAINT `app_trigger_type_check` CHECK (trigger_type IN ('PUSH', 'TAG')),
  CONSTRAINT `app_preview_cert_check` CHECK (preview_certificate_type IN ('LETSENCRYPT', 'NONE', 'CUSTOM'))
) STRICT;
-- create index "applications_app_name" to table: "applications"
CREATE UNIQUE INDEX `applications_app_name` ON `applications` (`app_name`);
-- create index "idx_applications_environment_id" to table: "applications"
CREATE INDEX `idx_applications_environment_id` ON `applications` (`environment_id`);
-- create index "idx_applications_server_id" to table: "applications"
CREATE INDEX `idx_applications_server_id` ON `applications` (`server_id`);
-- create index "idx_applications_app_status" to table: "applications"
CREATE INDEX `idx_applications_app_status` ON `applications` (`app_status`);
-- create trigger "applications_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `applications_updated_at` AFTER UPDATE ON `applications` FOR EACH ROW BEGIN
	UPDATE applications
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "compose_projects" table
CREATE TABLE `compose_projects` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `app_name` text NOT NULL,
  `description` text NULL,
  `env_var` text NULL,
  `compose_file` text NOT NULL DEFAULT '',
  `refresh_token` text NULL,
  `source_type` text NOT NULL DEFAULT 'GITHUB',
  `compose_type` text NOT NULL DEFAULT 'DOCKER-COMPOSE',
  `compose_status` text NOT NULL DEFAULT 'IDLE',
  `trigger_type` text NOT NULL DEFAULT 'PUSH',
  `repository` text NULL,
  `owner` text NULL,
  `branch` text NULL,
  `auto_deploy` integer NOT NULL DEFAULT 1,
  `gitlab_project_id` integer NULL,
  `gitlab_repository` text NULL,
  `gitlab_owner` text NULL,
  `gitlab_branch` text NULL,
  `gitlab_path_namespace` text NULL,
  `bitbucket_repository` text NULL,
  `bitbucket_repository_slug` text NULL,
  `bitbucket_owner` text NULL,
  `bitbucket_branch` text NULL,
  `gitea_repository` text NULL,
  `gitea_owner` text NULL,
  `gitea_branch` text NULL,
  `custom_git_url` text NULL,
  `custom_git_branch` text NULL,
  `custom_git_ssh_key_id` integer NULL,
  `command` text NOT NULL DEFAULT '',
  `enable_submodules` integer NOT NULL DEFAULT 0,
  `compose_path` text NOT NULL DEFAULT './docker-compose.yml',
  `suffix` text NOT NULL DEFAULT '',
  `randomize` integer NOT NULL DEFAULT 0,
  `isolated_deployment` integer NOT NULL DEFAULT 0,
  `isolated_deployments_volume` integer NOT NULL DEFAULT 0,
  `watch_paths` text NULL,
  `environment_id` integer NOT NULL,
  `server_id` integer NULL,
  `github_provider_id` integer NULL,
  `gitlab_provider_id` integer NULL,
  `gitea_provider_id` integer NULL,
  `bitbucket_provider_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`bitbucket_provider_id`) REFERENCES `bitbucket_providers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`gitea_provider_id`) REFERENCES `gitea_providers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `2` FOREIGN KEY (`gitlab_provider_id`) REFERENCES `gitlab_providers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `3` FOREIGN KEY (`github_provider_id`) REFERENCES `github_providers` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `4` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `5` FOREIGN KEY (`environment_id`) REFERENCES `environments` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `6` FOREIGN KEY (`custom_git_ssh_key_id`) REFERENCES `ssh_keys` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `compose_source_type_check` CHECK (source_type IN ('GIT', 'GITHUB', 'GITLAB', 'BITBUCKET', 'GITEA', 'RAW')),
  CONSTRAINT `compose_type_check` CHECK (compose_type IN ('DOCKER-COMPOSE', 'STACK')),
  CONSTRAINT `compose_status_check` CHECK (compose_status IN ('IDLE', 'RUNNING', 'DONE', 'ERROR')),
  CONSTRAINT `compose_trigger_type_check` CHECK (trigger_type IN ('PUSH', 'TAG'))
) STRICT;
-- create index "compose_projects_app_name" to table: "compose_projects"
CREATE UNIQUE INDEX `compose_projects_app_name` ON `compose_projects` (`app_name`);
-- create index "idx_compose_projects_environment_id" to table: "compose_projects"
CREATE INDEX `idx_compose_projects_environment_id` ON `compose_projects` (`environment_id`);
-- create index "idx_compose_projects_server_id" to table: "compose_projects"
CREATE INDEX `idx_compose_projects_server_id` ON `compose_projects` (`server_id`);
-- create trigger "compose_projects_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `compose_projects_updated_at` AFTER UPDATE ON `compose_projects` FOR EACH ROW BEGIN
	UPDATE compose_projects
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "domains" table
CREATE TABLE `domains` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `host` text NOT NULL,
  `https` integer NOT NULL DEFAULT 0,
  `port` integer NULL DEFAULT 3000,
  `path` text NULL DEFAULT '/',
  `internal_path` text NULL DEFAULT '/',
  `custom_entrypoint` text NULL,
  `service_name` text NULL,
  `custom_cert_resolver` text NULL,
  `strip_path` integer NOT NULL DEFAULT 0,
  `middlewares` text NOT NULL DEFAULT '[]',
  `domain_type` text NOT NULL DEFAULT 'APPLICATION',
  `certificate_type` text NOT NULL DEFAULT 'NONE',
  `application_id` integer NULL,
  `compose_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`compose_id`) REFERENCES `compose_projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `domain_cert_type_check` CHECK (certificate_type IN ('LETSENCRYPT', 'NONE', 'CUSTOM')),
  CONSTRAINT `domain_type_check` CHECK (domain_type IN ('APPLICATION', 'COMPOSE', 'PREVIEW'))
) STRICT;
-- create index "idx_domains_application_id" to table: "domains"
CREATE INDEX `idx_domains_application_id` ON `domains` (`application_id`);
-- create index "idx_domains_compose_id" to table: "domains"
CREATE INDEX `idx_domains_compose_id` ON `domains` (`compose_id`);
-- create index "idx_domains_host" to table: "domains"
CREATE INDEX `idx_domains_host` ON `domains` (`host`);
-- create trigger "domains_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `domains_updated_at` AFTER UPDATE ON `domains` FOR EACH ROW BEGIN
	UPDATE domains
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "patches" table
CREATE TABLE `patches` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `patch_type` text NOT NULL DEFAULT 'UPDATE',
  `file_path` text NOT NULL,
  `enabled` integer NOT NULL DEFAULT 1,
  `content` text NOT NULL,
  `application_id` integer NULL,
  `compose_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`compose_id`) REFERENCES `compose_projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `patch_type_check` CHECK (patch_type IN ('CREATE', 'UPDATE', 'DELETE'))
) STRICT;
-- create index "patches_file_path_application_id" to table: "patches"
CREATE UNIQUE INDEX `patches_file_path_application_id` ON `patches` (`file_path`, `application_id`);
-- create index "patches_file_path_compose_id" to table: "patches"
CREATE UNIQUE INDEX `patches_file_path_compose_id` ON `patches` (`file_path`, `compose_id`);
-- create index "idx_patches_application_id" to table: "patches"
CREATE INDEX `idx_patches_application_id` ON `patches` (`application_id`);
-- create index "idx_patches_compose_id" to table: "patches"
CREATE INDEX `idx_patches_compose_id` ON `patches` (`compose_id`);
-- create trigger "patches_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `patches_updated_at` AFTER UPDATE ON `patches` FOR EACH ROW BEGIN
	UPDATE patches
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "deployments" table
CREATE TABLE `deployments` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `title` text NOT NULL,
  `description` text NULL,
  `status` text NOT NULL DEFAULT 'RUNNING',
  `log_path` text NOT NULL,
  `pid` text NULL,
  `error_message` text NULL,
  `is_preview_deployment` integer NOT NULL DEFAULT 0,
  `started_at` integer NULL,
  `finished_at` integer NULL,
  `application_id` integer NULL,
  `compose_id` integer NULL,
  `server_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`compose_id`) REFERENCES `compose_projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `2` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `deployment_status_check` CHECK (status IN ('RUNNING', 'DONE', 'ERROR', 'CANCELLED'))
) STRICT;
-- create index "idx_deployments_status" to table: "deployments"
CREATE INDEX `idx_deployments_status` ON `deployments` (`status`);
-- create index "idx_deployments_created_at" to table: "deployments"
CREATE INDEX `idx_deployments_created_at` ON `deployments` (`created_at`);
-- create index "idx_deployments_compose_id" to table: "deployments"
CREATE INDEX `idx_deployments_compose_id` ON `deployments` (`compose_id`);
-- create index "idx_deployments_application_id" to table: "deployments"
CREATE INDEX `idx_deployments_application_id` ON `deployments` (`application_id`);
-- create "rollbacks" table
CREATE TABLE `rollbacks` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `deployment_id` integer NOT NULL,
  `version` integer NOT NULL DEFAULT 1,
  `image` text NULL,
  `full_context` text NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`deployment_id`) REFERENCES `deployments` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create index "idx_rollbacks_deployment_id" to table: "rollbacks"
CREATE INDEX `idx_rollbacks_deployment_id` ON `rollbacks` (`deployment_id`);
-- create "mounts" table
CREATE TABLE `mounts` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `mount_type` text NOT NULL DEFAULT 'VOLUME',
  `service_type` text NOT NULL DEFAULT 'APPLICATION',
  `host_path` text NULL,
  `volume_name` text NULL,
  `file_path` text NULL,
  `content` text NULL,
  `mount_path` text NOT NULL,
  `postgres_id` integer NULL,
  `mysql_id` integer NULL,
  `mariadb_id` integer NULL,
  `mongo_id` integer NULL,
  `redis_id` integer NULL,
  `libsql_id` integer NULL,
  `compose_id` integer NULL,
  `application_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`compose_id`) REFERENCES `compose_projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `2` FOREIGN KEY (`libsql_id`) REFERENCES `libsql_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `3` FOREIGN KEY (`redis_id`) REFERENCES `redis_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `4` FOREIGN KEY (`mongo_id`) REFERENCES `mongo_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `5` FOREIGN KEY (`mariadb_id`) REFERENCES `mariadb_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `6` FOREIGN KEY (`mysql_id`) REFERENCES `mysql_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `7` FOREIGN KEY (`postgres_id`) REFERENCES `postgres_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `mount_type_check` CHECK (mount_type IN ('BIND', 'VOLUME', 'FILE')),
  CONSTRAINT `mount_service_type_check` CHECK (
		service_type IN ('APPLICATION', 'COMPOSE', 'POSTGRES', 'MYSQL', 'MARIADB', 'MONGO', 'REDIS', 'LIBSQL')
	)
) STRICT;
-- create index "idx_mounts_application_id" to table: "mounts"
CREATE INDEX `idx_mounts_application_id` ON `mounts` (`application_id`);
-- create index "idx_mounts_compose_id" to table: "mounts"
CREATE INDEX `idx_mounts_compose_id` ON `mounts` (`compose_id`);
-- create index "idx_mounts_postgres_id" to table: "mounts"
CREATE INDEX `idx_mounts_postgres_id` ON `mounts` (`postgres_id`);
-- create index "idx_mounts_mysql_id" to table: "mounts"
CREATE INDEX `idx_mounts_mysql_id` ON `mounts` (`mysql_id`);
-- create index "idx_mounts_mariadb_id" to table: "mounts"
CREATE INDEX `idx_mounts_mariadb_id` ON `mounts` (`mariadb_id`);
-- create index "idx_mounts_mongo_id" to table: "mounts"
CREATE INDEX `idx_mounts_mongo_id` ON `mounts` (`mongo_id`);
-- create index "idx_mounts_redis_id" to table: "mounts"
CREATE INDEX `idx_mounts_redis_id` ON `mounts` (`redis_id`);
-- create index "idx_mounts_libsql_id" to table: "mounts"
CREATE INDEX `idx_mounts_libsql_id` ON `mounts` (`libsql_id`);
-- create trigger "mounts_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `mounts_updated_at` AFTER UPDATE ON `mounts` FOR EACH ROW BEGIN
	UPDATE mounts
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "certificates" table
CREATE TABLE `certificates` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `certificate_data` text NOT NULL,
  `private_key` text NOT NULL,
  `certificate_path` text NOT NULL,
  `auto_renew` integer NOT NULL DEFAULT 0,
  `server_id` integer NULL,
  `organization_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `auto_renew_check` CHECK (auto_renew IN (0, 1))
) STRICT;
-- create index "certificates_certificate_path" to table: "certificates"
CREATE UNIQUE INDEX `certificates_certificate_path` ON `certificates` (`certificate_path`);
-- create index "idx_certificates_server_id" to table: "certificates"
CREATE INDEX `idx_certificates_server_id` ON `certificates` (`server_id`);
-- create index "idx_certificates_organization_id" to table: "certificates"
CREATE INDEX `idx_certificates_organization_id` ON `certificates` (`organization_id`);
-- create trigger "certificates_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `certificates_updated_at` AFTER UPDATE ON `certificates` FOR EACH ROW BEGIN
	UPDATE certificates
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "destinations" table
CREATE TABLE `destinations` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `provider` text NOT NULL DEFAULT 'S3',
  `access_key` text NOT NULL,
  `secret_access_key` text NOT NULL,
  `bucket` text NOT NULL,
  `region` text NOT NULL,
  `endpoint` text NOT NULL,
  `additional_flags` text NULL,
  `organization_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `destination_provider_check` CHECK (provider IN ('S3', 'R2', 'BACKBLAZE', 'GCS', 'DO_SPACES'))
) STRICT;
-- create index "idx_destinations_organization_id" to table: "destinations"
CREATE INDEX `idx_destinations_organization_id` ON `destinations` (`organization_id`);
-- create trigger "destinations_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `destinations_updated_at` AFTER UPDATE ON `destinations` FOR EACH ROW BEGIN
	UPDATE destinations
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "backups" table
CREATE TABLE `backups` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `app_name` text NOT NULL,
  `schedule` text NOT NULL,
  `enabled` integer NOT NULL DEFAULT 1,
  `database_name` text NOT NULL,
  `prefix` text NOT NULL,
  `service_name` text NULL,
  `keep_latest_count` integer NULL,
  `backup_type` text NOT NULL DEFAULT 'DATABASE',
  `database_type` text NOT NULL,
  `metadata` text NULL,
  `compose_id` integer NULL,
  `postgres_id` integer NULL,
  `mysql_id` integer NULL,
  `mariadb_id` integer NULL,
  `mongo_id` integer NULL,
  `redis_id` integer NULL,
  `libsql_id` integer NULL,
  `destination_id` integer NOT NULL,
  `organization_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`destination_id`) REFERENCES `destinations` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `2` FOREIGN KEY (`libsql_id`) REFERENCES `libsql_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `3` FOREIGN KEY (`redis_id`) REFERENCES `redis_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `4` FOREIGN KEY (`mongo_id`) REFERENCES `mongo_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `5` FOREIGN KEY (`mariadb_id`) REFERENCES `mariadb_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `6` FOREIGN KEY (`mysql_id`) REFERENCES `mysql_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `7` FOREIGN KEY (`postgres_id`) REFERENCES `postgres_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `8` FOREIGN KEY (`compose_id`) REFERENCES `compose_projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `backup_type_check` CHECK (backup_type IN ('DATABASE', 'COMPOSE')),
  CONSTRAINT `backup_db_type_check` CHECK (database_type IN ('POSTGRES', 'MARIADB', 'MYSQL', 'MONGO', 'REDIS', 'LIBSQL'))
) STRICT;
-- create index "backups_app_name" to table: "backups"
CREATE UNIQUE INDEX `backups_app_name` ON `backups` (`app_name`);
-- create index "idx_backups_destination_id" to table: "backups"
CREATE INDEX `idx_backups_destination_id` ON `backups` (`destination_id`);
-- create index "idx_backups_organization_id" to table: "backups"
CREATE INDEX `idx_backups_organization_id` ON `backups` (`organization_id`);
-- create index "idx_backups_compose_id" to table: "backups"
CREATE INDEX `idx_backups_compose_id` ON `backups` (`compose_id`);
-- create index "idx_backups_postgres_id" to table: "backups"
CREATE INDEX `idx_backups_postgres_id` ON `backups` (`postgres_id`);
-- create index "idx_backups_mysql_id" to table: "backups"
CREATE INDEX `idx_backups_mysql_id` ON `backups` (`mysql_id`);
-- create index "idx_backups_mariadb_id" to table: "backups"
CREATE INDEX `idx_backups_mariadb_id` ON `backups` (`mariadb_id`);
-- create index "idx_backups_mongo_id" to table: "backups"
CREATE INDEX `idx_backups_mongo_id` ON `backups` (`mongo_id`);
-- create index "idx_backups_redis_id" to table: "backups"
CREATE INDEX `idx_backups_redis_id` ON `backups` (`redis_id`);
-- create index "idx_backups_libsql_id" to table: "backups"
CREATE INDEX `idx_backups_libsql_id` ON `backups` (`libsql_id`);
-- create trigger "backups_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `backups_updated_at` AFTER UPDATE ON `backups` FOR EACH ROW BEGIN
	UPDATE backups
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "volume_backups" table
CREATE TABLE `volume_backups` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `volume_name` text NOT NULL,
  `prefix` text NOT NULL,
  `service_type` text NOT NULL DEFAULT 'APPLICATION',
  `app_name` text NOT NULL,
  `service_name` text NULL,
  `turn_off` integer NOT NULL DEFAULT 0,
  `cron_expression` text NOT NULL,
  `keep_latest_count` integer NULL,
  `enabled` integer NOT NULL DEFAULT 1,
  `destination_id` integer NOT NULL,
  `organization_id` integer NOT NULL,
  `application_id` integer NULL,
  `postgres_id` integer NULL,
  `mysql_id` integer NULL,
  `mariadb_id` integer NULL,
  `mongo_id` integer NULL,
  `redis_id` integer NULL,
  `libsql_id` integer NULL,
  `compose_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`compose_id`) REFERENCES `compose_projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`libsql_id`) REFERENCES `libsql_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `2` FOREIGN KEY (`redis_id`) REFERENCES `redis_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `3` FOREIGN KEY (`mongo_id`) REFERENCES `mongo_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `4` FOREIGN KEY (`mariadb_id`) REFERENCES `mariadb_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `5` FOREIGN KEY (`mysql_id`) REFERENCES `mysql_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `6` FOREIGN KEY (`postgres_id`) REFERENCES `postgres_dbs` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `7` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `8` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `9` FOREIGN KEY (`destination_id`) REFERENCES `destinations` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `volume_backup_service_check` CHECK (
		service_type IN ('APPLICATION', 'COMPOSE', 'POSTGRES', 'MYSQL', 'MARIADB', 'MONGO', 'REDIS', 'LIBSQL')
	)
) STRICT;
-- create index "volume_backups_app_name" to table: "volume_backups"
CREATE UNIQUE INDEX `volume_backups_app_name` ON `volume_backups` (`app_name`);
-- create index "idx_volume_backups_destination_id" to table: "volume_backups"
CREATE INDEX `idx_volume_backups_destination_id` ON `volume_backups` (`destination_id`);
-- create index "idx_volume_backups_organization_id" to table: "volume_backups"
CREATE INDEX `idx_volume_backups_organization_id` ON `volume_backups` (`organization_id`);
-- create index "idx_volume_backups_application_id" to table: "volume_backups"
CREATE INDEX `idx_volume_backups_application_id` ON `volume_backups` (`application_id`);
-- create index "idx_volume_backups_compose_id" to table: "volume_backups"
CREATE INDEX `idx_volume_backups_compose_id` ON `volume_backups` (`compose_id`);
-- create trigger "volume_backups_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `volume_backups_updated_at` AFTER UPDATE ON `volume_backups` FOR EACH ROW BEGIN
	UPDATE volume_backups
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "notif_slack" table
CREATE TABLE `notif_slack` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `webhook_url` text NOT NULL,
  `channel` text NULL
) STRICT;
-- create "notif_telegram" table
CREATE TABLE `notif_telegram` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `bot_token` text NOT NULL,
  `chat_id` text NOT NULL,
  `message_thread_id` text NULL
) STRICT;
-- create "notif_discord" table
CREATE TABLE `notif_discord` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `webhook_url` text NOT NULL,
  `decoration` integer NOT NULL DEFAULT 0
) STRICT;
-- create "notif_email" table
CREATE TABLE `notif_email` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `smtp_server` text NOT NULL,
  `smtp_port` integer NOT NULL,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `from_address` text NOT NULL,
  `to_addresses` text NOT NULL DEFAULT '[]'
) STRICT;
-- create "notif_resend" table
CREATE TABLE `notif_resend` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `api_key` text NOT NULL,
  `from_address` text NOT NULL,
  `to_addresses` text NOT NULL DEFAULT '[]'
) STRICT;
-- create "notif_gotify" table
CREATE TABLE `notif_gotify` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `server_url` text NOT NULL,
  `app_token` text NOT NULL,
  `priority` integer NOT NULL DEFAULT 5,
  `decoration` integer NOT NULL DEFAULT 0
) STRICT;
-- create "notif_ntfy" table
CREATE TABLE `notif_ntfy` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `server_url` text NOT NULL,
  `topic` text NOT NULL,
  `access_token` text NULL,
  `priority` integer NOT NULL DEFAULT 3
) STRICT;
-- create "notif_mattermost" table
CREATE TABLE `notif_mattermost` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `webhook_url` text NOT NULL,
  `channel` text NULL,
  `username` text NULL
) STRICT;
-- create "notif_teams" table
CREATE TABLE `notif_teams` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `webhook_url` text NOT NULL
) STRICT;
-- create "notif_lark" table
CREATE TABLE `notif_lark` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `webhook_url` text NOT NULL
) STRICT;
-- create "notif_pushover" table
CREATE TABLE `notif_pushover` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `user_key` text NOT NULL,
  `api_token` text NOT NULL,
  `priority` integer NOT NULL DEFAULT 0,
  `retry` integer NULL,
  `expire` integer NULL
) STRICT;
-- create "notif_custom" table
CREATE TABLE `notif_custom` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `endpoint` text NOT NULL,
  `headers` text NULL
) STRICT;
-- create "notifications" table
CREATE TABLE `notifications` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `notification_type` text NOT NULL,
  `on_app_deploy` integer NOT NULL DEFAULT 0,
  `on_app_build_error` integer NOT NULL DEFAULT 0,
  `on_database_backup` integer NOT NULL DEFAULT 0,
  `on_volume_backup` integer NOT NULL DEFAULT 0,
  `on_panel_restart` integer NOT NULL DEFAULT 0,
  `on_docker_cleanup` integer NOT NULL DEFAULT 0,
  `on_server_threshold` integer NOT NULL DEFAULT 0,
  `slack_id` integer NULL,
  `telegram_id` integer NULL,
  `discord_id` integer NULL,
  `email_id` integer NULL,
  `resend_id` integer NULL,
  `gotify_id` integer NULL,
  `ntfy_id` integer NULL,
  `mattermost_id` integer NULL,
  `custom_id` integer NULL,
  `lark_id` integer NULL,
  `pushover_id` integer NULL,
  `teams_id` integer NULL,
  `organization_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`teams_id`) REFERENCES `notif_teams` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `2` FOREIGN KEY (`pushover_id`) REFERENCES `notif_pushover` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `3` FOREIGN KEY (`lark_id`) REFERENCES `notif_lark` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `4` FOREIGN KEY (`custom_id`) REFERENCES `notif_custom` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `5` FOREIGN KEY (`mattermost_id`) REFERENCES `notif_mattermost` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `6` FOREIGN KEY (`ntfy_id`) REFERENCES `notif_ntfy` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `7` FOREIGN KEY (`gotify_id`) REFERENCES `notif_gotify` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `8` FOREIGN KEY (`resend_id`) REFERENCES `notif_resend` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `9` FOREIGN KEY (`email_id`) REFERENCES `notif_email` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `10` FOREIGN KEY (`discord_id`) REFERENCES `notif_discord` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `11` FOREIGN KEY (`telegram_id`) REFERENCES `notif_telegram` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `12` FOREIGN KEY (`slack_id`) REFERENCES `notif_slack` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `notif_type_check` CHECK (
		notification_type IN ('SLACK', 'TELEGRAM', 'DISCORD', 'EMAIL', 'RESEND', 'GOTIFY', 'NTFY', 'MATTERMOST', 'PUSHOVER', 'CUSTOM', 'LARK', 'TEAMS')
	)
) STRICT;
-- create index "idx_notifications_organization_id" to table: "notifications"
CREATE INDEX `idx_notifications_organization_id` ON `notifications` (`organization_id`);
-- create trigger "notifications_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `notifications_updated_at` AFTER UPDATE ON `notifications` FOR EACH ROW BEGIN
	UPDATE notifications
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "schedules" table
CREATE TABLE `schedules` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `description` text NULL,
  `cron_expression` text NOT NULL,
  `app_name` text NOT NULL,
  `service_name` text NULL,
  `shell_type` text NOT NULL DEFAULT 'BASH',
  `schedule_type` text NOT NULL DEFAULT 'APPLICATION',
  `command` text NOT NULL,
  `script` text NULL,
  `timezone` text NULL,
  `enabled` integer NOT NULL DEFAULT 1,
  `application_id` integer NULL,
  `compose_id` integer NULL,
  `server_id` integer NULL,
  `organization_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `1` FOREIGN KEY (`server_id`) REFERENCES `servers` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `2` FOREIGN KEY (`compose_id`) REFERENCES `compose_projects` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `3` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `schedule_shell_type_check` CHECK (shell_type IN ('BASH', 'SH')),
  CONSTRAINT `schedule_type_check` CHECK (schedule_type IN ('APPLICATION', 'COMPOSE', 'SERVER', 'Goploy-SERVER'))
) STRICT;
-- create index "schedules_app_name" to table: "schedules"
CREATE UNIQUE INDEX `schedules_app_name` ON `schedules` (`app_name`);
-- create index "idx_schedules_application_id" to table: "schedules"
CREATE INDEX `idx_schedules_application_id` ON `schedules` (`application_id`);
-- create index "idx_schedules_compose_id" to table: "schedules"
CREATE INDEX `idx_schedules_compose_id` ON `schedules` (`compose_id`);
-- create index "idx_schedules_server_id" to table: "schedules"
CREATE INDEX `idx_schedules_server_id` ON `schedules` (`server_id`);
-- create index "idx_schedules_organization_id" to table: "schedules"
CREATE INDEX `idx_schedules_organization_id` ON `schedules` (`organization_id`);
-- create trigger "schedules_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `schedules_updated_at` AFTER UPDATE ON `schedules` FOR EACH ROW BEGIN
	UPDATE schedules
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "redirects" table
CREATE TABLE `redirects` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `regex` text NOT NULL,
  `replacement` text NOT NULL,
  `permanent` integer NOT NULL DEFAULT 0,
  `unique_config_key` integer NULL,
  `application_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create index "idx_redirects_application_id" to table: "redirects"
CREATE INDEX `idx_redirects_application_id` ON `redirects` (`application_id`);
-- create trigger "redirects_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `redirects_updated_at` AFTER UPDATE ON `redirects` FOR EACH ROW BEGIN
	UPDATE redirects
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "ports" table
CREATE TABLE `ports` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `published_port` integer NOT NULL,
  `target_port` integer NOT NULL,
  `protocol` text NOT NULL DEFAULT 'TCP',
  `publish_mode` text NOT NULL DEFAULT 'HOST',
  `application_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `port_protocol_check` CHECK (protocol IN ('TCP', 'UDP')),
  CONSTRAINT `port_publish_mode_check` CHECK (publish_mode IN ('INGRESS', 'HOST'))
) STRICT;
-- create index "idx_ports_application_id" to table: "ports"
CREATE INDEX `idx_ports_application_id` ON `ports` (`application_id`);
-- create "security" table
CREATE TABLE `security` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `username` text NOT NULL,
  `password` text NOT NULL,
  `application_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`application_id`) REFERENCES `applications` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE
) STRICT;
-- create index "security_username_application_id" to table: "security"
CREATE UNIQUE INDEX `security_username_application_id` ON `security` (`username`, `application_id`);
-- create index "idx_security_application_id" to table: "security"
CREATE INDEX `idx_security_application_id` ON `security` (`application_id`);
-- create "audit_logs" table
CREATE TABLE `audit_logs` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `user_email` text NOT NULL,
  `user_role` text NOT NULL,
  `action` text NOT NULL,
  `resource_type` text NOT NULL,
  `resource_id` text NULL,
  `resource_name` text NULL,
  `metadata` text NULL,
  `organization_id` integer NULL,
  `user_id` integer NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT `1` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE SET NULL
) STRICT;
-- create index "idx_audit_logs_user_id" to table: "audit_logs"
CREATE INDEX `idx_audit_logs_user_id` ON `audit_logs` (`user_id`);
-- create index "idx_audit_logs_created_at" to table: "audit_logs"
CREATE INDEX `idx_audit_logs_created_at` ON `audit_logs` (`created_at`);
-- create index "idx_audit_logs_organization_id" to table: "audit_logs"
CREATE INDEX `idx_audit_logs_organization_id` ON `audit_logs` (`organization_id`);
-- create "settings" table
CREATE TABLE `settings` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `server_ip` text NULL,
  `certificate_type` text NOT NULL DEFAULT 'NONE',
  `custom_cert_resolver` text NULL,
  `https` integer NOT NULL DEFAULT 0,
  `host` text NULL,
  `lets_encrypt_email` text NULL,
  `enable_docker_cleanup` integer NOT NULL DEFAULT 1,
  `log_cleanup_cron` text NULL DEFAULT '0 0 * * *',
  `metrics_config` text NOT NULL DEFAULT '',
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `settings_certificate_check` CHECK (certificate_type IN ('NONE', 'LETSENCRYPT', 'CUSTOM'))
) STRICT;
-- create trigger "settings_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `settings_updated_at` AFTER UPDATE ON `settings` FOR EACH ROW BEGIN
	UPDATE settings
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd
-- create "ai_settings" table
CREATE TABLE `ai_settings` (
  `id` integer NULL PRIMARY KEY AUTOINCREMENT,
  `name` text NOT NULL,
  `api_url` text NOT NULL,
  `api_key` text NOT NULL,
  `model` text NOT NULL,
  `is_enabled` integer NOT NULL DEFAULT 1,
  `organization_id` integer NOT NULL,
  `created_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  `updated_at` integer NOT NULL DEFAULT (strftime('%s', 'now')),
  CONSTRAINT `0` FOREIGN KEY (`organization_id`) REFERENCES `organization` (`id`) ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT `ai_enabled_check` CHECK (is_enabled IN (0, 1))
) STRICT;
-- create index "idx_ai_settings_organization_id" to table: "ai_settings"
CREATE INDEX `idx_ai_settings_organization_id` ON `ai_settings` (`organization_id`);
-- create trigger "ai_settings_updated_at"
-- +goose StatementBegin
CREATE TRIGGER `ai_settings_updated_at` AFTER UPDATE ON `ai_settings` FOR EACH ROW BEGIN
	UPDATE ai_settings
	SET updated_at = strftime('%s', 'now')
	WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
-- reverse: create trigger "ai_settings_updated_at"
DROP TRIGGER `ai_settings_updated_at`;
-- reverse: create index "idx_ai_settings_organization_id" to table: "ai_settings"
DROP INDEX `idx_ai_settings_organization_id`;
-- reverse: create "ai_settings" table
DROP TABLE `ai_settings`;
-- reverse: create trigger "settings_updated_at"
DROP TRIGGER `settings_updated_at`;
-- reverse: create "settings" table
DROP TABLE `settings`;
-- reverse: create index "idx_audit_logs_organization_id" to table: "audit_logs"
DROP INDEX `idx_audit_logs_organization_id`;
-- reverse: create index "idx_audit_logs_created_at" to table: "audit_logs"
DROP INDEX `idx_audit_logs_created_at`;
-- reverse: create index "idx_audit_logs_user_id" to table: "audit_logs"
DROP INDEX `idx_audit_logs_user_id`;
-- reverse: create "audit_logs" table
DROP TABLE `audit_logs`;
-- reverse: create index "idx_security_application_id" to table: "security"
DROP INDEX `idx_security_application_id`;
-- reverse: create index "security_username_application_id" to table: "security"
DROP INDEX `security_username_application_id`;
-- reverse: create "security" table
DROP TABLE `security`;
-- reverse: create index "idx_ports_application_id" to table: "ports"
DROP INDEX `idx_ports_application_id`;
-- reverse: create "ports" table
DROP TABLE `ports`;
-- reverse: create trigger "redirects_updated_at"
DROP TRIGGER `redirects_updated_at`;
-- reverse: create index "idx_redirects_application_id" to table: "redirects"
DROP INDEX `idx_redirects_application_id`;
-- reverse: create "redirects" table
DROP TABLE `redirects`;
-- reverse: create trigger "schedules_updated_at"
DROP TRIGGER `schedules_updated_at`;
-- reverse: create index "idx_schedules_organization_id" to table: "schedules"
DROP INDEX `idx_schedules_organization_id`;
-- reverse: create index "idx_schedules_server_id" to table: "schedules"
DROP INDEX `idx_schedules_server_id`;
-- reverse: create index "idx_schedules_compose_id" to table: "schedules"
DROP INDEX `idx_schedules_compose_id`;
-- reverse: create index "idx_schedules_application_id" to table: "schedules"
DROP INDEX `idx_schedules_application_id`;
-- reverse: create index "schedules_app_name" to table: "schedules"
DROP INDEX `schedules_app_name`;
-- reverse: create "schedules" table
DROP TABLE `schedules`;
-- reverse: create trigger "notifications_updated_at"
DROP TRIGGER `notifications_updated_at`;
-- reverse: create index "idx_notifications_organization_id" to table: "notifications"
DROP INDEX `idx_notifications_organization_id`;
-- reverse: create "notifications" table
DROP TABLE `notifications`;
-- reverse: create "notif_custom" table
DROP TABLE `notif_custom`;
-- reverse: create "notif_pushover" table
DROP TABLE `notif_pushover`;
-- reverse: create "notif_lark" table
DROP TABLE `notif_lark`;
-- reverse: create "notif_teams" table
DROP TABLE `notif_teams`;
-- reverse: create "notif_mattermost" table
DROP TABLE `notif_mattermost`;
-- reverse: create "notif_ntfy" table
DROP TABLE `notif_ntfy`;
-- reverse: create "notif_gotify" table
DROP TABLE `notif_gotify`;
-- reverse: create "notif_resend" table
DROP TABLE `notif_resend`;
-- reverse: create "notif_email" table
DROP TABLE `notif_email`;
-- reverse: create "notif_discord" table
DROP TABLE `notif_discord`;
-- reverse: create "notif_telegram" table
DROP TABLE `notif_telegram`;
-- reverse: create "notif_slack" table
DROP TABLE `notif_slack`;
-- reverse: create trigger "volume_backups_updated_at"
DROP TRIGGER `volume_backups_updated_at`;
-- reverse: create index "idx_volume_backups_compose_id" to table: "volume_backups"
DROP INDEX `idx_volume_backups_compose_id`;
-- reverse: create index "idx_volume_backups_application_id" to table: "volume_backups"
DROP INDEX `idx_volume_backups_application_id`;
-- reverse: create index "idx_volume_backups_organization_id" to table: "volume_backups"
DROP INDEX `idx_volume_backups_organization_id`;
-- reverse: create index "idx_volume_backups_destination_id" to table: "volume_backups"
DROP INDEX `idx_volume_backups_destination_id`;
-- reverse: create index "volume_backups_app_name" to table: "volume_backups"
DROP INDEX `volume_backups_app_name`;
-- reverse: create "volume_backups" table
DROP TABLE `volume_backups`;
-- reverse: create trigger "backups_updated_at"
DROP TRIGGER `backups_updated_at`;
-- reverse: create index "idx_backups_libsql_id" to table: "backups"
DROP INDEX `idx_backups_libsql_id`;
-- reverse: create index "idx_backups_redis_id" to table: "backups"
DROP INDEX `idx_backups_redis_id`;
-- reverse: create index "idx_backups_mongo_id" to table: "backups"
DROP INDEX `idx_backups_mongo_id`;
-- reverse: create index "idx_backups_mariadb_id" to table: "backups"
DROP INDEX `idx_backups_mariadb_id`;
-- reverse: create index "idx_backups_mysql_id" to table: "backups"
DROP INDEX `idx_backups_mysql_id`;
-- reverse: create index "idx_backups_postgres_id" to table: "backups"
DROP INDEX `idx_backups_postgres_id`;
-- reverse: create index "idx_backups_compose_id" to table: "backups"
DROP INDEX `idx_backups_compose_id`;
-- reverse: create index "idx_backups_organization_id" to table: "backups"
DROP INDEX `idx_backups_organization_id`;
-- reverse: create index "idx_backups_destination_id" to table: "backups"
DROP INDEX `idx_backups_destination_id`;
-- reverse: create index "backups_app_name" to table: "backups"
DROP INDEX `backups_app_name`;
-- reverse: create "backups" table
DROP TABLE `backups`;
-- reverse: create trigger "destinations_updated_at"
DROP TRIGGER `destinations_updated_at`;
-- reverse: create index "idx_destinations_organization_id" to table: "destinations"
DROP INDEX `idx_destinations_organization_id`;
-- reverse: create "destinations" table
DROP TABLE `destinations`;
-- reverse: create trigger "certificates_updated_at"
DROP TRIGGER `certificates_updated_at`;
-- reverse: create index "idx_certificates_organization_id" to table: "certificates"
DROP INDEX `idx_certificates_organization_id`;
-- reverse: create index "idx_certificates_server_id" to table: "certificates"
DROP INDEX `idx_certificates_server_id`;
-- reverse: create index "certificates_certificate_path" to table: "certificates"
DROP INDEX `certificates_certificate_path`;
-- reverse: create "certificates" table
DROP TABLE `certificates`;
-- reverse: create trigger "mounts_updated_at"
DROP TRIGGER `mounts_updated_at`;
-- reverse: create index "idx_mounts_libsql_id" to table: "mounts"
DROP INDEX `idx_mounts_libsql_id`;
-- reverse: create index "idx_mounts_redis_id" to table: "mounts"
DROP INDEX `idx_mounts_redis_id`;
-- reverse: create index "idx_mounts_mongo_id" to table: "mounts"
DROP INDEX `idx_mounts_mongo_id`;
-- reverse: create index "idx_mounts_mariadb_id" to table: "mounts"
DROP INDEX `idx_mounts_mariadb_id`;
-- reverse: create index "idx_mounts_mysql_id" to table: "mounts"
DROP INDEX `idx_mounts_mysql_id`;
-- reverse: create index "idx_mounts_postgres_id" to table: "mounts"
DROP INDEX `idx_mounts_postgres_id`;
-- reverse: create index "idx_mounts_compose_id" to table: "mounts"
DROP INDEX `idx_mounts_compose_id`;
-- reverse: create index "idx_mounts_application_id" to table: "mounts"
DROP INDEX `idx_mounts_application_id`;
-- reverse: create "mounts" table
DROP TABLE `mounts`;
-- reverse: create index "idx_rollbacks_deployment_id" to table: "rollbacks"
DROP INDEX `idx_rollbacks_deployment_id`;
-- reverse: create "rollbacks" table
DROP TABLE `rollbacks`;
-- reverse: create index "idx_deployments_application_id" to table: "deployments"
DROP INDEX `idx_deployments_application_id`;
-- reverse: create index "idx_deployments_compose_id" to table: "deployments"
DROP INDEX `idx_deployments_compose_id`;
-- reverse: create index "idx_deployments_created_at" to table: "deployments"
DROP INDEX `idx_deployments_created_at`;
-- reverse: create index "idx_deployments_status" to table: "deployments"
DROP INDEX `idx_deployments_status`;
-- reverse: create "deployments" table
DROP TABLE `deployments`;
-- reverse: create trigger "patches_updated_at"
DROP TRIGGER `patches_updated_at`;
-- reverse: create index "idx_patches_compose_id" to table: "patches"
DROP INDEX `idx_patches_compose_id`;
-- reverse: create index "idx_patches_application_id" to table: "patches"
DROP INDEX `idx_patches_application_id`;
-- reverse: create index "patches_file_path_compose_id" to table: "patches"
DROP INDEX `patches_file_path_compose_id`;
-- reverse: create index "patches_file_path_application_id" to table: "patches"
DROP INDEX `patches_file_path_application_id`;
-- reverse: create "patches" table
DROP TABLE `patches`;
-- reverse: create trigger "domains_updated_at"
DROP TRIGGER `domains_updated_at`;
-- reverse: create index "idx_domains_host" to table: "domains"
DROP INDEX `idx_domains_host`;
-- reverse: create index "idx_domains_compose_id" to table: "domains"
DROP INDEX `idx_domains_compose_id`;
-- reverse: create index "idx_domains_application_id" to table: "domains"
DROP INDEX `idx_domains_application_id`;
-- reverse: create "domains" table
DROP TABLE `domains`;
-- reverse: create trigger "compose_projects_updated_at"
DROP TRIGGER `compose_projects_updated_at`;
-- reverse: create index "idx_compose_projects_server_id" to table: "compose_projects"
DROP INDEX `idx_compose_projects_server_id`;
-- reverse: create index "idx_compose_projects_environment_id" to table: "compose_projects"
DROP INDEX `idx_compose_projects_environment_id`;
-- reverse: create index "compose_projects_app_name" to table: "compose_projects"
DROP INDEX `compose_projects_app_name`;
-- reverse: create "compose_projects" table
DROP TABLE `compose_projects`;
-- reverse: create trigger "applications_updated_at"
DROP TRIGGER `applications_updated_at`;
-- reverse: create index "idx_applications_app_status" to table: "applications"
DROP INDEX `idx_applications_app_status`;
-- reverse: create index "idx_applications_server_id" to table: "applications"
DROP INDEX `idx_applications_server_id`;
-- reverse: create index "idx_applications_environment_id" to table: "applications"
DROP INDEX `idx_applications_environment_id`;
-- reverse: create index "applications_app_name" to table: "applications"
DROP INDEX `applications_app_name`;
-- reverse: create "applications" table
DROP TABLE `applications`;
-- reverse: create trigger "bitbucket_providers_updated_at"
DROP TRIGGER `bitbucket_providers_updated_at`;
-- reverse: create "bitbucket_providers" table
DROP TABLE `bitbucket_providers`;
-- reverse: create trigger "gitea_providers_updated_at"
DROP TRIGGER `gitea_providers_updated_at`;
-- reverse: create "gitea_providers" table
DROP TABLE `gitea_providers`;
-- reverse: create trigger "gitlab_providers_updated_at"
DROP TRIGGER `gitlab_providers_updated_at`;
-- reverse: create "gitlab_providers" table
DROP TABLE `gitlab_providers`;
-- reverse: create trigger "github_providers_updated_at"
DROP TRIGGER `github_providers_updated_at`;
-- reverse: create "github_providers" table
DROP TABLE `github_providers`;
-- reverse: create trigger "git_providers_updated_at"
DROP TRIGGER `git_providers_updated_at`;
-- reverse: create "git_providers" table
DROP TABLE `git_providers`;
-- reverse: create trigger "servers_updated_at"
DROP TRIGGER `servers_updated_at`;
-- reverse: create index "servers_app_name" to table: "servers"
DROP INDEX `servers_app_name`;
-- reverse: create "servers" table
DROP TABLE `servers`;
-- reverse: create trigger "environments_updated_at"
DROP TRIGGER `environments_updated_at`;
-- reverse: create "environments" table
DROP TABLE `environments`;
-- reverse: create trigger "registries_updated_at"
DROP TRIGGER `registries_updated_at`;
-- reverse: create "registries" table
DROP TABLE `registries`;
-- reverse: create trigger "ssh_keys_updated_at"
DROP TRIGGER `ssh_keys_updated_at`;
-- reverse: create "ssh_keys" table
DROP TABLE `ssh_keys`;
-- reverse: create index "idx_container_metrics_name" to table: "container_metrics"
DROP INDEX `idx_container_metrics_name`;
-- reverse: create index "idx_container_metrics_timestamp" to table: "container_metrics"
DROP INDEX `idx_container_metrics_timestamp`;
-- reverse: create "container_metrics" table
DROP TABLE `container_metrics`;
-- reverse: create "server_metrics" table
DROP TABLE `server_metrics`;
-- reverse: create trigger "redis_updated_at"
DROP TRIGGER `redis_updated_at`;
-- reverse: create index "redis_dbs_app_name" to table: "redis_dbs"
DROP INDEX `redis_dbs_app_name`;
-- reverse: create "redis_dbs" table
DROP TABLE `redis_dbs`;
-- reverse: create trigger "mongo_updated_at"
DROP TRIGGER `mongo_updated_at`;
-- reverse: create index "mongo_dbs_app_name" to table: "mongo_dbs"
DROP INDEX `mongo_dbs_app_name`;
-- reverse: create "mongo_dbs" table
DROP TABLE `mongo_dbs`;
-- reverse: create trigger "mariadb_updated_at"
DROP TRIGGER `mariadb_updated_at`;
-- reverse: create index "mariadb_dbs_app_name" to table: "mariadb_dbs"
DROP INDEX `mariadb_dbs_app_name`;
-- reverse: create "mariadb_dbs" table
DROP TABLE `mariadb_dbs`;
-- reverse: create trigger "mysql_updated_at"
DROP TRIGGER `mysql_updated_at`;
-- reverse: create index "mysql_dbs_app_name" to table: "mysql_dbs"
DROP INDEX `mysql_dbs_app_name`;
-- reverse: create "mysql_dbs" table
DROP TABLE `mysql_dbs`;
-- reverse: create trigger "pg_updated_at"
DROP TRIGGER `pg_updated_at`;
-- reverse: create index "postgres_dbs_app_name" to table: "postgres_dbs"
DROP INDEX `postgres_dbs_app_name`;
-- reverse: create "postgres_dbs" table
DROP TABLE `postgres_dbs`;
-- reverse: create index "project_tags_project_id_tag_id" to table: "project_tags"
DROP INDEX `project_tags_project_id_tag_id`;
-- reverse: create "project_tags" table
DROP TABLE `project_tags`;
-- reverse: create index "tags_name_organization_id" to table: "tags"
DROP INDEX `tags_name_organization_id`;
-- reverse: create "tags" table
DROP TABLE `tags`;
-- reverse: create trigger "projects_updated_at"
DROP TRIGGER `projects_updated_at`;
-- reverse: create index "projects_name" to table: "projects"
DROP INDEX `projects_name`;
-- reverse: create "projects" table
DROP TABLE `projects`;
-- reverse: create index "organization_invites_token" to table: "organization_invites"
DROP INDEX `organization_invites_token`;
-- reverse: create "organization_invites" table
DROP TABLE `organization_invites`;
-- reverse: create trigger "organization_members_updated_at"
DROP TRIGGER `organization_members_updated_at`;
-- reverse: create "organization_members" table
DROP TABLE `organization_members`;
-- reverse: create trigger "organization_updated_at"
DROP TRIGGER `organization_updated_at`;
-- reverse: create index "organization_slug" to table: "organization"
DROP INDEX `organization_slug`;
-- reverse: create index "organization_name" to table: "organization"
DROP INDEX `organization_name`;
-- reverse: create "organization" table
DROP TABLE `organization`;
-- reverse: create trigger "jwt_tokens_updated_at"
DROP TRIGGER `jwt_tokens_updated_at`;
-- reverse: create "jwt_tokens" table
DROP TABLE `jwt_tokens`;
-- reverse: create "two_factor" table
DROP TABLE `two_factor`;
-- reverse: create trigger "users_updated_at"
DROP TRIGGER `users_updated_at`;
-- reverse: create index "users_email" to table: "users"
DROP INDEX `users_email`;
-- reverse: create "users" table
DROP TABLE `users`;
-- reverse: create "group_policy" table
DROP TABLE `group_policy`;
-- reverse: create index "policy_action" to table: "policy"
DROP INDEX `policy_action`;
-- reverse: create "policy" table
DROP TABLE `policy`;
-- reverse: create trigger "groups_updated_at"
DROP TRIGGER `groups_updated_at`;
-- reverse: create index "groups_name" to table: "groups"
DROP INDEX `groups_name`;
-- reverse: create "groups" table
DROP TABLE `groups`;
