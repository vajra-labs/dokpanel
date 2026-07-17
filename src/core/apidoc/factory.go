package apidoc

import (
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

type (
	Response map[string]*huma.Response
	Param    []*huma.Param
)

// JsonContent creates a JSON response entry.
func JsonContent(api huma.API, v any, desc string) *huma.Response {
	return &huma.Response{
		Description: desc,
		Content: map[string]*huma.MediaType{
			"application/json": {
				Schema: SchemaFor(api, v),
			},
		},
	}
}

// TextContent creates a text/plain response entry.
func TextContent(desc string) *huma.Response {
	return &huma.Response{
		Description: desc,
		Content: map[string]*huma.MediaType{
			"text/plain": {
				Schema: &huma.Schema{Type: "string"},
			},
		},
	}
}

// ErrContent creates an error response entry using custom HttpError schema.
func ErrContent(desc string) *huma.Response {
	return &huma.Response{
		Description: desc,
		Content: map[string]*huma.MediaType{
			"application/json": {
				Schema: &huma.Schema{Ref: "#/components/schemas/HttpError"},
			},
		},
	}
}

// ReqBody creates a request body schema.
func ReqBody(
	api huma.API,
	v any,
	desc string,
	required ...bool,
) *huma.RequestBody {
	req := true
	if len(required) > 0 {
		req = required[0]
	}
	return &huma.RequestBody{
		Required:    req,
		Description: desc,
		Content: map[string]*huma.MediaType{
			"application/json": {
				Schema: SchemaFor(api, v),
			},
		},
	}
}

// schemaFor generates a huma.Schema from any Go struct using the API registry.
func SchemaFor(api huma.API, v any) *huma.Schema {
	return api.OpenAPI().Components.Schemas.Schema(
		reflect.TypeOf(v), true, "",
	)
}

// IdParam generates a reusable path parameter definition for integer IDs.
func IdParam(name string, desc string) *huma.Param {
	return &huma.Param{
		Name:        name,
		In:          "path",
		Description: desc,
		Required:    true,
		Schema: &huma.Schema{
			Type: "integer",
		},
	}
}
