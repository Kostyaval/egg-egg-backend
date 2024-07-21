package service

import "gitlab.com/egg-be/egg-backend/internal/config"

type DBInterface interface {
	meDB
	jwtDB
	tapDB
}

type Service struct {
	cfg *config.Config
	db  DBInterface
}

func NewService(cfg *config.Config, db DBInterface) *Service {
	return &Service{
		cfg: cfg,
		db:  db,
	}
}
