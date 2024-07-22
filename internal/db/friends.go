package db

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db DB) ReadUserFriends(ctx context.Context, uid int64, limit int64, skip int64) ([]domain.Friend, int64, error) {
	list := make([]domain.Friend, 0, limit)

	opts := &options.FindOptions{}
	opts.SetProjection(bson.D{
		{Key: "_id", Value: 0},
		{Key: "profile", Value: 1},
		{Key: "level", Value: 1},
	})
	opts.SetLimit(limit)
	opts.SetSkip(skip)
	opts.SetSort(bson.M{"profile.createdAt": -1})

	c, err := db.users.Find(ctx, bson.M{
		"profile.ref":      uid,
		"profile.nickname": bson.D{{Key: "$ne", Value: nil}},
	}, opts)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		_ = c.Close(ctx)
	}()

	for c.Next(ctx) {
		var u struct {
			Profile domain.UserProfile `bson:"profile"`
			Level   domain.Level       `bson:"level"`
		}

		err := c.Decode(&u)
		if err != nil {
			return nil, 0, err
		}

		list = append(list, domain.Friend{
			Nickname:  u.Profile.Nickname,
			Level:     u.Level,
			IsPremium: u.Profile.Telegram.IsPremium,
			Points:    0,
		})
	}

	if err := c.Err(); err != nil {
		return nil, 0, err
	}

	count, err := db.users.CountDocuments(ctx, bson.M{
		"profile.ref":      uid,
		"profile.nickname": bson.D{{Key: "$ne", Value: nil}},
	})
	if err != nil {
		return nil, 0, err
	}

	return list, count, nil
}
