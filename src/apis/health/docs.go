package health

import (
	"net/http"

	"goploy/src/docs"

	"github.com/danielgtaylor/huma/v2"
)

var tags = []string{"System"}

// registerOpenApi registers all health-related paths into the OpenAPI spec.
func registerOpenApi(api huma.API) {
	r := api.OpenAPI()

	r.Paths["/api/ping"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        tags,
			OperationID: "get-ping",
			Summary:     "Ping Server",
			Description: "Checks whether the server is reachable and responding",
			Responses: docs.Responses(
				docs.TextContent(http.StatusOK, "Returns a simple pong response"),
			),
		},
	}

	r.Paths["/api/pong"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        tags,
			OperationID: "get-pong",
			Summary:     "Pong Server",
			Description: "Responds to a ping request to confirm server availability",
			Responses: docs.Responses(
				docs.TextContent(http.StatusOK, "Returns a simple ping response"),
			),
		},
	}

	r.Paths["/api/health"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        tags,
			OperationID: "get-health",
			Summary:     "Health Check",
			Description: "Provides detailed information about server health and runtime status",
			Responses: docs.Responses(
				docs.JsonContent(api, http.StatusOK, HealthRes{}, "Returns server uptime, environment, version, timestamp, and memory usage"),
				docs.ErrContent(http.StatusInternalServerError, "Internal server error"),
			),
		},
	}
}
