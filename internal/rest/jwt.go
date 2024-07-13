package rest

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"strconv"
	"time"
)

type jwtClaims struct {
	UID int64
}

func newJWTClaims(uid int64) *jwtClaims {
	return &jwtClaims{
		UID: uid,
	}
}

func jwtEncodeClaims(cfg *config.JWTConfig, c *jwtClaims) ([]byte, uuid.UUID, error) {
	now := time.Now().UTC().Truncate(time.Second)
	exp := now.Add(cfg.TTL)

	jti, err := uuid.NewRandom()
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("%w: %v", domain.ErrJWTEncode, err)
	}

	token, err := jwt.NewBuilder().
		Issuer(cfg.Iss).
		Subject(strconv.FormatInt(c.UID, 10)).
		NotBefore(now).
		IssuedAt(now).
		Expiration(exp).
		JwtID(jti.String()).
		Build()

	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("%w: %v", domain.ErrJWTEncode, err)
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, cfg.PrivateKey))
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("%w: %v", domain.ErrJWTEncode, err)
	}

	return signed, jti, nil
}

func jwtDecodeClaims(cfg *config.JWTConfig, token []byte) (*jwtClaims, error) {
	c := &jwtClaims{}
	now := time.Now().UTC().Truncate(time.Second)

	verifiedToken, err := jwt.Parse(token, jwt.WithKey(jwa.ES256, cfg.PublicKey))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrJWTDecode, err)
	}

	// iss - issuer
	if verifiedToken.Issuer() != cfg.Iss {
		return nil, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("invalid iss"))
	}

	// ndf - not before
	if !verifiedToken.NotBefore().Equal(now) && verifiedToken.NotBefore().After(now) {
		return nil, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("invalid nbf"))
	}

	// exp - expiration
	if !verifiedToken.Expiration().Equal(now) && verifiedToken.Expiration().Before(now) {
		return nil, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("invalid exp"))
	}

	// sub - subject
	if c.UID, err = strconv.ParseInt(verifiedToken.Subject(), 10, 64); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrJWTDecode, err)
	}

	return c, nil
}
