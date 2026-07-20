package provider

import (
	"context"
	"database/sql"
	"errors"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/db/repos"
)

type GitProviderService struct {
	queries *repos.Queries
}

func NewGitProviderService(queries *repos.Queries) *GitProviderService {
	return &GitProviderService{queries}
}

// CreateGitProvider creates the umbrella git_provider and initializes its empty specific provider configuration.
func (s *GitProviderService) CreateGitProvider(
	ctx context.Context,
	dto *dtos.CreateGitProviderDto,
) (*repos.GitProvider, error) {
	provider, err := s.queries.CreateGitProvider(
		ctx,
		repos.CreateGitProviderParams{
			Name:         dto.Name,
			ProviderType: dto.ProviderType,
			Shared:       dto.Shared,
		},
	)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create git provider record",
			"GIT_PROVIDER_CREATE_ERROR",
			throw.WithCause(err),
		)
	}

	// Initialize the empty child provider table depending on the type
	switch dto.ProviderType {
	case "GITHUB":
		_, err = s.queries.CreateGithubProvider(
			ctx,
			repos.CreateGithubProviderParams{
				GitProviderID: provider.ID,
			},
		)
	case "GITLAB":
		url := "https://gitlab.com"
		_, err = s.queries.CreateGitlabProvider(
			ctx,
			repos.CreateGitlabProviderParams{
				GitlabUrl:     url,
				GitProviderID: provider.ID,
			},
		)
	case "GITEA":
		url := "https://gitea.com"
		_, err = s.queries.CreateGiteaProvider(
			ctx,
			repos.CreateGiteaProviderParams{
				GiteaUrl:      url,
				GitProviderID: provider.ID,
			},
		)
	case "BITBUCKET":
		_, err = s.queries.CreateBitbucketProvider(
			ctx,
			repos.CreateBitbucketProviderParams{
				GitProviderID: provider.ID,
			},
		)
	}

	if err != nil {
		// Cleanup the created provider if child creation failed (simulate rollback)
		_ = s.queries.DeleteGitProvider(ctx, provider.ID)
		return nil, throw.InternalServerError(
			"Failed to initialize specific git provider configuration",
			"GIT_PROVIDER_CHILD_INIT_ERROR",
			throw.WithCause(err),
		)
	}

	return &provider, nil
}

// GetGitProviderByID retrieves a single git provider along with its typed configurations as a domain model.
func (s *GitProviderService) GetGitProviderByID(
	ctx context.Context,
	id int64,
) (*GitProviderDetails, error) {
	provider, err := s.queries.GetGitProviderByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"Git provider not found",
				"GIT_PROVIDER_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to fetch git provider",
			"GIT_PROVIDER_FETCH_ERROR",
			throw.WithCause(err),
		)
	}

	res := &GitProviderDetails{
		Provider: provider,
	}

	switch provider.ProviderType {
	case "GITHUB":
		child, err := s.queries.GetGithubProviderByGitProviderID(
			ctx,
			provider.ID,
		)
		if err == nil {
			res.Github = &child
		}
	case "GITLAB":
		child, err := s.queries.GetGitlabProviderByGitProviderID(
			ctx,
			provider.ID,
		)
		if err == nil {
			res.Gitlab = &child
		}
	case "GITEA":
		child, err := s.queries.GetGiteaProviderByGitProviderID(
			ctx,
			provider.ID,
		)
		if err == nil {
			res.Gitea = &child
		}
	case "BITBUCKET":
		child, err := s.queries.GetBitbucketProviderByGitProviderID(
			ctx,
			provider.ID,
		)
		if err == nil {
			res.Bitbucket = &child
		}
	}

	return res, nil
}

// ListGitProviders returns all git providers configured in the system.
func (s *GitProviderService) ListGitProviders(
	ctx context.Context,
) ([]*GitProviderDetails, error) {
	providers, err := s.queries.ListGitProviders(ctx)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to list git providers",
			"GIT_PROVIDER_LIST_ERROR",
			throw.WithCause(err),
		)
	}

	res := make([]*GitProviderDetails, 0, len(providers))
	for _, provider := range providers {
		// Populate child configs
		pDetails, err := s.GetGitProviderByID(ctx, provider.ID)
		if err == nil {
			res = append(res, pDetails)
		}
	}

	return res, nil
}

// UpdateGitProvider updates the name and shared status of the git provider.
func (s *GitProviderService) UpdateGitProvider(
	ctx context.Context,
	id int64,
	dto *dtos.UpdateGitProviderDto,
) (*repos.GitProvider, error) {
	provider, err := s.queries.UpdateGitProvider(
		ctx,
		repos.UpdateGitProviderParams{
			ID:     id,
			Name:   dto.Name,
			Shared: dto.Shared,
		},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"Git provider not found",
				"GIT_PROVIDER_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to update git provider",
			"GIT_PROVIDER_UPDATE_ERROR",
			throw.WithCause(err),
		)
	}
	return &provider, nil
}

// DeleteGitProvider deletes a git provider (which cascades to child credentials).
func (s *GitProviderService) DeleteGitProvider(
	ctx context.Context,
	id int64,
) error {
	err := s.queries.DeleteGitProvider(ctx, id)
	if err != nil {
		return throw.InternalServerError(
			"Failed to delete git provider",
			"GIT_PROVIDER_DELETE_ERROR",
			throw.WithCause(err),
		)
	}
	return nil
}
