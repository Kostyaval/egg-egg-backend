package tg

import (
	"gitlab.com/egg-be/egg-backend/internal/config"
	tele "gopkg.in/telebot.v3"
	"time"
)

type Tg struct {
	Bot *tele.Bot
}

func NewTelegramBot(cfg *config.Config) (*Tg, error) {
	pref := tele.Settings{
		Token:  cfg.TelegramToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	return &Tg{
		Bot: bot,
	}, err
}
