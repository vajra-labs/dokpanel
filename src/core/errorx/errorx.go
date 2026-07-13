package errorx

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
)

type HttpError struct {
	Status  int            `json:"status"`
	Name    string         `json:"error"`
	Code    string         `json:"code,omitempty"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta,omitempty"`
	cause   error
}

func NewError(
	status int,
	code string,
	message string,
	opts ...Option,
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

// Error Method
func (err *HttpError) Error() string {
	return err.Message
}

// toJSON Method
func (err *HttpError) ToJSON(ctx fiber.Ctx) error {
	return ctx.Status(err.Status).JSON(err)
}

// Cause Method
func (e *HttpError) Cause() error {
	return e.cause
}
