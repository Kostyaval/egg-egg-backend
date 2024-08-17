package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type tapDB interface {
	UpdateUserTap(ctx context.Context, uid int64, tap domain.UserTap, points int) (domain.UserDocument, error)
	UpdateUserTapBoost(ctx context.Context, uid int64, boost []int, points int) (domain.UserDocument, error)
	UpdateUserTapEnergyBoost(ctx context.Context, uid int64, boost []int, charge int, points int) (domain.UserDocument, error)
	UpdateUserTapEnergyRecharge(ctx context.Context, uid int64, available int, chargeMax int, points int) (domain.UserDocument, error)
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

	u.Calculate(s.cfg.Rules)

	if u.Tap.Energy.Charge == 0 || u.Tap.Energy.Charge < u.Tap.Points {
		return u, domain.ErrNoTapEnergy
	}

	energySpent := tapCount * u.Tap.Points
	if energySpent > u.Tap.Energy.Charge {
		tapCount = u.Tap.Energy.Charge / u.Tap.Points
		energySpent = tapCount * u.Tap.Points
	}

	u.Tap.Count += tapCount
	u.Tap.Energy.Charge = u.Tap.Energy.Charge - energySpent
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(time.Now().UTC().Truncate(time.Second))

	u, err = s.db.UpdateUserTap(ctx, uid, u.Tap, u.Points+energySpent)
	if err != nil {
		return u, err
	}

	if err := s.rdb.SetLeaderboardPlayerPoints(ctx, u.Profile.Telegram.ID, u.Level, u.Points); err != nil {
		return u, err
	}

	return u, nil
}

func (s Service) RechargeTapEnergy(ctx context.Context, uid int64) (domain.UserDocument, error) {
	now := time.Now().UTC().Truncate(time.Second)

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

	pts := u.Points
	u.Calculate(s.cfg.Rules)

	if u.Tap.Energy.RechargeAvailable == 0 {
		return u, domain.ErrNoEnergyRecharge
	}

	if s.cfg.Rules.Taps[u.Level].Energy.RechargeAvailableAfter.Seconds() > 0 &&
		u.Tap.Energy.RechargeAvailable < s.cfg.Rules.Taps[u.Level].Energy.RechargeAvailable &&
		now.Sub(u.Tap.Energy.RechargedAt.Time().UTC()).Seconds() < s.cfg.Rules.Taps[u.Level].Energy.RechargeAvailableAfter.Seconds() {
		return u, domain.ErrNoEnergyRecharge
	}

	if u.Points != pts {
		if err := s.rdb.SetLeaderboardPlayerPoints(ctx, uid, u.Level, u.Points); err != nil {
			return u, err
		}
	}

	return s.db.UpdateUserTapEnergyRecharge(ctx, uid, u.Tap.Energy.RechargeAvailable-1, u.TapEnergyChargeMax(s.cfg.Rules), u.Points)
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

	u.Calculate(s.cfg.Rules)

	if u.Tap.Boost[u.Level] >= s.cfg.Rules.Taps[u.Level].BoostAvailable {
		return u, domain.ErrNoBoost
	}

	if u.Points < s.cfg.Rules.Taps[u.Level].BoostCost {
		return u, domain.ErrNoPoints
	}

	u.Points -= s.cfg.Rules.Taps[u.Level].BoostCost
	u.Tap.Boost[u.Level]++

	u, err = s.db.UpdateUserTapBoost(ctx, uid, u.Tap.Boost, u.Points)
	if err != nil {
		return u, err
	}

	if err := s.rdb.SetLeaderboardPlayerPoints(ctx, uid, u.Level, u.Points); err != nil {
		return u, err
	}

	return u, nil
}

func (s Service) AddTapEnergyBoost(ctx context.Context, uid int64) (domain.UserDocument, error) {
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

	u.Calculate(s.cfg.Rules)

	if u.Tap.Energy.Boost[u.Level] >= s.cfg.Rules.Taps[u.Level].Energy.BoostChargeAvailable {
		return u, domain.ErrNoBoost
	}

	if u.Points < s.cfg.Rules.Taps[u.Level].Energy.BoostChargeCost {
		return u, domain.ErrNoPoints
	}

	u.Points -= s.cfg.Rules.Taps[u.Level].Energy.BoostChargeCost
	u.Tap.Energy.Boost[u.Level]++
	u.TapEnergyChargeMax(s.cfg.Rules)

	u, err = s.db.UpdateUserTapEnergyBoost(ctx, uid, u.Tap.Energy.Boost, u.TapEnergyChargeMax(s.cfg.Rules), u.Points)
	if err != nil {
		return u, err
	}

	if err := s.rdb.SetLeaderboardPlayerPoints(ctx, uid, u.Level, u.Points); err != nil {
		return u, err
	}

	return u, nil
}
