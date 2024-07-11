package db

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/config"
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

	return &DB{
		users: cli.Database("game").Collection("users"),
		cli:   cli,
	}, nil
}

func (db DB) Disconnect() error {
	return db.cli.Disconnect(context.Background())
}
