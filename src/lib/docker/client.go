package docker

import (
	"context"
	"os"

	"dokpanel/src/conf"

	"github.com/moby/moby/client"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
)

// newClient finds a reachable Docker socket and returns a client.
func newClient(cfg *conf.Config) *client.Client {
	for _, socketPath := range getCandidates(cfg) {
		fsPath := socketPath
		if len(socketPath) > 7 && socketPath[:7] == "unix://" {
			fsPath = socketPath[7:]
		}
		if _, err := os.Stat(fsPath); os.IsNotExist(err) {
			log.Debug().Str("socket", fsPath).Msg("Docker socket not found, skipping")
			continue
		}
		opts := []client.Opt{client.WithHost(socketPath)}
		if cfg.DOCKER_API_VERSION != "" {
			opts = append(opts, client.WithAPIVersion(cfg.DOCKER_API_VERSION))
		}
		c, err := client.New(opts...)
		if err != nil {
			log.Warn().Str("socket", socketPath).Err(err).Msg("Docker client init failed, trying next")
			continue
		}
		return c
	}
	log.Fatal().Msg("no reachable Docker socket found — is Docker running?")
	return nil
}

func provideClient(lc fx.Lifecycle, cfg *conf.Config) *client.Client {
	c := newClient(cfg)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if _, err := c.Ping(ctx, client.PingOptions{}); err != nil {
				return err
			}
			log.Info().Str("host", cfg.DOCKER_HOST).Msg("Docker connected")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return c.Close()
		},
	})

	return c
}

func providePaths(cfg *conf.Config) *AppPaths {
	return NewPaths(cfg)
}
