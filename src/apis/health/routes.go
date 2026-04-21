package health

import (
	"github.com/gofiber/fiber/v3"
)

func Router(app fiber.Router) {
	handler := NewHandler()
	// Register Health Endpoints
	app.Get("/ping", handler.Ping)
	app.Get("/pong", handler.Pong)
	app.Get("/health", handler.Health)
}
