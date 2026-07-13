package health

import (
	"runtime"
	"time"

	"goploy/src/conf"

	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	cfg *conf.Config
}

// NewHandler accepts *conf.Config via fx injection
func NewHandler(cfg *conf.Config) *Handler {
	return &Handler{cfg: cfg}
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
	uptime := humanize.Time(h.cfg.START_TIME)
	result := HealthRes{
		Uptime:    uptime,
		Version:   conf.VERSION,
		GoEnv:     h.cfg.GO_ENV,
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
