package docs

import (
	"fmt"
	"reflect"

	"github.com/danielgtaylor/huma/v2"
)

// ResponseEntry is a single status code + response pair.
type ResponseEntry struct {
	status   string
	response *huma.Response
}

// Responses builds a map[string]*huma.Response from entries.
// Same as TypeScript: responses: { [HttpStatus.OK]: jsonContent(...), ... }
//
// Usage:
//
//	Responses: lib.Responses(
//	    lib.JsonContent(http.StatusOK, schema, "description"),
//	    lib.TextContent(http.StatusOK, "description"),
//	    lib.ErrContent(http.StatusNotFound, "description"),
//	)
func Responses(entries ...ResponseEntry) map[string]*huma.Response {
	m := map[string]*huma.Response{}
	for _, e := range entries {
		m[e.status] = e.response
	}
	return m
}

// JsonContent creates a JSON response entry.
func JsonContent(api huma.API, status int, v any, description string) ResponseEntry {
	return ResponseEntry{
		status: fmt.Sprintf("%d", status),
		response: &huma.Response{
			Description: description,
			Content: map[string]*huma.MediaType{
				"application/json": {Schema: SchemaFor(api, v)},
			},
		},
	}
}

// TextContent creates a text/plain response entry.
func TextContent(status int, description string) ResponseEntry {
	return ResponseEntry{
		status: fmt.Sprintf("%d", status),
		response: &huma.Response{
			Description: description,
			Content: map[string]*huma.MediaType{
				"text/plain": {Schema: &huma.Schema{Type: "string"}},
			},
		},
	}
}

// ErrContent creates an error response entry using custom HttpError schema.
func ErrContent(status int, description string) ResponseEntry {
	return ResponseEntry{
		status: fmt.Sprintf("%d", status),
		response: &huma.Response{
			Description: description,
			Content: map[string]*huma.MediaType{
				"application/json": {
					Schema: &huma.Schema{Ref: "#/components/schemas/HttpError"},
				},
			},
		},
	}
}

// SchemaFor generates a huma.Schema from any Go struct using the API registry.
//
// Usage:
//
//	lib.JsonContent(http.StatusOK, lib.SchemaFor(api, MyDto{}), "description")
func SchemaFor(api huma.API, v any) *huma.Schema {
	return api.OpenAPI().Components.Schemas.Schema(
		reflect.TypeOf(v), true, "",
	)
}
