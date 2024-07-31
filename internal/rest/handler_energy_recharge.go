package rest

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

func (h handler) rechargeTapEnergy(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var res struct {
		domain.UserDocument
	}

	u, err := h.srv.RechargeTapEnergy(c.Context(), jwt.UID)

	if err != nil {
		log.Error("srv.RechargeTapEnergy", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return newHTTPError(fiber.StatusForbidden, err.Error())
		}

		if errors.Is(err, domain.ErrEnergyRechargeTooFast) {
			return newHTTPError(fiber.StatusBadRequest, "energy recharge too fast").withDetails(err)
		}

		if errors.Is(err, domain.ErrEnergyRechargeOverLimit) {
			return newHTTPError(fiber.StatusBadRequest, "energy recharge over limit").withDetails(err)
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	res.UserDocument = u

	return c.JSON(res)
}
