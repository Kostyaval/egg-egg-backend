package config

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

type JWTConfig struct {
	Iss        string
	PrivateKey jwk.Key
	PublicKey  jwk.Key
}

func newJWTConfig() (*JWTConfig, error) {
	var (
		err error
		ok  bool
	)

	cfg := &JWTConfig{}

	// Set iss
	iss, ok := os.LookupEnv("JWT_ISS")
	if !ok || iss == "" {
		cfg.Iss = "egg.one"
	} else {
		cfg.Iss = strings.TrimSpace(iss)
	}

	// Set private key
	privateKey, ok := os.LookupEnv("JWT_PRIVATE_KEY")
	if !ok || privateKey == "" {
		// deprecated
		privateKeyPath, ok := os.LookupEnv("JWT_PRIVATE_KEY_PATH")
		if !ok || privateKeyPath == "" {
			return nil, errors.New("env JWT_PRIVATE_KEY_PATH is not set")
		}

		privateKeyPath, err = filepath.Abs(privateKeyPath)
		if err != nil {
			return nil, err
		}

		privateKeyRaw, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, err
		}

		if len(privateKeyRaw) == 0 {
			return nil, errors.New(privateKeyPath + " is empty")
		}

		cfg.PrivateKey, err = jwk.ParseKey(privateKeyRaw)
		if err != nil {
			return nil, err
		}
	} else {
		cfg.PrivateKey, err = jwk.ParseKey([]byte(privateKey))
		if err != nil {
			return nil, err
		}
	}

	// Set public key
	publicKey, ok := os.LookupEnv("JWT_PUBLIC_KEY")
	if !ok || publicKey == "" {
		// deprecated
		publicKeyPath, ok := os.LookupEnv("JWT_PUBLIC_KEY_PATH")
		if !ok || publicKeyPath == "" {
			return nil, errors.New("env JWT_PUBLIC_KEY_PATH is not set")
		}

		publicKeyPath, err = filepath.Abs(publicKeyPath)
		if err != nil {
			return nil, err
		}

		publicKeyRaw, err := os.ReadFile(publicKeyPath)
		if err != nil {
			return nil, err
		}

		if len(publicKeyRaw) == 0 {
			return nil, errors.New(publicKeyPath + " is empty")
		}

		cfg.PublicKey, err = jwk.ParseKey(publicKeyRaw)
		if err != nil {
			return nil, err
		}
	} else {
		cfg.PublicKey, err = jwk.ParseKey([]byte(publicKey))
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func (cfg JWTConfig) Encode(c *domain.JWTClaims) ([]byte, error) {
	now := time.Now().UTC().Truncate(time.Second)
	exp := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)

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
		return nil, fmt.Errorf("%w: %v", domain.ErrJWTEncode, err)
	}

	signed, err := jwt.Sign(token, jwt.WithKey(jwa.ES256, cfg.PrivateKey))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrJWTEncode, err)
	}

	return signed, nil
}

func (cfg JWTConfig) Decode(token []byte) (domain.JWTClaims, error) {
	var (
		c   domain.JWTClaims
		now = time.Now().UTC().Truncate(time.Second)
	)

	verifiedToken, err := jwt.Parse(token, jwt.WithKey(jwa.ES256, cfg.PublicKey))
	if err != nil {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, err)
	}

	// iss - issuer
	if verifiedToken.Issuer() != cfg.Iss {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("invalid iss"))
	}

	// ndf - not before
	if !verifiedToken.NotBefore().Equal(now) && verifiedToken.NotBefore().After(now) {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("invalid nbf"))
	}

	// exp - expiration
	if !verifiedToken.Expiration().Equal(now) && verifiedToken.Expiration().Before(now) {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("invalid exp"))
	}

	// sub - subject
	if c.UID, err = strconv.ParseInt(verifiedToken.Subject(), 10, 64); err != nil {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, err)
	}

	// nickname
	nickname, ok := verifiedToken.PrivateClaims()["nickname"]
	if !ok {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("no nickname"))
	}

	if nickname == nil {
		c.Nickname = ""
	} else {
		if str, ok := nickname.(string); ok {
			c.Nickname = str
		} else {
			return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("invalid nickname"))
		}
	}

	// jti
	jti, ok := verifiedToken.Get("jti")
	if !ok {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("no jti"))
	}

	jtiStr, ok := jti.(string)
	if !ok {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, errors.New("invalid jti"))
	}

	c.JTI, err = uuid.Parse(jtiStr)
	if err != nil {
		return c, fmt.Errorf("%w: %v", domain.ErrJWTDecode, err)
	}

	return c, nil
}
