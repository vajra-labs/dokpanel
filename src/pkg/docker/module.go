package docker

import "go.uber.org/fx"

// Module provides *client.Client, *AppPaths via fx.
var Module = fx.Module(
	"docker",
	fx.Provide(providePaths),
	fx.Provide(provideClient),
)
