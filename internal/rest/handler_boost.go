package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type boostService interface {
	BoostTap(ctx context.Context, uid int64) (domain.UserDocument, error)
	BoostEnergy(ctx context.Context, uid int64) (domain.UserDocument, error)
}

func (h handler) addBoost(c *fiber.Ctx) error {
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

	log.Info("Subject")
	log.Info("Subject type", slog.String("type", req.Subject))

	if req.Subject != "tap" && req.Subject != "energy" {
		return newHTTPError(fiber.StatusBadRequest, "invalid subject").withDetails(errors.New("invalid subject"))
	}

	var (
		u   domain.UserDocument
		err error
	)

	switch req.Subject {
	case "tap":
		u, err = h.srv.BoostTap(c.Context(), jwt.UID)
	case "energy":
		u, err = h.srv.BoostEnergy(c.Context(), jwt.UID)
	}

	if err != nil {
		if errors.Is(err, domain.ErrInsufficientEggs) {
			return newHTTPError(fiber.StatusInternalServerError, "insufficient eggs").withDetails(err)
		}

		if errors.Is(err, domain.ErrBoostOverLimit) {
			return newHTTPError(fiber.StatusInternalServerError, "boost is over limit").withDetails(err)
		}

		log.Error("srv.Boost", slog.String("error", err.Error()))

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	var res struct {
		domain.UserDocument
	}

	res.UserDocument = u

	log.Info("tap", slog.String("count", req.Subject))

	return c.JSON(res)
}
