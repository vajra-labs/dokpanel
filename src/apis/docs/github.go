package docs

import (
	"goploy/src/apis/dtos"
	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var githubTags = []string{"GitHub App"}

// GithubOpenApi registers OpenAPI specifications for GitHub App provider endpoints.
func GithubOpenApi(api huma.API) {
	r := api.OpenAPI()

	r.Paths["/api/github"] = &huma.PathItem{
		Put: &huma.Operation{
			Tags:        githubTags,
			OperationID: "update-github-config",
			Summary:     "Update GitHub Settings",
			Description: "Save and update GitHub App parameters for a specific provider.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.SaveGithubDto{},
				"GitHub settings details",
				true,
			),
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.GitProviderResDto{},
					"GitHub settings updated successfully",
				),
				"400": apidoc.ErrContent("Invalid request body"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}

	r.Paths["/api/github/{id}/repos"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        githubTags,
			OperationID: "list-github-repos",
			Summary:     "List GitHub Repositories",
			Description: "List all accessible repositories for the GitHub App installation.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					[]dtos.GitRepoDto{},
					"List of GitHub repositories",
				),
				"400": apidoc.ErrContent(
					"Invalid ID or unconfigured integration",
				),
				"500": apidoc.ErrContent("Failed to query API from GitHub"),
			},
		},
	}

	r.Paths["/api/github/{id}/branches"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        githubTags,
			OperationID: "list-github-branches",
			Summary:     "List GitHub Branches",
			Description: "List branches in a specific GitHub repository.",
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
					"List of GitHub branches",
				),
				"400": apidoc.ErrContent("Missing query parameters"),
				"500": apidoc.ErrContent("Failed to query API from GitHub"),
			},
		},
	}
}
