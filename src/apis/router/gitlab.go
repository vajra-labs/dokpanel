package router

import (
	"goploy/src/apis/guard"
	"goploy/src/apis/handler"

	"github.com/gofiber/fiber/v3"
)

// GitlabRouter defines HTTP routes for GitLab provider endpoints.
func GitlabRouter(
	app fiber.Router,
	handler *handler.GitlabHandler,
	guard *guard.Guard,
) {
	gitlab := app.Group("/gitlab", guard.Auth())
	gitlab.Put("/", handler.Update)
	gitlab.Get("/:id/repos", handler.GetRepositories)
	gitlab.Get("/:id/branches", handler.GetBranches)
	// Callback is unprotected because it is called by GitLab's OAuth server redirection
	app.Get("/gitlab/:id/callback", handler.Callback)
}
