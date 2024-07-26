package service

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type meDB interface {
	GetUserDocumentWithID(ctx context.Context, uid int64) (domain.UserDocument, error)
	UpdateUserJWT(ctx context.Context, uid int64, jti uuid.UUID) error
	CheckUserNickname(ctx context.Context, nickname string) (bool, error)
	UpdateUserNickname(ctx context.Context, uid int64, nickname string, jti uuid.UUID) error
	IncPointsWithReferral(ctx context.Context, uid int64, points int) (int, error)
	IncPoints(ctx context.Context, uid int64, points int) (int, error)
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

	if u.Profile.JTI != nil {
		return u, nil, domain.ErrMultipleDevices
	}

	jwtClaims, err := domain.NewJWTClaims(u.Profile.Telegram.ID, u.Profile.Nickname)
	if err != nil {
		return u, nil, err
	}

	jwtBytes, err := jwtClaims.Encode(s.cfg.JWT)
	if err != nil {
		return u, nil, err
	}

	if err := s.db.UpdateUserJWT(ctx, uid, jwtClaims.JTI); err != nil {
		return u, nil, err
	}

	return u, jwtBytes, nil
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

	token, err := jwtClaims.Encode(s.cfg.JWT)
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
