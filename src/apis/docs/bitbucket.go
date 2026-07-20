package docs

import (
	"goploy/src/apis/dtos"
	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var bitbucketTags = []string{"Bitbucket settings"}

// BitbucketOpenApi registers OpenAPI specifications for Bitbucket provider endpoints.
func BitbucketOpenApi(api huma.API) {
	r := api.OpenAPI()

	r.Paths["/api/bitbucket"] = &huma.PathItem{
		Put: &huma.Operation{
			Tags:        bitbucketTags,
			OperationID: "update-bitbucket-config",
			Summary:     "Update Bitbucket Settings",
			Description: "Save and update Bitbucket app password or token configurations for a specific provider.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.SaveBitbucketDto{},
				"Bitbucket settings details",
				true,
			),
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.GitProviderResDto{},
					"Bitbucket settings updated successfully",
				),
				"400": apidoc.ErrContent("Invalid request body"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}

	r.Paths["/api/bitbucket/{id}/repos"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        bitbucketTags,
			OperationID: "list-bitbucket-repos",
			Summary:     "List Bitbucket Repositories",
			Description: "List all accessible repositories in the configured Bitbucket workspace.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					[]dtos.GitRepoDto{},
					"List of Bitbucket repositories",
				),
				"400": apidoc.ErrContent(
					"Invalid ID or unconfigured integration",
				),
				"500": apidoc.ErrContent("Failed to query API from Bitbucket"),
			},
		},
	}

	r.Paths["/api/bitbucket/{id}/branches"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        bitbucketTags,
			OperationID: "list-bitbucket-branches",
			Summary:     "List Bitbucket Branches",
			Description: "List repository branches in a specific Bitbucket repository.",
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
					"List of Bitbucket repository branches",
				),
				"400": apidoc.ErrContent("Missing query parameters"),
				"500": apidoc.ErrContent("Failed to query API from Bitbucket"),
			},
		},
	}
}
