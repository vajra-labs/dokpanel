package docker

import (
	"path/filepath"

	"goploy/src/conf"
)

type AppPaths struct {
	BASE_PATH               string // /etc/goploy  OR  ./.docker
	TRAEFIK_PATH            string // BASE/traefik
	TRAEFIK_DYN_PATH        string // BASE/traefik/dynamic
	CERTIFICATES_PATH       string // BASE/traefik/dynamic/certificates
	SSH_PATH                string // BASE/ssh
	LOGS_PATH               string // BASE/logs
	VOLUME_BACKUPS_PATH     string // BASE/volume-backups
	VOLUME_BACKUP_LOCK_PATH string // BASE/volume-backup-lock
	PATCH_REPOS_PATH        string // BASE/patch-repos
	APPLICATIONS_PATH       string // BASE/applications
	COMPOSE_PATH            string // BASE/compose
	REGISTRY_PATH           string // BASE/registry
}

// NewPaths builds AppPaths from config.
func providePaths(cfg *conf.Config) *AppPaths {
	base := "./.docker" // dev default
	if cfg.IS_PROD {
		base = "/etc/goploy"
	}
	// Docker bind mounts require absolute paths
	if abs, err := filepath.Abs(base); err == nil {
		base = abs
	}
	traefik := filepath.Join(base, "traefik")
	traefikDyn := filepath.Join(traefik, "dynamic")

	return &AppPaths{
		BASE_PATH:               base,
		TRAEFIK_PATH:            traefik,
		TRAEFIK_DYN_PATH:        traefikDyn,
		CERTIFICATES_PATH:       filepath.Join(traefikDyn, "certificates"),
		SSH_PATH:                filepath.Join(base, "ssh"),
		LOGS_PATH:               filepath.Join(base, "logs"),
		VOLUME_BACKUPS_PATH:     filepath.Join(base, "volume-backups"),
		VOLUME_BACKUP_LOCK_PATH: filepath.Join(base, "volume-backup-lock"),
		PATCH_REPOS_PATH:        filepath.Join(base, "patch-repos"),
		APPLICATIONS_PATH:       filepath.Join(base, "applications"),
		COMPOSE_PATH:            filepath.Join(base, "compose"),
		REGISTRY_PATH:           filepath.Join(base, "registry"),
	}
}
