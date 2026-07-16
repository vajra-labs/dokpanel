package db

import (
	"goploy/src/db/seeds"

	"go.uber.org/fx"
)

// Module is the fx module for database dependencies.
var Module = fx.Module(
	"database",
	fx.Provide(providerPool, provideQueries),
	fx.Invoke(seeds.SeedGroup),
)
