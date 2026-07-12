package docs

import (
	"encoding/json"
	"fmt"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v3"

	"dokpanel/src/conf"
)

type Handler struct {
	specJSON string
	cfg      *conf.Config
}

func Router(app fiber.Router, h *Handler) {
	app.Get("/api/docs", h.docs)
}

func newHandler(api huma.API, cfg *conf.Config) *Handler {
	b, err := json.Marshal(api.OpenAPI())
	if err != nil {
		panic(fmt.Sprintf("failed to marshal openapi spec: %v", err))
	}
	return &Handler{specJSON: string(b), cfg: cfg}
}

func (h *Handler) docs(c fiber.Ctx) error {
	html, err := scalar.ApiReferenceHTML(
		scalar.DefaultOptions(scalar.Options{
			SpecContent: h.specJSON,
			Theme:       scalar.ThemeDefault,
			Layout:      scalar.LayoutClassic,
			DarkMode:    false,
			CustomOptions: scalar.CustomOptions{
				PageTitle: fmt.Sprintf("%s API Docs", h.cfg.NAME),
			},
		}),
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}
