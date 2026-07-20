package handler

import (
	"strconv"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/service/provider"

	"github.com/gofiber/fiber/v3"
)

type GithubHandler struct {
	github      *provider.GithubProviderService
	gitProvider *provider.GitProviderService
}

func NewGithubHandler(
	github *provider.GithubProviderService,
	gitProvider *provider.GitProviderService,
) *GithubHandler {
	return &GithubHandler{
		github:      github,
		gitProvider: gitProvider,
	}
}

// Update handles PUT /api/github
func (h *GithubHandler) Update(ctx fiber.Ctx) error {
	var body dtos.SaveGithubDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	_, err := h.github.UpdateGithubProvider(ctx.Context(), &body)
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

// GetRepositories handles GET /api/github/:id/repos
func (h *GithubHandler) GetRepositories(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	list, err := h.github.GetGithubRepositories(ctx.Context(), id)
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

// GetBranches handles GET /api/github/:id/branches
func (h *GithubHandler) GetBranches(ctx fiber.Ctx) error {
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
	list, err := h.github.GetGithubBranches(ctx.Context(), id, owner, repo)
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
