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
	CreateUser(ctx context.Context, u *domain.UserDocument) ([]byte, error)
	SetMeReferral(ctx context.Context, u *domain.UserDocument, ref string) error
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
			log.Info("registration", slog.Int64("uid", data.User.ID))

			u = domain.NewUserDocument(h.cfg.Rules)
			u.Profile.Telegram.ID = data.User.ID
			u.Profile.Telegram.FirstName = data.User.FirstName
			u.Profile.Telegram.LastName = data.User.LastName
			u.Profile.Telegram.Username = data.User.Username
			u.Profile.Telegram.Language = data.User.LanguageCode
			u.Profile.Telegram.IsPremium = data.User.IsPremium
			u.Profile.Telegram.AllowsWriteToPm = data.User.AllowsWriteToPm

			if data.StartParam != "" {
				if err := h.srv.SetMeReferral(ctx, &u, data.StartParam); err != nil {
					log.Error("srv.SetMeReferral", slog.String("error", err.Error()))
				}
			}

			jwt, err := h.srv.CreateUser(ctx, &u)
			if err != nil {
				log.Error("srv.CreateUser", slog.String("error", err.Error()))
				return c.Status(fiber.StatusInternalServerError).Send(nil)
			}

			res.UserDocument = u
			res.Token = string(jwt)

			return c.JSON(res)
		}

		log.Error("srv.GetMe", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	if u.Profile.Nickname == nil && data.StartParam != "" {
		if err := h.srv.SetMeReferral(ctx, &u, data.StartParam); err != nil {
			log.Error("srv.SetMeReferral", slog.String("error", err.Error()))
		}
	}

	res.UserDocument = u
	res.Token = string(jwt)

	log.Info("me", slog.Int64("uid", u.Profile.Telegram.ID))

	return c.JSON(res)
}
