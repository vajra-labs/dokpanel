package docs

import (
	"goploy/src/apis/dtos"
	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var giteaTags = []string{"Gitea OAuth"}

// GiteaOpenApi registers OpenAPI specifications for Gitea provider endpoints.
func GiteaOpenApi(api huma.API) {
	r := api.OpenAPI()

	r.Paths["/api/gitea"] = &huma.PathItem{
		Put: &huma.Operation{
			Tags:        giteaTags,
			OperationID: "update-gitea-config",
			Summary:     "Update Gitea Settings",
			Description: "Save and update Gitea OAuth client parameters for a specific provider.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.SaveGiteaDto{},
				"Gitea settings details",
				true,
			),
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.GitProviderResDto{},
					"Gitea settings updated successfully",
				),
				"400": apidoc.ErrContent("Invalid request body"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}

	r.Paths["/api/gitea/{id}/repos"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        giteaTags,
			OperationID: "list-gitea-repos",
			Summary:     "List Gitea Repositories",
			Description: "List all accessible repositories for the authenticated Gitea user.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					[]dtos.GitRepoDto{},
					"List of Gitea repositories",
				),
				"400": apidoc.ErrContent(
					"Invalid ID or unconfigured integration",
				),
				"500": apidoc.ErrContent("Failed to query API from Gitea"),
			},
		},
	}

	r.Paths["/api/gitea/{id}/branches"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        giteaTags,
			OperationID: "list-gitea-branches",
			Summary:     "List Gitea Branches",
			Description: "List repository branches in a specific Gitea repository.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
				{
					Name:        "owner",
					In:          "query",
					Description: "Repository owner",
					Required:    true,
					Schema: &huma.Schema{
						Type: "string",
					},
				},
				{
					Name:        "repo",
					In:          "query",
					Description: "Repository name",
					Required:    true,
					Schema: &huma.Schema{
						Type: "string",
					},
				},
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					[]dtos.GitBranchDto{},
					"List of Gitea repository branches",
				),
				"400": apidoc.ErrContent("Missing query parameters"),
				"500": apidoc.ErrContent("Failed to query API from Gitea"),
			},
		},
	}
}
