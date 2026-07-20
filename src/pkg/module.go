package pkg

import (
	"goploy/src/pkg/docker"
	"goploy/src/pkg/jwt"
	"goploy/src/pkg/shellx"

	"go.uber.org/fx"
)

// Module bundles all utility sub-modules into a single fx module.
var Module = fx.Module(
	"pkg",
	docker.Module,
	fx.Provide(
		shellx.NewSSHPool,
		jwt.NewJwtToken,
	),
)
