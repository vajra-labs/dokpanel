package apis

import (
	"dokpanel/src/apis/auth"
	"dokpanel/src/apis/health"
	"dokpanel/src/apis/monitor"
	"dokpanel/src/middle"

	"github.com/gofiber/fiber/v3"
)

func Router(app fiber.Router) {
	// Api's Endpoint
	app.Route("/", health.Router, "health")
	app.Route("/auth", auth.Router, "auth")
	app.Route("/monitor", monitor.Router, "monitor")
	// 404 for unknown API routes
	app.Use(middle.NotFoundHandler)
}
