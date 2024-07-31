package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type tapService interface {
	AddTap(ctx context.Context, uid int64, tapCount int) (domain.UserDocument, error)
	AddTapBoost(ctx context.Context, uid int64) (domain.UserDocument, error)
	AddEnergyBoost(ctx context.Context, uid int64) (domain.UserDocument, error)
	RechargeTapEnergy(ctx context.Context, uid int64) (domain.UserDocument, error)
}

func (h handler) addTap(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var req struct {
		Count int `json:"count" validate:"required,min=1"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Error("failed to parse request body: ", slog.String("string", err.Error()))
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	if err := validate.Struct(req); err != nil {
		log.Error("validate request body", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "invalid request body").withValidator(err)
	}

	u, err := h.srv.AddTap(c.Context(), jwt.UID, req.Count)
	if err != nil {
		log.Error("srv.AddTap", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return newHTTPError(fiber.StatusForbidden, err.Error())
		}

		if !errors.Is(err, domain.ErrTapOverLimit) && !errors.Is(err, domain.ErrTapTooFast) {
			return newHTTPError(fiber.StatusInternalServerError, "add tap").withDetails(err)
		}
	}

	var res struct {
		domain.UserDocument
	}

	res.UserDocument = u

	log.Info("tap", slog.Int("count", req.Count))

	return c.JSON(res)
}

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

		if errors.Is(err, domain.ErrNoPoints) {
			return newHTTPError(fiber.StatusBadRequest, "no points").withDetails(err)
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
