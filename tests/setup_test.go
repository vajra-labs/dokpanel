package tests

import (
	"os"
	"testing"

	"dokpanel/src"
	"dokpanel/src/db"
	_ "dokpanel/src/logger"

	"github.com/gofiber/fiber/v3"
)

var App *fiber.App

func TestMain(m *testing.M) {
	App = src.App()

	// Run all tests cases
	code := m.Run()

	// Your cleanup tasks go here
	db.Pool.Close()
	os.Exit(code)
}
