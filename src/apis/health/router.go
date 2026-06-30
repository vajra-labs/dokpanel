package health

import (
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func Router(app fiber.Router, handler *Handler) {
	app.Get("/ping", handler.Ping)
	app.Get("/pong", handler.Pong)
	app.Get("/health", handler.Health)
}

var Module = fx.Module("health",
	fx.Provide(NewHandler),
)
