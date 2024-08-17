package db

import (
	"context"
	"errors"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (db DB) IncPointsWithReferral(ctx context.Context, uid int64, points int, incNewUser bool) (int, error) {
	var result struct {
		Points int `bson:"points"`
	}

	incRefCount := 0
	if incNewUser {
		incRefCount = 1
	}

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$inc", Value: bson.M{"points": points, "referralPoints": points, "referralCount": incRefCount}},
	}, opt).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, domain.ErrNoUser
		}

		return 0, err
	}

	return result.Points, nil
}

func (db DB) IncPoints(ctx context.Context, uid int64, points int) (int, error) {
	var result struct {
		Points int `bson:"points"`
	}

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$inc", Value: bson.M{"points": points}},
	}, opt).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, domain.ErrNoUser
		}

		return 0, err
	}

	return result.Points, nil
}

func (db DB) SetPoints(ctx context.Context, uid int64, points int) error {
	res, err := db.users.UpdateOne(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{"points": points}},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount != 1 {
		return domain.ErrNoUser
	}

	return nil
}

func (db DB) SetDailyReward(ctx context.Context, uid int64, points int, reward *domain.DailyReward) error {
	reward.ReceivedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	_, err := db.users.UpdateOne(ctx,
		bson.D{
			bson.E{Key: "profile.telegram.id", Value: uid},
			{Key: "profile.hasBan", Value: false},
			{Key: "profile.isGhost", Value: false},
		},
		bson.M{"$set": bson.M{
			"dailyReward": reward,
			"points":      points,
			"playedAt":    primitive.NewDateTimeFromTime(time.Now().UTC()),
		}},
	)

	if err != nil {
		return err
	}

	return nil
}
