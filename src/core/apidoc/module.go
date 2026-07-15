// Package docs provides OpenAPI 3.1 spec generation and Scalar UI rendering.
// It uses Huma for spec building and go-scalar-api-reference for the UI.
package apidoc

import "go.uber.org/fx"

var Module = fx.Module(
	"apidoc",
	fx.Provide(provideOpenAPI),
	fx.Invoke(invokeRouter),
)
