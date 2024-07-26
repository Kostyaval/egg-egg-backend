package service

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"time"
)

type tapDB interface {
	UpdateUserTapCount(ctx context.Context, uid int64, count int) error
	UpdateUserTapBoostCount(ctx context.Context, uid int64, cost int) error
	UpdateUserEnergyBoostCount(ctx context.Context, uid int64, cost int) error
	UpdateUserEnergyCount(ctx context.Context, uid int64, energyCount int) error
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

	inactiveTime := time.Now().Sub(u.Taps.PlayedAt.Time())
	if inactiveTime.Seconds() == 0 {
		return u, domain.ErrTapTooFast
	}

	pointsPerTap := u.Taps.TapBoostCount + 1

	energyNeededForTaps := tapCount * pointsPerTap
	log.Info("Inactive time")
	log.Info(inactiveTime.Seconds())
	accumulatedEnergy := (int(inactiveTime.Seconds())*36)/10 + u.Taps.EnergyCount
	levelParams := s.cfg.Rules.Taps[u.Level]
	log.Info("Accumulated energy: ", accumulatedEnergy)
	log.Info("Energy needed for taps: ", energyNeededForTaps)

	if accumulatedEnergy > levelParams.Energy {
		accumulatedEnergy = levelParams.Energy
	}

	if energyNeededForTaps > accumulatedEnergy {
		energyNeededForTaps = accumulatedEnergy
		tapCount = (energyNeededForTaps / levelParams.EnergyRechargeSeconds / 10) * 10
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
