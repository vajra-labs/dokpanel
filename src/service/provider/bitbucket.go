package provider

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/db/repos"
)

type BitbucketProviderService struct {
	queries *repos.Queries
}

func NewBitbucketProviderService(
	queries *repos.Queries,
) *BitbucketProviderService {
	return &BitbucketProviderService{queries}
}

// UpdateBitbucketProvider updates the Bitbucket app password/token configuration.
func (s *BitbucketProviderService) UpdateBitbucketProvider(
	ctx context.Context,
	dto *dtos.SaveBitbucketDto,
) (*repos.BitbucketProvider, error) {
	child, err := s.queries.GetBitbucketProviderByGitProviderID(
		ctx,
		dto.GitProviderID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"Bitbucket provider configuration not found",
				"BITBUCKET_PROVIDER_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to fetch Bitbucket provider config",
			"BITBUCKET_PROVIDER_FETCH_ERROR",
			throw.WithCause(err),
		)
	}

	updated, err := s.queries.UpdateBitbucketProvider(
		ctx,
		repos.UpdateBitbucketProviderParams{
			ID:                     child.ID,
			BitbucketUsername:      dto.BitbucketUsername,
			BitbucketEmail:         dto.BitbucketEmail,
			AppPassword:            dto.AppPassword,
			ApiToken:               dto.ApiToken,
			BitbucketWorkspaceName: dto.BitbucketWorkspaceName,
		},
	)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to update Bitbucket provider config",
			"BITBUCKET_PROVIDER_UPDATE_ERROR",
			throw.WithCause(err),
		)
	}

	return &updated, nil
}

type bitbucketRepoValues struct {
	FullName string `json:"full_name"`
	Links    struct {
		Clone []struct {
			Name string `json:"name"`
			Href string `json:"href"`
		} `json:"clone"`
	} `json:"links"`
}

type bitbucketReposRes struct {
	Values []bitbucketRepoValues `json:"values"`
}

// GetBitbucketRepositories lists all Bitbucket repositories in the configured workspace.
func (s *BitbucketProviderService) GetBitbucketRepositories(
	ctx context.Context,
	providerID int64,
) ([]GitRepository, error) {
	child, err := s.queries.GetBitbucketProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return nil, throw.NotFoundError(
			"Bitbucket provider not found",
			"BITBUCKET_PROVIDER_NOT_FOUND",
		)
	}

	workspace := ""
	if child.BitbucketWorkspaceName != nil {
		workspace = *child.BitbucketWorkspaceName
	}
	if workspace == "" {
		return nil, throw.BadRequestError(
			"Bitbucket Workspace Name is required to list repositories",
			"BITBUCKET_WORKSPACE_REQUIRED",
		)
	}

	urlStr := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s?pagelen=100",
		workspace,
	)
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create HTTP request",
			"HTTP_REQUEST_ERROR",
			throw.WithCause(err),
		)
	}

	// Auth: Prefer API Token (Bearer), fallback to Username + App Password (Basic)
	if child.ApiToken != nil && *child.ApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+*child.ApiToken)
	} else if child.BitbucketUsername != nil && child.AppPassword != nil && *child.BitbucketUsername != "" && *child.AppPassword != "" {
		req.SetBasicAuth(*child.BitbucketUsername, *child.AppPassword)
	} else {
		return nil, throw.BadRequestError(
			"Bitbucket credentials are not fully configured (App Password or API Token is required)",
			"BITBUCKET_NOT_CONFIGURED",
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to fetch Bitbucket repositories",
			"BITBUCKET_REPOS_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, throw.BadRequestError(
			"Failed to fetch repositories from Bitbucket (check credentials/workspace name)",
			"BITBUCKET_REPOS_FETCH_FAILED",
		)
	}

	var result bitbucketReposRes
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, throw.InternalServerError(
			"Failed to parse Bitbucket repositories response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	reposList := make([]GitRepository, len(result.Values))
	for i, r := range result.Values {
		cloneUrl := ""
		for _, c := range r.Links.Clone {
			if c.Name == "https" {
				cloneUrl = c.Href
				break
			}
		}
		if cloneUrl == "" && len(r.Links.Clone) > 0 {
			cloneUrl = r.Links.Clone[0].Href
		}
		reposList[i] = GitRepository{
			Name: r.FullName,
			URL:  cloneUrl,
		}
	}

	return reposList, nil
}

type bitbucketBranch struct {
	Name string `json:"name"`
}

type bitbucketBranchesRes struct {
	Values []bitbucketBranch `json:"values"`
}

// GetBitbucketBranches lists all branches in the selected Bitbucket repository.
func (s *BitbucketProviderService) GetBitbucketBranches(
	ctx context.Context,
	providerID int64,
	owner, repo string,
) ([]GitBranch, error) {
	child, err := s.queries.GetBitbucketProviderByGitProviderID(ctx, providerID)
	if err != nil {
		return nil, throw.NotFoundError(
			"Bitbucket provider not found",
			"BITBUCKET_PROVIDER_NOT_FOUND",
		)
	}

	urlStr := fmt.Sprintf(
		"https://api.bitbucket.org/2.0/repositories/%s/%s/refs/branches?pagelen=100",
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

	// Auth: Prefer API Token (Bearer), fallback to Username + App Password (Basic)
	if child.ApiToken != nil && *child.ApiToken != "" {
		req.Header.Set("Authorization", "Bearer "+*child.ApiToken)
	} else if child.BitbucketUsername != nil && child.AppPassword != nil && *child.BitbucketUsername != "" && *child.AppPassword != "" {
		req.SetBasicAuth(*child.BitbucketUsername, *child.AppPassword)
	} else {
		return nil, throw.BadRequestError(
			"Bitbucket credentials are not fully configured (App Password or API Token is required)",
			"BITBUCKET_NOT_CONFIGURED",
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to fetch Bitbucket branches",
			"BITBUCKET_BRANCHES_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, throw.BadRequestError(
			"Failed to fetch branches from Bitbucket",
			"BITBUCKET_BRANCHES_FETCH_FAILED",
		)
	}

	var result bitbucketBranchesRes
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, throw.InternalServerError(
			"Failed to parse Bitbucket branches response",
			"JSON_PARSE_ERROR",
			throw.WithCause(err),
		)
	}

	branchesList := make([]GitBranch, len(result.Values))
	for i, b := range result.Values {
		branchesList[i] = GitBranch{
			Name: b.Name,
		}
	}

	return branchesList, nil
}
