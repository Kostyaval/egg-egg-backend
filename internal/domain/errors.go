package domain

import "errors"

var (
	ErrInvalidUserType         = errors.New("invalid user type")
	ErrBannedUser              = errors.New("banned")
	ErrGhostUser               = errors.New("ghost")
	ErrNoUser                  = errors.New("user is not found")
	ErrConflictNickname        = errors.New("conflict nickname")
	ErrMultipleDevices         = errors.New("multiple devices")
	ErrJWTEncode               = errors.New("jwt encode")
	ErrJWTDecode               = errors.New("jwt decode")
	ErrNoJWT                   = errors.New("jwt is not found")
	ErrCorruptJWT              = errors.New("jwt corrupt")
	ErrTapOverLimit            = errors.New("tap over limit")
	ErrTapTooFast              = errors.New("tap too fast")
	ErrInsufficientEggs        = errors.New("insufficient eggs")
	ErrBoostOverLimit          = errors.New("boost over limit")
	ErrEnergyRechargeOverLimit = errors.New("energy recharge over limit")
	ErrEnergyRechargeTooFast   = errors.New("energy recharge too fast")
	ErrNoPoints                = errors.New("no points")
	ErrNoLevel                 = errors.New("no level")
	ErrHasAutoClicker          = errors.New("has auto clicker")
	ErrHasNoAutoClicker        = errors.New("has no auto clicker")
)
