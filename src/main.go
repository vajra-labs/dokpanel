package main

import (
	"fmt"

	"dokpanel/src/conf"
	"dokpanel/src/db"
	"dokpanel/src/logger"
	"dokpanel/src/server"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func main() {
	conf.Init()
	logger.Init()

	// Connect to database
	db.Connect()
	defer db.Disconnect()

	// Initialize server
	app := server.NewApp()
	addr := fmt.Sprintf("%s:%d", conf.Env.HOST, conf.Env.PORT)

	// Listen and serve
	if err := app.Listen(addr, fiber.ListenConfig{
		EnablePrefork: false,
	}); err != nil {
		log.Fatal().Err(err).Str("addr", addr).Msg("Failed to listen")
	}
}
