package rdb

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"strconv"
)

func (r Redis) SetLeaderboardPlayerPoints(ctx context.Context, uid int64, level domain.Level, points int) error {
	var err error

	err = r.leaderboardClient.ZAdd(ctx, "global", redis.Z{Score: float64(points), Member: uid}).Err()
	if err != nil {
		return err
	}

	err = r.leaderboardClient.ZAdd(ctx, level.String(), redis.Z{Score: float64(points), Member: uid}).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r Redis) DeleteAllUsers(ctx context.Context) error {
	return r.leaderboardClient.FlushDB(ctx).Err()
}

func (r Redis) ReadLevelLeaderboardPlayerRank(ctx context.Context, uid int64, level domain.Level) (int64, error) {
	rank, err := r.leaderboardClient.ZRevRank(ctx, level.String(), strconv.FormatInt(uid, 10)).Result()
	if err != nil {
		return 0, err
	}

	return rank + int64(1), nil
}

func (r Redis) ReadGlobalLeaderboardPlayerRank(ctx context.Context, uid int64) (int64, error) {
	rank, err := r.leaderboardClient.ZRevRank(ctx, "global", strconv.FormatInt(uid, 10)).Result()
	if err != nil {
		return 0, err
	}

	return rank + int64(1), nil
}

func (r Redis) ReadLevelLeaderboardRanks(ctx context.Context, level domain.Level, limit int64, skip int64) ([]int64, error) {
	uids, err := r.leaderboardClient.ZRevRange(ctx, level.String(), skip, skip+limit).Result()
	if err != nil {
		return nil, err
	}

	result := make([]int64, len(uids))

	for i, uid := range uids {
		r, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			return nil, err
		}

		result[i] = r
	}

	return result, nil
}

func (r Redis) ReadGlobalLeaderboardRanks(ctx context.Context, limit int64, skip int64) ([]int64, error) {
	uids, err := r.leaderboardClient.ZRevRange(ctx, "global", skip, skip+limit).Result()
	if err != nil {
		return nil, err
	}

	result := make([]int64, len(uids))

	for i, uid := range uids {
		r, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			return nil, err
		}

		result[i] = r
	}

	return result, nil
}
