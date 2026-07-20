package docs

import (
	"goploy/src/apis/dtos"
	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var gitlabTags = []string{"GitLab OAuth"}

// GitlabOpenApi registers OpenAPI specifications for GitLab provider endpoints.
func GitlabOpenApi(api huma.API) {
	r := api.OpenAPI()

	r.Paths["/api/gitlab"] = &huma.PathItem{
		Put: &huma.Operation{
			Tags:        gitlabTags,
			OperationID: "update-gitlab-config",
			Summary:     "Update GitLab Settings",
			Description: "Save and update GitLab OAuth client parameters for a specific provider.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.SaveGitlabDto{},
				"GitLab settings details",
				true,
			),
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.GitProviderResDto{},
					"GitLab settings updated successfully",
				),
				"400": apidoc.ErrContent("Invalid request body"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}

	r.Paths["/api/gitlab/{id}/repos"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        gitlabTags,
			OperationID: "list-gitlab-repos",
			Summary:     "List GitLab Projects",
			Description: "List all accessible projects/repositories for the authenticated GitLab user.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					[]dtos.GitRepoDto{},
					"List of GitLab projects",
				),
				"400": apidoc.ErrContent(
					"Invalid ID or unconfigured integration",
				),
				"500": apidoc.ErrContent("Failed to query API from GitLab"),
			},
		},
	}

	r.Paths["/api/gitlab/{id}/branches"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        gitlabTags,
			OperationID: "list-gitlab-branches",
			Summary:     "List GitLab Branches",
			Description: "List repository branches in a specific GitLab project.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
				{
					Name:        "projectPath",
					In:          "query",
					Description: "Full project path namespace (e.g. group/project)",
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
					"List of GitLab project branches",
				),
				"400": apidoc.ErrContent("Missing query parameters"),
				"500": apidoc.ErrContent("Failed to query API from GitLab"),
			},
		},
	}
}
