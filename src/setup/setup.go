package setup

import (
	"context"
	"fmt"
)

func RunSetup(ctx context.Context) error {
	// Setup required directories
	if err := SetupDirectories(); err != nil {
		return fmt.Errorf("directories setup failed: %w", err)
	}
	//  Setup Docker Swarm is initialized
	if err := SetupSwarm(ctx); err != nil {
		return fmt.Errorf("swarm setup failed: %w", err)
	}
	// Setup Overlay Network exists
	if err := SetupNetwork(ctx); err != nil {
		return fmt.Errorf("network setup failed: %w", err)
	}
	// Setup Traefik configuration files exist
	if err := WriteTraefikConfig(); err != nil {
		return fmt.Errorf("traefik config setup failed: %w", err)
	}
	// Setup Traefik container is running
	if err := SetupTraefik(ctx); err != nil {
		return fmt.Errorf("traefik service setup failed: %w", err)
	}
	return nil
}

func RunTeardown(ctx context.Context) error {
	// Teardown Traefik container
	if err := TeardownTraefik(ctx); err != nil {
		return fmt.Errorf("traefik teardown failed: %w", err)
	}
	// Teardown Overlay Network
	if err := TeardownNetwork(ctx); err != nil {
		return fmt.Errorf("network teardown failed: %w", err)
	}
	// Teardown Docker Swarm
	if err := TeardownSwarm(ctx); err != nil {
		return fmt.Errorf("swarm teardown failed: %w", err)
	}
	// Teardown directories
	if err := TeardownDirectories(); err != nil {
		return fmt.Errorf("directories teardown failed: %w", err)
	}
	return nil
}
