package jwt

import "go.uber.org/fx"

// Module provides *JwtToken via fx dependency injection.
var Module = fx.Module("jwt", fx.Provide(New))
