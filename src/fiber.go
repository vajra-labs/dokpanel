package src

import (
	"time"

	"dokpanel/src/conf"
	"dokpanel/src/middle"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func Fiber(cfg *conf.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         cfg.NAME,
		BodyLimit:       cfg.BODY_LIMIT,
		ErrorHandler:    middle.ErrorHandler(cfg),
		IdleTimeout:     5 * time.Second,
		StructValidator: middle.NewStructValidator(),
	})

	// Stack trace only in dev
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.IS_DEV,
	}))
	// Logger Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))
	// Secure Headers
	app.Use(helmet.New(helmet.Config{}))
	// Rate Limiting
	app.Use(middle.RateLimit(middle.RateOption{
		Limit:  cfg.RATE_LIMIT_MAX_REQ,
		Window: cfg.RATE_LIMIT_WINDOWS,
	}))
	// CORS
	app.Use(cors.New(cors.Config{
		MaxAge:           86400,
		AllowOrigins:     []string{cfg.CORS_ORIGIN},
		AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Content-Length"},
	}))

	return app
}
