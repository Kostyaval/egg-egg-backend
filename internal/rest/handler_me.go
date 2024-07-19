package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
	"time"
)

type meService interface {
	GetMe(ctx context.Context, uid int64) (domain.UserDocument, []byte, error)
}

func (h handler) me(c *fiber.Ctx) error {
	log := h.log.HTTPRequest(c)

	exp := 30 * time.Second
	if h.cfg.Runtime == config.RuntimeDevelopment {
		exp = 24 * time.Hour
	}

	if h.cfg.Runtime == config.RuntimeProduction {
		if err := initdata.Validate(string(c.Request().URI().QueryString()), h.cfg.TelegramToken, exp); err != nil {
			log.Error("validate initial data", slog.String("error", err.Error()))
			return c.Status(fiber.StatusForbidden).Send(nil)
		}
	}

	data, err := initdata.Parse(string(c.Request().URI().QueryString()))
	if err != nil {
		log.Error("parse initial data", slog.String("error", err.Error()))
		return c.Status(fiber.StatusBadRequest).Send(nil)
	}

	u, jwt, err := h.srv.GetMe(c.Context(), data.User.ID)
	if err != nil {
		log.Error("srv.GetMe", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) ||
			errors.Is(err, domain.ErrGhostUser) ||
			errors.Is(err, domain.ErrBannedUser) ||
			errors.Is(err, domain.ErrMultipleDevices) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	var res struct {
		domain.UserDocument
		Token string `json:"token"`
	}

	res.UserDocument = u
	res.Token = string(jwt)

	log.Info("me", slog.Int64("uid", u.Profile.Telegram.ID))

	return c.JSON(res)
}
