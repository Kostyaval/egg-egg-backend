package tg

import (
	"gitlab.com/egg-be/egg-backend/internal/domain"
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
	log   *handlerLogger
	rules *domain.Rules
	db    DBInterface
	rdb   RedisInterface
}

func newHandler(logger *slog.Logger, rules *domain.Rules, db DBInterface, rdb RedisInterface) *handler {
	return &handler{
		log:   &handlerLogger{log: logger},
		rules: rules,
		db:    db,
		rdb:   rdb,
	}
}
