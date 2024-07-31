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
	CreateUser(ctx context.Context, user *domain.UserProfile, levelCount int) error
	UpdateReferralUserProfile(ctx context.Context, uid int64, ref *domain.ReferralUserProfile) error
}

var regexpUserReferral = regexp.MustCompile(`^(?i)/start [0-9]+$`)

func (h handler) start(c tele.Context) error {
	log := h.log.Message(c)
	ctx := context.Background()

	if c.Sender().IsBot || c.Sender().IsForum {
		log.Error(domain.ErrInvalidUserType.Error())
		return c.Send("Invalid Telegram user type")
	}

	u, err := h.db.GetUserProfileWithID(ctx, c.Sender().ID)
	if err != nil {
		// registration
		if errors.Is(err, domain.ErrNoUser) {
			user := &domain.UserProfile{
				Nickname:  nil,
				CreatedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				UpdatedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
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
					refUser, err := h.db.GetUserProfileWithID(ctx, refID)
					if err != nil {
						log.Error("db.GetUserProfileWithID refUser", slog.String("error", err.Error()))
					} else {
						if !refUser.IsGhost && !refUser.HasBan && refUser.Nickname != nil {
							user.Referral = &domain.ReferralUserProfile{
								ID:       refUser.Telegram.ID,
								Nickname: *refUser.Nickname,
							}
						}
					}
				}
			}

			if err := h.db.CreateUser(ctx, user, len(h.cfg.Rules.Taps)+1); err != nil {
				log.Error("registration", slog.String("error", err.Error()))
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

	if u.Nickname == nil && regexpUserReferral.MatchString(c.Text()) {
		split := strings.Split(c.Text(), " ")

		refID, err := strconv.ParseInt(split[1], 10, 64)
		if err == nil {
			if u.Referral == nil || (u.Referral != nil && u.Referral.ID != refID) {
				refUser, err := h.db.GetUserProfileWithID(ctx, refID)
				if err != nil {
					log.Error("db.GetUserProfileWithID refUser", slog.String("error", err.Error()))
				} else {
					if !refUser.IsGhost && !refUser.HasBan && refUser.Nickname != nil {
						u.Referral = &domain.ReferralUserProfile{
							ID:       refUser.Telegram.ID,
							Nickname: *refUser.Nickname,
						}

						if err := h.db.UpdateReferralUserProfile(ctx, u.Telegram.ID, u.Referral); err != nil {
							log.Error("db.UpdateReferralUserProfile", slog.String("error", err.Error()))
						}
					}
				}
			}
		}
	}

	log.Info("exists")

	return nil
}
