package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"log/slog"
	"time"
)

type ServiceInterface interface{}

func NewREST(cfg *config.Config, log *slog.Logger, srv ServiceInterface) *fiber.App {
	// Create fiber app
	app := fiber.New(fiber.Config{
		ProxyHeader:           fiber.HeaderXForwardedFor,
		DisableStartupMessage: true,
		IdleTimeout:           5 * time.Second,
		BodyLimit:             1 * 1024 * 1024, // max size 1MB
	})

	if cfg.Runtime == config.RuntimeProduction {
		app.Use(recover.New())
	}

	return app
}
