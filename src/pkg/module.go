package pkg

import (
	"goploy/src/pkg/docker"
	"goploy/src/pkg/jwt"
	"goploy/src/pkg/shell"

	"go.uber.org/fx"
)

// Module bundles all utility sub-modules into a single fx module.
var Module = fx.Module(
	"pkg",
	jwt.Module,
	docker.Module,
	fx.Provide(shell.NewSSHPool),
)
