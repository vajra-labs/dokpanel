package apis

import (
	"goploy/src/apis/docs"
	"goploy/src/apis/guard"
	"goploy/src/apis/handler"

	"go.uber.org/fx"
)

// Module provides API handlers and registers routing lifecycles in Fx.
var Module = fx.Module(
	"apis",
	fx.Provide(
		handler.NewHealthHandler,
		handler.NewAuthHandler,
		handler.NewSshKeyHandler,
		guard.NewGuard,
	),
	fx.Invoke(
		docs.HealthOpenApi,
		docs.AuthOpenApi,
		docs.SshKeyOpenApi,
		Register,
	),
)
