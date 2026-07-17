package docs

import (
	"goploy/src/apis/dtos"
	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var sshTags = []string{"SSH Keys"}

// SshKeyOpenApi registers OpenAPI 3.1 specifications for SSH Key management endpoints.
func SshKeyOpenApi(api huma.API) {
	r := api.OpenAPI()
	r.Paths["/api/ssh-keys"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        sshTags,
			OperationID: "create-ssh-key",
			Summary:     "Create SSH Key",
			Description: "Save a new SSH key pair.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.CreateSshKeyDto{},
				"SSH Key creation details",
				true,
			),
			Responses: apidoc.Response{
				"201": apidoc.JsonContent(
					api,
					dtos.SshKeyResDto{},
					"SSH Key created successfully",
				),
				"400": apidoc.ErrContent("Invalid request body"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
		Get: &huma.Operation{
			Tags:        sshTags,
			OperationID: "list-ssh-keys",
			Summary:     "List SSH Keys",
			Description: "List all registered SSH keys.",
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					[]dtos.SshKeyResDto{},
					"List of SSH Keys",
				),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
	r.Paths["/api/ssh-keys/mini"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        sshTags,
			OperationID: "list-ssh-keys-mini",
			Summary:     "List SSH Keys (Mini)",
			Description: "List SSH key names without private keys.",
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					[]dtos.SshKeyMiniResDto{},
					"Minimal list of SSH Keys",
				),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
	r.Paths["/api/ssh-keys/generate"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        sshTags,
			OperationID: "generate-ssh-key",
			Summary:     "Generate SSH Key Pair",
			Description: "Generate a new secure SSH key pair.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.GenSshKeyDto{},
				"SSH Key Generation parameters",
				true,
			),
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.GenSshKeyResDto{},
					"SSH Key Pair generated successfully",
				),
				"400": apidoc.ErrContent("Invalid parameters"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
	r.Paths["/api/ssh-keys/{id}"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        sshTags,
			OperationID: "get-ssh-key",
			Summary:     "Get SSH Key",
			Description: "Get details of an SSH key.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "SSH Key ID"),
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.SshKeyResDto{},
					"SSH Key details",
				),
				"400": apidoc.ErrContent("Invalid SSH Key ID"),
				"404": apidoc.ErrContent("SSH Key not found"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
		Put: &huma.Operation{
			Tags:        sshTags,
			OperationID: "update-ssh-key",
			Summary:     "Update SSH Key",
			Description: "Update an existing SSH key.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "SSH Key ID"),
			},
			RequestBody: apidoc.ReqBody(
				api,
				dtos.UpdateSshKeyDto{},
				"SSH Key update details",
				true,
			),
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.SshKeyResDto{},
					"SSH Key updated successfully",
				),
				"400": apidoc.ErrContent("Invalid request body or ID"),
				"404": apidoc.ErrContent("SSH Key not found"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
		Delete: &huma.Operation{
			Tags:        sshTags,
			OperationID: "delete-ssh-key",
			Summary:     "Delete SSH Key",
			Description: "Delete an SSH key.",
			Parameters: apidoc.Param{
				apidoc.IdParam("id", "SSH Key ID"),
			},
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.SshKeyResDto{},
					"SSH Key deleted successfully",
				),
				"400": apidoc.ErrContent("Invalid SSH Key ID"),
				"404": apidoc.ErrContent("SSH Key not found"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
}
