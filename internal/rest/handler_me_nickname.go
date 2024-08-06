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
	CreateUserNickname(ctx context.Context, uid int64, nickname string) ([]byte, *domain.UserDocument, *domain.ReferralBonus, error)
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

	var res struct {
		Nickname  string `json:"nickname"`
		Available bool   `json:"isAvailable"`
	}

	res.Nickname = req.Nickname
	res.Available = ok

	return c.JSON(res)
}

func (h handler) createUserNickname(c *fiber.Ctx) error {
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

	token, user, ref, err := h.srv.CreateUserNickname(c.Context(), jwt.UID, req.Nickname)
	if err != nil {
		log.Error("srv.CreateUserNickname", slog.String("error", err.Error()))

		if errors.Is(err, domain.ErrConflictNickname) {
			return c.Status(fiber.StatusConflict).Send(nil)
		}

		if errors.Is(err, domain.ErrNoUser) {
			return c.Status(fiber.StatusForbidden).Send(nil)
		}

		return c.Status(fiber.StatusInternalServerError).Send(nil)
	}

	log.Info("create nickname")

	if ref != nil {
		log.Info("referral bonus", slog.Int("points", ref.UserPoints))

		log.Info(
			"referral bonus",
			slog.Int64("ref", ref.ReferralUserID),
			slog.Int("points", ref.ReferralUserPoints),
		)
	}

	var res struct {
		*domain.UserDocument
		Token string `json:"token"`
	}

	res.UserDocument = user
	res.Token = string(token)

	return c.JSON(res)
}
