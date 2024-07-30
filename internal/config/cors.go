package config

import (
	"os"
	"strconv"
	"strings"
)

type CORSConfig struct {
	Origins   string
	MaxAge    int
	IsEnabled bool
}

func newCORSConfig() (*CORSConfig, error) {
	var (
		err error
		cfg = &CORSConfig{
			Origins: strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS")),
			MaxAge:  0,
		}
	)

	maxAge := strings.TrimSpace(os.Getenv("CORS_MAX_AGE"))
	if maxAge != "" {
		if cfg.MaxAge, err = strconv.Atoi(maxAge); err != nil {
			return nil, err
		}
	}

	cfg.IsEnabled = cfg.Origins != ""

	return cfg, nil
}
