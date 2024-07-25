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
	opts.Registry = newMongoRegistry()

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

	// create unique referral id index
	indexReferralID := mongo.IndexModel{
		Keys:    bson.M{"profile.ref.id": 1},
		Options: options.Index().SetUnique(false),
	}

	// create unique user nickname index
	indexUserNickname := mongo.IndexModel{
		Keys: bson.M{"profile.nickname": 1},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.M{"profile.nickname": bson.M{"$type": "string"}}).
			SetCollation(&options.Collation{
				Locale:   "en",
				Strength: 2, // case insensitive
			}),
	}

	_, err = usersCol.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		indexUserID,
		indexReferralID,
		indexUserNickname,
	})
	if err != nil {
		return nil, err
	}

	// create index for auto-delete jwt token
	ctx := context.Background()

	cur, err := usersCol.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = cur.Close(ctx)
	}()

	var ixs []bson.M
	if err = cur.All(ctx, &ixs); err != nil {
		return nil, err
	}

	hasIndexJTI := false

	for _, ix := range ixs {
		if ix["name"] == "profile.jti_1" {
			hasIndexJTI = true
			break
		}
	}

	if hasIndexJTI {
		var result bson.M

		err := cli.Database("game").RunCommand(
			context.Background(),
			bson.D{
				bson.E{Key: "collMod", Value: "users"},
				bson.E{Key: "index", Value: bson.D{
					bson.E{Key: "name", Value: "profile.jti_1"},
					bson.E{Key: "expireAfterSeconds", Value: int32(cfg.JWT.TTL.Seconds())},
				}},
			}).Decode(&result)
		if err != nil {
			return nil, err
		}
	} else {
		indexJTI := mongo.IndexModel{
			Keys: bson.M{"profile.jti": 1},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{"profile.jti": bson.M{"$type": "string"}}).
				SetExpireAfterSeconds(int32(cfg.JWT.TTL.Seconds())),
		}

		_, err = usersCol.Indexes().CreateOne(context.Background(), indexJTI)
		if err != nil {
			return nil, err
		}
	}

	return &DB{
		users: usersCol,
		cli:   cli,
	}, nil
}

func (db DB) Disconnect() error {
	return db.cli.Disconnect(context.Background())
}
