package core

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
)

type HttpError struct {
	Status  int            `json:"status"`
	Name    string         `json:"error"`
	Code    string         `json:"code,omitempty"`
	Message string         `json:"message"`
	Cause   error          `json:"-"`
	Meta    map[string]any `json:"meta,omitempty"`
}

type option func(*HttpError)

func NewError(
	status int,
	code string,
	message string,
	opts ...option,
) *HttpError {
	if status < 400 || status > 599 {
		status = fiber.StatusInternalServerError
	}
	err := &HttpError{
		Status:  status,
		Code:    code,
		Message: message,
		Name:    utils.StatusMessage(status),
		Meta:    make(map[string]any),
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}

// WithCause Method
func WithCause(err error) option {
	return func(e *HttpError) {
		e.Cause = err
	}
}

// WithMeta Method
func WithMeta(key string, value any) option {
	return func(e *HttpError) {
		if e.Meta == nil {
			e.Meta = make(map[string]any)
		}
		e.Meta[key] = value
	}
}

// IsHttpError Method
func IsHttpError(err error) (*HttpError, bool) {
	e, ok := err.(*HttpError)
	return e, ok
}

func (err *HttpError) Error() string {
	return err.Message
}

func (err *HttpError) ToJSON(ctx fiber.Ctx) error {
	return ctx.Status(err.Status).JSON(err)
}

// Create Method
func create(status int) func(message string, code string, opts ...option) *HttpError {
	return func(message string, code string, opts ...option) *HttpError {
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
