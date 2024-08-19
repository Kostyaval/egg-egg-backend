package service

import (
	"context"
	"errors"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"time"
)

type meDB interface {
	GetUserDocumentWithID(ctx context.Context, uid int64) (domain.UserDocument, error)
	CheckUserNickname(ctx context.Context, nickname string) (bool, error)
	UpdateUserNickname(ctx context.Context, uid int64, nickname string) error
	IncPointsWithReferral(ctx context.Context, uid int64, points int, incNewUser bool) (int, error)
	IncPoints(ctx context.Context, uid int64, points int) (int, error)
	SetPoints(ctx context.Context, uid int64, points int) error
	SetDailyReward(ctx context.Context, uid int64, points int, reward *domain.DailyReward) error
	CreateUserAutoClicker(ctx context.Context, uid int64, cost int) (domain.UserDocument, error)
	UpdateUserAutoClicker(ctx context.Context, uid int64, isEnabled bool) (domain.UserDocument, error)
	UpdateUserLevel(ctx context.Context, uid int64, level int, cost int) (domain.UserDocument, error)
	CreateUser(ctx context.Context, user *domain.UserDocument) error
	UpdateUserQuests(ctx context.Context, uid int64, quests domain.UserQuests) error
	UpdateUserDocument(ctx context.Context, u *domain.UserDocument) error
}

type meRedis interface {
	SetLeaderboardPlayerPoints(ctx context.Context, uid int64, level domain.Level, points int) error
}

func (s Service) GetMe(ctx context.Context, uid int64) (domain.UserDocument, []byte, error) {
	u, err := s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, nil, err
	}

	if u.Profile.IsGhost {
		return u, nil, domain.ErrGhostUser
	}

	if u.Profile.HasBan {
		return u, nil, domain.ErrBannedUser
	}

	jwtClaims, err := domain.NewJWTClaims(u.Profile.Telegram.ID)
	if err != nil {
		return u, nil, err
	}

	jwtBytes, err := s.cfg.JWT.Encode(jwtClaims)
	if err != nil {
		return u, nil, err
	}

	u.Calculate(s.cfg.Rules)

	if err := s.db.UpdateUserDocument(ctx, &u); err != nil {
		return u, nil, err
	}

	_ = s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, u.Points)

	return u, jwtBytes, nil
}

func (s Service) CreateUser(ctx context.Context, u *domain.UserDocument, ref string) ([]byte, error) {
	var (
		isFreeNickname bool
		err            error
	)

	if u.Profile.Telegram.Username != "" {
		isFreeNickname, err = s.db.CheckUserNickname(ctx, u.Profile.Telegram.Username)
		if err != nil {
			return nil, err
		}

		if isFreeNickname {
			u.Profile.Nickname = u.Profile.Telegram.Username
		}
	}

	if !isFreeNickname {
		randNickname, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 8)
		if err != nil {
			return nil, err
		}

		u.Profile.Nickname = "_" + randNickname
	}

	// set jwt
	jwtClaims, err := domain.NewJWTClaims(u.Profile.Telegram.ID)
	if err != nil {
		return nil, err
	}

	jwtBytes, err := s.cfg.JWT.Encode(jwtClaims)
	if err != nil {
		return nil, err
	}

	// set referral
	refUser, err := s.setReferral(ctx, u, ref)
	if err != nil {
		return nil, err
	}

	if u.Profile.Referral != nil && len(s.cfg.Rules.Referral) > 0 {
		if u.Profile.Telegram.IsPremium {
			u.Points += s.cfg.Rules.Referral[0].Recipient.Premium
		} else {
			u.Points += s.cfg.Rules.Referral[0].Recipient.Plain
		}
	}

	// try to save here
	if err := s.db.CreateUser(ctx, u); err != nil {
		return nil, err
	}

	// no need to check error, because always will be updated with taps
	_ = s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, u.Points)

	// indicate that user has not played yet
	u.PlayedAt = primitive.NewDateTimeFromTime(time.Unix(0, 0).UTC())

	// save referral bonus
	if refUser != nil && len(s.cfg.Rules.Referral) > 0 {
		inc := 0

		if u.Profile.Telegram.IsPremium {
			inc = s.cfg.Rules.Referral[0].Sender.Premium
		} else {
			inc = s.cfg.Rules.Referral[0].Sender.Plain
		}

		if inc > 0 {
			if refUserPoints, err := s.db.IncPointsWithReferral(ctx, refUser.Profile.Telegram.ID, inc, true); err == nil {
				// no need to check error, because always will be updated with taps
				_ = s.rdb.SetLeaderboardPlayerPoints(ctx, refUser.Profile.Telegram.ID, refUser.Level, refUserPoints)
			}
		}
	}

	return jwtBytes, nil
}

func (s Service) setReferral(ctx context.Context, u *domain.UserDocument, ref string) (*domain.UserDocument, error) {
	if ref == "" {
		return nil, nil
	}

	if u.Profile.Referral != nil {
		return nil, nil
	}

	// plain referral parameter is user telegram id
	refID, err := strconv.ParseInt(ref, 10, 64)
	if err == nil {
		// prevent referral to self
		if refID == u.Profile.Telegram.ID {
			return nil, nil
		}

		refUser, err := s.db.GetUserDocumentWithID(ctx, refID)
		if err != nil {
			if errors.Is(err, domain.ErrNoUser) {
				return nil, nil
			}

			return nil, err
		}

		if !refUser.Profile.IsGhost && !refUser.Profile.HasBan && refUser.Profile.Nickname != "" {
			u.Profile.Referral = &domain.ReferralUserProfile{
				ID:       refUser.Profile.Telegram.ID,
				Nickname: refUser.Profile.Nickname,
			}
		}

		return &refUser, nil
	}

	// TODO implement another referral program
	return nil, nil
}

func (s Service) CheckUserNickname(ctx context.Context, nickname string) (bool, error) {
	return s.db.CheckUserNickname(ctx, nickname)
}

func (s Service) UpdateUserNickname(ctx context.Context, uid int64, nickname string) error {
	ok, err := s.db.CheckUserNickname(ctx, nickname)
	if err != nil {
		return err
	}

	if !ok {
		return domain.ErrConflictNickname
	}

	return s.db.UpdateUserNickname(ctx, uid, nickname)
}

func (s Service) CreateAutoClicker(ctx context.Context, uid int64) (domain.UserDocument, error) {
	user, err := s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return user, err
	}

	if user.Profile.IsGhost {
		return user, domain.ErrGhostUser
	}

	if user.Profile.HasBan {
		return user, domain.ErrBannedUser
	}

	if user.AutoClicker.IsAvailable {
		return user, domain.ErrHasAutoClicker
	}

	if user.Level < s.cfg.Rules.AutoClicker.MinLevel {
		return user, domain.ErrNoLevel
	}

	if user.Points < s.cfg.Rules.AutoClicker.Cost {
		return user, domain.ErrNoPoints
	}

	user, err = s.db.CreateUserAutoClicker(ctx, uid, s.cfg.Rules.AutoClicker.Cost)
	if err != nil {
		return user, err
	}

	if err := s.rdb.SetLeaderboardPlayerPoints(ctx, user.Profile.Telegram.ID, user.Level, user.Points); err != nil {
		return user, err
	}

	return user, nil
}

func (s Service) UpdateAutoClicker(ctx context.Context, uid int64) (domain.UserDocument, error) {
	user, err := s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return user, err
	}

	if user.Profile.IsGhost {
		return user, domain.ErrGhostUser
	}

	if user.Profile.HasBan {
		return user, domain.ErrBannedUser
	}

	if !user.AutoClicker.IsAvailable {
		return user, domain.ErrHasNoAutoClicker
	}

	return s.db.UpdateUserAutoClicker(ctx, uid, !user.AutoClicker.IsEnabled)
}

func (s Service) UpgradeLevel(ctx context.Context, uid int64) (domain.UserDocument, error) {
	u, err := s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	if u.Profile.IsGhost {
		return u, domain.ErrGhostUser
	}

	if u.Profile.HasBan {
		return u, domain.ErrBannedUser
	}

	u.Calculate(s.cfg.Rules)

	if !u.IsNextLevelAvailable {
		return u, domain.ErrNextLevelNotAvailable
	}

	u, err = s.db.UpdateUserLevel(ctx, uid, int(u.Level)+1, s.cfg.Rules.Taps[u.Level].NextLevel.Cost)
	if err != nil {
		return u, err
	}

	_ = s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, u.Points)

	// add referral points
	if u.Profile.Referral != nil {
		pts := s.cfg.Rules.Referral[u.Level].Sender.Plain
		if u.Profile.Telegram.IsPremium {
			pts = s.cfg.Rules.Referral[u.Level].Sender.Premium
		}

		pts, err = s.db.IncPointsWithReferral(ctx, u.Profile.Referral.ID, pts, false)
		if err == nil {
			_ = s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Referral.ID, u.Level, pts)
		}
	}

	return u, nil
}

func (s Service) StartQuest(ctx context.Context, uid int64, questName string) error {
	user, err := s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return err
	}

	if user.Profile.IsGhost {
		return domain.ErrGhostUser
	}

	if user.Profile.HasBan {
		return domain.ErrBannedUser
	}

	switch questName {
	case "telegram":
		if user.Quests.Telegram != 0 {
			return domain.ErrReplay
		}

		user.Quests.Telegram = -1
		user.Quests.TelegramStartedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	case "youtube":
		if user.Quests.Youtube != 0 {
			return domain.ErrReplay
		}

		user.Quests.Youtube = -1
		user.Quests.YoutubeStartedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	case "x":
		if user.Quests.X != 0 {
			return domain.ErrReplay
		}

		user.Quests.X = -1
		user.Quests.XStartedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	default:
		return domain.ErrInvalidQuest
	}

	if err := s.db.UpdateUserQuests(ctx, uid, user.Quests); err != nil {
		return err
	}

	return nil
}
