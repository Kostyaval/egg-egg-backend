package rest

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type friendsService interface {
	ReadUserFriends(ctx context.Context, uid int64, limit int64, skip int64) ([]domain.Friend, int64, error)
}

func (h handler) readUserFriends(c *fiber.Ctx) error {
	var err error

	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var req struct {
		Limit int64 `query:"l" validate:"omitempty,oneof=10 25 50 100"`
		Skip  int64 `query:"s" validate:"omitempty,min=0,max=1000000"`
	}

	if err := c.QueryParser(&req); err != nil {
		log.Error("QueryParser", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "query string parse error").withDetails(err)
	}

	if err := validate.Struct(req); err != nil {
		log.Error("validate query string", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "invalid query string").withValidator(err)
	}

	if req.Limit == 0 {
		req.Limit = 50
	}

	var resp struct {
		List   []domain.Friend `json:"list"`
		Amount int64           `json:"amount"`
		Limit  int64           `json:"limit"`
		Skip   int64           `json:"skip"`
	}

	resp.List, resp.Amount, err = h.srv.ReadUserFriends(c.Context(), jwt.UID, req.Limit, req.Skip)
	if err != nil {
		log.Error("ReadUserFriends", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusInternalServerError, "read error").withDetails(err)
	}

	resp.Limit = req.Limit
	resp.Skip = req.Skip

	return c.JSON(resp)
}
