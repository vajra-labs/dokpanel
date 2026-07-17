package service

import "go.uber.org/fx"

// Module provides all application services via fx.
var Module = fx.Module(
	"service",
	fx.Provide(
		NewTokenService,
		NewAuthService,
		NewSshKeyService,
	),
)
