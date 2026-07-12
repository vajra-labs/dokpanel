package sqldb

import (
	"database/sql"
	"embed"

	"dokpanel/src/conf"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
)

//go:embed migrate/*.sql
var embedMigrations embed.FS

// Migrate runs all pending goose migrations against the given DB pool.
// Pass isDev=true to enable goose query logs.
func Migrate(db *sql.DB, cfg *conf.Config) {
	goose.SetBaseFS(embedMigrations)

	if cfg.IS_PROD {
		goose.SetLogger(goose.NopLogger())
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		log.Fatal().Err(err).Msg("Failed to set goose dialect")
	}

	if err := goose.Up(db, "migrate"); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
}
