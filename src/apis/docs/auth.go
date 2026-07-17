package docs

import (
	"goploy/src/apis/dtos"
	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var authTags = []string{"Authentication"}

// AuthOpenApi registers OpenAPI 3.1 specifications for auth endpoints.
func AuthOpenApi(api huma.API) {
	r := api.OpenAPI()

	r.Paths["/api/auth/setup"] = &huma.PathItem{
		Get: &huma.Operation{
			Tags:        authTags,
			OperationID: "auth-setup",
			Summary:     "Check Setup Status",
			Description: "Check if the initial owner registration is done.",
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.SetupStatusResDto{},
					"Setup status",
				),
			},
		},
	}
	r.Paths["/api/auth/register"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        authTags,
			OperationID: "register-user",
			Summary:     "Register Owner",
			Description: "Register the first user as owner of the app.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.RegisterDto{},
				"Owner registration details",
				true,
			),
			Responses: apidoc.Response{
				"201": apidoc.JsonContent(
					api,
					dtos.LoginRes{},
					"Owner registered and logged in successfully",
				),
				"400": apidoc.ErrContent("Invalid request body"),
				"409": apidoc.ErrContent("Owner already exists"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
	r.Paths["/api/auth/login"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        authTags,
			OperationID: "login-user",
			Summary:     "Login",
			Description: "Authenticate user and set cookies for session.",
			RequestBody: apidoc.ReqBody(
				api,
				dtos.LoginDto{},
				"Login credentials",
				true,
			),
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.LoginRes{},
					"Logged in successfully, cookies set",
				),
				"400": apidoc.ErrContent("Invalid request body"),
				"401": apidoc.ErrContent("Invalid email or password"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
	r.Paths["/api/auth/refresh"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        authTags,
			OperationID: "refresh-token",
			Summary:     "Refresh Access Token",
			Description: "Refresh session cookie using the refresh token.",
			Responses: apidoc.Response{
				"200": apidoc.JsonContent(
					api,
					dtos.TokenDto{},
					"Access token refreshed successfully",
				),
				"400": apidoc.ErrContent("Refresh token is required"),
				"401": apidoc.ErrContent("Refresh token is invalid or expired"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
	r.Paths["/api/auth/logout"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        authTags,
			OperationID: "logout-user",
			Summary:     "Logout",
			Description: "Clear authentication cookies and log out.",
			Parameters: apidoc.Param{
				{
					Name:        "all",
					In:          "query",
					Description: "Logout from all devices",
					Schema: &huma.Schema{
						Type: "boolean",
					},
				},
			},
			Responses: apidoc.Response{
				"200": apidoc.TextContent("Logged out successfully"),
				"400": apidoc.ErrContent("Refresh token is required"),
				"500": apidoc.ErrContent("Internal server error"),
			},
		},
	}
}
