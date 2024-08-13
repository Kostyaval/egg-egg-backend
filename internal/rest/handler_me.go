package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
	"time"
)

type meService interface {
	GetMe(ctx context.Context, uid int64) (domain.UserDocument, []byte, error)
	CreateUser(ctx context.Context, u *domain.UserDocument, ref string) ([]byte, error)
}

func (h handler) me(c *fiber.Ctx) error {
	var (
		ctx = context.Background()
		log = h.log.HTTPRequest(c)
		res struct {
			domain.UserDocument
			Token string `json:"token"`
		}
	)

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

	if data.User.ID <= 0 {
		log.Error("parse initial data", slog.String("error", fmt.Sprintf("user id is incorrect - %d", data.User.ID)))
		return newHTTPError(fiber.StatusBadRequest, "incorrect user id or no initial data")
	}

	u, jwt, err := h.srv.GetMe(ctx, data.User.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoUser) {
			u = domain.NewUserDocument(h.cfg.Rules)
			u.Profile.Telegram.ID = data.User.ID
			u.Profile.Telegram.FirstName = data.User.FirstName
			u.Profile.Telegram.LastName = data.User.LastName
			u.Profile.Telegram.Username = data.User.Username
			u.Profile.Telegram.Language = data.User.LanguageCode
			u.Profile.Telegram.IsPremium = data.User.IsPremium
			u.Profile.Telegram.AllowsWriteToPm = data.User.AllowsWriteToPm

			jwt, err := h.srv.CreateUser(ctx, &u, data.StartParam)
			if err != nil {
				log.Error(
					"registration",
					slog.Int64("uid", u.Profile.Telegram.ID),
					slog.String("error", err.Error()),
				)

				return c.Status(fiber.StatusInternalServerError).Send(nil)
			}

			if u.Profile.Referral != nil {
				log.Info(
					"registration",
					slog.Int64("uid", u.Profile.Telegram.ID),
					slog.Int("pts", u.Points),
					slog.Int("nrg", u.Tap.Energy.Charge),
					slog.Int64("ref", u.Profile.Referral.ID),
				)
			} else {
				log.Info(
					"registration",
					slog.Int64("uid", u.Profile.Telegram.ID),
					slog.Int("pts", u.Points),
					slog.Int("nrg", u.Tap.Energy.Charge),
				)
			}

			res.UserDocument = u
			res.Token = string(jwt)

			return c.JSON(res)
		}

		log.Error(
			"initialization",
			slog.Int64("uid", data.User.ID),
			slog.String("error", err.Error()),
		)

		if errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	res.UserDocument = u
	res.Token = string(jwt)

	log.Info(
		"initialization",
		slog.Int64("uid", u.Profile.Telegram.ID),
		slog.Int("pts", u.Points),
		slog.Int("nrg", u.Tap.Energy.Charge),
	)

	return c.JSON(res)
}
