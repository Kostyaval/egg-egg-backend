package rest

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

func (h handler) addTapBoost(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var req struct {
		Subject string `query:"s" validate:"omitempty,oneof=tap energy"`
	}

	if err := c.QueryParser(&req); err != nil {
		log.Error("QueryParser", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "query string parse error").withDetails(err)
	}

	if err := validate.Struct(req); err != nil {
		log.Error("validate query string", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "invalid query string").withValidator(err)
	}

	var (
		u   domain.UserDocument
		err error
	)

	switch req.Subject {
	case "tap":
		u, err = h.srv.AddTapBoost(c.Context(), jwt.UID)
	case "energy":
		u, err = h.srv.AddEnergyBoost(c.Context(), jwt.UID)
	default:
		return newHTTPError(fiber.StatusBadRequest, "invalid subject").withDetails(errors.New("invalid subject"))
	}

	if err != nil {
		log.Error("srv.Boost", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return newHTTPError(fiber.StatusForbidden, err.Error())
		}

		if errors.Is(err, domain.ErrInsufficientEggs) {
			return newHTTPError(fiber.StatusBadRequest, "insufficient eggs").withDetails(err)
		}

		if errors.Is(err, domain.ErrBoostOverLimit) {
			return newHTTPError(fiber.StatusBadRequest, "boost is over limit").withDetails(err)
		}

		return newHTTPError(fiber.StatusInternalServerError, "tap boost").withDetails(err)
	}

	var res struct {
		domain.UserDocument
	}

	res.UserDocument = u

	log.Info("tap", slog.String("count", req.Subject))

	return c.JSON(res)
}
