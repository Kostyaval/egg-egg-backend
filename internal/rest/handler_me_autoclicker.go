package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type autoClickerService interface {
	CreateAutoClicker(ctx context.Context, uid int64) (domain.UserDocument, error)
}

func (h handler) createAutoClicker(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	doc, err := h.srv.CreateAutoClicker(context.Background(), jwt.UID)
	if err != nil {
		log.Error("srv.CreateAutoClicker", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) {
			return newHTTPError(fiber.StatusForbidden, err.Error())
		}

		if errors.Is(err, domain.ErrNoPoints) || errors.Is(err, domain.ErrNoLevel) || errors.Is(err, domain.ErrHasAutoClicker) {
			return newHTTPError(fiber.StatusBadRequest, err.Error())
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	log.Info("created autoclicker")

	return c.JSON(doc)
}
