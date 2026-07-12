package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"dokpanel/src"
	"dokpanel/src/apis"
	"dokpanel/src/conf"
	"dokpanel/src/db"
	"dokpanel/src/docs"
	"dokpanel/src/logger"
	"dokpanel/web"

	"github.com/gofiber/fiber/v3"
	zerolog "github.com/rs/zerolog/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// StartServer starts the Fiber server and manages its lifecycle with fx.
func StartServer(lc fx.Lifecycle, app *fiber.App, cfg *conf.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			uri := fmt.Sprintf("%s:%d", cfg.HOST, cfg.PORT)
			go func() {
				if err := app.Listen(uri, fiber.ListenConfig{
					EnablePrefork: false,
				}); err != nil && !errors.Is(err, net.ErrClosed) {
					zerolog.Panic().Err(err).Msg("Failed to start server")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			zerolog.Info().Msg("Gracefully shutting down...")
			if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
				zerolog.Error().Err(err).Msg("Shutdown error")
			}
			return nil
		},
	})
}

// FxLogger is a helper function to provide a logger for fx.
var FxLogger = fx.WithLogger(func(cfg *conf.Config) fxevent.Logger {
	if cfg.IS_PROD {
		return fxevent.NopLogger
	}
	return &fxevent.ConsoleLogger{W: log.Writer()}
})

func main() {
	app := fx.New(
		FxLogger,
		conf.Module,
		logger.Module,
		db.Module,
		apis.Module,
		docs.Module,
		fx.Provide(src.Fiber),
		fx.Invoke(web.ServeSPA),
		fx.Invoke(StartServer),
	)
	app.Run()
}
