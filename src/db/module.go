package db

import (
	"dokpanel/src/db/repos"

	"go.uber.org/fx"
)

var Module = fx.Module("database",
	fx.Provide(providerPool, provideQueries),
	// Force eager initialization — ensures OnStart lifecycle hook always runs
	// even when no handler currently injects *repos.Queries
	fx.Invoke(func(_ *repos.Queries) {}),
)
