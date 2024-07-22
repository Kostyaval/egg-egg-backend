package service

import (
	"context"
	"errors"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type leaderboardDB interface {
	ReadLeaderboardPlayer(ctx context.Context, uid int64) (domain.LeaderboardPlayer, error)
	ReadFriendsLeaderboardPlayers(ctx context.Context, uid int64, limit int64, skip int64) ([]domain.LeaderboardPlayer, error)
	ReadLevelLeaderboardPlayers(ctx context.Context, level int, excludeUID int64, limit int64, skip int64) ([]domain.LeaderboardPlayer, error)
	ReadGlobalLeaderboardPlayers(ctx context.Context, excludeUID int64, limit int64, skip int64) ([]domain.LeaderboardPlayer, error)
}

func (s Service) ReadLeaderboard(ctx context.Context, uid int64, tab string, limit int64, skip int64) (domain.LeaderboardPlayer, []domain.LeaderboardPlayer, int64, error) {
	player, err := s.db.ReadLeaderboardPlayer(ctx, uid)
	if err != nil {
		return player, nil, 0, err
	}

	if tab == "friends" {
		friends, err := s.db.ReadFriendsLeaderboardPlayers(ctx, uid, limit, skip)
		if err != nil {
			return player, nil, 0, err
		}

		return player, friends, int64(len(friends)), nil
	}

	if tab == "level" {
		// TODO
		return player, nil, 0, errors.New("not implemented")
	}

	if tab == "global" {
		// TODO
		return player, nil, 0, errors.New("not implemented")
	}

	return player, nil, 0, errors.New("invalid tab")
}
