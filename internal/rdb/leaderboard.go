package rdb

import (
	"context"
	"github.com/redis/go-redis/v9"
	"gitlab.com/egg-be/egg-backend/internal/domain"
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
