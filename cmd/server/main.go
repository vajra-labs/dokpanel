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
	"dokpanel/src/logger"
	"dokpanel/web"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

func StartServer(lc fx.Lifecycle, app *fiber.App, cfg *conf.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			uri := fmt.Sprintf("%s:%d", cfg.HOST, cfg.PORT)
			go func() {
				if err := app.Listen(uri, fiber.ListenConfig{
					EnablePrefork: false,
				}); err != nil && !errors.Is(err, net.ErrClosed) {
					log.Panic(err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Gracefully shutting down...")
			if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
				log.Printf("Shutdown error: %v\n", err)
			}
			return nil
		},
	})
}

func main() {
	app := fx.New(
		fx.NopLogger,
		conf.Module,
		logger.Module,
		db.Module,
		apis.Module,
		fx.Provide(src.Fiber),
		fx.Invoke(web.ServeSPA),
		fx.Invoke(StartServer),
	)
	app.Run()
}
