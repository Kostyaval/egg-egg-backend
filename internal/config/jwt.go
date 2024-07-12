package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

type JWTConfig struct {
	Iss        string
	TTL        time.Duration
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

	// Set TTL for tokens
	if ttl, ok := os.LookupEnv("JWT_TTL"); !ok || ttl == "" {
		cfg.TTL, _ = time.ParseDuration("15m")
	} else {
		if cfg.TTL, err = time.ParseDuration(ttl); err != nil {
			return nil, err
		}
	}

	// Set private key
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

	// Set public key
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

	return cfg, nil
}
