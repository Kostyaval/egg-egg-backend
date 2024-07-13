package domain

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"strconv"
	"time"
)

type JWTClaims struct {
	UID      int64
	Nickname *string
	JTI      uuid.UUID
}

func NewJWTClaims(uid int64, nickname *string) (*JWTClaims, error) {
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

func (c *JWTClaims) Encode(cfg *config.JWTConfig) ([]byte, error) {
	now := time.Now().UTC().Truncate(time.Second)
	exp := now.Add(cfg.TTL)

	token, err := jwt.NewBuilder().
		Issuer(cfg.Iss).
		Subject(strconv.FormatInt(c.UID, 10)).
		NotBefore(now).
		IssuedAt(now).
		Expiration(exp).
		JwtID(c.JTI.String()).
		Claim("nickname", c.Nickname).
		Build()

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJWTEncode, err)
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, cfg.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrJWTEncode, err)
	}

	return signed, nil
}

func (c *JWTClaims) Decode(cfg *config.JWTConfig, token []byte) error {
	now := time.Now().UTC().Truncate(time.Second)

	verifiedToken, err := jwt.Parse(token, jwt.WithKey(jwa.ES256, cfg.PublicKey))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJWTDecode, err)
	}

	// iss - issuer
	if verifiedToken.Issuer() != cfg.Iss {
		return fmt.Errorf("%w: %v", ErrJWTDecode, errors.New("invalid iss"))
	}

	// ndf - not before
	if !verifiedToken.NotBefore().Equal(now) && verifiedToken.NotBefore().After(now) {
		return fmt.Errorf("%w: %v", ErrJWTDecode, errors.New("invalid nbf"))
	}

	// exp - expiration
	if !verifiedToken.Expiration().Equal(now) && verifiedToken.Expiration().Before(now) {
		return fmt.Errorf("%w: %v", ErrJWTDecode, errors.New("invalid exp"))
	}

	// sub - subject
	if c.UID, err = strconv.ParseInt(verifiedToken.Subject(), 10, 64); err != nil {
		return fmt.Errorf("%w: %v", ErrJWTDecode, err)
	}

	// nickname
	nickname, ok := verifiedToken.PrivateClaims()["nickname"]
	if !ok {
		return fmt.Errorf("%w: %v", ErrJWTDecode, errors.New("no nickname"))
	}

	if nickname == nil {
		c.Nickname = nil
	} else {
		if str, ok := nickname.(string); ok {
			c.Nickname = &str
		} else {
			return fmt.Errorf("%w: %v", ErrJWTDecode, errors.New("invalid nickname"))
		}
	}

	// jti
	jti, ok := verifiedToken.Get("jti")
	if !ok {
		return fmt.Errorf("%w: %v", ErrJWTDecode, errors.New("no jti"))
	}

	jtiStr, ok := jti.(string)
	if !ok {
		return fmt.Errorf("%w: %v", ErrJWTDecode, errors.New("invalid jti"))
	}

	c.JTI, err = uuid.Parse(jtiStr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJWTDecode, err)
	}

	return nil
}
