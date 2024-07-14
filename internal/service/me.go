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
	UpdateUserNickname(ctx context.Context, uid int64, nickname string, jti uuid.UUID) error
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

func (s Service) CreateUserNickname(ctx context.Context, uid int64, nickname string) ([]byte, error) {
	ok, err := s.db.CheckUserNickname(ctx, nickname)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, domain.ErrConflictNickname
	}

	jwtClaims, err := domain.NewJWTClaims(uid, &nickname)
	if err != nil {
		return nil, err
	}

	token, err := jwtClaims.Encode(s.cfg.JWT)
	if err != nil {
		return nil, err
	}

	if err := s.db.UpdateUserNickname(ctx, uid, nickname, jwtClaims.JTI); err != nil {
		return nil, err
	}

	return token, nil
}
