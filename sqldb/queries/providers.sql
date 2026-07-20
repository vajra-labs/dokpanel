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

-- name: CreateGitProvider :one
INSERT INTO git_providers (
	name, provider_type, shared
) VALUES (?, ?, ?)
RETURNING *;

-- name: GetGitProviderByID :one
SELECT * FROM git_providers WHERE id = ? LIMIT 1;

-- name: ListGitProviders :many
SELECT * FROM git_providers ORDER BY created_at DESC;

-- name: UpdateGitProvider :one
UPDATE git_providers
SET name = ?, shared = ?, updated_at = (strftime('%s', 'now'))
WHERE id = ?
RETURNING *;

-- name: DeleteGitProvider :exec
DELETE FROM git_providers WHERE id = ?;

-- name: CreateGithubProvider :one
INSERT INTO github_providers (
	github_app_name, github_app_id, github_client_id, github_client_secret,
	github_installation_id, github_private_key, github_webhook_secret, git_provider_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateGithubProvider :one
UPDATE github_providers
SET github_app_name = ?, github_app_id = ?, github_client_id = ?, github_client_secret = ?,
	github_installation_id = ?, github_private_key = ?, github_webhook_secret = ?, updated_at = (strftime('%s', 'now'))
WHERE id = ?
RETURNING *;

-- name: CreateGitlabProvider :one
INSERT INTO gitlab_providers (
	gitlab_url, gitlab_internal_url, application_id, redirect_uri, secret,
	access_token, refresh_token, group_name, expires_at, git_provider_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateGitlabProvider :one
UPDATE gitlab_providers
SET gitlab_url = ?, gitlab_internal_url = ?, application_id = ?, redirect_uri = ?, secret = ?,
	access_token = ?, refresh_token = ?, group_name = ?, expires_at = ?, updated_at = (strftime('%s', 'now'))
WHERE id = ?
RETURNING *;

-- name: CreateGiteaProvider :one
INSERT INTO gitea_providers (
	gitea_url, gitea_internal_url, redirect_uri, client_id, client_secret,
	access_token, refresh_token, expires_at, scopes, last_authenticated_at, git_provider_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateGiteaProvider :one
UPDATE gitea_providers
SET gitea_url = ?, gitea_internal_url = ?, redirect_uri = ?, client_id = ?, client_secret = ?,
	access_token = ?, refresh_token = ?, expires_at = ?, scopes = ?, last_authenticated_at = ?, updated_at = (strftime('%s', 'now'))
WHERE id = ?
RETURNING *;

-- name: CreateBitbucketProvider :one
INSERT INTO bitbucket_providers (
	bitbucket_username, bitbucket_email, app_password, api_token, bitbucket_workspace_name, git_provider_id
) VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateBitbucketProvider :one
UPDATE bitbucket_providers
SET bitbucket_username = ?, bitbucket_email = ?, app_password = ?, api_token = ?, bitbucket_workspace_name = ?, updated_at = (strftime('%s', 'now'))
WHERE id = ?
RETURNING *;
