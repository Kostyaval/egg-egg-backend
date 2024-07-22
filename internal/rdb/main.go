// Package rdb it is in-memory storage
package rdb

import (
	"github.com/redis/go-redis/v9"
	"gitlab.com/egg-be/egg-backend/internal/config"
)

type Redis struct {
	Client *redis.Client
}

func NewRedis(cfg *config.Config) (*Redis, error) {
	opt, err := redis.ParseURL(cfg.RedisURI)
	if err != nil {
		return nil, err
	}

	return &Redis{
		Client: redis.NewClient(opt),
	}, nil
}
