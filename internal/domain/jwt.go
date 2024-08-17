package domain

import (
	"github.com/google/uuid"
)

type JWTClaims struct {
	UID      int64
	Nickname string
	JTI      uuid.UUID
}

func NewJWTClaims(uid int64, nickname string) (*JWTClaims, error) {
	jti, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &JWTClaims{
		UID:      uid,
		Nickname: nickname,
		JTI:      jti,
	}, nil
}
