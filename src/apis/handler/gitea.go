package handler

import (
	"strconv"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/service/provider"

	"github.com/gofiber/fiber/v3"
)

type GiteaHandler struct {
	gitea       *provider.GiteaProviderService
	gitProvider *provider.GitProviderService
}

func NewGiteaHandler(
	gitea *provider.GiteaProviderService,
	gitProvider *provider.GitProviderService,
) *GiteaHandler {
	return &GiteaHandler{
		gitea:       gitea,
		gitProvider: gitProvider,
	}
}

// Update handles PUT /api/gitea
func (h *GiteaHandler) Update(ctx fiber.Ctx) error {
	var body dtos.SaveGiteaDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	_, err := h.gitea.UpdateGiteaProvider(ctx.Context(), &body)
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

// GetRepositories handles GET /api/gitea/:id/repos
func (h *GiteaHandler) GetRepositories(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	list, err := h.gitea.GetGiteaRepositories(ctx.Context(), id)
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

// GetBranches handles GET /api/gitea/:id/branches
func (h *GiteaHandler) GetBranches(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	owner := ctx.Query("owner")
	repo := ctx.Query("repo")
	if owner == "" || repo == "" {
		return throw.BadRequestError(
			"owner and repo query parameters are required",
			"MISSING_QUERY_PARAMS",
		)
	}
	list, err := h.gitea.GetGiteaBranches(ctx.Context(), id, owner, repo)
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

// Callback handles GET /api/gitea/:id/callback
func (h *GiteaHandler) Callback(ctx fiber.Ctx) error {
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
	if err := h.gitea.ExchangeOAuthToken(ctx.Context(), id, code); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"success": true})
}
