package rest

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
	tapService
	friendsService
	leaderboardService
	autoClickerService
	levelService
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

	if cfg.CORS.IsEnabled {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.CORS.Origins,
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
			AllowCredentials: cfg.CORS.Origins != "*",
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		}))
	}

	h := newHandler(cfg, logger, srv)
	app.Get("/ping", h.ping)

	api := app.Group("/api")
	api.Get("/me", h.me)

	api.Use(middlewareJWT(&middlewareJWTConfig{log: h.log, cfg: cfg.JWT, mustNickname: false}))
	api.Put("/me/token", h.jwtRefresh)
	api.Delete("/me/token", h.jwtDelete)
	api.Get("/me/nickname", h.checkUserNickname)
	api.Post("/me/nickname", h.createUserNickname)
	api.Put("/me/level", h.upgradeLevel)

	api.Use(middlewareJWT(&middlewareJWTConfig{log: h.log, cfg: cfg.JWT, mustNickname: true}))
	api.Put("/me/tap", h.addTap)
	api.Put("/me/tap/boost", h.addTapBoost)
	api.Put("/me/tap/energy", h.rechargeTapEnergy)
	api.Put("/me/tap/energy/boost", h.addTapEnergyBoost)
	api.Post("/me/tap/autoclicker", h.createAutoClicker)
	api.Put("/me/tap/autoclicker", h.updateAutoClicker)

	api.Get("/me/friends", h.readUserFriends)
	api.Get("/leaderboard", h.leaderboard)

	return app
}
