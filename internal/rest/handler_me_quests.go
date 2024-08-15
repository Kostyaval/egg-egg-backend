package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type questsService interface {
	StartQuest(ctx context.Context, uid int64, questName string) error
}

func (h handler) updateQuest(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var req struct {
		Name string `query:"n" validate:"required,oneof=telegram youtube x"`
	}

	if err := c.QueryParser(&req); err != nil {
		log.Error("QueryParser", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "query string parse error").withDetails(err)
	}

	if err := validate.Struct(req); err != nil {
		log.Error("validate query string", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "invalid query string").withValidator(err)
	}

	if err := h.srv.StartQuest(context.Background(), jwt.UID, req.Name); err != nil {
		log.Error("quests", slog.String("name", req.Name), slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) || errors.Is(err, domain.ErrGhostUser) || errors.Is(err, domain.ErrBannedUser) {
			return newHTTPError(fiber.StatusForbidden, err.Error())
		}

		if errors.Is(err, domain.ErrReplay) || errors.Is(err, domain.ErrInvalidQuest) {
			return newHTTPError(fiber.StatusBadRequest, err.Error())
		}

		return newHTTPError(fiber.StatusInternalServerError, "internal server error")
	}

	log.Info("quests", slog.String("name", req.Name))

	return nil
}
