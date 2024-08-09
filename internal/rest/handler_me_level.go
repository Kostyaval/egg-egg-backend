package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type levelService interface {
	UpgradeLevel(ctx context.Context, uid int64) (domain.UserDocument, error)
}

func (h handler) upgradeLevel(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	doc, err := h.srv.UpgradeLevel(context.Background(), jwt.UID)
	if err != nil {
		log.Error("srv.UpgradeLevel", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return newHTTPError(fiber.StatusForbidden, err.Error())
		}

		if errors.Is(err, domain.ErrNoPoints) || errors.Is(err, domain.ErrReachedLevelLimit) ||
			errors.Is(err, domain.ErrNotFollowedTelegramChannel) || errors.Is(err, domain.ErrNotEnoughReferrals) {
			return newHTTPError(fiber.StatusBadRequest, err.Error())
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	log.Info("upgrade level")

	return c.JSON(doc)
}
