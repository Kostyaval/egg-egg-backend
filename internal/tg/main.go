package tg

import (
	"gitlab.com/egg-be/egg-backend/internal/config"
	tele "gopkg.in/telebot.v3"
	"log/slog"
	"time"
)

type DBInterface interface {
	startHandlerDB
}

type Tg struct {
	Bot *tele.Bot
}

func NewTelegramBot(cfg *config.Config, logger *slog.Logger, db DBInterface) (*Tg, error) {
	pref := tele.Settings{
		Token:  cfg.TelegramToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	h := newHandler(logger, db)
	bot.Handle("/start", h.start)

	return &Tg{
		Bot: bot,
	}, nil
}
