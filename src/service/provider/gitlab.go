package provider

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/db/repos"
)

type GitlabProviderService struct {
	queries *repos.Queries
}

func NewGitlabProviderService(queries *repos.Queries) *GitlabProviderService {
	return &GitlabProviderService{queries}
}

// UpdateGitlabProvider updates the GitLab configuration.
func (s *GitlabProviderService) UpdateGitlabProvider(
	ctx context.Context,
	dto *dtos.SaveGitlabDto,
) (*repos.GitlabProvider, error) {
	child, err := s.queries.GetGitlabProviderByGitProviderID(
		ctx,
		dto.GitProviderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"GitLab provider configuration not found",
				"GITLAB_PROVIDER_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to fetch GitLab provider config",
			"GITLAB_PROVIDER_FETCH_ERROR",
			throw.WithCause(err),
		)
	}

	updated, err := s.queries.UpdateGitlabProvider(
		ctx,
		repos.UpdateGitlabProviderParams{
			ID:                child.ID,
			GitlabUrl:         dto.GitlabUrl,
			GitlabInternalUrl: dto.GitlabInternalUrl,
			ApplicationID:     dto.ApplicationID,
			RedirectUri:       dto.RedirectUri,
			Secret:            dto.Secret,
			GroupName:         dto.GroupName,
			AccessToken:       child.AccessToken,
			RefreshToken:      child.RefreshToken,
			ExpiresAt:         child.ExpiresAt,
		},
	)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to update GitLab provider config",
			"GITLAB_PROVIDER_UPDATE_ERROR",
			throw.WithCause(err),
		)
	}

	return &updated, nil
}

type tokenExchangeResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// ExchangeOAuthToken exchanges the authorization code for access and refresh tokens.
func (s *GitlabProviderService) ExchangeOAuthToken(
	ctx context.Context,
	providerID int64,
	code string,
) error {
	child, err := s.queries.GetGitlabProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return throw.NotFoundError(
			"GitLab provider not found",
			"GITLAB_PROVIDER_NOT_FOUND",
		)
	}

	baseUrl := child.GitlabUrl
	if child.GitlabInternalUrl != nil {
		baseUrl = *child.GitlabInternalUrl
	}

	tokenUrl := fmt.Sprintf("%s/oauth/token", strings.TrimSuffix(baseUrl, "/"))

	form := url.Values{}
	form.Set("client_id", *child.ApplicationID)
	form.Set("client_secret", *child.Secret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	redirectUri := fmt.Sprintf("%s?gitlabId=%d", *child.RedirectUri, providerID)
	form.Set("redirect_uri", redirectUri)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		tokenUrl,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return throw.InternalServerError(
			"Failed to execute OAuth token exchange request",
			"GITLAB_TOKEN_EXCHANGE_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return throw.BadRequestError(
			fmt.Sprintf(
				"GitLab token exchange failed: %v",
				errData["error_description"],
			),
			"GITLAB_TOKEN_EXCHANGE_FAILED",
		)
	}

	var tokenRes tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return throw.InternalServerError(
			"Failed to parse GitLab token response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	expiresAt := time.Now().Unix() + tokenRes.ExpiresIn
	_, err = s.queries.UpdateGitlabProvider(
		ctx,
		repos.UpdateGitlabProviderParams{
			ID:                child.ID,
			GitlabUrl:         child.GitlabUrl,
			GitlabInternalUrl: child.GitlabInternalUrl,
			ApplicationID:     child.ApplicationID,
			RedirectUri:       child.RedirectUri,
			Secret:            child.Secret,
			GroupName:         child.GroupName,
			AccessToken:       &tokenRes.AccessToken,
			RefreshToken:      &tokenRes.RefreshToken,
			ExpiresAt:         &expiresAt,
		},
	)
	if err != nil {
		return throw.InternalServerError(
			"Failed to save GitLab access tokens",
			"GITLAB_TOKEN_SAVE_ERROR",
			throw.WithCause(err),
		)
	}

	return nil
}

// RefreshAccessToken checks if the GitLab token is expired and refreshes it if needed.
func (s *GitlabProviderService) RefreshAccessToken(
	ctx context.Context,
	providerID int64,
) (string, error) {
	child, err := s.queries.GetGitlabProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return "", throw.NotFoundError(
			"GitLab provider not found",
			"GITLAB_PROVIDER_NOT_FOUND",
		)
	}

	if child.AccessToken == nil || child.RefreshToken == nil {
		return "", throw.BadRequestError(
			"GitLab provider is not authenticated. Please login via OAuth.",
			"GITLAB_NOT_AUTHENTICATED",
		)
	}

	now := time.Now().Unix()
	safetyMargin := int64(300) // 5 minutes
	if child.ExpiresAt != nil && now+safetyMargin < *child.ExpiresAt {
		return *child.AccessToken, nil
	}

	// Refresh token
	baseUrl := child.GitlabUrl
	if child.GitlabInternalUrl != nil {
		baseUrl = *child.GitlabInternalUrl
	}

	tokenUrl := fmt.Sprintf("%s/oauth/token", strings.TrimSuffix(baseUrl, "/"))

	form := url.Values{}
	form.Set("client_id", *child.ApplicationID)
	form.Set("client_secret", *child.Secret)
	form.Set("refresh_token", *child.RefreshToken)
	form.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		tokenUrl,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", throw.InternalServerError(
			"Failed to refresh GitLab token",
			"GITLAB_TOKEN_REFRESH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", throw.BadRequestError(
			"GitLab token refresh failed. Please re-authenticate.",
			"GITLAB_TOKEN_REFRESH_FAILED",
		)
	}

	var tokenRes tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", throw.InternalServerError(
			"Failed to parse GitLab token refresh response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	expiresAt := time.Now().Unix() + tokenRes.ExpiresIn
	_, err = s.queries.UpdateGitlabProvider(
		ctx,
		repos.UpdateGitlabProviderParams{
			ID:                child.ID,
			GitlabUrl:         child.GitlabUrl,
			GitlabInternalUrl: child.GitlabInternalUrl,
			ApplicationID:     child.ApplicationID,
			RedirectUri:       child.RedirectUri,
			Secret:            child.Secret,
			GroupName:         child.GroupName,
			AccessToken:       &tokenRes.AccessToken,
			RefreshToken:      &tokenRes.RefreshToken,
			ExpiresAt:         &expiresAt,
		},
	)
	if err != nil {
		return "", throw.InternalServerError(
			"Failed to save refreshed GitLab access tokens",
			"GITLAB_TOKEN_SAVE_ERROR",
			throw.WithCause(err),
		)
	}

	return tokenRes.AccessToken, nil
}

type gitlabProject struct {
	ID                int64  `json:"id"`
	PathWithNamespace string `json:"path_with_namespace"`
	HttpUrlToRepo     string `json:"http_url_to_repo"`
}

// GetGitlabRepositories lists all GitLab projects accessible to the authenticated user.
func (s *GitlabProviderService) GetGitlabRepositories(
	ctx context.Context,
	providerID int64,
) ([]GitRepository, error) {
	token, err := s.RefreshAccessToken(ctx, providerID)
	if err != nil {
		return nil, err
	}

	child, err := s.queries.GetGitlabProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	// Fetch projects from GitLab
	urlStr := fmt.Sprintf(
		"%s/api/v4/projects?membership=true&per_page=100&simple=true",
		strings.TrimSuffix(child.GitlabUrl, "/"),
	)
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to fetch GitLab repositories",
			"GITLAB_REPOS_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, throw.BadRequestError(
			"Failed to fetch repositories from GitLab",
			"GITLAB_REPOS_FETCH_FAILED",
		)
	}

	var projects []gitlabProject
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, throw.InternalServerError(
			"Failed to parse GitLab projects response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	reposList := make([]GitRepository, 0, len(projects))
	groupName := ""
	if child.GroupName != nil {
		groupName = strings.ToLower(*child.GroupName)
	}

	for _, p := range projects {
		// Filter by group name if configured (matches Dokploy filter)
		if groupName != "" &&
			!strings.HasPrefix(
				strings.ToLower(p.PathWithNamespace),
				groupName+"/",
			) {
			continue
		}
		reposList = append(reposList, GitRepository{
			Name: p.PathWithNamespace,
			URL:  p.HttpUrlToRepo,
		})
	}

	return reposList, nil
}

type gitlabBranch struct {
	Name string `json:"name"`
}

// GetGitlabBranches lists all branches in the selected GitLab project.
func (s *GitlabProviderService) GetGitlabBranches(
	ctx context.Context,
	providerID int64,
	projectPath string,
) ([]GitBranch, error) {
	token, err := s.RefreshAccessToken(ctx, providerID)
	if err != nil {
		return nil, err
	}

	child, err := s.queries.GetGitlabProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	// Escape path namespace (e.g. owner/repo -> owner%2Frepo)
	escapedPath := url.PathEscape(projectPath)

	urlStr := fmt.Sprintf(
		"%s/api/v4/projects/%s/repository/branches?per_page=100",
		strings.TrimSuffix(child.GitlabUrl, "/"),
		escapedPath,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to fetch GitLab branches",
			"GITLAB_BRANCHES_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, throw.BadRequestError(
			"Failed to fetch branches from GitLab",
			"GITLAB_BRANCHES_FETCH_FAILED",
		)
	}

	var branches []gitlabBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, throw.InternalServerError(
			"Failed to parse GitLab branches response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	branchesList := make([]GitBranch, len(branches))
	for i, b := range branches {
		branchesList[i] = GitBranch{
			Name: b.Name,
		}
	}

	return branchesList, nil
}
