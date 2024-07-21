package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type friendsDB interface {
	ReadUserFriends(ctx context.Context, uid int64, limit int64, skip int64) ([]domain.Friend, int64, error)
}

func (s Service) ReadUserFriends(ctx context.Context, uid int64, limit int64, skip int64) ([]domain.Friend, int64, error) {
	list, count, err := s.db.ReadUserFriends(ctx, uid, limit, skip)
	if err != nil {
		return nil, 0, err
	}

	for i := 0; i < len(list); i++ {
		if list[i].Level < len(s.cfg.Rules.Referral) {
			for k := 0; k <= list[i].Level; k++ {
				if list[i].IsPremium {
					list[i].Points += s.cfg.Rules.Referral[k].Sender.Premium
				} else {
					list[i].Points += s.cfg.Rules.Referral[k].Sender.Plain
				}
			}
		}
	}

	return list, count, nil
}
