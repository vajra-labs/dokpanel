package logger

import (
	"io"
	"os"
	"sync"

	"dokpanel/src/conf"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

func init() {
	once.Do(func() {
		var writers []io.Writer
		// Dev: Colored console only
		console := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: TIMESTAMP,
		}
		writers = append(writers, console)
		if conf.Env.IS_PROD {
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
	})
}
