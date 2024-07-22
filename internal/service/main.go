package service

import "gitlab.com/egg-be/egg-backend/internal/config"

type DBInterface interface {
	meDB
	jwtDB
	friendsDB
}

type RedisInterface interface{}

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
