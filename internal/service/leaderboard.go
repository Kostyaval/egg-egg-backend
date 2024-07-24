package service

import (
	"context"
	"errors"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type leaderboardDB interface {
	ReadLeaderboardPlayer(ctx context.Context, uid int64) (domain.LeaderboardPlayer, error)
	ReadFriendsLeaderboardPlayers(ctx context.Context, uid int64, limit int64, skip int64) ([]domain.LeaderboardPlayer, error)
	ReadFriendsLeaderboardTotalPlayers(ctx context.Context, uid int64) (int64, error)
	ReadLevelLeaderboardTotalPlayers(ctx context.Context, level domain.Level) (int64, error)
	ReadGlobalLeaderboardTotalPlayers(ctx context.Context) (int64, error)
	ReadLeaderboardPlayers(ctx context.Context, uids []int64) ([]domain.LeaderboardPlayer, error)
}

type leaderboardRedis interface {
	ReadLevelLeaderboardPlayerRank(ctx context.Context, uid int64, level domain.Level) (int64, error)
	ReadGlobalLeaderboardPlayerRank(ctx context.Context, uid int64) (int64, error)
	ReadLevelLeaderboardRanks(ctx context.Context, level domain.Level, limit int64, skip int64) ([]int64, error)
	ReadGlobalLeaderboardRanks(ctx context.Context, limit int64, skip int64) ([]int64, error)
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

		total, err := s.db.ReadFriendsLeaderboardTotalPlayers(ctx, uid)
		if err != nil {
			return me, nil, 0, err
		}

		var isMeInRange, isMeFound bool

		for i := int64(0); i < int64(len(list)); i++ {
			if i == 0 {
				isMeInRange = me.Points <= list[0].Points && me.Points >= list[int64(len(list)-1)].Points
			}

			if isMeInRange {
				if isMeFound {
					list[i].Rank = skip + i + 2
				} else {
					if me.Points >= list[i].Points {
						isMeFound = true
						me.Rank = skip + i + 1
						list[i].Rank = skip + i + 2
					} else {
						list[i].Rank = skip + i + 1
					}
				}
			} else {
				list[i].Rank = skip + i + 1
			}
		}

		return me, list, total, nil
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
		me.Rank, err = s.rdb.ReadGlobalLeaderboardPlayerRank(ctx, uid)
		if err != nil {
			return me, nil, 0, err
		}

		uids, err := s.rdb.ReadGlobalLeaderboardRanks(ctx, limit, skip)
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

		total, err := s.db.ReadGlobalLeaderboardTotalPlayers(ctx)
		if err != nil {
			return me, nil, 0, err
		}

		return me, list, total, nil
	}

	return me, nil, 0, errors.New("invalid tab")
}
