package apis

import (
	"goploy/src/apis/guard"
	"goploy/src/apis/handler"
	"goploy/src/apis/router"
	"goploy/src/core/middle"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

// RouterParams wraps Fiber router dependencies for Fx injection.
type RouterParams struct {
	fx.In
	App                *fiber.App
	HealthHandler      *handler.HealthHandler
	AuthHandler        *handler.AuthHandler
	SshKeyHandler      *handler.SshKeyHandler
	GitProviderHandler *handler.GitProviderHandler
	GithubHandler      *handler.GithubHandler
	GitlabHandler      *handler.GitlabHandler
	GiteaHandler       *handler.GiteaHandler
	BitbucketHandler   *handler.BitbucketHandler
	Guard              *guard.Guard
}

// Register setups top-level API routes and handles 404 fallbacks.
func Register(p RouterParams) {
	api := p.App.Group("/api")
	router.AuthRouter(api, p.AuthHandler)
	router.HealthRouter(api, p.HealthHandler)
	router.SshKeyRouter(api, p.SshKeyHandler, p.Guard)
	router.GitProviderRouter(api, p.GitProviderHandler, p.Guard)
	router.GithubRouter(api, p.GithubHandler, p.Guard)
	router.GitlabRouter(api, p.GitlabHandler, p.Guard)
	router.GiteaRouter(api, p.GiteaHandler, p.Guard)
	router.BitbucketRouter(api, p.BitbucketHandler, p.Guard)
	api.Use(middle.NotFoundHandler)
}
