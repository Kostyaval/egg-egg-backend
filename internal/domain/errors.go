package domain

import "errors"

var (
	ErrInvalidUserType            = errors.New("invalid user type")
	ErrBannedUser                 = errors.New("banned")
	ErrGhostUser                  = errors.New("ghost")
	ErrNoUser                     = errors.New("user is not found")
	ErrConflictNickname           = errors.New("conflict nickname")
	ErrJWTEncode                  = errors.New("jwt encode")
	ErrJWTDecode                  = errors.New("jwt decode")
	ErrNoJWT                      = errors.New("jwt is not found")
	ErrCorruptJWT                 = errors.New("jwt corrupt")
	ErrTapOverLimit               = errors.New("tap over limit")
	ErrNoTapEnergy                = errors.New("no tap energy")
	ErrNoBoost                    = errors.New("no boost")
	ErrNoEnergyRecharge           = errors.New("no energy recharge")
	ErrNoPoints                   = errors.New("no points")
	ErrNoLevel                    = errors.New("no level")
	ErrHasAutoClicker             = errors.New("has auto clicker")
	ErrHasNoAutoClicker           = errors.New("has no auto clicker")
	ErrNotAllowedTelegramChat     = errors.New("not allowed telegram chat")
	ErrNotEnoughReferrals         = errors.New("not enough referrals")
	ErrNotFollowedTelegramChannel = errors.New("not followed telegram channel")
	ErrReachedLevelLimit          = errors.New("reached level limit")
)
