package db

import (
	"go.uber.org/fx"
)

// Module is the fx module for database dependencies.
var Module = fx.Module(
	"database",
	fx.Provide(providerPool, provideQueries),
)
