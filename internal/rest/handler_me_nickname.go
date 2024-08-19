package rest

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
	"regexp"
)

type nicknameService interface {
	CheckUserNickname(ctx context.Context, nickname string) (bool, error)
	UpdateUserNickname(ctx context.Context, uid int64, nickname string) error
}

var regexpNickname = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{3,30}[a-zA-Z0-9]$`)

func (h handler) checkUserNickname(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
	}

	var req struct {
		Nickname string `query:"n" validate:"required,min=4,max=32"`
	}

	if err := c.QueryParser(&req); err != nil {
		log.Error("QueryParser", slog.String("error", err.Error()))
		return newHTTPError(fiber.StatusBadRequest, "query string parser").withDetails(err)
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

	log.Info("check nickname", slog.String("nickname", req.Nickname))

	var res struct {
		Nickname  string `json:"nickname"`
		Available bool   `json:"isAvailable"`
	}

	res.Nickname = req.Nickname
	res.Available = ok

	return c.JSON(res)
}

func (h handler) updateUserNickname(c *fiber.Ctx) error {
	log, jwt := h.log.AuthorizedHTTPRequest(c)
	if jwt == nil {
		log.Debug("jwt is null")
		return c.Status(fiber.StatusUnauthorized).Send(nil)
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

	err := h.srv.UpdateUserNickname(c.Context(), jwt.UID, req.Nickname)
	if err != nil {
		log.Error("srv.UpdateUserNickname", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrConflictNickname) {
			return c.Status(fiber.StatusConflict).Send(nil)
		}

		if errors.Is(err, domain.ErrNoUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	log.Info("update nickname", slog.String("nickname", req.Nickname))

	return c.Status(fiber.StatusOK).Send(nil)
}
