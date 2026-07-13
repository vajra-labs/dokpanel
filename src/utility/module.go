package utility

import (
	"goploy/src/utility/docker"
	"goploy/src/utility/jwt"

	"go.uber.org/fx"
)

// Module bundles all utility sub-modules into a single fx module.
var Module = fx.Module("lib", jwt.Module, docker.Module)
