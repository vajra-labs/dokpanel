package router

import (
	"goploy/src/apis/guard"
	"goploy/src/apis/handler"

	"github.com/gofiber/fiber/v3"
)

// GithubRouter defines HTTP routes for GitHub provider endpoints.
func GithubRouter(
	app fiber.Router,
	handler *handler.GithubHandler,
	guard *guard.Guard,
) {
	github := app.Group("/github", guard.Auth())
	github.Put("/", handler.Update)
	github.Get("/:id/repos", handler.GetRepositories)
	github.Get("/:id/branches", handler.GetBranches)
}
