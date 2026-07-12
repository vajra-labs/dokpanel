package docker

import "go.uber.org/fx"

// Module provides *client.Client and *AppPaths via fx.
var Module = fx.Module(
	"docker",
	fx.Provide(provideClient, providePaths),
)
