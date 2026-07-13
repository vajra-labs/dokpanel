package db

import (
	"context"
	"database/sql"

	"goploy/sqldb"
	"goploy/src/conf"
	"goploy/src/db/repos"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/mattn/go-sqlite3"
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

func init() {
	// Register sqlite3_nanoid
	sql.Register("sqlite3_nanoid", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			// nanoid()
			if err := conn.RegisterFunc("nanoid", func() (string, error) {
				return gonanoid.New()
			}, true); err != nil {
				return err
			}
			// nanoid(length)
			if err := conn.RegisterFunc("nanoid", func(length int) (string, error) {
				return gonanoid.New(length)
			}, true); err != nil {
				return err
			}
			return nil
		},
	})
	// Register gen_app_name
	sql.Register("gen_app_name", &sqlite3.SQLiteDriver{
		ConnectHook: func(sc *sqlite3.SQLiteConn) error {
			return nil
		},
	})
}

func providerPool(lc fx.Lifecycle, cfg *conf.Config) *sql.DB {
	// sql.Open only validates the driver name and DSN — no actual connection yet
	pool, err := sql.Open("sqlite3", cfg.DB_PATH)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open DB")
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// SQLite performance + safety settings
			for _, q := range pragmas {
				if _, err := pool.ExecContext(ctx, q); err != nil {
					return err
				}
			}
			// Verify connection is alive
			if err := pool.PingContext(ctx); err != nil {
				return err
			}
			// Connection pool limits
			pool.SetMaxOpenConns(10)
			pool.SetMaxIdleConns(5)
			// Run embedded migrations
			sqldb.Migrate(pool, cfg)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info().Msg("Closing database connection")
			// Close the connection pool
			return pool.Close()
		},
	})

	return pool
}

func provideQueries(pool *sql.DB) *repos.Queries {
	return repos.New(pool)
}
