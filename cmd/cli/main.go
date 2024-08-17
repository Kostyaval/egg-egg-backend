package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/urfave/cli/v2"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"gitlab.com/egg-be/egg-backend/internal/db"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"gitlab.com/egg-be/egg-backend/internal/rdb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"
	"math/big"
	"os"
	"time"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

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
		if err := redis.Close(); err != nil {
			logger.Error("redis close", slog.String("error", err.Error()))
		}
	}()

	app := &cli.App{
		Name:                 "cli",
		Usage:                "cli [command]",
		EnableBashCompletion: false,
		Commands: []*cli.Command{
			{
				Name:        "reset",
				Usage:       "egg reset",
				Description: "Delete all users from database and create new random users",
				Action: func(*cli.Context) error {
					ctx := context.Background()
					now := time.Now().UTC()
					limit := 100000
					users := make([]domain.UserDocument, limit)

					if err := mongodb.DeleteAllUsers(ctx); err != nil {
						log.Error("mongodb.DeleteAllUsers", slog.String("error", err.Error()))
						return err
					}

					if err := redis.DeleteAllUsers(ctx); err != nil {
						log.Error("redis.DeleteAllUsers", slog.String("error", err.Error()))
						return err
					}

					// utils -- random premium user
					premiumUsersCount := limit / 2
					isPremiumUser := func() bool {
						if premiumUsersCount == 0 {
							return false
						}

						n, err := rand.Int(rand.Reader, big.NewInt(2))
						if err != nil {
							return false
						}

						b := n.Int64() == 1
						if b {
							premiumUsersCount--
						}

						return b
					}

					// utils -- random user language
					language := func() string {
						l := []string{"en", "ru", "uk", "de", "es", "fr", "it", "pt", "zh", "ja"}

						n, err := rand.Int(rand.Reader, big.NewInt(int64(len(l))))
						if err != nil {
							return l[0]
						}

						return l[int(n.Int64())]
					}

					// utils -- random user level
					level := func() domain.Level {
						l := []domain.Level{domain.Lv0, domain.Lv1, domain.Lv2, domain.Lv3, domain.Lv4, domain.Lv5}

						n, err := rand.Int(rand.Reader, big.NewInt(int64(len(l))))
						if err != nil {
							return l[0]
						}

						return l[int(n.Int64())]
					}

					// utils -- random int
					randInt := func(min, max int) int {
						if min == max {
							return min
						}

						n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
						if err != nil {
							return 0
						}

						return int(n.Int64()) + min
					}

					// utils -- random time from range
					randTime := func(start, end time.Time) time.Time {
						if start.After(end) {
							return time.Now().UTC()
						}

						duration := end.Sub(start)

						n, err := rand.Int(rand.Reader, big.NewInt(duration.Nanoseconds()))
						if err != nil {
							return time.Now().UTC()
						}

						return start.Add(time.Duration(n.Int64()))
					}

					// utils -- random bool
					randBool := func() bool {
						n, err := rand.Int(rand.Reader, big.NewInt(2))
						if err != nil {
							return false
						}

						return n.Int64() == 1
					}

					// utils -- random nickname based on id
					nickname := func() (string, error) {
						randNickname, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", randInt(8, 16))
						if err != nil {
							return "", err
						}

						return randNickname, nil
					}

					// create users
					for i := 0; i < limit; i++ {
						createdAt := randTime(now.Add(-time.Hour*24*14), now)
						users[i] = domain.NewUserDocument(cfg.Rules)

						users[i].Profile = domain.UserProfile{
							Telegram: domain.TelegramUserProfile{
								ID:        int64(400_000_000 + i),
								Username:  fmt.Sprintf("id%d", i),
								Language:  language(),
								IsPremium: isPremiumUser(),
							},
							CreatedAt: primitive.NewDateTimeFromTime(createdAt),
							UpdatedAt: primitive.NewDateTimeFromTime(randTime(createdAt, now)),
						}

						users[i].Profile.Nickname, err = nickname()
						if err != nil {
							return err
						}

						users[i].Points = randInt(0, 1_000_000)
						users[i].Level = level()
						users[i].PlayedAt = primitive.NewDateTimeFromTime(randTime(now.Add(-time.Hour*48), now))
						users[i].Tap.PlayedAt = users[i].PlayedAt
						users[i].Tap.Energy.Charge = randInt(0, cfg.Rules.TapsBaseEnergyCharge)

						if users[i].Level > 0 {
							users[i].Profile.Channel.ID = cfg.Rules.TelegramBotAllowedChannels[0]
						} else {
							if randBool() {
								users[i].Profile.Channel.ID = cfg.Rules.TelegramBotAllowedChannels[0]
							}
						}

						for j := 0; j <= int(users[i].Level); j++ {
							users[i].Tap.Boost[j] = randInt(0, cfg.Rules.Taps[j].BoostAvailable)
							users[i].Tap.Points += users[i].Tap.Boost[j]
							users[i].Tap.Energy.Boost[j] = randInt(0, cfg.Rules.Taps[j].Energy.BoostChargeAvailable)
							users[i].Tap.Energy.Charge += cfg.Rules.Taps[j].Energy.BoostCharge
							users[i].Tap.Energy.RechargeAvailable = cfg.Rules.Taps[j].Energy.BoostChargeAvailable
							users[i].Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(randTime(
								time.Date(users[i].PlayedAt.Time().Year(), users[i].PlayedAt.Time().Month(), users[i].PlayedAt.Time().Day(), 0, 0, 0, 0, time.UTC),
								users[i].PlayedAt.Time(),
							))
						}
					}

					// set daily rewards and autoclicker
					for i := 0; i < limit; i++ {
						users[i].DailyReward = domain.DailyReward{
							ReceivedAt: primitive.NewDateTimeFromTime(randTime(users[i].PlayedAt.Time().Add(-time.Hour*12), users[i].PlayedAt.Time())),
							Day:        randInt(1, 10),
						}

						if users[i].Level >= cfg.Rules.AutoClicker.MinLevel {
							users[i].AutoClicker.IsAvailable = randBool()
							if users[i].AutoClicker.IsAvailable {
								users[i].AutoClicker.IsEnabled = randBool()
							}
						}
					}

					// set referrals
					for i := 0; i < limit; i++ {
						refLimit := randInt(1, 100)

						for k := 0; k < refLimit; k++ {
							ix := randInt(0, limit)

							if users[ix].Profile.Referral != nil {
								continue
							}

							users[ix].Profile.Referral = &domain.ReferralUserProfile{
								ID:       users[i].Profile.Telegram.ID,
								Nickname: users[i].Profile.Nickname,
							}

							users[i].ReferralCount++

							if users[ix].Profile.Telegram.IsPremium {
								users[i].ReferralPoints += cfg.Rules.Referral[0].Sender.Premium
								users[i].Points += cfg.Rules.Referral[0].Sender.Premium
								users[ix].Points += cfg.Rules.Referral[0].Recipient.Premium
							} else {
								users[i].ReferralPoints += cfg.Rules.Referral[0].Sender.Plain
								users[i].Points += cfg.Rules.Referral[0].Sender.Plain
								users[ix].Points += cfg.Rules.Referral[0].Recipient.Plain
							}

							if users[ix].Level > 0 {
								for j := users[ix].Level; j > 0; j-- {
									if users[ix].Profile.Telegram.IsPremium {
										users[i].ReferralPoints += cfg.Rules.Referral[j].Sender.Premium
										users[ix].Points += cfg.Rules.Referral[j].Sender.Premium
									} else {
										users[i].ReferralPoints += cfg.Rules.Referral[j].Sender.Plain
										users[ix].Points += cfg.Rules.Referral[j].Sender.Plain
									}
								}
							}
						}
					}

					// set leaderboard
					for i := 0; i < limit; i++ {
						n := users[i].Profile.Nickname

						for j := 0; j < limit; j++ {
							if users[i].Profile.Telegram.ID != users[j].Profile.Telegram.ID {
								if users[j].Profile.Nickname == n {
									return fmt.Errorf("nickname collision")
								}
							}
						}
					}

					if err := mongodb.CreateUsers(ctx, users); err != nil {
						return err
					}

					for _, u := range users {
						if err := redis.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, u.Points); err != nil {
							return err
						}
					}

					log.Info("egg reset")

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
