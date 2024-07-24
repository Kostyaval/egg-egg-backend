package db

import (
	"context"
	"errors"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db DB) IncPointsWithReferral(ctx context.Context, uid int64, points int) (int, error) {
	var result struct {
		Points int `bson:"points"`
	}

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$inc", Value: bson.M{"points": points, "referralPoints": points}},
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
