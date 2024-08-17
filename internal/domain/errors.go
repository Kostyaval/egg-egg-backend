package domain

import "errors"

var (
	ErrBannedUser            = errors.New("banned")
	ErrGhostUser             = errors.New("ghost")
	ErrNoUser                = errors.New("user is not found")
	ErrConflictNickname      = errors.New("conflict nickname")
	ErrJWTEncode             = errors.New("jwt encode")
	ErrJWTDecode             = errors.New("jwt decode")
	ErrTapOverLimit          = errors.New("tap over limit")
	ErrNoTapEnergy           = errors.New("no tap energy")
	ErrNoBoost               = errors.New("no boost")
	ErrNoEnergyRecharge      = errors.New("no energy recharge")
	ErrNoPoints              = errors.New("no points")
	ErrNoLevel               = errors.New("no level")
	ErrNextLevelNotAvailable = errors.New("next level not available")
	ErrHasAutoClicker        = errors.New("has auto clicker")
	ErrHasNoAutoClicker      = errors.New("has no auto clicker")
	ErrReplay                = errors.New("replay")
	ErrInvalidQuest          = errors.New("invalid quest")
)
