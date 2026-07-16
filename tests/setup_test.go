package tests

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"goploy/src/apis"
	"goploy/src/conf"
	"goploy/src/core"
	"goploy/src/db"
	"goploy/src/db/repos"
	"goploy/src/pkg"
	"goploy/src/service"

	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
)

var (
	App     *fiber.App
	DBConn  *sql.DB
	Queries *repos.Queries
)

func TestMain(m *testing.M) {
	var fiberApp *fiber.App
	var dbConn *sql.DB
	var queries *repos.Queries

	fxApp := fx.New(
		fx.NopLogger,
		conf.Module,
		core.Module,
		db.Module,
		apis.Module,
		pkg.Module,
		service.Module,
		fx.Populate(&fiberApp, &dbConn, &queries),
	)

	ctx := context.Background()
	if err := fxApp.Start(ctx); err != nil {
		panic("fx startup failed: " + err.Error())
	}

	App = fiberApp
	DBConn = dbConn
	Queries = queries
	code := m.Run()

	_ = fxApp.Stop(ctx)
	os.Exit(code)
}
