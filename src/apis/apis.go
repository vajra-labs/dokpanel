package apis

import (
	"dokpanel/src/apis/auth"
	"dokpanel/src/apis/health"
	"dokpanel/src/docs"
	"dokpanel/src/middle"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

type RouterParams struct {
	fx.In
	App           *fiber.App
	HealthHandler *health.Handler
	AuthHandler   *auth.Handler
	DocsHandler   *docs.Handler
}

func Register(p RouterParams) {
	api := p.App.Group("/api")
	health.Router(api, p.HealthHandler)
	auth.Router(api, p.AuthHandler)
	docs.Router(api, p.DocsHandler)
	api.Use(middle.NotFoundHandler)
}

var Module = fx.Module(
	"apis",
	auth.Module,
	health.Module,
	fx.Invoke(Register),
)
