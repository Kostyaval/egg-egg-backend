package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type tapService interface {
	AddTap(ctx context.Context, uid int64, tapCount int) (domain.UserDocument, error)
	AddTapBoost(ctx context.Context, uid int64) (domain.UserDocument, error)
	AddTapEnergyBoost(ctx context.Context, uid int64) (domain.UserDocument, error)
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
		log.Error("BodyParser", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "body parser").withDetails(err)
	}

	if err := validate.Struct(req); err != nil {
		log.Error("validate request body", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "invalid body").withValidator(err)
	}

	u, err := h.srv.AddTap(c.Context(), jwt.UID, req.Count)
	if err != nil {
		log.Error("srv.AddTap", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		if errors.Is(err, domain.ErrTapOverLimit) || errors.Is(err, domain.ErrNoTapEnergy) {
			// for difficult to develop scripts for automatic tapping
			return c.JSON(u)
		}

		return newHTTPError(fiber.StatusInternalServerError, "error").withDetails(err)
	}

	log.Info("ok", slog.Int("count", req.Count))

	return c.JSON(u)
}

func (h handler) addTapBoost(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var (
		u   domain.UserDocument
		err error
	)

	u, err = h.srv.AddTapBoost(c.Context(), jwt.UID)
	if err != nil {
		log.Error("srv.AddTapBoost", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		if errors.Is(err, domain.ErrNoPoints) || errors.Is(err, domain.ErrNoBoost) {
			return newHTTPError(fiber.StatusBadRequest, err.Error())
		}

		return newHTTPError(fiber.StatusInternalServerError, "error").withDetails(err)
	}

	log.Info("ok", slog.String("boost", fmt.Sprintf("%+v", u.Tap.Boost)), slog.Int("pts", u.Tap.Points))

	return c.JSON(u)
}

func (h handler) addTapEnergyBoost(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	u, err := h.srv.AddTapEnergyBoost(c.Context(), jwt.UID)
	if err != nil {
		log.Error("srv.AddTapEnergyBoost", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		if errors.Is(err, domain.ErrNoPoints) || errors.Is(err, domain.ErrNoBoost) {
			return newHTTPError(fiber.StatusBadRequest, err.Error())
		}

		return newHTTPError(fiber.StatusInternalServerError, "error").withDetails(err)
	}

	log.Info("ok", slog.String("boost", fmt.Sprintf("%+v", u.Tap.Energy.Boost)))

	return c.JSON(u)
}

func (h handler) rechargeTapEnergy(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	u, err := h.srv.RechargeTapEnergy(c.Context(), jwt.UID)
	if err != nil {
		log.Error("srv.RechargeTapEnergy", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		if errors.Is(err, domain.ErrNoEnergyRecharge) {
			return newHTTPError(fiber.StatusBadRequest, err.Error())
		}

		return newHTTPError(fiber.StatusInternalServerError, "error").withDetails(err)
	}

	log.Info("ok", slog.Int("available", u.Tap.Energy.RechargeAvailable))

	return c.JSON(u)
}
