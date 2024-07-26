package service

import (
	"context"
	"github.com/gofiber/fiber/v2/log"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

func (s Service) BoostTap(ctx context.Context, uid int64) (domain.UserDocument, error) {
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

	levelParams := s.cfg.Rules.Taps[u.Level]

	log.Info("tap boosts")
	log.Info(u.Taps.LevelTapBoosts)
	log.Info(u.Level)
	log.Info(len(levelParams.EnergyBoosts))

	if u.Taps.LevelTapBoosts == len(levelParams.EnergyBoosts) {
		return u, domain.ErrBoostOverLimit
	}
	if u.Taps.TapCount < levelParams.EnergyBoostCost {
		return u, domain.ErrInsufficientEggs
	}

	err = s.db.UpdateUserTapBoostCount(ctx, uid, levelParams.EnergyBoostCost)

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (s Service) BoostEnergy(ctx context.Context, uid int64) (domain.UserDocument, error) {
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

	levelParams := s.cfg.Rules.Taps[u.Level]

	if u.Taps.LevelTapBoosts == len(levelParams.EnergyBoosts) {
		return u, domain.ErrBoostOverLimit
	}
	if u.Taps.TapCount < levelParams.EnergyBoostCost {
		return u, domain.ErrInsufficientEggs
	}

	err = s.db.UpdateUserEnergyBoostCount(ctx, uid, levelParams.EnergyBoostCost)

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	return u, nil
}
