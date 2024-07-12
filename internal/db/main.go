package db

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	users *mongo.Collection
	cli   *mongo.Client
}

func NewMongoDB(cfg *config.Config) (*DB, error) {
	opts := options.Client().ApplyURI(cfg.MongoURI)

	cli, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	usersCol := cli.Database("game").Collection("users")

	// create unique user id index
	indexUserID := mongo.IndexModel{
		Keys:    bson.M{"profile.telegram.id": 1},
		Options: options.Index().SetUnique(true),
	}

	indexUserNickname := mongo.IndexModel{
		Keys: bson.M{"profile.nickname": 1},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.M{"profile.nickname": bson.M{"$type": "string"}}),
	}

	_, err = usersCol.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexUserID,
		indexUserNickname,
	})
	if err != nil {
		return nil, err
	}

	return &DB{
		users: usersCol,
		cli:   cli,
	}, nil
}

func (db DB) Disconnect() error {
	return db.cli.Disconnect(context.Background())
}
