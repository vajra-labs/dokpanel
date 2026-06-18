package setup

import (
	"fmt"
	"os"

	"dokpanel/src/docker"
)

// Setup all necessary directories.
func SetupDirectories() error {
	p := docker.Paths

	dirs := []string{
		p.BASE_PATH,
		p.TRAEFIK_PATH,
		p.TRAEFIK_DYN_PATH,
		p.CERTIFICATES_PATH,
		p.SSH_PATH,
		p.LOGS_PATH,
		p.VOLUME_BACKUPS_PATH,
	}

	for _, dir := range dirs {
		if err := ensureDir(dir); err != nil {
			return err
		}
	}

	// SSH dir needs strict permissions (700) — private keys protection
	if err := os.Chmod(p.SSH_PATH, 0o700); err != nil {
		fmt.Printf("warn: failed to chmod ssh dir %s: %v\n", p.SSH_PATH, err)
	}

	fmt.Println("Directories ensured")
	return nil
}

// Removes the base directory and all its contents.
func TeardownDirectories() error {
	base := docker.Paths.BASE_PATH

	if _, err := os.Stat(base); os.IsNotExist(err) {
		fmt.Printf("Directory %s not found, skipping\n", base)
		return nil
	}

	if err := os.RemoveAll(base); err != nil {
		return err
	}

	fmt.Printf("Directory %s removed\n", base)
	return nil
}

// Creates a directory if it doesn't already exist.
func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			fmt.Printf("error: failed to create directory %s: %v\n", path, err)
			return err
		}
		fmt.Printf("   created: %s\n", path)
		return nil
	}
	fmt.Printf("   exists:  %s\n", path)
	return nil
}
