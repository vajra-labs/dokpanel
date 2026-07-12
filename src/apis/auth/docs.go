package auth

import (
	"net/http"

	"dokpanel/src/docs"

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
			RequestBody: docs.Body(api, RegisterDto{}, true, "User registration information"),
			Responses: docs.Responses(
				docs.TextContent(http.StatusCreated, "User registered successfully"),
				docs.ErrContent(http.StatusBadRequest, "Invalid request body"),
				docs.ErrContent(http.StatusConflict, "Email already exists"),
				docs.ErrContent(http.StatusInternalServerError, "Internal server error"),
			),
		},
	}

	r.Paths["/api/auth/login"] = &huma.PathItem{
		Post: &huma.Operation{
			Tags:        tags,
			OperationID: "login-user",
			Summary:     "Login User",
			Description: "Authenticates a user using email and password",
			RequestBody: docs.Body(api, LoginDto{}, true, "User login credentials"),
			Responses: docs.Responses(
				docs.JsonContent(api, http.StatusOK, LoginRes{}, "User logged in successfully"),
				docs.ErrContent(http.StatusBadRequest, "Invalid request body"),
				docs.ErrContent(http.StatusUnauthorized, "Invalid email or password"),
				docs.ErrContent(http.StatusInternalServerError, "Internal server error"),
			),
		},
	}
}
