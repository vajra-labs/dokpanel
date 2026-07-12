package db

import (
	"database/sql"

	"go.uber.org/fx"
)

// Module is the fx module for database dependencies.
var Module = fx.Module(
	"database",
	fx.Provide(providerPool, provideQueries),
	// Force database pool initialization to register its lifecycle hooks
	// (like pragmas, ping, and migrations) even if no other module injects it.
	fx.Invoke(func(_ *sql.DB) {}),
)
