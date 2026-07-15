package utils

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	baseURL      = "https://templates.dokploy.com"
	fetchTimeout = 10 * time.Second
)

// TemplateMetadata is the metadata for a single template from meta.json
type TemplateMetadata struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Version     string        `json:"version"`
	Logo        string        `json:"logo"`
	Tags        []string      `json:"tags"`
	Links       TemplateLinks `json:"links"`
}

type TemplateLinks struct {
	Github  string `json:"github"`
	Website string `json:"website,omitempty"`
	Docs    string `json:"docs,omitempty"`
}

// CompleteTemplate is the parsed template.toml structure
type CompleteTemplate struct {
	Metadata  TemplateMetadata  `toml:"metadata"`
	Variables map[string]string `toml:"variables"`
	Config    TemplateConfig    `toml:"config"`
}

// TemplateConfig holds the runtime config of a template
type TemplateConfig struct {
	Isolated bool           `toml:"isolated"`
	Domains  []DomainConfig `toml:"domains"`
	Env      map[string]any `toml:"env"`
	Mounts   []MountConfig  `toml:"mounts"`
}

// DomainConfig holds a single domain entry
type DomainConfig struct {
	ServiceName string `toml:"serviceName"`
	Port        int    `toml:"port"`
	Path        string `toml:"path,omitempty"`
	Host        string `toml:"host,omitempty"`
}

// MountConfig holds a single mount entry
type MountConfig struct {
	FilePath string `toml:"filePath"`
	Content  string `toml:"content"`
}

// TemplateFiles holds fetched template files
type TemplateFiles struct {
	Config        *CompleteTemplate
	DockerCompose string
}

var httpClient = &http.Client{Timeout: fetchTimeout}

// FetchTemplatesList fetches the list of available templates from CDN
func FetchTemplatesList(ctx context.Context) ([]TemplateMetadata, error) {
	resp, err := fetchBytes(ctx, fmt.Sprintf("%s/meta.json", baseURL))
	if err != nil {
		return nil, fmt.Errorf("fetch templates list: %w", err)
	}

	var templates []TemplateMetadata
	if err := json.Unmarshal(resp, &templates); err != nil {
		return nil, fmt.Errorf("decode templates list: %w", err)
	}
	return templates, nil
}

// FetchTemplateFiles fetches template.toml + docker-compose.yml from CDN concurrently
func FetchTemplateFiles(ctx context.Context, templateID string) (*TemplateFiles, error) {
	type result struct {
		data []byte
		err  error
	}

	tomlCh := make(chan result, 1)
	composeCh := make(chan result, 1)

	tomlURL := fmt.Sprintf("%s/blueprints/%s/template.toml", baseURL, templateID)
	composeURL := fmt.Sprintf("%s/blueprints/%s/docker-compose.yml", baseURL, templateID)

	go func() {
		data, err := fetchBytes(ctx, tomlURL)
		tomlCh <- result{data, err}
	}()
	go func() {
		data, err := fetchBytes(ctx, composeURL)
		composeCh <- result{data, err}
	}()

	tomlResult := <-tomlCh
	composeResult := <-composeCh

	if tomlResult.err != nil {
		return nil, fmt.Errorf("fetch template.toml: %w", tomlResult.err)
	}
	if composeResult.err != nil {
		return nil, fmt.Errorf("fetch docker-compose.yml: %w", composeResult.err)
	}

	// Parse TOML config
	var cfg CompleteTemplate
	if err := toml.Unmarshal(tomlResult.data, &cfg); err != nil {
		return nil, fmt.Errorf("parse template.toml: %w", err)
	}

	return &TemplateFiles{
		Config:        &cfg,
		DockerCompose: string(composeResult.data),
	}, nil
}

// fetchBytes does a GET with context and returns body bytes
func fetchBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s for %s", resp.Status, url)
	}
	return io.ReadAll(resp.Body)
}
