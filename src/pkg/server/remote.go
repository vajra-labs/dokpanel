package server

import (
	"context"
	"net"

	"goploy/src/pkg/shell"

	"github.com/moby/moby/client"
)

// GetRemoteDocker returns a Docker client tunneled over SSH.
func GetRemoteDocker(ctx context.Context, pool *shell.SSHPool, serverId int64) (*client.Client, error) {
	sc, err := pool.Get(ctx, serverId)
	if err != nil {
		return nil, err
	}
	dialCtx := func(ctx context.Context, _, _ string) (net.Conn, error) {
		return sc.Client().
			DialContext(ctx, "unix", "/var/run/docker.sock")
	}
	return client.New(client.WithDialContext(dialCtx))
}
