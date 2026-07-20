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

type GiteaProviderService struct {
	queries *repos.Queries
}

func NewGiteaProviderService(queries *repos.Queries) *GiteaProviderService {
	return &GiteaProviderService{queries}
}

// UpdateGiteaProvider updates the Gitea configuration.
func (s *GiteaProviderService) UpdateGiteaProvider(
	ctx context.Context,
	dto *dtos.SaveGiteaDto,
) (*repos.GiteaProvider, error) {
	child, err := s.queries.GetGiteaProviderByGitProviderID(
		ctx,
		dto.GitProviderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"Gitea provider configuration not found",
				"GITEA_PROVIDER_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to fetch Gitea provider config",
			"GITEA_PROVIDER_FETCH_ERROR",
			throw.WithCause(err),
		)
	}

	updated, err := s.queries.UpdateGiteaProvider(
		ctx,
		repos.UpdateGiteaProviderParams{
			ID:                  child.ID,
			GiteaUrl:            dto.GiteaUrl,
			GiteaInternalUrl:    dto.GiteaInternalUrl,
			RedirectUri:         dto.RedirectUri,
			ClientID:            dto.ClientID,
			ClientSecret:        dto.ClientSecret,
			AccessToken:         child.AccessToken,
			RefreshToken:        child.RefreshToken,
			ExpiresAt:           child.ExpiresAt,
			Scopes:              child.Scopes,
			LastAuthenticatedAt: child.LastAuthenticatedAt,
		},
	)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to update Gitea provider config",
			"GITEA_PROVIDER_UPDATE_ERROR",
			throw.WithCause(err),
		)
	}

	return &updated, nil
}

// ExchangeOAuthToken exchanges the authorization code for access and refresh tokens.
func (s *GiteaProviderService) ExchangeOAuthToken(
	ctx context.Context,
	providerID int64,
	code string,
) error {
	child, err := s.queries.GetGiteaProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return throw.NotFoundError(
			"Gitea provider not found",
			"GITEA_PROVIDER_NOT_FOUND",
		)
	}

	baseUrl := child.GiteaUrl
	if child.GiteaInternalUrl != nil {
		baseUrl = *child.GiteaInternalUrl
	}

	tokenUrl := fmt.Sprintf(
		"%s/login/oauth/access_token",
		strings.TrimSuffix(baseUrl, "/"),
	)

	form := url.Values{}
	form.Set("client_id", *child.ClientID)
	form.Set("client_secret", *child.ClientSecret)
	form.Set("code", code)
	form.Set("grant_type", "authorization_code")
	redirectUri := fmt.Sprintf("%s?giteaId=%d", *child.RedirectUri, providerID)
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
			"Failed to execute Gitea OAuth token exchange request",
			"GITEA_TOKEN_EXCHANGE_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return throw.BadRequestError(
			fmt.Sprintf(
				"Gitea token exchange failed: %v",
				errData["error_description"],
			),
			"GITEA_TOKEN_EXCHANGE_FAILED",
		)
	}

	var tokenRes tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return throw.InternalServerError(
			"Failed to parse Gitea token response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	expiresAt := time.Now().Unix() + tokenRes.ExpiresIn
	now := time.Now().Unix()
	_, err = s.queries.UpdateGiteaProvider(ctx, repos.UpdateGiteaProviderParams{
		ID:                  child.ID,
		GiteaUrl:            child.GiteaUrl,
		GiteaInternalUrl:    child.GiteaInternalUrl,
		RedirectUri:         child.RedirectUri,
		ClientID:            child.ClientID,
		ClientSecret:        child.ClientSecret,
		AccessToken:         &tokenRes.AccessToken,
		RefreshToken:        &tokenRes.RefreshToken,
		ExpiresAt:           &expiresAt,
		Scopes:              child.Scopes,
		LastAuthenticatedAt: &now,
	})
	if err != nil {
		return throw.InternalServerError(
			"Failed to save Gitea access tokens",
			"GITEA_TOKEN_SAVE_ERROR",
			throw.WithCause(err),
		)
	}

	return nil
}

// RefreshAccessToken checks if the Gitea token is expired and refreshes it if needed.
func (s *GiteaProviderService) RefreshAccessToken(
	ctx context.Context,
	providerID int64,
) (string, error) {
	child, err := s.queries.GetGiteaProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return "", throw.NotFoundError(
			"Gitea provider not found",
			"GITEA_PROVIDER_NOT_FOUND",
		)
	}

	if child.AccessToken == nil || child.RefreshToken == nil {
		return "", throw.BadRequestError(
			"Gitea provider is not authenticated. Please login via OAuth.",
			"GITEA_NOT_AUTHENTICATED",
		)
	}

	now := time.Now().Unix()
	safetyMargin := int64(300) // 5 minutes
	if child.ExpiresAt != nil && now+safetyMargin < *child.ExpiresAt {
		return *child.AccessToken, nil
	}

	// Refresh token
	baseUrl := child.GiteaUrl
	if child.GiteaInternalUrl != nil {
		baseUrl = *child.GiteaInternalUrl
	}

	tokenUrl := fmt.Sprintf(
		"%s/login/oauth/access_token",
		strings.TrimSuffix(baseUrl, "/"),
	)

	form := url.Values{}
	form.Set("client_id", *child.ClientID)
	form.Set("client_secret", *child.ClientSecret)
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
			"Failed to refresh Gitea token",
			"GITEA_TOKEN_REFRESH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", throw.BadRequestError(
			"Gitea token refresh failed. Please re-authenticate.",
			"GITEA_TOKEN_REFRESH_FAILED",
		)
	}

	var tokenRes tokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", throw.InternalServerError(
			"Failed to parse Gitea token refresh response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	expiresAt := time.Now().Unix() + tokenRes.ExpiresIn
	nowUnix := time.Now().Unix()
	_, err = s.queries.UpdateGiteaProvider(ctx, repos.UpdateGiteaProviderParams{
		ID:                  child.ID,
		GiteaUrl:            child.GiteaUrl,
		GiteaInternalUrl:    child.GiteaInternalUrl,
		RedirectUri:         child.RedirectUri,
		ClientID:            child.ClientID,
		ClientSecret:        child.ClientSecret,
		AccessToken:         &tokenRes.AccessToken,
		RefreshToken:        &tokenRes.RefreshToken,
		ExpiresAt:           &expiresAt,
		Scopes:              child.Scopes,
		LastAuthenticatedAt: &nowUnix,
	})
	if err != nil {
		return "", throw.InternalServerError(
			"Failed to save refreshed Gitea access tokens",
			"GITEA_TOKEN_SAVE_ERROR",
			throw.WithCause(err),
		)
	}

	return tokenRes.AccessToken, nil
}

type giteaRepo struct {
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
}

// GetGiteaRepositories lists all Gitea repositories accessible to the authenticated user.
func (s *GiteaProviderService) GetGiteaRepositories(
	ctx context.Context,
	providerID int64,
) ([]GitRepository, error) {
	token, err := s.RefreshAccessToken(ctx, providerID)
	if err != nil {
		return nil, err
	}

	child, err := s.queries.GetGiteaProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf(
		"%s/api/v1/user/repos?per_page=100",
		strings.TrimSuffix(child.GiteaUrl, "/"),
	)
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}

	req.Header.Set("Authorization", "token "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to fetch Gitea repositories",
			"GITEA_REPOS_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, throw.BadRequestError(
			"Failed to fetch repositories from Gitea",
			"GITEA_REPOS_FETCH_FAILED",
		)
	}

	var reposList []giteaRepo
	if err := json.NewDecoder(resp.Body).Decode(&reposList); err != nil {
		return nil, throw.InternalServerError(
			"Failed to parse Gitea repositories response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	result := make([]GitRepository, len(reposList))
	for i, r := range reposList {
		result[i] = GitRepository{
			Name: r.FullName,
			URL:  r.CloneURL,
		}
	}

	return result, nil
}

type giteaBranch struct {
	Name string `json:"name"`
}

// GetGiteaBranches lists all branches in the Gitea repository.
func (s *GiteaProviderService) GetGiteaBranches(
	ctx context.Context,
	providerID int64,
	owner, repo string,
) ([]GitBranch, error) {
	token, err := s.RefreshAccessToken(ctx, providerID)
	if err != nil {
		return nil, err
	}

	child, err := s.queries.GetGiteaProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf(
		"%s/api/v1/repos/%s/%s/branches?per_page=100",
		strings.TrimSuffix(child.GiteaUrl, "/"),
		owner,
		repo,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}

	req.Header.Set("Authorization", "token "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to fetch Gitea branches",
			"GITEA_BRANCHES_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, throw.BadRequestError(
			"Failed to fetch branches from Gitea",
			"GITEA_BRANCHES_FETCH_FAILED",
		)
	}

	var branches []giteaBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, throw.InternalServerError(
			"Failed to parse Gitea branches response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	result := make([]GitBranch, len(branches))
	for i, b := range branches {
		result[i] = GitBranch{
			Name: b.Name,
		}
	}

	return result, nil
}
