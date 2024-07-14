package rest

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"log/slog"
	"time"
)

type ServiceInterface interface {
	meService
	jwtRefreshService
	jwtDeleteService
	nicknameService
}

func NewREST(cfg *config.Config, logger *slog.Logger, srv ServiceInterface) *fiber.App {
	setupValidator()

	// Create fiber app
	app := fiber.New(fiber.Config{
		ProxyHeader:           fiber.HeaderXForwardedFor,
		DisableStartupMessage: true,
		IdleTimeout:           5 * time.Second,
		BodyLimit:             1 * 1024 * 1024, // max size 1MB
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				return ctx.Status(fiberErr.Code).JSON(
					newHTTPError(fiber.StatusInternalServerError, "unexpected error").withDetails(fiberErr))
			}

			var ee *httpError
			if errors.As(err, &ee) {
				return ctx.Status(ee.Status).JSON(ee)
			}

			return nil
		},
	})

	if cfg.Runtime == config.RuntimeProduction {
		app.Use(recover.New())
	}

	h := newHandler(cfg.JWT, logger, srv)
	app.Get("/ping", h.ping)
	app.Get("/me", h.me)

	app.Use(middlewareJWT(&middlewareJWTConfig{log: h.log, cfg: cfg.JWT, mustNickname: false}))
	app.Put("/me/token", h.jwtRefresh)
	app.Delete("/me/token", h.jwtDelete)
	app.Get("/me/nickname", h.checkUserNickname)

	return app
}
