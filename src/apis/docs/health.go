package docs

import (
	"goploy/src/apis/dtos"
	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var healthTags = []string{"System"}

// HealthOpenApi registers OpenAPI 3.1 specifications for health endpoints.
func HealthOpenApi(api huma.API) {
	r := api.OpenAPI()
	r.Paths["/api/ping"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        healthTags,
			OperationID: "get-ping",
			Summary:     "Ping Server",
			Description: "Checks whether the server is reachable and responding",
			Responses: apidoc.Response{
				"200": apidoc.TextContent("Returns a simple pong response"),
			},
		},
	}
	r.Paths["/api/pong"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        healthTags,
			OperationID: "get-pong",
			Summary:     "Pong Server",
			Description: "Responds to a ping request to confirm server availability",
			Responses: apidoc.Response{
				"200": apidoc.TextContent("Returns a simple ping response"),
			},
		},
	}
	r.Paths["/api/health"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        healthTags,
			OperationID: "get-health",
			Summary:     "Health Check",
			Description: "Provides detailed information about server health and runtime status",
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.HealthRes{},
					"Returns server uptime, environment, version, timestamp, and memory usage",
				),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
}
