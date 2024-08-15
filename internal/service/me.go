package service

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
	"math/big"
	"strconv"
	"time"
)

type meDB interface {
	GetUserDocumentWithID(ctx context.Context, uid int64) (domain.UserDocument, error)
	UpdateUserJWT(ctx context.Context, uid int64, jti uuid.UUID) error
	CheckUserNickname(ctx context.Context, nickname string) (bool, error)
	UpdateUserNickname(ctx context.Context, uid int64, nickname string, jti uuid.UUID) error
	IncPointsWithReferral(ctx context.Context, uid int64, points int) (int, error)
	IncPoints(ctx context.Context, uid int64, points int) (int, error)
	SetPoints(ctx context.Context, uid int64, points int) error
	SetDailyReward(ctx context.Context, uid int64, points int, reward *domain.DailyReward) error
	CreateUserAutoClicker(ctx context.Context, uid int64, cost int) (domain.UserDocument, error)
	UpdateUserAutoClicker(ctx context.Context, uid int64, isEnabled bool) (domain.UserDocument, error)
	ReadTotalUserReferrals(ctx context.Context, uid int64) (int64, error)
	UpdateUserLevel(ctx context.Context, uid int64, level int, cost int) (domain.UserDocument, error)
	CreateUser(ctx context.Context, user *domain.UserDocument) error
	UpdateUserQuests(ctx context.Context, uid int64, quests domain.UserQuests) error
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

	jwtClaims, err := domain.NewJWTClaims(u.Profile.Telegram.ID, u.Profile.Nickname)
	if err != nil {
		return u, nil, err
	}

	jwtBytes, err := s.cfg.JWT.Encode(jwtClaims)
	if err != nil {
		return u, nil, err
	}

	// TODO optimize it for unite with getMe
	if s.checkUserQuests(&u) {
		if err := s.db.SetPoints(ctx, u.Profile.Telegram.ID, u.Points); err != nil {
			return u, nil, err
		}

		if err := s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, u.Points); err != nil {
			return u, nil, err
		}

		if err := s.db.UpdateUserQuests(ctx, u.Profile.Telegram.ID, u.Quests); err != nil {
			return u, nil, err
		}
	}

	// TODO optimize it for unite with autoclicker
	dailyReward, withDailyRewardPoints := s.checkDailyReward(&u)
	if dailyReward != nil {
		if err := s.db.SetDailyReward(ctx, u.Profile.Telegram.ID, withDailyRewardPoints, dailyReward); err != nil {
			return u, nil, err
		}

		if u.Points != withDailyRewardPoints {
			u.Points = withDailyRewardPoints

			if err := s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, withDailyRewardPoints); err != nil {
				return u, nil, err
			}
		}
	}

	// TODO optimize it for unite with daily reward
	withAutoclickerPoints := s.checkAutoClicker(&u)
	if withAutoclickerPoints != u.Points {
		u.Points = withAutoclickerPoints

		if err := s.db.SetPoints(ctx, uid, withAutoclickerPoints); err != nil {
			return u, nil, err
		}

		if err := s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, withAutoclickerPoints); err != nil {
			return u, nil, err
		}
	}

	charge, _ := s.userTapEnergy(&u)
	u.Tap.Energy.Charge = charge

	if u.Tap.Energy.RechargedAt.Time().UTC().Day() != time.Now().UTC().Day() {
		prevPlayedAt := u.PlayedAt
		u, err = s.db.UpdateUserTapEnergyRecharge(
			ctx,
			uid,
			s.cfg.Rules.Taps[u.Level].Energy.RechargeAvailable,
			charge,
			u.Points,
		)

		if err != nil {
			return u, nil, err
		}

		u.PlayedAt = prevPlayedAt
	}

	if err := s.db.UpdateUserJWT(ctx, uid, jwtClaims.JTI); err != nil {
		return u, nil, err
	}

	return u, jwtBytes, nil
}

func (s Service) checkDailyReward(u *domain.UserDocument) (*domain.DailyReward, int) {
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfYesterday := startOfToday.AddDate(0, 0, -1)

	if u.DailyReward.ReceivedAt.Time().After(startOfToday) || u.DailyReward.ReceivedAt.Time().Equal(startOfToday) {
		return nil, u.Points
	}

	dr := &domain.DailyReward{Day: u.DailyReward.Day}
	pts := u.Points

	if u.DailyReward.ReceivedAt.Time().After(startOfYesterday) || u.DailyReward.ReceivedAt.Time().Equal(startOfYesterday) {
		if u.DailyReward.Day >= len(s.cfg.Rules.DailyRewards) {
			dr.Day = 1
		} else {
			dr.Day++
		}

		pts += s.cfg.Rules.DailyRewards[dr.Day-1]
	} else {
		dr.Day = 1
	}

	return dr, pts
}

func (s Service) checkAutoClicker(u *domain.UserDocument) int {
	if !u.AutoClicker.IsAvailable || !u.AutoClicker.IsEnabled {
		return u.Points
	}

	delta := time.Now().Truncate(time.Second).UTC().Sub(u.PlayedAt.Time()).Seconds()
	if delta <= 0 {
		return u.Points
	}

	if delta >= s.cfg.Rules.AutoClicker.TTL.Seconds() {
		return u.Points + int(math.Floor(s.cfg.Rules.AutoClicker.TTL.Seconds()/s.cfg.Rules.AutoClicker.Speed.Seconds()))
	}

	return u.Points + int(math.Floor(delta/s.cfg.Rules.AutoClicker.Speed.Seconds()))
}

func (s Service) checkUserQuests(u *domain.UserDocument) bool {
	var (
		hasUpdate      bool
		now            = time.Now().UTC()
		solvedRandTime = func(startedAt time.Time) bool {
			if startedAt.After(now.Add(-2 * time.Hour)) {
				return false
			}

			if startedAt.Add(24 * time.Hour).Before(now) {
				return true
			}

			d := 24 * time.Hour
			n, err := rand.Int(rand.Reader, big.NewInt(d.Nanoseconds()))
			if err != nil {
				return false
			}

			return startedAt.Add(time.Duration(n.Int64())).Before(now)
		}
	)

	if u.Quests.Telegram == -1 && solvedRandTime(u.Quests.TelegramStartedAt.Time()) {
		u.Points += s.cfg.Rules.Quests.Telegram
		u.Quests.Telegram = 1
		hasUpdate = true
	}

	if u.Quests.Youtube == -1 && solvedRandTime(u.Quests.YoutubeStartedAt.Time()) {
		u.Points += s.cfg.Rules.Quests.Youtube
		u.Quests.Youtube = 1
		hasUpdate = true
	}

	if u.Quests.X == -1 && solvedRandTime(u.Quests.XStartedAt.Time()) {
		u.Points += s.cfg.Rules.Quests.X
		u.Quests.X = 1
		hasUpdate = true
	}

	return hasUpdate
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
			u.Profile.Nickname = &u.Profile.Telegram.Username
		}
	}

	if !isFreeNickname {
		randNickname, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 8)
		if err != nil {
			return nil, err
		}

		randNickname = "_" + randNickname
		u.Profile.Nickname = &randNickname
	}

	// set jwt
	jwtClaims, err := domain.NewJWTClaims(u.Profile.Telegram.ID, u.Profile.Nickname)
	if err != nil {
		return nil, err
	}

	jwtBytes, err := s.cfg.JWT.Encode(jwtClaims)
	if err != nil {
		return nil, err
	}

	u.Profile.JTI = &jwtClaims.JTI

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
			if refUserPoints, err := s.db.IncPointsWithReferral(ctx, refUser.Profile.Telegram.ID, inc); err == nil {
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

		if !refUser.Profile.IsGhost && !refUser.Profile.HasBan && refUser.Profile.Nickname != nil {
			u.Profile.Referral = &domain.ReferralUserProfile{
				ID:       refUser.Profile.Telegram.ID,
				Nickname: *refUser.Profile.Nickname,
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

// CreateUserNickname update a user profile nickname from null to a `nickname` if no conflict with another user.
func (s Service) CreateUserNickname(ctx context.Context, uid int64, nickname string) ([]byte, *domain.UserDocument, error) {
	ok, err := s.db.CheckUserNickname(ctx, nickname)
	if err != nil {
		return nil, nil, err
	}

	if !ok {
		return nil, nil, domain.ErrConflictNickname
	}

	jwtClaims, err := domain.NewJWTClaims(uid, &nickname)
	if err != nil {
		return nil, nil, err
	}

	token, err := s.cfg.JWT.Encode(jwtClaims)
	if err != nil {
		return nil, nil, err
	}

	if err := s.db.UpdateUserNickname(ctx, uid, nickname, jwtClaims.JTI); err != nil {
		return nil, nil, err
	}

	user, err := s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return nil, nil, err
	}

	return token, &user, nil
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

	if int(user.Level)+1 >= len(s.cfg.Rules.Taps) {
		return user, domain.ErrReachedLevelLimit
	}

	if user.Points < s.cfg.Rules.Taps[user.Level].NextLevel.Cost {
		return user, domain.ErrNoPoints
	}

	if len(s.cfg.Rules.Taps[user.Level].NextLevel.Tasks.Telegram) > 0 {
		if len(user.Tasks.Telegram) != len(s.cfg.Rules.Taps[user.Level].NextLevel.Tasks.Telegram) {
			return user, domain.ErrNotFollowedTelegramChannel
		}

		telegramTasksMap := make(map[int]bool)
		for _, value := range s.cfg.Rules.Taps[user.Level].NextLevel.Tasks.Telegram {
			telegramTasksMap[value] = true
		}

		for _, value := range user.Tasks.Telegram {
			if !telegramTasksMap[value] {
				return user, domain.ErrNotFollowedTelegramChannel
			}
		}
	}

	totalReferrals, err := s.db.ReadTotalUserReferrals(ctx, uid)
	if err != nil {
		return user, err
	}

	if totalReferrals < int64(s.cfg.Rules.Taps[user.Level].NextLevel.Tasks.Referral) {
		return user, domain.ErrNotEnoughReferrals
	}

	user, err = s.db.UpdateUserLevel(ctx, uid, int(user.Level)+1, s.cfg.Rules.Taps[user.Level].NextLevel.Cost)
	if err != nil {
		return user, err
	}

	return user, nil
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
