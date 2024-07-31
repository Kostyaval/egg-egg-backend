package tg

import (
	"bytes"
	"fmt"
	"gitlab.com/egg-be/egg-backend/internal/config"
	tele "gopkg.in/telebot.v3"
	"log/slog"
	"sync"
	"time"
)

type DBInterface interface {
	startHandlerDB
	resetHandlerDB
}

type RedisInterface interface {
	resetHandlerRedis
}

type Tg struct {
	Bot   *tele.Bot
	Rules *config.Rules
}

func NewTelegramBot(cfg *config.Config, logger *slog.Logger, db DBInterface, rdb RedisInterface) (*Tg, error) {
	pref := tele.Settings{
		Token:  cfg.TelegramToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	h := newHandler(logger, cfg.Rules, db, rdb)
	bot.Handle("/start", h.start)

	if cfg.Runtime == config.RuntimeDevelopment {
		bot.Handle("/reset", h.reset)
	}

	return &Tg{
		Bot: bot,
	}, nil
}

type tgWriter struct {
	timeout time.Duration
	timer   *time.Timer
	buf     bytes.Buffer
	mu      sync.Mutex
	c       tele.Context
}

func newTgWriter(c tele.Context) *tgWriter {
	tw := &tgWriter{
		timeout: 1 * time.Second,
		c:       c,
	}

	return tw
}

func (tw *tgWriter) Write(p []byte) (n int, err error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	tw.resetTimer()
	return tw.buf.Write(p)
}

func (tw *tgWriter) resetTimer() {
	if tw.timer != nil {
		tw.timer.Stop()
	}

	tw.timer = time.AfterFunc(tw.timeout, func() {
		tw.mu.Lock()
		defer tw.mu.Unlock()

		_ = tw.c.Send(fmt.Sprintf("`%s`", tw.buf.String()), tele.ModeMarkdownV2)

		tw.buf.Reset()
	})
}
