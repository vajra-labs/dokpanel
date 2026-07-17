-- name: GetGithubProviderByID :one
SELECT * FROM github_providers WHERE id = ? LIMIT 1;

-- name: GetGithubProviderByGitProviderID :one
SELECT * FROM github_providers WHERE git_provider_id = ? LIMIT 1;

-- name: GetGitlabProviderByID :one
SELECT * FROM gitlab_providers WHERE id = ? LIMIT 1;

-- name: GetGitlabProviderByGitProviderID :one
SELECT * FROM gitlab_providers WHERE git_provider_id = ? LIMIT 1;

-- name: GetGiteaProviderByID :one
SELECT * FROM gitea_providers WHERE id = ? LIMIT 1;

-- name: GetGiteaProviderByGitProviderID :one
SELECT * FROM gitea_providers WHERE git_provider_id = ? LIMIT 1;

-- name: GetBitbucketProviderByID :one
SELECT * FROM bitbucket_providers WHERE id = ? LIMIT 1;

-- name: GetBitbucketProviderByGitProviderID :one
SELECT * FROM bitbucket_providers WHERE git_provider_id = ? LIMIT 1;

-- name: CreateGithubProvider :one
INSERT INTO github_providers (
	github_app_name, github_app_id, github_client_id, github_client_secret,
	github_installation_id, github_private_key, github_webhook_secret, git_provider_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;
