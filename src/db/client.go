package db

import (
	"database/sql"
	"dokpanel/src/conf"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

var (
	Pool *sql.DB
	once sync.Once
)

func init() {
	once.Do(func() {
		// Open connection (creates DB file if not exists)
		con, err := sql.Open("sqlite3", conf.Env.DB_PATH)
		if err != nil {
			log.Fatal().Err(err).Msg("❌ Failed to open DB")
		}
		// Verify connection
		if err = con.Ping(); err != nil {
			log.Fatal().Err(err).Msg("❌ Failed to connect to DB")
		}
		// Optional: SQLite works best with 1 writer
		Pool = con
		Pool.SetMaxOpenConns(1)
	})
}
