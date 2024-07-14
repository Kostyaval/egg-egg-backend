package service

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type meDB interface {
	GetUserWithID(ctx context.Context, uid int64) (*domain.UserProfile, error)
	UpdateUserJWT(ctx context.Context, uid int64, jti uuid.UUID) error
	CheckUserNickname(ctx context.Context, nickname string) (bool, error)
}

func (s Service) GetMe(ctx context.Context, uid int64) (*domain.UserProfile, []byte, error) {
	u, err := s.db.GetUserWithID(ctx, uid)
	if err != nil {
		return nil, nil, err
	}

	if u == nil {
		return nil, nil, domain.ErrNoUser
	}

	if u.IsGhost {
		return nil, nil, domain.ErrGhostUser
	}

	if u.HasBan {
		return nil, nil, domain.ErrBannedUser
	}

	if u.JTI != nil {
		return nil, nil, domain.ErrMultipleDevices
	}

	jwtClaims, err := domain.NewJWTClaims(u.Telegram.ID, u.Nickname)
	if err != nil {
		return nil, nil, err
	}

	jwtBytes, err := jwtClaims.Encode(s.cfg.JWT)
	if err != nil {
		return nil, nil, err
	}

	if err := s.db.UpdateUserJWT(ctx, uid, jwtClaims.JTI); err != nil {
		return nil, nil, err
	}

	return u, jwtBytes, nil
}

func (s Service) CheckUserNickname(ctx context.Context, nickname string) (bool, error) {
	return s.db.CheckUserNickname(ctx, nickname)
}
