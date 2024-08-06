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
	UpdateUserTapEnergyRecharge(ctx context.Context, uid int64, available int, at time.Time, chargeMax int, points int) (domain.UserDocument, error)
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

	energyAvailable, _ := s.userTapEnergy(&u)
	if energyAvailable == 0 || energyAvailable < u.Tap.Points {
		return u, domain.ErrNoTapEnergy
	}

	energySpent := tapCount * u.Tap.Points
	if energySpent > energyAvailable {
		tapCount = energyAvailable / u.Tap.Points
		energySpent = tapCount * u.Tap.Points
	}

	u.Tap.Count += tapCount
	u.Tap.Energy.Charge = energyAvailable - energySpent
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

// userTapEnergy returns available energy charge and max energy.
func (s Service) userTapEnergy(u *domain.UserDocument) (int, int) {
	maxCharge := s.cfg.Rules.TapsBaseEnergyCharge

	for i, v := range u.Tap.Energy.Boost {
		if i < len(s.cfg.Rules.Taps) {
			maxCharge += s.cfg.Rules.Taps[i].Energy.BoostCharge * v
		}
	}

	// when used recharge
	if u.Tap.Energy.Charge >= maxCharge {
		return maxCharge, maxCharge
	}

	// when left 1 day from last tap request
	now := time.Now().UTC()
	ago := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)

	if u.Tap.PlayedAt.Time().UTC().Before(ago) {
		return maxCharge, maxCharge
	}

	delta := now.Sub(u.Tap.PlayedAt.Time().UTC()).Milliseconds()
	if delta < s.cfg.Rules.Taps[u.Level].Energy.ChargeTimeSegment.Milliseconds() {
		return u.Tap.Energy.Charge, maxCharge
	}

	charge := int(delta / s.cfg.Rules.Taps[u.Level].Energy.ChargeTimeSegment.Milliseconds())
	if charge > maxCharge {
		return maxCharge, maxCharge
	}

	charge += u.Tap.Energy.Charge
	if charge > maxCharge {
		charge = maxCharge
	}

	return charge, maxCharge
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

	if u.Tap.Energy.RechargeAvailable == 0 {
		return u, domain.ErrNoEnergyRecharge
	}

	if s.cfg.Rules.Taps[u.Level].Energy.RechargeAvailableAfter.Seconds() > 0 &&
		now.Sub(u.Tap.Energy.RechargedAt.Time().UTC()).Seconds() < s.cfg.Rules.Taps[u.Level].Energy.RechargeAvailableAfter.Seconds() {
		return u, domain.ErrNoEnergyRecharge
	}

	pts := s.checkAutoClicker(&u)
	_, chargeMax := s.userTapEnergy(&u)

	if u.Points != pts {
		if err := s.rdb.SetLeaderboardPlayerPoints(ctx, uid, u.Level, pts); err != nil {
			return u, err
		}
	}

	return s.db.UpdateUserTapEnergyRecharge(ctx, uid, u.Tap.Energy.RechargeAvailable-1, now, chargeMax, pts)
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

	if u.Tap.Boost[u.Level] >= s.cfg.Rules.Taps[u.Level].BoostAvailable {
		return u, domain.ErrNoBoost
	}

	pts := s.checkAutoClicker(&u)
	if pts < s.cfg.Rules.Taps[u.Level].BoostCost {
		return u, domain.ErrNoPoints
	}

	pts -= s.cfg.Rules.Taps[u.Level].BoostCost
	u.Tap.Boost[u.Level]++

	u, err = s.db.UpdateUserTapBoost(ctx, uid, u.Tap.Boost, pts)
	if err != nil {
		return u, err
	}

	if err := s.rdb.SetLeaderboardPlayerPoints(ctx, uid, u.Level, pts); err != nil {
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

	if u.Tap.Energy.Boost[u.Level] >= s.cfg.Rules.Taps[u.Level].Energy.BoostChargeAvailable {
		return u, domain.ErrNoBoost
	}

	pts := s.checkAutoClicker(&u)
	if pts < s.cfg.Rules.Taps[u.Level].Energy.BoostChargeCost {
		return u, domain.ErrNoPoints
	}

	pts -= s.cfg.Rules.Taps[u.Level].Energy.BoostChargeCost
	u.Tap.Energy.Boost[u.Level]++
	_, chargeMax := s.userTapEnergy(&u)

	u, err = s.db.UpdateUserTapEnergyBoost(ctx, uid, u.Tap.Energy.Boost, chargeMax, pts)
	if err != nil {
		return u, err
	}

	if err := s.rdb.SetLeaderboardPlayerPoints(ctx, uid, u.Level, pts); err != nil {
		return u, err
	}

	return u, nil
}
