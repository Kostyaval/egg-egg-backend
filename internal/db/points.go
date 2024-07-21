package db

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
)

func (db DB) IncPointsWithReferral(ctx context.Context, uid int64, points int) error {
	res, err := db.users.UpdateOne(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$inc", Value: bson.M{"points": points, "referralPoints": points}},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return domain.ErrNoUser
	}

	return nil
}

func (db DB) IncPoints(ctx context.Context, uid int64, points int) error {
	res, err := db.users.UpdateOne(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$inc", Value: bson.M{"points": points}},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return domain.ErrNoUser
	}

	return nil
}
