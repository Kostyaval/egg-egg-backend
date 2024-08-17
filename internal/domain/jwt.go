package domain

import (
	"github.com/google/uuid"
)

type JWTClaims struct {
	UID int64
	JTI uuid.UUID
}

func NewJWTClaims(uid int64) (*JWTClaims, error) {
	jti, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &JWTClaims{
		UID: uid,
		JTI: jti,
	}, nil
}
