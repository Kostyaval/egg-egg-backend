package tg

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/speps/go-hashids/v2"
	"github.com/urfave/cli/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	tele "gopkg.in/telebot.v3"
	"log/slog"
	"math/big"
	"strings"
	"time"
	"unicode"
)

type resetHandlerDB interface {
	DeleteAllUsers(ctx context.Context) error
	CreateUsers(ctx context.Context, users []domain.UserDocument) error
}

type resetHandlerRedis interface {
	DeleteAllUsers(ctx context.Context) error
	SetLeaderboardPlayerPoints(ctx context.Context, uid int64, level domain.Level, points int) error
}

func (h handler) reset(c tele.Context) error {
	var err error

	log := h.log.Message(c)
	msg := c.Text()[1:] // remove command name `/` prefix
	msg = strings.ReplaceAll(msg, "—", "--")

	if c.Sender().IsBot || c.Sender().IsForum {
		log.Error(domain.ErrInvalidUserType.Error())
		return c.Send("Invalid Telegram user type")
	}

	app := &cli.App{
		Name:                 "reset",
		Usage:                "/reset [command]",
		EnableBashCompletion: false,
		CommandNotFound: func(cCtx *cli.Context, command string) {
			log.Error("command not found")

			_, err = fmt.Fprintf(cCtx.App.Writer, "No command %q\nTry /reset --help", command)
			if err != nil {
				log.Error(err.Error())
				_ = c.Send(err.Error(), tele.ModeMarkdownV2)
			}
		},
		OnUsageError: func(cCtx *cli.Context, err error, isSubcommand bool) error {
			log.Error("usage error")

			if isSubcommand {
				return err
			}

			_, err = fmt.Fprintf(cCtx.App.Writer, "WRONG: %#v\n", err)
			if err != nil {
				log.Error(err.Error())
				_ = c.Send(err.Error(), tele.ModeMarkdownV2)
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:        "all",
				Usage:       "/reset all",
				Description: "Delete all users from database and create new random users",
				Action: func(*cli.Context) error {
					ctx := context.Background()
					now := time.Now().UTC()
					limit := 100000
					users := make([]domain.UserDocument, limit)

					_ = c.Send(fmt.Sprintf("`start: %s. Please wait for a message about its done`", msg), tele.ModeMarkdownV2)

					if err := h.db.DeleteAllUsers(ctx); err != nil {
						log.Error("db.DeleteAllUsers", slog.String("error", err.Error()))
						return err
					}

					if err := h.rdb.DeleteAllUsers(ctx); err != nil {
						log.Error("rdb.DeleteAllUsers", slog.String("error", err.Error()))
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
					hd := hashids.NewData()
					hd.Salt = "spam"
					hd.MinLength = 4
					hd.Alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

					nickname := func(id int) *string {
						h, _ := hashids.NewWithData(hd)

						e, err := h.Encode([]int{id})
						if err != nil {
							return nil
						}

						if unicode.IsDigit(rune(e[0])) {
							e = "a" + e
						}

						return &e
					}

					// create users
					for i := 0; i < limit; i++ {
						createdAt := randTime(now.Add(-time.Hour*24*14), now)
						users[i] = domain.NewUserDocument(h.rules)

						users[i].Profile = domain.UserProfile{
							Telegram: domain.TelegramUserProfile{
								ID:        int64(400_000_000 + i),
								Username:  fmt.Sprintf("id%d", i),
								Language:  language(),
								IsPremium: isPremiumUser(),
							},
							Nickname:  nickname(400_000_000 + i),
							CreatedAt: primitive.NewDateTimeFromTime(createdAt),
							UpdatedAt: primitive.NewDateTimeFromTime(randTime(createdAt, now)),
						}

						users[i].Points = randInt(0, 1_000_000)
						users[i].Level = level()
						users[i].PlayedAt = primitive.NewDateTimeFromTime(randTime(now.Add(-time.Hour*48), now))
						users[i].Tap.PlayedAt = users[i].PlayedAt
						users[i].Tap.Energy.Charge = randInt(0, h.rules.TapsBaseEnergyCharge)

						if users[i].Level > 0 {
							users[i].Tasks.Telegram = h.rules.Taps[0].NextLevel.Tasks.Telegram
						} else {
							if randBool() {
								users[i].Tasks.Telegram = h.rules.Taps[0].NextLevel.Tasks.Telegram
							}
						}

						for j := 0; j <= int(users[i].Level); j++ {
							users[i].Tap.Boost[j] = randInt(0, h.rules.Taps[j].BoostAvailable)
							users[i].Tap.Points += users[i].Tap.Boost[j]
							users[i].Tap.Energy.Boost[j] = randInt(0, h.rules.Taps[j].Energy.BoostChargeAvailable)
							users[i].Tap.Energy.Charge += h.rules.Taps[j].Energy.BoostCharge
							users[i].Tap.Energy.RechargeAvailable = h.rules.Taps[j].Energy.BoostChargeAvailable
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

						if users[i].Level >= h.rules.AutoClicker.MinLevel {
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
								Nickname: *users[i].Profile.Nickname,
							}

							if users[ix].Profile.Telegram.IsPremium {
								users[i].ReferralPoints += h.rules.Referral[0].Sender.Premium
								users[i].Points += h.rules.Referral[0].Sender.Premium
								users[ix].Points += h.rules.Referral[0].Recipient.Premium
							} else {
								users[i].ReferralPoints += h.rules.Referral[0].Sender.Plain
								users[i].Points += h.rules.Referral[0].Sender.Plain
								users[ix].Points += h.rules.Referral[0].Recipient.Plain
							}

							if users[ix].Level > 0 {
								for j := users[ix].Level; j > 0; j-- {
									if users[ix].Profile.Telegram.IsPremium {
										users[i].ReferralPoints += h.rules.Referral[j].Sender.Premium
										users[ix].Points += h.rules.Referral[j].Sender.Premium
									} else {
										users[i].ReferralPoints += h.rules.Referral[j].Sender.Plain
										users[ix].Points += h.rules.Referral[j].Sender.Plain
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

					if err := h.db.CreateUsers(ctx, users); err != nil {
						return err
					}

					for _, u := range users {
						if err := h.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, u.Points); err != nil {
							return err
						}
					}

					log.Info("reset all")

					return c.Send(fmt.Sprintf("`ok: %s`", msg), tele.ModeMarkdownV2)
				},
			},
		},
	}

	app.Writer = newTgWriter(c)
	app.ErrWriter = newTgWriter(c)

	if err := app.Run(strings.Split(msg, " ")); err != nil {
		log.Error(err.Error())
		return c.Send(fmt.Sprintf("`%v`", err), tele.ModeMarkdownV2)
	}

	return nil
}
