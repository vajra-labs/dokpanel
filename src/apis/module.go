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
		handler.NewGitProviderHandler,
		handler.NewGithubHandler,
		handler.NewGitlabHandler,
		handler.NewGiteaHandler,
		handler.NewBitbucketHandler,
		guard.NewGuard,
	),
	fx.Invoke(
		docs.HealthOpenApi,
		docs.AuthOpenApi,
		docs.SshKeyOpenApi,
		docs.GitProviderOpenApi,
		docs.GithubOpenApi,
		docs.GitlabOpenApi,
		docs.GiteaOpenApi,
		docs.BitbucketOpenApi,
		Register,
	),
)
