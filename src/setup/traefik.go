package setup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"goploy/src/conf"
	"goploy/src/utility/docker"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

const (
	traefikVersion   = "3.7.5"
	traefikImage     = "traefik:v" + traefikVersion
	traefikContainer = "goploy-traefik"
)

const traefikYMLTemplate = `global:
  sendAnonymousUsage: false
providers:
{{- if .IsDev }}
  docker:
    defaultRule: {{ .DefaultRule }}
{{- else }}
  swarm:
    exposedByDefault: false
    watch: true
  docker:
    exposedByDefault: false
    watch: true
    network: "{{ .NetworkName }}"
{{- end }}
  file:
    directory: "/etc/goploy/traefik/dynamic"
    watch: true
entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"
    http3:
      advertisedPort: 443
{{- if not .IsDev }}
    http:
      tls:
        certResolver: "letsencrypt"
{{- end }}
api:
  insecure: true
{{- if not .IsDev }}
certificatesResolvers:
  letsencrypt:
    acme:
      email: "test@localhost.com"
      storage: "/etc/goploy/traefik/dynamic/acme.json"
      httpChallenge:
        entryPoint: "web"
{{- end }}
`

const middlewaresYMLTemplate = `http:
  middlewares:
    redirect-to-https:
      redirectScheme:
        scheme: https
        permanent: true
`

const goployYMLTemplate = `http:
  routers:
    {{ .Name }}-router-app:
      rule: {{ .DokpanelRule }}
      service: {{ .Name }}-service-app
      entryPoints:
        - web
  services:
    {{ .Name }}-service-app:
      loadBalancer:
        servers:
          - url: "http://{{ .Name }}:{{ .Port }}"
        passHostHeader: true
`

type TraefikTemplateData struct {
	IsDev        bool
	Port         int
	Name         string
	NetworkName  string
	DefaultRule  string
	DokpanelRule string
}

// WriteTraefikConfig writes traefik.yml, middlewares.yml, and goploy.yml.
func writeTraefikConfig(cfg *conf.Config, p *docker.AppPaths) error {
	acmePath := filepath.Join(p.TRAEFIK_DYN_PATH, "acme.json")
	if _, err := os.Stat(acmePath); err == nil {
		if err := os.Chmod(acmePath, 0o600); err != nil {
			fmt.Printf("warn: failed to chmod acme.json %s: %v\n", acmePath, err)
		}
	}

	appName := strings.ToLower(cfg.NAME)
	netName := getNetworkName(cfg)

	data := TraefikTemplateData{
		IsDev:        cfg.IS_DEV,
		Port:         cfg.PORT,
		Name:         appName,
		NetworkName:  netName,
		DefaultRule:  "Host(`{{ trimPrefix `/` .Name }}.docker.localhost`)",
		DokpanelRule: fmt.Sprintf("Host(`%s.docker.localhost`) && PathPrefix(`/`)", appName),
	}

	if err := writeTemplateFile(filepath.Join(p.TRAEFIK_PATH, "traefik.yml"), traefikYMLTemplate, &data); err != nil {
		return fmt.Errorf("write traefik.yml: %w", err)
	}
	if err := writeTemplateFile(filepath.Join(p.TRAEFIK_DYN_PATH, "middlewares.yml"), middlewaresYMLTemplate, &data); err != nil {
		return fmt.Errorf("write middlewares.yml: %w", err)
	}
	if err := writeTemplateFile(filepath.Join(p.TRAEFIK_DYN_PATH, appName+".yml"), goployYMLTemplate, &data); err != nil {
		return fmt.Errorf("write %s.yml: %w", appName, err)
	}

	fmt.Println("Traefik configuration files written")
	return nil
}

func writeTemplateFile(filePath, tmplContent string, data *TraefikTemplateData) error {
	if info, err := os.Stat(filePath); err == nil {
		if info.IsDir() {
			fmt.Printf("warn: path is a directory, removing: %s\n", filePath)
			if err := os.RemoveAll(filePath); err != nil {
				return err
			}
		} else {
			fmt.Printf("   exists:  %s\n", filePath)
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	tmpl, err := template.New(filepath.Base(filePath)).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	fmt.Printf("   created: %s\n", filePath)
	return nil
}

// SetupTraefik pulls the Traefik image and starts the container.
func setupTraefik(ctx context.Context, c *client.Client, cfg *conf.Config, p *docker.AppPaths) error {
	fmt.Printf("Pulling Traefik image: %s\n", traefikImage)
	rc, err := c.ImagePull(ctx, traefikImage, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("pull traefik image: %w", err)
	}
	defer func() { _ = rc.Close() }()
	_, _ = io.Copy(io.Discard, rc)
	fmt.Println("Traefik image pulled")

	// Remove existing container if any
	existing, err := c.ContainerInspect(ctx, traefikContainer, client.ContainerInspectOptions{})
	if err == nil {
		fmt.Println("Removing existing Traefik container...")
		stopTimeout := 15
		_, _ = c.ContainerStop(ctx, existing.Container.ID, client.ContainerStopOptions{Timeout: &stopTimeout})
		_, _ = c.ContainerRemove(ctx, existing.Container.ID, client.ContainerRemoveOptions{Force: true})
		time.Sleep(3 * time.Second)
	}

	hostConfig := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{Name: "always"},
		Binds: []string{
			filepath.Join(p.TRAEFIK_PATH, "traefik.yml") + ":/etc/traefik/traefik.yml",
			p.TRAEFIK_DYN_PATH + ":/etc/goploy/traefik/dynamic",
			"/var/run/docker.sock:/var/run/docker.sock",
		},
		PortBindings: network.PortMap{
			network.MustParsePort("80/tcp"):  {{HostPort: "80"}},
			network.MustParsePort("443/tcp"): {{HostPort: "443"}},
			network.MustParsePort("443/udp"): {{HostPort: "443"}},
		},
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			getNetworkName(cfg): {},
		},
	}

	containerConfig := &container.Config{
		Image: traefikImage,
		ExposedPorts: network.PortSet{
			network.MustParsePort("80/tcp"):  {},
			network.MustParsePort("443/tcp"): {},
			network.MustParsePort("443/udp"): {},
		},
	}

	resp, err := c.ContainerCreate(ctx, client.ContainerCreateOptions{
		Name:             traefikContainer,
		Config:           containerConfig,
		HostConfig:       hostConfig,
		NetworkingConfig: networkConfig,
	})
	if err != nil {
		return fmt.Errorf("create traefik container: %w", err)
	}

	if _, err := c.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("start traefik container: %w", err)
	}

	fmt.Printf("Traefik container started (id: %s)\n", resp.ID[:12])
	return nil
}

// TeardownTraefik stops and removes the Traefik container.
func teardownTraefik(ctx context.Context, c *client.Client) error {
	existing, err := c.ContainerInspect(ctx, traefikContainer, client.ContainerInspectOptions{})
	if err != nil {
		fmt.Println("Traefik container not found, skipping")
		return nil
	}

	stopTimeout := 10
	_, _ = c.ContainerStop(ctx, existing.Container.ID, client.ContainerStopOptions{Timeout: &stopTimeout})
	if _, err := c.ContainerRemove(ctx, existing.Container.ID, client.ContainerRemoveOptions{Force: true}); err != nil {
		return err
	}

	fmt.Println("Traefik container removed")
	return nil
}
