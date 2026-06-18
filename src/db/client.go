package db

import (
	"database/sql"
	"dokpanel/src/conf"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

var (
	Pool *sql.DB
)

func init() {
	// Open DB connection
	con, err := sql.Open("sqlite3", conf.Env.DB_PATH)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open DB")
	}
	// SQLite settings
	if _, err := con.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		log.Fatal().Err(err).Msg("Failed to set WAL mode")
	}
	if _, err := con.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		log.Fatal().Err(err).Msg("Failed to set busy timeout")
	}
	if _, err := con.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		log.Fatal().Err(err).Msg("Failed to enable foreign keys")
	}
	// Verify connection
	if err = con.Ping(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to DB")
	}
	Pool = con
	// Connection pool
	Pool.SetMaxOpenConns(10)
	Pool.SetMaxIdleConns(5)
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}
