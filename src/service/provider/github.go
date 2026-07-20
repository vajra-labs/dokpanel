package provider

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/db/repos"

	"github.com/golang-jwt/jwt/v5"
)

type GithubProviderService struct {
	queries *repos.Queries
}

func NewGithubProviderService(queries *repos.Queries) *GithubProviderService {
	return &GithubProviderService{queries}
}

// UpdateGithubProvider updates the GitHub app configuration.
func (s *GithubProviderService) UpdateGithubProvider(
	ctx context.Context,
	dto *dtos.SaveGithubDto,
) (*repos.GithubProvider, error) {
	child, err := s.queries.GetGithubProviderByGitProviderID(
		ctx,
		dto.GitProviderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"GitHub provider configuration not found",
				"GITHUB_PROVIDER_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to fetch GitHub provider config",
			"GITHUB_PROVIDER_FETCH_ERROR",
			throw.WithCause(err),
		)
	}

	updated, err := s.queries.UpdateGithubProvider(
		ctx,
		repos.UpdateGithubProviderParams{
			ID:                   child.ID,
			GithubAppName:        dto.GithubAppName,
			GithubAppID:          dto.GithubAppID,
			GithubClientID:       dto.GithubClientID,
			GithubClientSecret:   dto.GithubClientSecret,
			GithubInstallationID: dto.GithubInstallationID,
			GithubPrivateKey:     dto.GithubPrivateKey,
			GithubWebhookSecret:  dto.GithubWebhookSecret,
		},
	)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to update GitHub provider config",
			"GITHUB_PROVIDER_UPDATE_ERROR",
			throw.WithCause(err),
		)
	}

	return &updated, nil
}

// GetInstallationToken generates a JWT signed with the GitHub App private key and exchanges it for a temporary Installation Access Token.
func (s *GithubProviderService) GetInstallationToken(
	ctx context.Context,
	providerID int64,
) (string, error) {
	child, err := s.queries.GetGithubProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return "", throw.NotFoundError(
			"GitHub app configuration not found",
			"GITHUB_PROVIDER_NOT_FOUND",
		)
	}

	if child.GithubAppID == nil || child.GithubPrivateKey == nil ||
		child.GithubInstallationID == nil {
		return "", throw.BadRequestError(
			"GitHub app is not fully configured (App ID, Private Key and Installation ID are required)",
			"GITHUB_APP_NOT_CONFIGURED",
		)
	}

	// 1. Generate JWT Token
	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(*child.GithubPrivateKey))
	if err != nil {
		return "", throw.BadRequestError(
			"Invalid GitHub private key format (PEM expected)",
			"GITHUB_INVALID_PRIVATE_KEY",
			throw.WithCause(err),
		)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Unix() - 60,
		"exp": now.Add(10 * time.Minute).Unix(),
		"iss": *child.GithubAppID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	jwtString, err := token.SignedString(key)
	if err != nil {
		return "", throw.InternalServerError(
			"Failed to sign GitHub JWT token",
			"GITHUB_JWT_SIGN_ERROR",
			throw.WithCause(err),
		)
	}

	// 2. Exchange JWT for Installation Access Token
	url := fmt.Sprintf(
		"https://api.github.com/app/installations/%s/access_tokens",
		*child.GithubInstallationID,
	)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}

	req.Header.Set("Authorization", "Bearer "+jwtString)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "goploy")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", throw.InternalServerError(
			"Failed to exchange GitHub JWT",
			"GITHUB_TOKEN_EXCHANGE_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		var errData map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&errData)
		return "", throw.BadRequestError(
			fmt.Sprintf("GitHub token exchange failed: %v", errData["message"]),
			"GITHUB_TOKEN_EXCHANGE_FAILED",
		)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", throw.InternalServerError(
			"Failed to parse GitHub token response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	return result.Token, nil
}

type githubRepo struct {
	FullName string `json:"full_name"`
	CloneURL string `json:"clone_url"`
}

// GetGithubRepositories lists all repositories accessible to the GitHub App installation.
func (s *GithubProviderService) GetGithubRepositories(
	ctx context.Context,
	providerID int64,
) ([]GitRepository, error) {
	token, err := s.GetInstallationToken(ctx, providerID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://api.github.com/installation/repositories?per_page=100",
		nil,
	)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "goploy")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to fetch GitHub repositories",
			"GITHUB_REPOS_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, throw.BadRequestError(
			"Failed to fetch repositories from GitHub",
			"GITHUB_REPOS_FETCH_FAILED",
		)
	}

	var result struct {
		Repositories []githubRepo `json:"repositories"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, throw.InternalServerError(
			"Failed to parse GitHub repositories response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	reposList := make([]GitRepository, len(result.Repositories))
	for i, r := range result.Repositories {
		reposList[i] = GitRepository{
			Name: r.FullName,
			URL:  r.CloneURL,
		}
	}

	return reposList, nil
}

type githubBranch struct {
	Name string `json:"name"`
}

// GetGithubBranches lists all branches in the selected GitHub repository.
func (s *GithubProviderService) GetGithubBranches(
	ctx context.Context,
	providerID int64,
	owner, repo string,
) ([]GitBranch, error) {
	token, err := s.GetInstallationToken(ctx, providerID)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/branches?per_page=100",
		owner,
		repo,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "goploy")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to fetch GitHub branches",
			"GITHUB_BRANCHES_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, throw.BadRequestError(
			"Failed to fetch branches from GitHub",
			"GITHUB_BRANCHES_FETCH_FAILED",
		)
	}

	var result []githubBranch
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, throw.InternalServerError(
			"Failed to parse GitHub branches response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	branchesList := make([]GitBranch, len(result))
	for i, b := range result {
		branchesList[i] = GitBranch{
			Name: b.Name,
		}
	}

	return branchesList, nil
}
