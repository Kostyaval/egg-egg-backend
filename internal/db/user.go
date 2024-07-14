package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (db DB) GetUserWithID(ctx context.Context, uid int64) (*domain.UserProfile, error) {
	var result struct {
		Profile *domain.UserProfile `bson:"profile"`
	}

	opts := &options.FindOneOptions{}
	opts.SetProjection(bson.D{
		bson.E{Key: "_id", Value: 0},
		bson.E{Key: "profile", Value: 1},
	})

	r := db.users.FindOne(ctx, bson.M{"profile.telegram.id": uid}, opts)
	if err := r.Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return result.Profile, nil
}

func (db DB) RegisterUser(ctx context.Context, user *domain.UserProfile) error {
	_, err := db.users.InsertOne(ctx, bson.D{bson.E{Key: "profile", Value: user}})
	if err != nil {
		return err
	}

	return nil
}

func (db DB) CheckUserNickname(ctx context.Context, nickname string) (bool, error) {
	rx := primitive.Regex{Pattern: fmt.Sprintf("^%s$", nickname), Options: "i"}
	count, err := db.users.CountDocuments(ctx, bson.D{{"profile.nickname", rx}})
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (db DB) UpdateUserNickname(ctx context.Context, uid int64, nickname string, jti uuid.UUID) error {
	res, err := db.users.UpdateOne(ctx, bson.D{
		{"profile.telegram.id", uid},
		{"profile.hasBan", false},
		{"profile.isGhost", false},
	}, bson.D{
		{"$set", bson.M{
			"profile.jti":       jti,
			"profile.nickname":  nickname,
			"profile.updatedAt": primitive.NewDateTimeFromTime(time.Now()),
		}},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return domain.ErrNoUser
	}

	if res.ModifiedCount != 1 {
		return domain.ErrConflictNickname
	}

	return nil
}
