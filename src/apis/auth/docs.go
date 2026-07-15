package auth

import (
	"net/http"

	"goploy/src/core/apidoc"

	"github.com/danielgtaylor/huma/v2"
)

var tags = []string{"Authentication"}

// registerOpenApi registers all authentication-related paths into the OpenAPI spec.
func registerOpenApi(api huma.API) {
	r := api.OpenAPI()

	r.Paths["/api/auth/register"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        tags,
			OperationID: "register-user",
			Summary:     "Register User",
			Description: "Creates a new user account using the provided registration details",
			RequestBody: apidoc.Body(api, RegisterDto{}, true, "User registration information"),
			Responses: apidoc.Responses(
				apidoc.TextContent(http.StatusCreated, "User registered successfully"),
				apidoc.ErrContent(http.StatusBadRequest, "Invalid request body"),
				apidoc.ErrContent(http.StatusConflict, "Email already exists"),
				apidoc.ErrContent(http.StatusInternalServerError, "Internal server error"),
			),
		},
	}

	r.Paths["/api/auth/login"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        tags,
			OperationID: "login-user",
			Summary:     "Login User",
			Description: "Authenticates a user using email and password",
			RequestBody: apidoc.Body(api, LoginDto{}, true, "User login credentials"),
			Responses: apidoc.Responses(
				apidoc.JsonContent(api, http.StatusOK, LoginRes{}, "User logged in successfully"),
				apidoc.ErrContent(http.StatusBadRequest, "Invalid request body"),
				apidoc.ErrContent(http.StatusUnauthorized, "Invalid email or password"),
				apidoc.ErrContent(http.StatusInternalServerError, "Internal server error"),
			),
		},
	}
}
