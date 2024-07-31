package tg

import (
	"gitlab.com/egg-be/egg-backend/internal/config"
	tele "gopkg.in/telebot.v3"
	"log/slog"
)

type handlerLogger struct {
	log *slog.Logger
}

func (l handlerLogger) Message(c tele.Context) *slog.Logger {
	attr := slog.Group("msg",
		slog.String("txt", c.Text()),
		slog.Int64("uid", c.Sender().ID),
	)

	return l.log.With(attr)
}

type handler struct {
	cfg   *config.Config
	log   *handlerLogger
	rules *config.Rules
	db    DBInterface
}

func newHandler(logger *slog.Logger, rules *config.Rules, db DBInterface) *handler {
	return &handler{
		log:   &handlerLogger{log: logger},
		rules: rules,
		db:    db,
	}
}
