package rest

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type jwtDeleteService interface {
	DeleteJWT(ctx context.Context, jwtClaims *domain.JWTClaims) error
}

func (h handler) jwtDelete(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	if err := h.srv.DeleteJWT(c.Context(), jwt); err != nil {
		log.Error("srv.DeleteJWT", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	return nil
}
