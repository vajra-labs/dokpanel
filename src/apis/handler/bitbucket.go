package handler

import (
	"strconv"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/service/provider"

	"github.com/gofiber/fiber/v3"
)

type BitbucketHandler struct {
	bitbucket   *provider.BitbucketProviderService
	gitProvider *provider.GitProviderService
}

func NewBitbucketHandler(
	bitbucket *provider.BitbucketProviderService,
	gitProvider *provider.GitProviderService,
) *BitbucketHandler {
	return &BitbucketHandler{
		bitbucket:   bitbucket,
		gitProvider: gitProvider,
	}
}

// Update handles PUT /api/bitbucket
func (h *BitbucketHandler) Update(ctx fiber.Ctx) error {
	var body dtos.SaveBitbucketDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	_, err := h.bitbucket.UpdateBitbucketProvider(ctx.Context(), &body)
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

// GetRepositories handles GET /api/bitbucket/:id/repos
func (h *BitbucketHandler) GetRepositories(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError(
			"Invalid Git Provider ID",
			"INVALID_GIT_PROVIDER_ID",
		)
	}
	list, err := h.bitbucket.GetBitbucketRepositories(ctx.Context(), id)
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

// GetBranches handles GET /api/bitbucket/:id/branches
func (h *BitbucketHandler) GetBranches(ctx fiber.Ctx) error {
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
	list, err := h.bitbucket.GetBitbucketBranches(
		ctx.Context(),
		id,
		owner,
		repo,
	)
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
