package apis

import (
	"goploy/src/apis/auth"
	"goploy/src/apis/health"
	"goploy/src/core/middle"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In
	App           *fiber.App
	HealthHandler *health.Handler
	AuthHandler   *auth.Handler
}

func Register(p RouterParams) {
	api := p.App.Group("/api")
	auth.Router(api, p.AuthHandler)
	health.Router(api, p.HealthHandler)
	api.Use(middle.NotFoundHandler)
}

var Module = fx.Module(
	"apis",
	auth.Module,
	health.Module,
	fx.Invoke(Register),
)
