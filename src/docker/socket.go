package docker

import (
	"dokpanel/src/conf"
	"fmt"
	"os"
)

func getCandidates() []string {
	var candidates []string

	// Priority 1: from .env / conf (e.g. "unix:///var/run/docker.sock")
	if conf.Env.DOCKER_HOST != "" {
		candidates = append(candidates, conf.Env.DOCKER_HOST)
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
