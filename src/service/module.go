package service

import (
	"goploy/src/service/provider"

	"go.uber.org/fx"
)

// Module provides all application services via fx.
var Module = fx.Module(
	"service",
	fx.Provide(
		NewAuthService,
		NewTokenService,
		NewSshKeyService,
		NewDockerService,
		provider.NewGitProviderService,
		provider.NewGithubProviderService,
		provider.NewGitlabProviderService,
		provider.NewGiteaProviderService,
		provider.NewBitbucketProviderService,
	),
)
