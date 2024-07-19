package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type jwtDB interface {
	DeleteUserJWT(ctx context.Context, uid int64) error
	GetUserProfileWithID(ctx context.Context, uid int64) (domain.UserProfile, error)
}

func (s Service) RefreshJWT(ctx context.Context, jwtClaims *domain.JWTClaims) ([]byte, error) {
	u, err := s.db.GetUserProfileWithID(ctx, jwtClaims.UID)
	if err != nil {
		return nil, err
	}

	if u.IsGhost {
		return nil, domain.ErrGhostUser
	}

	if u.HasBan {
		return nil, domain.ErrBannedUser
	}

	if u.JTI == nil {
		return nil, domain.ErrNoJWT
	}

	if *u.JTI != jwtClaims.JTI {
		return nil, domain.ErrCorruptJWT
	}

	newJWTClaims, err := domain.NewJWTClaims(u.Telegram.ID, u.Nickname)
	if err != nil {
		return nil, err
	}

	newJWTBytes, err := newJWTClaims.Encode(s.cfg.JWT)
	if err != nil {
		return nil, err
	}

	if err := s.db.UpdateUserJWT(ctx, u.Telegram.ID, newJWTClaims.JTI); err != nil {
		return nil, err
	}

	return newJWTBytes, nil
}

func (s Service) DeleteJWT(ctx context.Context, jwtClaims *domain.JWTClaims) error {
	return s.db.DeleteUserJWT(ctx, jwtClaims.UID)
}
