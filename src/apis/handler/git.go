package handler

import (
	"strconv"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/service/provider"

	"github.com/gofiber/fiber/v3"
)

type GitProviderHandler struct {
	gitProvider *provider.GitProviderService
}

func NewGitProviderHandler(
	gitProvider *provider.GitProviderService,
) *GitProviderHandler {
	return &GitProviderHandler{
		gitProvider: gitProvider,
	}
}

// Create handles POST /api/git-providers
func (h *GitProviderHandler) Create(ctx fiber.Ctx) error {
	var body dtos.CreateGitProviderDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	res, err := h.gitProvider.CreateGitProvider(ctx.Context(), &body)
	if err != nil {
		return err
	}
	fullDetails, err := h.gitProvider.GetGitProviderByID(ctx.Context(), res.ID)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(mapGitProviderRes(fullDetails))
}

// List handles GET /api/git-providers
func (h *GitProviderHandler) List(ctx fiber.Ctx) error {
	list, err := h.gitProvider.ListGitProviders(ctx.Context())
	if err != nil {
		return err
	}
	res := make([]*dtos.GitProviderResDto, len(list))
	for i, item := range list {
		res[i] = mapGitProviderRes(item)
	}
	return ctx.JSON(res)
}

// Get handles GET /api/git-providers/:id
func (h *GitProviderHandler) Get(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	res, err := h.gitProvider.GetGitProviderByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(mapGitProviderRes(res))
}

// Update handles PUT /api/git-providers/:id
func (h *GitProviderHandler) Update(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	var body dtos.UpdateGitProviderDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	res, err := h.gitProvider.UpdateGitProvider(ctx.Context(), id, &body)
	if err != nil {
		return err
	}
	fullDetails, err := h.gitProvider.GetGitProviderByID(ctx.Context(), res.ID)
	if err != nil {
		return err
	}
	return ctx.JSON(mapGitProviderRes(fullDetails))
}

// Delete handles DELETE /api/git-providers/:id
func (h *GitProviderHandler) Delete(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	if err := h.gitProvider.DeleteGitProvider(ctx.Context(), id); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"success": true})
}

func mapGitProviderRes(
	details *provider.GitProviderDetails,
) *dtos.GitProviderResDto {
	res := &dtos.GitProviderResDto{
		ID:           details.Provider.ID,
		Name:         details.Provider.Name,
		ProviderType: details.Provider.ProviderType,
		Shared:       details.Provider.Shared,
		CreatedAt:    details.Provider.CreatedAt,
		UpdatedAt:    details.Provider.UpdatedAt,
	}

	if details.Github != nil {
		isConfigured := details.Github.GithubPrivateKey != nil &&
			details.Github.GithubAppID != nil &&
			details.Github.GithubInstallationID != nil
		res.Github = &dtos.GithubProviderDto{
			ID:                   details.Github.ID,
			GithubAppName:        details.Github.GithubAppName,
			GithubAppID:          details.Github.GithubAppID,
			GithubClientID:       details.Github.GithubClientID,
			GithubInstallationID: details.Github.GithubInstallationID,
			IsConfigured:         isConfigured,
		}
	}

	if details.Gitlab != nil {
		isConfigured := details.Gitlab.AccessToken != nil &&
			details.Gitlab.RefreshToken != nil
		res.Gitlab = &dtos.GitlabProviderDto{
			ID:            details.Gitlab.ID,
			GitlabUrl:     details.Gitlab.GitlabUrl,
			ApplicationID: details.Gitlab.ApplicationID,
			GroupName:     details.Gitlab.GroupName,
			IsConfigured:  isConfigured,
		}
	}

	if details.Gitea != nil {
		isConfigured := details.Gitea.AccessToken != nil &&
			details.Gitea.RefreshToken != nil
		res.Gitea = &dtos.GiteaProviderDto{
			ID:           details.Gitea.ID,
			GiteaUrl:     details.Gitea.GiteaUrl,
			ClientID:     details.Gitea.ClientID,
			IsConfigured: isConfigured,
		}
	}

	if details.Bitbucket != nil {
		isConfigured := details.Bitbucket.AppPassword != nil ||
			details.Bitbucket.ApiToken != nil
		res.Bitbucket = &dtos.BitbucketProviderDto{
			ID:                details.Bitbucket.ID,
			BitbucketUsername: details.Bitbucket.BitbucketUsername,
			IsConfigured:      isConfigured,
		}
	}

	return res
}
