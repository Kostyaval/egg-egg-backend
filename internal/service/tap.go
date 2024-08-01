package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"math"
	"time"
)

type tapDB interface {
	UpdateUserTapCount(ctx context.Context, uid int64, count int) error
	UpdateUserPointsCount(ctx context.Context, uid int64, count int) error
	UpdateUserTapBoostCount(ctx context.Context, uid int64, cost int) error
	UpdateUserEnergyBoostCount(ctx context.Context, uid int64, cost int, level domain.Level) error
	UpdateUserEnergyCount(ctx context.Context, uid int64, energyCount int) error
	UpdateUserEnergyRechargeCount(ctx context.Context, uid int64) error
	ResetUserEnergyRechargeCount(ctx context.Context, uid int64) error
}

func getEnergyBoostCount(u domain.UserDocument) (int, error) {
	sum := 0
	for _, value := range u.Taps.TotalEnergyBoosts {
		sum += value
	}

	return sum, nil
}

func getUserMaxEnergy(u domain.UserDocument, boostPackage int) (int, error) {

	boostCount, err := getEnergyBoostCount(u)
	if err != nil {
		return 0, err
	}

	result := boostCount*boostPackage + boostPackage

	return result, nil
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

	inactiveTime := time.Since(u.Taps.PlayedAt.Time().UTC())
	if inactiveTime.Seconds() == 0 {
		return u, domain.ErrTapTooFast
	}

	pointsPerTap := u.Taps.TotalTapBoosts + 1

	energyNeededForTaps := tapCount * pointsPerTap

	levelParams := s.cfg.Rules.Taps[u.Level]

	userMaxEnergy, err := getUserMaxEnergy(u, levelParams.Energy.BoostPackage)
	if err != nil {
		return u, err
	}

	accumulatedEnergy := (int(math.Floor(inactiveTime.Seconds() * levelParams.Energy.RechargeSeconds))) + u.Taps.EnergyCount

	if accumulatedEnergy > userMaxEnergy {
		accumulatedEnergy = userMaxEnergy
	}

	if energyNeededForTaps > accumulatedEnergy {
		energyNeededForTaps = accumulatedEnergy
		tapCount = int(math.Floor(float64(energyNeededForTaps) / (levelParams.Energy.RechargeSeconds)))
	}

	totalPoints := tapCount * pointsPerTap

	err = s.db.UpdateUserTapCount(ctx, uid, tapCount)
	if err != nil {
		return u, err
	}

	err = s.db.UpdateUserPointsCount(ctx, uid, totalPoints)
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

	if err := s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, u.Points); err != nil {
		return u, err
	}

	return u, nil
}

func (s Service) RechargeTapEnergy(ctx context.Context, uid int64) (domain.UserDocument, error) {
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

	levelParams := s.cfg.Rules.Taps[u.Level]

	if u.Taps.EnergyRechargedAt.Time().UTC().Day() != time.Now().UTC().Day() {
		err = s.db.ResetUserEnergyRechargeCount(ctx, uid)
		if err != nil {
			return u, err
		}

		u.Taps.EnergyRechargeCount = 0
	}

	if int(time.Since(u.Taps.EnergyRechargedAt.Time().UTC()).Seconds()) < levelParams.Energy.FullRechargeDelaySeconds {
		return u, domain.ErrEnergyRechargeTooFast
	}

	if u.Taps.EnergyRechargeCount == levelParams.Energy.FullRechargeCount {
		return u, domain.ErrEnergyRechargeOverLimit
	}

	err = s.db.UpdateUserEnergyRechargeCount(ctx, uid)
	if err != nil {
		return u, err
	}

	userMaxEnergy, err := getUserMaxEnergy(u, levelParams.Energy.BoostPackage)
	if err != nil {
		return u, err
	}

	err = s.db.UpdateUserEnergyCount(ctx, uid, userMaxEnergy)
	if err != nil {
		return u, err
	}

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (s Service) AddTapBoost(ctx context.Context, uid int64) (domain.UserDocument, error) {
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

	levelParams := s.cfg.Rules.Taps[u.Level]

	if u.Taps.LevelTapBoosts == levelParams.Energy.BoostLimit {
		return u, domain.ErrBoostOverLimit
	}

	if u.Points < levelParams.Energy.BoostCost {
		return u, domain.ErrInsufficientEggs
	}

	err = s.db.UpdateUserTapBoostCount(ctx, uid, levelParams.Energy.BoostCost)
	if err != nil {
		return u, err
	}

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (s Service) AddEnergyBoost(ctx context.Context, uid int64) (domain.UserDocument, error) {
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

	levelParams := s.cfg.Rules.Taps[u.Level]

	if u.Taps.TotalEnergyBoosts[u.Level] == levelParams.Energy.BoostLimit {
		return u, domain.ErrBoostOverLimit
	}

	if u.Points < levelParams.Energy.BoostCost {
		return u, domain.ErrNoPoints
	}

	err = s.db.UpdateUserEnergyBoostCount(ctx, uid, levelParams.Energy.BoostCost, u.Level)
	if err != nil {
		return u, err
	}

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	return u, nil
}
