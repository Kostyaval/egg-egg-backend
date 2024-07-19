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
