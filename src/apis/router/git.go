package router

import (
	"goploy/src/apis/guard"
	"goploy/src/apis/handler"

	"github.com/gofiber/fiber/v3"
)

// GitProviderRouter defines HTTP routes for Git Provider umbrella endpoints.
func GitProviderRouter(
	app fiber.Router,
	handler *handler.GitProviderHandler,
	guard *guard.Guard,
) {
	providers := app.Group("/git-providers", guard.Auth())
	providers.Post("/", handler.Create)
	providers.Get("/", handler.List)
	providers.Get("/:id", handler.Get)
	providers.Put("/:id", handler.Update)
	providers.Delete("/:id", handler.Delete)
}
