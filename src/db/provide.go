package db

import (
	"context"
	"database/sql"

	"goploy/sqldb"
	"goploy/src/conf"
	"goploy/src/db/repos"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
)

var pragmas = []string{
	"PRAGMA foreign_keys=ON;",
	"PRAGMA journal_mode=WAL;",
	"PRAGMA synchronous=NORMAL;",
	"PRAGMA busy_timeout=5000;",
}

func providerPool(lc fx.Lifecycle, cfg *conf.Config) (*sql.DB, error) {
	pool, err := sql.Open("sqlite3", cfg.DB_PATH)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open DB")
	}
	ctx := context.Background()
	// SQLite performance + safety settings
	for _, q := range pragmas {
		if _, err := pool.ExecContext(ctx, q); err != nil {
			return nil, err
		}
	}
	// Verify connection is alive
	if err := pool.PingContext(ctx); err != nil {
		return nil, err
	}
	// Connection pool limits
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(5)
	// Run embedded migrations
	sqldb.Migrate(pool, cfg)
	// Close Database
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			log.Info().Msg("Closing database connection")
			// Close the connection pool
			return pool.Close()
		},
	})
	return pool, nil
}

func provideQueries(pool *sql.DB) *repos.Queries {
	return repos.New(pool)
}
