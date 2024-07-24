package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type leaderboardService interface {
	ReadLeaderboard(ctx context.Context, uid int64, tab string, limit int64, skip int64) (domain.LeaderboardPlayer, []domain.LeaderboardPlayer, int64, error)
}

func (h handler) leaderboard(c *fiber.Ctx) error {
	var err error

	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var req struct {
		Limit int64  `query:"l" validate:"omitempty,oneof=10 25 50 100"`
		Skip  int64  `query:"s" validate:"omitempty,min=0,max=1000000000"`
		Tab   string `query:"t" validate:"omitempty,oneof=friends level global"`
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

	if req.Tab == "" {
		req.Tab = "friends"
	}

	if req.Tab == "friends" {
		req.Skip = 0
		req.Limit = 500
	}

	var res struct {
		Me    domain.LeaderboardPlayer   `json:"me"`
		List  []domain.LeaderboardPlayer `json:"list"`
		Total int64                      `json:"total"`
		Limit int64                      `json:"limit"`
		Skip  int64                      `json:"skip"`
	}

	res.Skip = req.Skip
	res.Limit = req.Limit

	res.Me, res.List, res.Total, err = h.srv.ReadLeaderboard(c.Context(), jwt.UID, req.Tab, req.Limit, req.Skip)
	if err != nil {
		log.Error("ReadLeaderboard", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrNoUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		return newHTTPError(fiber.StatusInternalServerError, "read leaderboard").withDetails(err)
	}

	return c.JSON(res)
}
