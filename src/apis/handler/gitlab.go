package handler

import (
	"strconv"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/service/provider"

	"github.com/gofiber/fiber/v3"
)

type GitlabHandler struct {
	gitlab      *provider.GitlabProviderService
	gitProvider *provider.GitProviderService
}

func NewGitlabHandler(
	gitlab *provider.GitlabProviderService,
	gitProvider *provider.GitProviderService,
) *GitlabHandler {
	return &GitlabHandler{
		gitlab:      gitlab,
		gitProvider: gitProvider,
	}
}

// Update handles PUT /api/gitlab
func (h *GitlabHandler) Update(ctx fiber.Ctx) error {
	var body dtos.SaveGitlabDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	_, err := h.gitlab.UpdateGitlabProvider(ctx.Context(), &body)
	if err != nil {
		return err
	}
	fullDetails, err := h.gitProvider.GetGitProviderByID(
		ctx.Context(),
		body.GitProviderID,
	)
	if err != nil {
		return err
	}
	return ctx.JSON(mapGitProviderRes(fullDetails))
}

// GetRepositories handles GET /api/gitlab/:id/repos
func (h *GitlabHandler) GetRepositories(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	list, err := h.gitlab.GetGitlabRepositories(ctx.Context(), id)
	if err != nil {
		return err
	}
	res := make([]dtos.GitRepoDto, len(list))
	for i, r := range list {
		res[i] = dtos.GitRepoDto{
			Name: r.Name,
			URL:  r.URL,
		}
	}
	return ctx.JSON(res)
}

// GetBranches handles GET /api/gitlab/:id/branches
func (h *GitlabHandler) GetBranches(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	projectPath := ctx.Query("projectPath")
	if projectPath == "" {
		return throw.BadRequestError(
			"projectPath query parameter is required",
			"MISSING_QUERY_PARAMS",
		)
	}
	list, err := h.gitlab.GetGitlabBranches(ctx.Context(), id, projectPath)
	if err != nil {
		return err
	}
	res := make([]dtos.GitBranchDto, len(list))
	for i, b := range list {
		res[i] = dtos.GitBranchDto{
			Name: b.Name,
		}
	}
	return ctx.JSON(res)
}

// Callback handles GET /api/gitlab/:id/callback
func (h *GitlabHandler) Callback(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	code := ctx.Query("code")
	if code == "" {
		return throw.BadRequestError(
			"OAuth authorization code is required",
			"OAUTH_CODE_REQUIRED",
		)
	}
	if err := h.gitlab.ExchangeOAuthToken(ctx.Context(), id, code); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"success": true})
}
