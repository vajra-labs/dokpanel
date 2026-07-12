package auth

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Router(app fiber.Router, handler *Handler) {
	auth := app.Group("/auth")
	auth.Post("/login", handler.login)
	auth.Post("/register", handler.register)
}

var Module = fx.Module(
	"auth",
	fx.Provide(NewHandler),
	fx.Invoke(registerOpenApi),
)
