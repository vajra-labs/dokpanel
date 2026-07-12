package lib

import (
	"dokpanel/src/lib/docker"
	"dokpanel/src/lib/jwt"

	"go.uber.org/fx"
)

// Module bundles all lib sub-modules into a single fx module.
var Module = fx.Module("lib", jwt.Module, docker.Module)
