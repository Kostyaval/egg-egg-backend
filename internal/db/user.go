package db

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (db DB) UpdateUserJWT(ctx context.Context, uid int64, jti uuid.UUID) error {
	res, err := db.users.UpdateOne(ctx,
		bson.D{
			bson.E{Key: "profile.telegram.id", Value: uid},
			bson.E{Key: "profile.hasBan", Value: false},
			bson.E{Key: "profile.isGhost", Value: false},
		},
		bson.M{"$set": bson.M{"profile.jti": jti}},
	)

	if err != nil {
		return err
	}

	if res.MatchedCount != 1 || res.ModifiedCount != 1 {
		return domain.ErrNoUser
	}

	return nil
}
