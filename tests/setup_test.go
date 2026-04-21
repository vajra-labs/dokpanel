package tests

import (
	"dokpanel/src/conf"
	"dokpanel/src/db"
	"dokpanel/src/logger"
	"dokpanel/src/server"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestMain(m *testing.M) {
	conf.Init()
	logger.Init()
	// Connect to database
	db.Connect()
	// Run all tests cases
	code := m.Run()
	// Disconnect from database
	db.Disconnect()
	// Exit with the test result code
	os.Exit(code)
}

func SetupApp() *fiber.App {
	return server.NewApp()
}
