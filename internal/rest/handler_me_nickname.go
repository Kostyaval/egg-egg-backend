package rest

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"regexp"
)

type nicknameService interface {
	CheckUserNickname(ctx context.Context, nickname string) (bool, error)
}

var regexpNickname = regexp.MustCompile(`^(?i)[a-z][a-z0-9]{3,31}$`)

func (h handler) checkUserNickname(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	if jwt.Nickname != nil {
		log.Error("already has a nickname")
		return c.Status(fiber.StatusBadRequest).Send(nil)
	}

	var req struct {
		Nickname string `query:"n" validate:"required,min=4,max=32"`
	}

	if err := c.QueryParser(&req); err != nil {
		log.Error("QueryParser", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "query string parse error").withDetails(err)
	}

	if err := validate.Struct(req); err != nil {
		log.Error("validate query string", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "invalid query string").withValidator(err)
	}

	if !regexpNickname.MatchString(req.Nickname) {
		log.Error("invalid nickname format")
		return newHTTPError(fiber.StatusBadRequest, "invalid nickname format")
	}

	ok, err := h.srv.CheckUserNickname(c.Context(), req.Nickname)
	if err != nil {
		log.Error("srv.HasNickname", slog.String("error", err.Error()))
		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	var res struct {
		Nickname  string `json:"nickname"`
		Available bool   `json:"isAvailable"`
	}

	res.Nickname = req.Nickname
	res.Available = ok

	return c.JSON(res)
}
