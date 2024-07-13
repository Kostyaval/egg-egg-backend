package service

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type meDB interface {
	GetUserWithID(ctx context.Context, uid int64) (*domain.UserProfile, error)
	UpdateUserJWT(ctx context.Context, uid int64, jti uuid.UUID) error
}

func (s Service) GetMe(ctx context.Context, uid int64, jti uuid.UUID) (*domain.UserProfile, error) {
	u, err := s.db.GetUserWithID(ctx, uid)
	if err != nil {
		return nil, err
	}

	if u == nil {
		return nil, domain.ErrNoUser
	}

	if u.IsGhost {
		return nil, domain.ErrGhostUser
	}

	if u.HasBan {
		return nil, domain.ErrBannedUser
	}

	if u.JTI != nil {
		return nil, domain.ErrMultipleDevices
	}

	if err := s.db.UpdateUserJWT(ctx, uid, jti); err != nil {
		return nil, err
	}

	return u, nil
}
