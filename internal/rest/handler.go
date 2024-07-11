package rest

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type handlerLogger struct {
	log *slog.Logger
}

func (l handlerLogger) HTTPRequest(c *fiber.Ctx) *slog.Logger {
	attr := slog.Group("req",
		slog.String("m", c.Method()),
		slog.String("uri", string(c.Request().RequestURI())),
		slog.String("ip", c.IP()),
	)

	return l.log.With(attr)
}

type handler struct {
	log *handlerLogger
	srv ServiceInterface
}

func newHandler(srv ServiceInterface, logger *slog.Logger) *handler {
	return &handler{
		log: &handlerLogger{log: logger},
		srv: srv,
	}
}
