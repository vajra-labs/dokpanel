package setup

import (
	"goploy/src/pkg/docker"

	"go.uber.org/fx"
)

// Module provides *Runner via fx.
var Module = fx.Module(
	"setup",
	docker.Module,
	fx.Provide(newRunner),
)
