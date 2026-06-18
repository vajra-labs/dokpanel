package monitor

import (
	"github.com/gofiber/fiber/v3"
)

func Router(app fiber.Router) {
	handler := NewHandler()
	// Register Health Endpoints
	app.Get("/metrics", handler.Metrics)
	app.Get("/containers", handler.Containers)
}
