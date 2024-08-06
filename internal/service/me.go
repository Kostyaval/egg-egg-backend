package service

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"math"
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

func (s Service) CheckUserNickname(ctx context.Context, nickname string) (bool, error) {
	return s.db.CheckUserNickname(ctx, nickname)
}

// CreateUserNickname update a user profile nickname from null to a `nickname` if no conflict with another user.
func (s Service) CreateUserNickname(ctx context.Context, uid int64, nickname string) ([]byte, *domain.UserDocument, *domain.ReferralBonus, error) {
	ok, err := s.db.CheckUserNickname(ctx, nickname)
	if err != nil {
		return nil, nil, nil, err
	}

	if !ok {
		return nil, nil, nil, domain.ErrConflictNickname
	}

	jwtClaims, err := domain.NewJWTClaims(uid, &nickname)
	if err != nil {
		return nil, nil, nil, err
	}

	token, err := s.cfg.JWT.Encode(jwtClaims)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := s.db.UpdateUserNickname(ctx, uid, nickname, jwtClaims.JTI); err != nil {
		return nil, nil, nil, err
	}

	user, err := s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(s.cfg.Rules.Referral) > 0 {
		if user.Profile.Referral != nil {
			refUser, err := s.db.GetUserDocumentWithID(ctx, user.Profile.Referral.ID)
			if err != nil {
				return token, &user, nil, err
			}

			if !refUser.Profile.HasBan && !refUser.Profile.IsGhost && refUser.Profile.Nickname != nil {
				ref := &domain.ReferralBonus{
					UserID:         user.Profile.Telegram.ID,
					ReferralUserID: refUser.Profile.Telegram.ID,
				}

				if user.Profile.Telegram.IsPremium {
					ref.UserPoints = s.cfg.Rules.Referral[0].Recipient.Premium
					ref.ReferralUserPoints = s.cfg.Rules.Referral[0].Sender.Premium
				} else {
					ref.UserPoints = s.cfg.Rules.Referral[0].Recipient.Plain
					ref.ReferralUserPoints = s.cfg.Rules.Referral[0].Sender.Plain
				}

				if _, err := s.db.IncPoints(context.Background(), user.Profile.Telegram.ID, ref.UserPoints); err != nil {
					ref.UserPoints = 0
					return token, &user, nil, err
				}

				// no need to check error, because always will be updated with taps
				_ = s.rdb.SetLeaderboardPlayerPoints(ctx, user.Profile.Telegram.ID, user.Level, ref.UserPoints)
				user.Points = ref.UserPoints

				if _, err := s.db.IncPointsWithReferral(context.Background(), refUser.Profile.Telegram.ID, ref.ReferralUserPoints); err != nil {
					ref.ReferralUserPoints = 0
					return token, &user, nil, err
				}

				// no need to check error, because always will be updated with taps
				_ = s.rdb.SetLeaderboardPlayerPoints(ctx, refUser.Profile.Telegram.ID, refUser.Level, ref.ReferralUserPoints+refUser.Points)

				return token, &user, ref, nil
			}
		}
	}

	return token, &user, nil, nil
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
