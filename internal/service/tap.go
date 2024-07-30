package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"time"
)

type tapDB interface {
	UpdateUserTapCount(ctx context.Context, uid int64, count int) error
	UpdateUserTapBoostCount(ctx context.Context, uid int64, cost int) error
	UpdateUserEnergyBoostCount(ctx context.Context, uid int64, cost int) error
	UpdateUserEnergyCount(ctx context.Context, uid int64, energyCount int) error
	UpdateUserEnergyRechargeCount(ctx context.Context, uid int64) error
	ResetUserEnergyRechargeCount(ctx context.Context, uid int64) error
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

	//if u.Profile.JTI != nil {
	//	return u, domain.ErrMultipleDevices
	//}

	if u.Taps.TapCount == 24000 {
		return u, domain.ErrTapOverLimit
	}

	inactiveTime := time.Since(u.Taps.PlayedAt.Time())
	if inactiveTime.Seconds() == 0 {
		return u, domain.ErrTapTooFast
	}

	pointsPerTap := u.Taps.TapBoostCount + 1

	energyNeededForTaps := tapCount * pointsPerTap

	levelParams := s.cfg.Rules.Taps[u.Level]
	userMaxEnergy := u.Taps.EnergyBoostCount*500 + 500
	accumulatedEnergy := (int(inactiveTime.Seconds())*levelParams.EnergyRechargeSeconds*10)/10 + u.Taps.EnergyCount

	if accumulatedEnergy > userMaxEnergy {
		accumulatedEnergy = userMaxEnergy
	}

	if energyNeededForTaps > accumulatedEnergy {
		energyNeededForTaps = accumulatedEnergy
		tapCount = energyNeededForTaps / (levelParams.EnergyRechargeSeconds * 10) / 10
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

	u.Taps.EnergyCount = accumulatedEnergy - energyNeededForTaps

	err = s.db.UpdateUserEnergyCount(ctx, uid, u.Taps.EnergyCount)
	if err != nil {
		return u, err
	}

	return u, nil
}
