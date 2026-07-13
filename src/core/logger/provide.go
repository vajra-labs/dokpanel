package logger

import (
	"io"
	"os"

	"goploy/src/conf"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	TIMESTAMP = "2006/01/02 15:04:05.00"
	LOG_PATH  = "./logs/server.log"
)

func configLogger(cfg *conf.Config) {
	var writers []io.Writer
	// Dev: Colored console only
	console := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: TIMESTAMP,
	}
	writers = append(writers, console)
	if cfg.IS_PROD {
		// Production: Stdout + Rotating file
		fileWriter := &lumberjack.Logger{
			Filename:   LOG_PATH,
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
		writers = append(writers, fileWriter)
	}
	// Optimized for Zerolog specifically
	multi := zerolog.MultiLevelWriter(writers...)
	log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()
	zerolog.TimeFieldFormat = TIMESTAMP
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}
