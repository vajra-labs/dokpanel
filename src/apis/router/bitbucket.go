package router

import (
	"goploy/src/apis/guard"
	"goploy/src/apis/handler"

	"github.com/gofiber/fiber/v3"
)

// BitbucketRouter defines HTTP routes for Bitbucket provider endpoints.
func BitbucketRouter(
	app fiber.Router,
	handler *handler.BitbucketHandler,
	guard *guard.Guard,
) {
	bitbucket := app.Group("/bitbucket", guard.Auth())
	bitbucket.Put("/", handler.Update)
	bitbucket.Get("/:id/repos", handler.GetRepositories)
	bitbucket.Get("/:id/branches", handler.GetBranches)
}
