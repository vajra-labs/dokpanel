package apis

import (
	"dokpanel/src/apis/health"
	"dokpanel/src/middle"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In
	App           *fiber.App
	HealthHandler *health.Handler
}

func Register(p RouterParams) {
	api := p.App.Group("/api")
	health.Router(api, p.HealthHandler)
	api.Use(middle.NotFoundHandler)
}

var Module = fx.Module("apis",
	health.Module,
	fx.Invoke(Register),
)
