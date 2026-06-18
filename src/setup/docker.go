package setup

import (
	"context"
	"fmt"
	"strings"

	"dokpanel/src/conf"
	"dokpanel/src/docker"

	"github.com/moby/moby/client"
)

// Setup Docker Swarm if not already initialized.
func SetupSwarm(ctx context.Context) error {
	c := docker.Client

	// Check if Swarm is initialized by inspecting it
	_, err := c.SwarmInspect(ctx, client.SwarmInspectOptions{})
	if err == nil {
		fmt.Println("Swarm is already initialized")
		return nil
	}

	// If SwarmInspect failed, initialize the swarm
	req := client.SwarmInitOptions{
		AdvertiseAddr: "127.0.0.1",
		ListenAddr:    "0.0.0.0",
	}

	_, err = c.SwarmInit(ctx, req)
	if err != nil {
		return err
	}

	fmt.Println("Swarm was initialized")
	return nil
}

// Dynamic overlay network name based on conf.Env.NAME.
func GetNetworkName() string {
	return strings.ToLower(conf.Env.NAME) + "-network"
}

// Setup docker network if it doesn't exist.
func SetupNetwork(ctx context.Context) error {
	c := docker.Client
	netName := GetNetworkName()

	// Check if network already exists
	_, err := c.NetworkInspect(ctx, netName, client.NetworkInspectOptions{})
	if err == nil {
		fmt.Printf("Network %q is already initialized\n", netName)
		return nil
	}

	// Create network
	opts := client.NetworkCreateOptions{
		Driver:     "overlay",
		Attachable: true,
	}

	_, err = c.NetworkCreate(ctx, netName, opts)
	if err != nil {
		return err
	}

	fmt.Printf("Network %q was initialized\n", netName)
	return nil
}

// Removes the overlay network.
func TeardownNetwork(ctx context.Context) error {
	c := docker.Client
	netName := GetNetworkName()

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

// Leaves Docker Swarm.
func TeardownSwarm(ctx context.Context) error {
	c := docker.Client

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
