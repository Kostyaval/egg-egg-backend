package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type energyRechargeService interface {
	RechargeEnergy(ctx context.Context, uid int64) (domain.UserDocument, error)
}

func (h handler) rechargeEnergy(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var res struct {
		domain.UserDocument
	}

	u, err := h.srv.RechargeEnergy(c.Context(), jwt.UID)

	if err != nil {
		if errors.Is(err, domain.ErrEnergyRechargeTooFast) {
			return newHTTPError(fiber.StatusInternalServerError, "energy recharge too fast").withDetails(err)
		}

		if errors.Is(err, domain.ErrEnergyRechargeOverLimit) {
			return newHTTPError(fiber.StatusInternalServerError, "energy recharge over limit").withDetails(err)
		}

		log.Error("srv.Recharge", slog.String("error", err.Error()))

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	res.UserDocument = u

	log.Info("energy recharge", slog.Int64("count", jwt.UID))

	return c.JSON(res)
}
