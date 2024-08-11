package tg

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	tele "gopkg.in/telebot.v3"
	"log/slog"
)

type startHandlerDB interface {
	GetUserProfileWithID(ctx context.Context, uid int64) (domain.UserProfile, error)
	SetUserIsTelegramChannelMember(ctx context.Context, uid int64, channelID int64) error
	SetUserIsTelegramChannelLeft(ctx context.Context, uid int64, channelID int64) error
}

func (h handler) onChatMemberUpdate(c tele.Context) error {
	log := h.log.Message(c)
	ctx := context.Background()

	isAllowedChat := false

	for _, v := range h.rules.TelegramBotAllowedChannels {
		if int64(v) == c.Chat().ID {
			isAllowedChat = true
			break
		}
	}

	if !isAllowedChat {
		log.Error(domain.ErrNotAllowedTelegramChat.Error())
		return nil
	}

	if c.Sender().IsBot || c.Sender().IsForum {
		log.Error(domain.ErrInvalidUserType.Error())
		return nil
	}

	u, err := h.db.GetUserProfileWithID(ctx, c.Sender().ID)
	if err != nil {
		log.Error("db.GetUserProfileWithID", slog.String("error", err.Error()))
		return nil
	}

	if u.HasBan {
		log.Error(domain.ErrBannedUser.Error())
		return nil
	}

	if u.IsGhost {
		log.Error(domain.ErrGhostUser.Error())
		return nil
	}

	if c.ChatMember().NewChatMember.Role == tele.Member {
		log.Info("new channel member")

		if err := h.db.SetUserIsTelegramChannelMember(ctx, c.Sender().ID, c.Chat().ID); err != nil {
			log.Error("db.SetUserIsTelegramChannelMember", slog.String("error", err.Error()))
		}
	}

	if c.ChatMember().NewChatMember.Role == tele.Left {
		log.Info("member left channel")

		if err := h.db.SetUserIsTelegramChannelLeft(ctx, c.Sender().ID, c.Chat().ID); err != nil {
			log.Error("db.SetUserIsTelegramChannelLeft", slog.String("error", err.Error()))
		}
	}

	return nil
}
