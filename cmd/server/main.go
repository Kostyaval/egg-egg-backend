package main

import (
	"gitlab.com/egg-be/egg-backend/internal/config"
	"gitlab.com/egg-be/egg-backend/internal/db"
	"gitlab.com/egg-be/egg-backend/internal/rdb"
	"gitlab.com/egg-be/egg-backend/internal/rest"
	"gitlab.com/egg-be/egg-backend/internal/service"
	"gitlab.com/egg-be/egg-backend/internal/tg"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Setup logger
	var logger *slog.Logger

	if cfg.Runtime == config.RuntimeDevelopment {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}

		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	} else {
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}

		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	slog.SetDefault(logger)

	// Setup MongoDB
	mongodb, err := db.NewMongoDB(cfg)
	if err != nil {
		logger.Error("new mongodb", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		if err := mongodb.Disconnect(); err != nil {
			logger.Error("mongodb disconnect", slog.String("error", err.Error()))
		}
	}()

	// Setup Redis
	redis, err := rdb.NewRedis(cfg)
	if err != nil {
		logger.Error("new redis", slog.String("error", err.Error()))
		os.Exit(1)
	}

	defer func() {
		if err := redis.Client.Close(); err != nil {
			logger.Error("redis close", slog.String("error", err.Error()))
		}
	}()

	// Setup Telegram bot
	bot, err := tg.NewTelegramBot(cfg, logger, mongodb)
	if err != nil {
		logger.Error("new telegram bot", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		logger.Info("start telegram bot")
		bot.Bot.Start()
	}()

	// Setup service
	srv := service.NewService(cfg, mongodb, redis)

	// Setup REST
	restApp := rest.NewREST(cfg, logger, srv)
	restAddr := "0.0.0.0:8000"

	go func() {
		logger.With(slog.String("addr", restAddr)).Info("start REST")

		if err := restApp.Listen(restAddr); err != nil {
			logger.Error("listen REST", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("service shutdown")

	if err := restApp.Shutdown(); err != nil {
		logger.Error("shutdown REST", slog.String("error", err.Error()))
		os.Exit(1)
	}

	bot.Bot.Stop()

	logger.Info("good bye")
}
