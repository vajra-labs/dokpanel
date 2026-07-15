package throw

import "github.com/gofiber/fiber/v3"

// Create Method
func create(status int) func(message string, code string, opts ...Option) *HttpError {
	return func(message string, code string, opts ...Option) *HttpError {
		return NewError(status, code, message, opts...)
	}
}

// Common Global HttpErrors
var (
	BadRequestError      = create(fiber.StatusBadRequest)
	ConflictError        = create(fiber.StatusConflict)
	ForbiddenError       = create(fiber.StatusForbidden)
	NotFoundError        = create(fiber.StatusNotFound)
	UnauthorizedError    = create(fiber.StatusUnauthorized)
	InternalServerError  = create(fiber.StatusInternalServerError)
	ContentTooLargeError = create(fiber.StatusRequestEntityTooLarge)
)
