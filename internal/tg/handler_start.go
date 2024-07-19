package tg

import (
	"context"
	"errors"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	tele "gopkg.in/telebot.v3"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type startHandlerDB interface {
	GetUserProfileWithID(ctx context.Context, uid int64) (domain.UserProfile, error)
	RegisterUser(ctx context.Context, user *domain.UserProfile, points int) error
	IncReferralPoints(ctx context.Context, uid int64, points int) error
}

var regexpUserReferral = regexp.MustCompile(`^(?i)/start [0-9]+$`)

func (h handler) start(c tele.Context) error {
	log := h.log.Message(c)

	if c.Sender().IsBot || c.Sender().IsForum {
		log.Error(domain.ErrInvalidUserType.Error())
		return c.Send("Invalid Telegram user type")
	}

	u, err := h.db.GetUserProfileWithID(context.Background(), c.Sender().ID)
	if err != nil {
		// registration
		if errors.Is(err, domain.ErrNoUser) {
			userPoints := 0
			refUserPoints := 0
			user := &domain.UserProfile{
				Nickname:  nil,
				CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
				UpdatedAt: primitive.NewDateTimeFromTime(time.Now()),
				HasBan:    false,
				IsGhost:   false,
				Referral:  nil,
				Telegram: domain.TelegramUserProfile{
					ID:        c.Sender().ID,
					IsPremium: c.Sender().IsPremium,
					Firstname: c.Sender().FirstName,
					Lastname:  c.Sender().LastName,
					Language:  c.Sender().LanguageCode,
					Username:  c.Sender().Username,
				},
			}

			// check for reference id
			if regexpUserReferral.MatchString(c.Text()) {
				split := strings.Split(c.Text(), " ")

				refID, err := strconv.ParseInt(split[1], 10, 64)
				if err == nil {
					refUser, err := h.db.GetUserProfileWithID(context.Background(), refID)
					if err != nil {
						log.Error("db.GetUserProfileWithID refUser", slog.String("error", err.Error()))
					} else {
						if !refUser.IsGhost && !refUser.HasBan {
							user.Referral = &refID

							if len(h.rules.Referral) > 0 {
								if user.Telegram.IsPremium {
									userPoints = h.rules.Referral[0].Recipient.Premium
									refUserPoints = h.rules.Referral[0].Sender.Premium
								} else {
									userPoints = h.rules.Referral[0].Recipient.Plain
									refUserPoints = h.rules.Referral[0].Sender.Plain
								}

								if err := h.db.IncReferralPoints(context.Background(), refID, refUserPoints); err != nil {
									log.Error("db.AddUserPoints referral", slog.String("error", err.Error()))
									return c.Send("Oops! Something went wrong. Please try again later")
								}
							}
						}
					}
				}
			}

			if err := h.db.RegisterUser(context.Background(), user, userPoints); err != nil {
				log.Error("registration", slog.String("error", err.Error()))

				if refUserPoints > 0 {
					if err := h.db.IncReferralPoints(context.Background(), *user.Referral, -refUserPoints); err != nil {
						log.Error("db.SubUserPoints referral", slog.String("error", err.Error()))
					}
				}

				return c.Send("Oops! Something went wrong. Please try again later")
			}

			log.Info("registration")

			return nil
		}

		log.Error("db.GetUserProfileWithID", slog.String("error", err.Error()))

		return c.Send("Oops! Something went wrong. Please try again later")
	}

	if u.HasBan {
		log.Error(domain.ErrBannedUser.Error())
		return c.Send("You has been banned")
	}

	if u.IsGhost {
		log.Error(domain.ErrGhostUser.Error())
		return c.Send("Your player profile was deleted")
	}

	log.Info("exists")

	return nil
}
