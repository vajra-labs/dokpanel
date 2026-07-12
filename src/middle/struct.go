package middle

import (
	"dokpanel/src/errx"

	"github.com/go-playground/validator/v10"
)

// StructValidator wraps go-playground/validator for Fiber's StructValidator interface.
// Register once in fiber.Config.StructValidator — validation runs automatically on Bind().
type StructValidator struct {
	validate *validator.Validate
}

func NewStructValidator() *StructValidator {
	return &StructValidator{validate: validator.New()}
}

func (v *StructValidator) Validate(out any) error {
	err := v.validate.Struct(out)
	if err != nil {
		return errx.BadRequestError(err.Error(), "VALIDATION_ERROR")
	}
	return nil
}
