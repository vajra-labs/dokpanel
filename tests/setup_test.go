package tests

import (
	"os"
	"testing"

	"dokpanel/src/db"
	_ "dokpanel/src/logger"
	"dokpanel/src/server"

	"github.com/gofiber/fiber/v3"
)

var App *fiber.App

func TestMain(m *testing.M) {
	App = server.New()

	// Run all tests cases
	code := m.Run()

	// Your cleanup tasks go here
	db.Pool.Close()
	os.Exit(code)
}
