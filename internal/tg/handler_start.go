package tg

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	tele "gopkg.in/telebot.v3"
	"log/slog"
	"time"
)

type startHandlerDB interface {
	GetUserWithID(ctx context.Context, uid int64) (*domain.UserProfile, error)
	RegisterUser(ctx context.Context, user *domain.UserProfile) error
}

func (h handler) start(c tele.Context) error {
	log := h.log.Message(c)

	if c.Sender().IsBot || c.Sender().IsForum {
		log.Error(domain.ErrInvalidUserType.Error())
		return c.Send("Invalid Telegram user type")
	}

	u, err := h.db.GetUserWithID(context.Background(), c.Sender().ID)
	if err != nil {
		log.Error("db.GetUserWithID", slog.String("error", err.Error()))
		return c.Send("Oops! Something went wrong. Please try again later")
	}

	// registration
	if u == nil {
		user := &domain.UserProfile{
			Nickname:  nil,
			CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
			UpdatedAt: primitive.NewDateTimeFromTime(time.Now()),
			HasBan:    false,
			IsGhost:   false,
			Telegram: domain.TelegramUserProfile{
				ID:        c.Sender().ID,
				IsPremium: c.Sender().IsPremium,
				Firstname: c.Sender().FirstName,
				Lastname:  c.Sender().LastName,
				Language:  c.Sender().LanguageCode,
				Username:  c.Sender().Username,
			},
		}

		if err := h.db.RegisterUser(context.Background(), user); err != nil {
			log.Error("registration", slog.String("error", err.Error()))
			return c.Send("Oops! Something went wrong. Please try again later")
		}

		log.Info("registration")

		return nil
	}

	if u.HasBan {
		log.Error(domain.ErrBannedUser.Error())
		return c.Send("Your player profile has been banned")
	}

	if u.IsGhost {
		log.Error(domain.ErrGhostUser.Error())
		return c.Send("Your player profile is ghost")
	}

	log.Info("exists")

	return nil
}
