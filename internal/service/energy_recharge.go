package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"time"
)

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
		_ = s.db.ResetUserEnergyRechargeCount(ctx, uid)
		u.Taps.EnergyRechargeCount = 0
	}

	if int(time.Since(u.Taps.EnergyRechargedAt.Time().UTC()).Seconds()) < levelParams.EnergyFullRechargeDelaySeconds {
		return u, domain.ErrEnergyRechargeTooFast
	}

	if u.Taps.EnergyRechargeCount == levelParams.EnergyFullRechargeCount {
		return u, domain.ErrEnergyRechargeOverLimit
	}

	_ = s.db.UpdateUserEnergyRechargeCount(ctx, uid)
	_ = s.db.UpdateUserEnergyCount(ctx, uid, u.Taps.EnergyBoostCount*500+500)

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	return u, nil
}
