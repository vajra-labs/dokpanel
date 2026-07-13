package auth

import (
	"goploy/src/service"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	tokens *service.TokenService
}

// NewHandler creates a new auth handler
func NewHandler(tokens *service.TokenService) *Handler {
	return &Handler{tokens: tokens}
}

// Login Handler
func (h *Handler) login(ctx fiber.Ctx) error {
	var body LoginDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	// TODO: implement login logic
	return ctx.JSON(fiber.Map{
		"token": "token-come",
	})
}

// Register Handler
func (h *Handler) register(ctx fiber.Ctx) error {
	var body RegisterDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	// TODO: implement register logic
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}
