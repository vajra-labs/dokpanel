package monitor

import (
	"github.com/gofiber/fiber/v3"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Metrics(ctx fiber.Ctx) error { return nil }

func (h *Handler) Containers(ctx fiber.Ctx) error { return nil }
