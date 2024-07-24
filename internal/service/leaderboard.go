package service

import (
	"context"
	"errors"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type leaderboardDB interface {
	ReadLeaderboardPlayer(ctx context.Context, uid int64) (domain.LeaderboardPlayer, error)
	ReadFriendsLeaderboardPlayers(ctx context.Context, uid int64, limit int64, skip int64) ([]domain.LeaderboardPlayer, error)
	ReadLevelLeaderboardTotalPlayers(ctx context.Context, level domain.Level) (int64, error)
	ReadGlobalLeaderboardTotalPlayers(ctx context.Context) (int64, error)
	ReadLeaderboardPlayers(ctx context.Context, uids []int64) ([]domain.LeaderboardPlayer, error)
}

type leaderboardRedis interface {
	ReadLevelLeaderboardPlayerRank(ctx context.Context, uid int64, level domain.Level) (int64, error)
	ReadGlobalLeaderboardPlayerRank(ctx context.Context, uid int64) (int64, error)
	ReadLevelLeaderboardRanks(ctx context.Context, level domain.Level, limit int64, skip int64) ([]int64, error)
}

func (s Service) ReadLeaderboard(ctx context.Context, uid int64, tab string, limit int64, skip int64) (domain.LeaderboardPlayer, []domain.LeaderboardPlayer, int64, error) {
	var err error

	me, err := s.db.ReadLeaderboardPlayer(ctx, uid)
	if err != nil {
		return me, nil, 0, err
	}

	if tab == "friends" {
		list, err := s.db.ReadFriendsLeaderboardPlayers(ctx, uid, limit, skip)
		if err != nil {
			return me, nil, 0, err
		}

		return me, list, int64(len(list)), nil
	}

	if tab == "level" {
		me.Rank, err = s.rdb.ReadLevelLeaderboardPlayerRank(ctx, uid, me.Level)
		if err != nil {
			return me, nil, 0, err
		}

		uids, err := s.rdb.ReadLevelLeaderboardRanks(ctx, me.Level, limit, skip)
		if err != nil {
			return me, nil, 0, err
		}

		list, err := s.db.ReadLeaderboardPlayers(ctx, uids)
		if err != nil {
			return me, nil, 0, err
		}

		for i := int64(0); i < int64(len(list)); i++ {
			list[i].Rank = skip + i + 1
		}

		total, err := s.db.ReadLevelLeaderboardTotalPlayers(ctx, me.Level)
		if err != nil {
			return me, nil, 0, err
		}

		return me, list, total, nil
	}

	if tab == "global" {
		// TODO
		return me, nil, 0, errors.New("not implemented")
	}

	return me, nil, 0, errors.New("invalid tab")
}
