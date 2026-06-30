package setup

import (
	"context"
	"fmt"
	"strings"

	"dokpanel/src/conf"

	"github.com/moby/moby/client"
)

// SetupSwarm initializes Docker Swarm if not already active.
func setupSwarm(ctx context.Context, c *client.Client) error {
	_, err := c.SwarmInspect(ctx, client.SwarmInspectOptions{})
	if err == nil {
		fmt.Println("Swarm is already initialized")
		return nil
	}

	_, err = c.SwarmInit(ctx, client.SwarmInitOptions{
		AdvertiseAddr: "127.0.0.1",
		ListenAddr:    "0.0.0.0",
	})
	if err != nil {
		return err
	}

	fmt.Println("Swarm was initialized")
	return nil
}

// GetNetworkName returns the overlay network name derived from config.
func getNetworkName(cfg *conf.Config) string {
	return strings.ToLower(cfg.NAME) + "-network"
}

// SetupNetwork creates the overlay network if it doesn't exist.
func setupNetwork(ctx context.Context, c *client.Client, cfg *conf.Config) error {
	netName := getNetworkName(cfg)

	_, err := c.NetworkInspect(ctx, netName, client.NetworkInspectOptions{})
	if err == nil {
		fmt.Printf("Network %q is already initialized\n", netName)
		return nil
	}

	_, err = c.NetworkCreate(ctx, netName, client.NetworkCreateOptions{
		Driver:     "overlay",
		Attachable: true,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Network %q was initialized\n", netName)
	return nil
}

// TeardownNetwork removes the overlay network.
func teardownNetwork(ctx context.Context, c *client.Client, cfg *conf.Config) error {
	netName := getNetworkName(cfg)

	if _, err := c.NetworkInspect(ctx, netName, client.NetworkInspectOptions{}); err != nil {
		fmt.Printf("Network %q not found, skipping\n", netName)
		return nil
	}

	if _, err := c.NetworkRemove(ctx, netName, client.NetworkRemoveOptions{}); err != nil {
		return err
	}

	fmt.Printf("Network %q removed\n", netName)
	return nil
}

// TeardownSwarm leaves Docker Swarm.
func teardownSwarm(ctx context.Context, c *client.Client) error {
	if _, err := c.SwarmInspect(ctx, client.SwarmInspectOptions{}); err != nil {
		fmt.Println("Swarm not active, skipping")
		return nil
	}

	if _, err := c.SwarmLeave(ctx, client.SwarmLeaveOptions{Force: true}); err != nil {
		return err
	}

	fmt.Println("Swarm left")
	return nil
}
