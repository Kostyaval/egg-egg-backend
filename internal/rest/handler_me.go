package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
	"strconv"
	"strings"
)

type meService interface {
	GetMe(ctx context.Context, uid int64, jti uuid.UUID) (*domain.UserProfile, error)
}

func (h handler) me(c *fiber.Ctx) error {
	log := h.log.HTTPRequest(c)
	log.Info("me")

	xTgID := c.Get("X-Telegram-Id")
	xTgID = strings.TrimSpace(xTgID)
	if xTgID == "" {
		log.Error("X-Telegram-Id", slog.String("error", "empty header"))
		return newHTTPError(fiber.StatusBadRequest, "empty telegram user id")
	}

	tgID, err := strconv.ParseInt(xTgID, 10, 64)
	if err != nil {
		log.Error("strconv.ParseInt", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "telegram user id bad format")
	}

	if tgID <= 0 {
		log.Error("bad telegram id")
		return newHTTPError(fiber.StatusBadRequest, "telegram user id bad format")
	}

	jwt, jti, err := jwtEncodeClaims(h.jwt, newJWTClaims(tgID))
	if err != nil {
		log.Error("jwtEncodeClaims", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	u, err := h.srv.GetMe(c.Context(), tgID, jti)
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
		UID      int64   `json:"uid"`
		Nickname *string `json:"nickname"`
		Language string  `json:"language"`
		Token    string  `json:"token"`
	}

	res.UID = u.Telegram.ID
	res.Nickname = u.Nickname
	res.Token = string(jwt)
	res.Language = u.Telegram.Language

	return c.JSON(res)
}
