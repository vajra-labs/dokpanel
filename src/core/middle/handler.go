package middle

import (
	"errors"
	"net/http"
	"runtime/debug"

	"goploy/src/conf"
	"goploy/src/core/throw"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// Global Error Handler
func ErrorHandler(cfg *conf.Config) fiber.ErrorHandler {
	return func(ctx fiber.Ctx, err error) error {
		// Handle known HttpError
		if httpErr, ok := throw.IsHttpError(err); ok {
			cause := httpErr.Cause()
			if cause != nil {
				log.Error().
					Err(cause).
					Str("code", httpErr.Code).
					Int("status", httpErr.Status).
					Str("path", ctx.Path()).
					Str("method", ctx.Method()).
					Msg(httpErr.Message)
			}
			return httpErr.ToJSON(ctx)
		}
		// Handle Fiber built-in errors
		if e, ok := errors.AsType[*fiber.Error](err); ok {
			return ctx.Status(e.Code).JSON(fiber.Map{
				"status":  e.Code,
				"error":   http.StatusText(e.Code),
				"message": e.Message,
			})
		}
		// Unknown errors
		log.Error().
			Err(err).
			Str("path", ctx.Path()).
			Str("method", ctx.Method()).
			Msg("Unhandled error in ErrorHandler")
		message := "Something went wrong"
		stack := interface{}(nil)
		if cfg.IS_DEV {
			message = err.Error()
			stack = string(debug.Stack())
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  fiber.StatusInternalServerError,
			"error":   "InternalServerError",
			"message": message,
			"stack":   stack,
		})
	}
}

// NotFound Handler
func NotFoundHandler(ctx fiber.Ctx) error {
	path := ctx.Path()
	method := ctx.Method()
	// BadRequest Error
	return throw.BadRequestError(
		"Wrong Path",
		"NOT_FOUND",
		throw.WithMeta("path", path),
		throw.WithMeta("method", method),
	)
}
