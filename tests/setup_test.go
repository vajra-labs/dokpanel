package tests

import (
	"context"
	"os"
	"testing"

	"dokpanel/src"
	"dokpanel/src/apis"
	"dokpanel/src/conf"
	"dokpanel/src/db"
	"dokpanel/src/logger"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

var App *fiber.App

func TestMain(m *testing.M) {
	var fiberApp *fiber.App

	fxApp := fx.New(
		fx.NopLogger,
		conf.Module,
		logger.Module,
		db.Module,
		apis.Module,
		fx.Provide(src.Fiber),
		fx.Populate(&fiberApp),
	)

	ctx := context.Background()
	if err := fxApp.Start(ctx); err != nil {
		panic("fx startup failed: " + err.Error())
	}

	App = fiberApp
	code := m.Run()

	_ = fxApp.Stop(ctx)
	os.Exit(code)
}
