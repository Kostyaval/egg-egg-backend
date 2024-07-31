package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

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

	if u.Taps.LevelTapBoosts == len(levelParams.EnergyBoosts) {
		return u, domain.ErrBoostOverLimit
	}

	if u.Points < levelParams.EnergyBoostCost {
		return u, domain.ErrInsufficientEggs
	}

	_ = s.db.UpdateUserTapBoostCount(ctx, uid, levelParams.EnergyBoostCost)

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

	if u.Taps.LevelTapBoosts == len(levelParams.EnergyBoosts) {
		return u, domain.ErrBoostOverLimit
	}

	if u.Points < levelParams.EnergyBoostCost {
		return u, domain.ErrInsufficientEggs
	}

	_ = s.db.UpdateUserEnergyBoostCount(ctx, uid, levelParams.EnergyBoostCost)

	u, err = s.db.GetUserDocumentWithID(ctx, uid)
	if err != nil {
		return u, err
	}

	return u, nil
}
