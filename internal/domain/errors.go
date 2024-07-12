package domain

import "errors"

var (
	ErrInvalidUserType = errors.New("invalid user type")
	ErrBannedUser      = errors.New("banned")
	ErrGhostUser       = errors.New("ghost")
)
