package domain

import "errors"

var (
	ErrInvalidUserType = errors.New("invalid user type")
	ErrBannedUser      = errors.New("banned")
	ErrGhostUser       = errors.New("ghost")
	ErrNoUser          = errors.New("user is not found")
	ErrMultipleDevices = errors.New("multiple devices")
	ErrJWTEncode       = errors.New("jwt encode")
	ErrJWTDecode       = errors.New("jwt decode")
	ErrNoJWT           = errors.New("jwt is not found")
	ErrCorruptJWT      = errors.New("jwt corrupt")
)
