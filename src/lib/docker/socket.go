package docker

import (
	"fmt"
	"os"

	"dokpanel/src/conf"
)

func getCandidates(cfg *conf.Config) []string {
	var candidates []string

	// Priority 1: from config (e.g. "unix:///var/run/docker.sock")
	if cfg.DOCKER_HOST != "" {
		candidates = append(candidates, cfg.DOCKER_HOST)
	}
	// Priority 2 & 3: Rancher Desktop + Colima — both need home dir
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			fmt.Sprintf("unix://%s/.rd/docker.sock", home),             // Rancher Desktop
			fmt.Sprintf("unix://%s/.colima/default/docker.sock", home), // Colima
		)
	}
	// Priority 4: Standard Docker socket (fallback)
	candidates = append(candidates, "unix:///var/run/docker.sock")

	return candidates
}
