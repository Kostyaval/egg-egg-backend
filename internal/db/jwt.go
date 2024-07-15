package db

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (db DB) UpdateUserJWT(ctx context.Context, uid int64, jti uuid.UUID) error {
	res, err := db.users.UpdateOne(ctx,
		bson.D{
			bson.E{Key: "profile.telegram.id", Value: uid},
			bson.E{Key: "profile.hasBan", Value: false},
			bson.E{Key: "profile.isGhost", Value: false},
		},
		bson.M{"$set": bson.M{
			"profile.jti":       jti,
			"activity.onlineAt": primitive.NewDateTimeFromTime(time.Now()),
		}},
	)

	if err != nil {
		return err
	}

	if res.MatchedCount != 1 || res.ModifiedCount != 1 {
		return domain.ErrNoUser
	}

	return nil
}

func (db DB) DeleteUserJWT(ctx context.Context, uid int64) error {
	_, err := db.users.UpdateOne(ctx,
		bson.D{
			bson.E{Key: "profile.telegram.id", Value: uid},
		},
		bson.M{"$set": bson.M{
			"profile.jti":        nil,
			"activity.offlineAt": primitive.NewDateTimeFromTime(time.Now()),
		}},
	)

	if err != nil {
		return err
	}

	return nil
}
