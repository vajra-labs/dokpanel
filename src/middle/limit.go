package middle

import (
	"time"

	"dokpanel/src/errx"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

type RateOption struct {
	Limit   int
	Window  time.Duration
	Message string
	Code    string
}

func RateLimit(opts RateOption) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        opts.Limit,
		Expiration: opts.Window,
		KeyGenerator: func(c fiber.Ctx) string {
			ip := c.IP()
			if ip == "" {
				return "unknown"
			}
			return ip
		},
		LimitReached: func(c fiber.Ctx) error {
			msg := opts.Message
			if msg == "" {
				msg = "There are too many requests."
			}
			code := opts.Code
			if code == "" {
				code = "RATE_LIMIT_EXCEEDED"
			}
			return errx.NewError(
				fiber.StatusTooManyRequests,
				code,
				msg,
				errx.WithMeta("limitReq", opts.Limit),
				errx.WithMeta("windowMs", opts.Window.String()),
			)
		},
	})
}
