package apis

import (
	"dokpanel/src/apis/health"
	"dokpanel/src/middle"

	"github.com/gofiber/fiber/v3"
)

func Router(app fiber.Router) {
	// Health Endpoint
	app.Route("/", health.Router, "health")
	// 404 for unknown API routes
	app.Use(middle.NotFoundHandler)
}
