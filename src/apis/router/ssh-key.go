package router

import (
	"goploy/src/apis/guard"
	"goploy/src/apis/handler"

	"github.com/gofiber/fiber/v3"
)

// SshKeyRouter defines HTTP routes for SSH keys endpoints.
func SshKeyRouter(
	app fiber.Router,
	handler *handler.SshKeyHandler,
	guard *guard.Guard,
) {
	keys := app.Group("/ssh-keys", guard.Auth())
	keys.Post("/", handler.Create)
	keys.Get("/", handler.List)
	keys.Get("/mini", handler.ListMini)
	keys.Post("/generate", handler.Generate)
	keys.Get("/:id", handler.Get)
	keys.Put("/:id", handler.Update)
	keys.Delete("/:id", handler.Delete)
}
