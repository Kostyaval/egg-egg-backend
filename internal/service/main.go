package service

import "gitlab.com/egg-be/egg-backend/internal/config"

type DBInterface interface{}

type TgInterface interface{}

type Service struct {
	cfg *config.Config
	db  DBInterface
	tg  TgInterface
}

func NewService(cfg *config.Config, db DBInterface, tg TgInterface) *Service {
	return &Service{
		cfg: cfg,
		db:  db,
		tg:  tg,
	}
}
