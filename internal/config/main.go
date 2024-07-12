package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	RuntimeProduction  = "production"
	RuntimeDevelopment = "development"
)

type Config struct {
	Runtime       string
	MongoURI      string
	TelegramToken string
	JWT           *JWTConfig
}

func NewConfig() (*Config, error) {
	var (
		ok  bool
		err error
	)

	time.Local = time.UTC

	if err := os.Setenv("TZ", "UTC"); err != nil {
		return nil, err
	}

	cfg := &Config{
		Runtime: RuntimeProduction,
	}

	// Setup runtime
	runtime, ok := os.LookupEnv("RUNTIME")
	if ok && runtime != "" {
		if runtime == RuntimeDevelopment {
			cfg.Runtime = RuntimeDevelopment
		} else if runtime != RuntimeProduction {
			return nil, fmt.Errorf("env RUNTIME=%s has unknown value", runtime)
		}
	}

	// Setup MongoDB URI
	cfg.MongoURI, ok = os.LookupEnv("MONGODB_URI")
	if !ok || cfg.MongoURI == "" {
		return nil, errors.New("env MONGODB_URI is not set")
	}

	// Setup Telegram Token
	cfg.TelegramToken, ok = os.LookupEnv("TELEGRAM_TOKEN")
	if !ok || cfg.TelegramToken == "" {
		return nil, errors.New("env TELEGRAM_TOKEN is not set")
	}

	// Setup JWT config
	cfg.JWT, err = newJWTConfig()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
