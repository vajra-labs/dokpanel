package setup

import (
	"context"
	"fmt"

	"goploy/src/conf"
	"goploy/src/utility/docker"

	"github.com/moby/moby/client"
)

// Runner holds injected deps for setup/teardown operations.
type Runner struct {
	cfg    *conf.Config
	docker *client.Client
	paths  *docker.AppPaths
}

func newRunner(cfg *conf.Config, dc *client.Client, paths *docker.AppPaths) *Runner {
	return &Runner{cfg: cfg, docker: dc, paths: paths}
}

func (r *Runner) Setup(ctx context.Context) error {
	if err := setupDirectories(r.paths); err != nil {
		return fmt.Errorf("directories setup failed: %w", err)
	}
	if err := setupSwarm(ctx, r.docker); err != nil {
		return fmt.Errorf("swarm setup failed: %w", err)
	}
	if err := setupNetwork(ctx, r.docker, r.cfg); err != nil {
		return fmt.Errorf("network setup failed: %w", err)
	}
	if err := writeTraefikConfig(r.cfg, r.paths); err != nil {
		return fmt.Errorf("traefik config setup failed: %w", err)
	}
	if err := setupTraefik(ctx, r.docker, r.cfg, r.paths); err != nil {
		return fmt.Errorf("traefik service setup failed: %w", err)
	}
	return nil
}

func (r *Runner) Teardown(ctx context.Context) error {
	if err := teardownTraefik(ctx, r.docker); err != nil {
		return fmt.Errorf("traefik teardown failed: %w", err)
	}
	if err := teardownNetwork(ctx, r.docker, r.cfg); err != nil {
		return fmt.Errorf("network teardown failed: %w", err)
	}
	if err := teardownSwarm(ctx, r.docker); err != nil {
		return fmt.Errorf("swarm teardown failed: %w", err)
	}
	if err := teardownDirectories(r.paths); err != nil {
		return fmt.Errorf("directories teardown failed: %w", err)
	}
	return nil
}
