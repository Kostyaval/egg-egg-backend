package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type tapDB interface {
	UpdateUserTapCount(ctx context.Context, uid int64, count int64) error
}

func (s Service) AddTap(ctx context.Context, uid int64, count int64) (domain.UserDocument, error) {
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

	if u.Profile.JTI != nil {
		return u, domain.ErrMultipleDevices
	}

	err = s.db.UpdateUserTapCount(ctx, uid, count)
	if err != nil {
		return u, err
	}

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	return u, nil
}
