package rest

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type tapService interface {
	AddTap(ctx context.Context, uid int64, tapCount int) (domain.UserDocument, error)
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
		log.Error("failed to parse request body: ", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	if err := validate.Struct(req); err != nil {
		log.Error("validate request body", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "invalid request body").withValidator(err)
	}

	u, err := h.srv.AddTap(c.Context(), jwt.UID, req.Count)
	if err != nil {
		log.Error("srv.AddTap", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	var res struct {
		domain.UserDocument
	}

	res.UserDocument = u

	log.Info("tap", slog.Int("count", req.Count))

	return c.JSON(res)
}
