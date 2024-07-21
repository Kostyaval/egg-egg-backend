package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type tapService interface {
	AddTap(ctx context.Context, uid int64, tapCount int64) (domain.UserDocument, error)
}

func (h handler) addTap(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var req struct {
		Count int64 `json:"count" validate:"required,min=1,max=1000"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Error("failed to parse request body: ", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	if err := validate.Struct(req); err != nil {
		log.Error("validate request body", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "invalid request body").withValidator(err)
	}

	u, _, err := h.srv.GetMe(c.Context(), jwt.UID)

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

	u, err = h.srv.AddTap(c.Context(), jwt.UID, req.Count)
	if err != nil {
		log.Error("srv.AddTap", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	var res struct {
		domain.UserDocument
	}

	res.UserDocument = u

	log.Info("me", slog.Int64("uid", u.Profile.Telegram.ID))

	return c.JSON(res)
}
