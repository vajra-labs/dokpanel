package tests

import (
	"context"
	"os"
	"testing"

	"goploy/src/apis"
	"goploy/src/conf"
	"goploy/src/core"
	"goploy/src/db"
	"goploy/src/pkg"
	"goploy/src/service"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

var App *fiber.App

func TestMain(m *testing.M) {
	var fiberApp *fiber.App

	fxApp := fx.New(
		fx.NopLogger,
		conf.Module,
		db.Module,
		core.Module,
		apis.Module,
		pkg.Module,
		service.Module,
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
