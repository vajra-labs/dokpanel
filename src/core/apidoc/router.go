package apidoc

import (
	"encoding/json/v2"
	"fmt"
	"sync"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/danielgtaylor/huma/v2"
	"github.com/gofiber/fiber/v3"

	"goploy/src/conf"
)

func invokeRouter(app *fiber.App, api huma.API, cfg *conf.Config) {
	var (
		once sync.Once
		html string
	)

	app.Get("/api/docs", func(ctx fiber.Ctx) error {
		var renderErr error

		once.Do(func() {
			b, err := json.Marshal(api.OpenAPI())
			if err != nil {
				renderErr = err
				return
			}
			html, renderErr = scalar.ApiReferenceHTML(
				scalar.DefaultOptions(scalar.Options{
					SpecContent: string(b),
					Theme:       scalar.ThemeDefault,
					Layout:      scalar.LayoutClassic,
					DarkMode:    false,
					PageTitle:   fmt.Sprintf("%s API Docs", cfg.NAME),
				}),
			)
		})

		if renderErr != nil {
			once = sync.Once{} // reset so next request retries
			return ctx.Status(fiber.StatusInternalServerError).SendString(renderErr.Error())
		}

		ctx.Set("Content-Type", "text/html")
		return ctx.SendString(html)
	})
}
