package service

import (
	"gitlab.com/egg-be/egg-backend/internal/config"
)

type DBInterface interface {
	meDB
	tapDB
	friendsDB
	leaderboardDB
}

type RedisInterface interface {
	meRedis
	leaderboardRedis
}

type Service struct {
	cfg *config.Config
	db  DBInterface
	rdb RedisInterface
}

func NewService(cfg *config.Config, db DBInterface, rdb RedisInterface) *Service {
	return &Service{
		cfg: cfg,
		db:  db,
		rdb: rdb,
	}
}
