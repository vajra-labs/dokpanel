package router

import (
	"goploy/src/apis/guard"
	"goploy/src/apis/handler"

	"github.com/gofiber/fiber/v3"
)

// GiteaRouter defines HTTP routes for Gitea provider endpoints.
func GiteaRouter(
	app fiber.Router,
	handler *handler.GiteaHandler,
	guard *guard.Guard,
) {
	gitea := app.Group("/gitea", guard.Auth())
	gitea.Put("/", handler.Update)
	gitea.Get("/:id/repos", handler.GetRepositories)
	gitea.Get("/:id/branches", handler.GetBranches)
	// Callback is unprotected because it is called by Gitea's OAuth server redirection
	app.Get("/gitea/:id/callback", handler.Callback)
}
