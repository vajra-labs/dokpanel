package docs

import (
	"fmt"
	"reflect"

	"goploy/src/conf"
	"goploy/src/core/errorx"

	"github.com/danielgtaylor/huma/v2"
)

// tags for the OpenAPI spec.
var tags = []*huma.Tag{
	{
		Name:        "System",
		Description: "Endpoints related to system operations and health checks",
	},
	{
		Name:        "Authentication",
		Description: "Endpoints for user authentication and session management",
	},
}

// provideOpenAPI creates and returns a huma.API with all routes registered.
func provideOpenAPI(cfg *conf.Config) huma.API {
	api := huma.NewAPI(huma.Config{
		OpenAPI: &huma.OpenAPI{
			OpenAPI: "3.1.0",
			Info: &huma.Info{
				Title:       fmt.Sprintf("%s API", cfg.NAME),
				Version:     conf.VERSION,
				Description: fmt.Sprintf("Complete API documentation for %s - manage applications, databases, and orchestrate your infrastructure.", cfg.NAME),
				Contact: &huma.Contact{
					Name: fmt.Sprintf("%s Team", cfg.NAME),
					URL:  "https://goploy.com",
				},
				License: &huma.License{
					Name: "MIT",
					URL:  "https://github.com/vajra-labs/goploy/blob/canary/LICENSE",
				},
			},
			Tags:  tags,
			Paths: map[string]*huma.PathItem{},
		},
		Formats: huma.DefaultFormats,
	}, nil)

	r := api.OpenAPI()

	// Security schemes
	r.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"apiKey": {
			Type:        "apiKey",
			In:          "header",
			Name:        "x-api-key",
			Description: fmt.Sprintf("API key authentication. Generate an API key from your %s dashboard under Settings > API Keys.", cfg.NAME),
		},
	}

	// Register custom HttpError in components/schemas
	r.Components.Schemas.Schema(reflect.TypeOf(errorx.HttpError{}), true, "HttpError")

	return api
}
