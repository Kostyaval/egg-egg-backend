package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"time"
)

type tapDB interface {
	UpdateUserTapCount(ctx context.Context, uid int64, count int) error
}

func (s Service) AddTap(ctx context.Context, uid int64, tapCount int) (domain.UserDocument, error) {
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

	if u.Taps.TapCount == 24000 {
		return u, domain.ErrTapOverLimit
	}

	inactiveTime := time.Now().Sub(u.PlayedAt.Time())
	if inactiveTime.Seconds() == 0 {
		return u, domain.ErrTapTooFast
	}

	pointsPerTap := u.Taps.TapBoostCount + 1

	energyUsedForTaps := tapCount * pointsPerTap
	accumulatedEnergy := (int(inactiveTime.Seconds()) * 36) / 10

	if energyUsedForTaps > accumulatedEnergy {
		energyUsedForTaps = accumulatedEnergy
		tapCount = (energyUsedForTaps / 36) * 10
	}

	totalPoints := tapCount * pointsPerTap

	err = s.db.UpdateUserTapCount(ctx, uid, totalPoints)
	if err != nil {
		return u, err
	}

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	u.Taps.EnergyCount = accumulatedEnergy - energyUsedForTaps

	return u, nil
}
