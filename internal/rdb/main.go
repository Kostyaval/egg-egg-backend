// Package rdb it is in-memory storage
package rdb

import (
	"github.com/redis/go-redis/v9"
	"gitlab.com/egg-be/egg-backend/internal/config"
)

type Redis struct {
	leaderboardClient *redis.Client
}

func NewRedis(cfg *config.Config) (*Redis, error) {
	opt, err := redis.ParseURL(cfg.RedisURI)
	if err != nil {
		return nil, err
	}

	opt.DB = 0
	leaderboardClient := redis.NewClient(opt)

	return &Redis{
		leaderboardClient: leaderboardClient,
	}, nil
}

func (r Redis) Close() error {
	return r.leaderboardClient.Close()
}
