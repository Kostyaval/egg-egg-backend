package config

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	RuntimeProduction  = "production"
	RuntimeDevelopment = "development"
)

type Config struct {
	Runtime       string
	MongoURI      string
	RedisURI      string
	TelegramToken string
	Rules         *Rules
	JWT           *JWTConfig
	CORS          *CORSConfig
	APIKey        string
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

	// Setup Redis URI
	cfg.RedisURI, ok = os.LookupEnv("REDIS_URI")
	if !ok || cfg.RedisURI == "" {
		return nil, errors.New("env REDIS_URI is not set")
	}

	// Setup Telegram Token
	cfg.TelegramToken, ok = os.LookupEnv("TELEGRAM_TOKEN")
	if !ok || cfg.TelegramToken == "" {
		return nil, errors.New("env TELEGRAM_TOKEN is not set")
	}

	// Setup game rules path
	rulesPath, ok := os.LookupEnv("RULES_PATH")
	if !ok || rulesPath == "" {
		return nil, errors.New("env RULES_PATH is not set")
	}

	rulesPath, err = filepath.Abs(rulesPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &cfg.Rules)
	if err != nil {
		return nil, err
	}

	// Setup JWT config
	cfg.JWT, err = newJWTConfig()
	if err != nil {
		return nil, err
	}

	// Setup CORS config
	cfg.CORS, err = newCORSConfig()
	if err != nil {
		return nil, err
	}

	// Setup API key
	cfg.APIKey = strings.TrimSpace(os.Getenv("API_KEY"))

	return cfg, nil
}
