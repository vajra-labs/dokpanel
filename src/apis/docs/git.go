package docs

import (
	"goploy/src/apis/dtos"
	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var providerTags = []string{"Git Providers"}

// GitProviderOpenApi registers OpenAPI specifications for Git Provider umbrella endpoints.
func GitProviderOpenApi(api huma.API) {
	r := api.OpenAPI()

	r.Paths["/api/git-providers"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        providerTags,
			OperationID: "create-git-provider",
			Summary:     "Create Git Provider",
			Description: "Create a generic Git provider configuration and initialize its configurations.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.CreateGitProviderDto{},
				"Git provider configuration details",
				true,
			),
			Responses: apidoc.Response{
				"201": apidoc.JsonContent(
					api,
					dtos.GitProviderResDto{},
					"Git Provider created successfully",
				),
				"400": apidoc.ErrContent("Invalid request body"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
		Get: &huma.Operation{
			Tags:        providerTags,
			OperationID: "list-git-providers",
			Summary:     "List Git Providers",
			Description: "List all configured Git providers and their settings.",
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					[]dtos.GitProviderResDto{},
					"List of Git Providers",
				),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}

	r.Paths["/api/git-providers/{id}"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        providerTags,
			OperationID: "get-git-provider",
			Summary:     "Get Git Provider",
			Description: "Get details and settings of a specific Git provider.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.GitProviderResDto{},
					"Git Provider configuration details",
				),
				"400": apidoc.ErrContent("Invalid Git Provider ID"),
				"404": apidoc.ErrContent("Git Provider not found"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
		Put: &huma.Operation{
			Tags:        providerTags,
			OperationID: "update-git-provider",
			Summary:     "Update Git Provider",
			Description: "Update the name or sharing status of a Git provider.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
			},
			RequestBody: apidoc.ReqBody(
				api,
				dtos.UpdateGitProviderDto{},
				"Update details",
				true,
			),
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.GitProviderResDto{},
					"Git Provider updated successfully",
				),
				"400": apidoc.ErrContent("Invalid request body or ID"),
				"404": apidoc.ErrContent("Git Provider not found"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
		Delete: &huma.Operation{
			Tags:        providerTags,
			OperationID: "delete-git-provider",
			Summary:     "Delete Git Provider",
			Description: "Delete a Git provider configuration.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "Git Provider ID"),
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					map[string]bool{},
					"Git Provider deleted successfully",
				),
				"400": apidoc.ErrContent("Invalid Git Provider ID"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
}
