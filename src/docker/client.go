package docker

import (
	"os"
	"sync"

	"dokpanel/src/conf"

	"github.com/moby/moby/client"
	"github.com/rs/zerolog/log"
)

var (
	Client *client.Client
	once   sync.Once
)

// Initializes the Docker client singleton.
func Init() {
	once.Do(func() {
		for _, socketPath := range getCandidates() {
			// Strip "unix://" prefix for os.Stat check
			fsPath := socketPath
			if len(socketPath) > 7 && socketPath[:7] == "unix://" {
				fsPath = socketPath[7:]
			}
			// Check if socket file exists on disk
			if _, err := os.Stat(fsPath); os.IsNotExist(err) {
				log.Debug().Str("socket", fsPath).Msg("Docker socket not found, skipping")
				continue
			}
			opts := []client.Opt{client.WithHost(socketPath)}
			// Use API version from .env only if explicitly set
			if conf.Env.DOCKER_API_VERSION != "" {
				opts = append(opts, client.WithAPIVersion(conf.Env.DOCKER_API_VERSION))
			}
			client, err := client.New(opts...)
			if err != nil {
				log.Warn().Str("socket", socketPath).Err(err).Msg("Docker client init failed, trying next")
				continue
			}
			Client = client
			return
		}
		log.Fatal().Msg("no reachable Docker socket found — is Docker running?")
	})
}

// Closes the Docker client connection if it exists.
func Close() {
	if Client != nil {
		Client.Close()
	}
}
