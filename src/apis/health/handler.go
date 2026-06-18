package health

import (
	"runtime"
	"time"

	"dokpanel/src/conf"

	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	version *string
	env     *string
	startAt *time.Time
}

// Health Method
func NewHandler() *Handler {
	return &Handler{
		version: &conf.VERSION,
		env:     &conf.Env.GO_ENV,
		startAt: &conf.Env.START_TIME,
	}
}

// Server Ping Handler
func (h *Handler) Ping(ctx fiber.Ctx) error {
	return ctx.SendString("Pong!")
}

// Server Pong Handler
func (h *Handler) Pong(ctx fiber.Ctx) error {
	return ctx.SendString("Ping!")
}

// Server Health Handler
func (h *Handler) Health(ctx fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	uptime := humanize.Time(*h.startAt)
	result := HealthRes{
		Uptime:    uptime,
		Version:   *h.version,
		GoEnv:     *h.env,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Memory: MemoryUsage{
			Alloc: MemoryInfo{
				Bytes: m.Alloc,
				Human: humanize.Bytes(m.Alloc),
			},
			TotalAlloc: MemoryInfo{
				Bytes: m.TotalAlloc,
				Human: humanize.Bytes(m.TotalAlloc),
			},
			Sys: MemoryInfo{
				Bytes: m.Sys,
				Human: humanize.Bytes(m.Sys),
			},
			HeapAlloc: MemoryInfo{
				Bytes: m.HeapAlloc,
				Human: humanize.Bytes(m.HeapAlloc),
			},
			HeapSys: MemoryInfo{
				Bytes: m.HeapSys,
				Human: humanize.Bytes(m.HeapSys),
			},
		},
	}
	return ctx.JSON(result)
}
