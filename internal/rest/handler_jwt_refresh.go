package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type jwtRefreshService interface {
	RefreshJWT(ctx context.Context, jwtClaims *domain.JWTClaims) ([]byte, error)
}

func (h handler) jwtRefresh(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	newJWT, err := h.srv.RefreshJWT(c.Context(), jwt)
	if err != nil {
		log.Error("srv.RefreshJWT", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) ||
			errors.Is(err, domain.ErrGhostUser) ||
			errors.Is(err, domain.ErrBannedUser) ||
			errors.Is(err, domain.ErrNoJWT) ||
			errors.Is(err, domain.ErrCorruptJWT) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	var res struct {
		Token string `json:"token"`
	}

	res.Token = string(newJWT)
	log.Info("jwt refresh")

	return c.JSON(res)

}
