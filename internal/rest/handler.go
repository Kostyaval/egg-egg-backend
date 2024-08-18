package rest

import (
	"github.com/gofiber/fiber/v2"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"log/slog"
)

type handlerLogger struct {
	log *slog.Logger
}

func (l handlerLogger) HTTPRequest(c *fiber.Ctx) *slog.Logger {
	attr := slog.Group("req",
		slog.String("m", c.Method()),
		slog.String("uri", string(c.Request().URI().Path())),
		slog.String("ip", c.IP()),
	)

	return l.log.With(attr)
}

func (l handlerLogger) AuthorizedHTTPRequest(c *fiber.Ctx) (*slog.Logger, *domain.JWTClaims) {
	log := l.HTTPRequest(c)

	jwt, ok := c.Locals("jwt").(*domain.JWTClaims)
	if ok {
		return log.With(slog.Int64("uid", jwt.UID), slog.String("sid", jwt.JTI.String())), jwt
	}

	return log, nil
}

type handler struct {
	log *handlerLogger
	srv ServiceInterface
	cfg *config.Config
}

func newHandler(cfg *config.Config, logger *slog.Logger, srv ServiceInterface) *handler {
	return &handler{
		log: &handlerLogger{log: logger},
		srv: srv,
		cfg: cfg,
	}
}
