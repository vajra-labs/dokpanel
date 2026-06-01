package middle

import (
	"net/http"
	"runtime/debug"

	"dokpanel/src/conf"
	"dokpanel/src/lib/core"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// Global Error Handler
func ErrorHandler(ctx fiber.Ctx, err error) error {
	// Handle known HttpError
	if httpErr, ok := core.IsHttpError(err); ok {
		return httpErr.ToJSON(ctx)
	}
	// Handle Fiber built-in errors
	if e, ok := err.(*fiber.Error); ok {
		return ctx.Status(e.Code).JSON(fiber.Map{
			"status":  e.Code,
			"error":   http.StatusText(e.Code),
			"message": e.Message,
		})
	}
	// Unknown errors
	log.Error().
		Err(err).
		Str("path", string(ctx.Request().URI().Path())).
		Str("method", ctx.Method()).
		Msg("Unhandled error in ErrorHandler")
	message := "Something went wrong"
	stack := interface{}(nil)
	if conf.Env.IS_DEV {
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

// NotFound Handler
func NotFoundHandler(ctx fiber.Ctx) error {
	path := ctx.Path()
	method := ctx.Method()
	// BadRequest Error
	return core.BadRequestError(
		"Wrong Path",
		"NOT_FOUND",
		core.WithMeta("path", path),
		core.WithMeta("method", method),
	)
}
