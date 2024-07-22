package db

import (
	"context"
	"errors"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type leaderboardPlayerBSON struct {
	Profile domain.UserProfile `bson:"profile"`
	Level   domain.Level       `bson:"level"`
	Points  int                `bson:"points"`
}

func (db DB) ReadLeaderboardPlayer(ctx context.Context, uid int64) (domain.LeaderboardPlayer, error) {
	var (
		result domain.LeaderboardPlayer
		doc    leaderboardPlayerBSON
	)

	opts := &options.FindOneOptions{}
	opts.SetProjection(bson.D{
		{Key: "_id", Value: 0},
		{Key: "playedAt", Value: 0},
		{Key: "referralPoints", Value: 0},
	})

	r := db.users.FindOne(ctx, bson.M{
		"profile.telegram.id": uid,
		"profile.hasBan":      false,
		"profile.isGhost":     false,
		"profile.nickname":    bson.D{{Key: "$ne", Value: nil}},
	}, opts)
	if err := r.Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return result, domain.ErrNoUser
		}

		return result, err
	}

	result.Nickname = *doc.Profile.Nickname
	result.IsPremium = doc.Profile.Telegram.IsPremium
	result.Points = doc.Points
	result.Level = doc.Level

	return result, nil
}

func (db DB) ReadFriendsLeaderboardPlayers(ctx context.Context, uid int64, limit int64, skip int64) ([]domain.LeaderboardPlayer, error) {
	list := make([]domain.LeaderboardPlayer, 0, limit)

	opts := &options.FindOptions{}
	opts.SetProjection(bson.D{
		{Key: "_id", Value: 0},
		{Key: "playedAt", Value: 0},
		{Key: "referralPoints", Value: 0},
	})
	opts.SetLimit(limit)
	opts.SetSkip(skip)
	opts.SetSort(bson.M{"points": -1})

	c, err := db.users.Find(ctx, bson.M{
		"profile.ref":      uid,
		"profile.isGhost":  false,
		"profile.nickname": bson.D{{Key: "$ne", Value: nil}},
	}, opts)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = c.Close(ctx)
	}()

	for c.Next(ctx) {
		var doc leaderboardPlayerBSON

		err := c.Decode(&doc)
		if err != nil {
			return nil, err
		}

		list = append(list, domain.LeaderboardPlayer{
			Nickname:  *doc.Profile.Nickname,
			Level:     doc.Level,
			IsPremium: doc.Profile.Telegram.IsPremium,
			Points:    doc.Points,
		})
	}

	if err := c.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (db DB) ReadLevelLeaderboardPlayers(ctx context.Context, level domain.Level, excludeUID int64, limit int64, skip int64) ([]domain.LeaderboardPlayer, error) {
	list := make([]domain.LeaderboardPlayer, 0, limit)

	opts := &options.FindOptions{}
	opts.SetProjection(bson.D{
		{Key: "_id", Value: 0},
		{Key: "playedAt", Value: 0},
		{Key: "referralPoints", Value: 0},
	})
	opts.SetLimit(limit)
	opts.SetSkip(skip)
	opts.SetSort(bson.M{"points": -1})

	c, err := db.users.Find(ctx, bson.M{
		"profile.telegram.id": bson.D{{Key: "$ne", Value: excludeUID}},
		"profile.nickname":    bson.D{{Key: "$ne", Value: nil}},
		"profile.isGhost":     false,
		"level":               level,
	}, opts)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = c.Close(ctx)
	}()

	for c.Next(ctx) {
		var doc leaderboardPlayerBSON

		err := c.Decode(&doc)
		if err != nil {
			return nil, err
		}

		list = append(list, domain.LeaderboardPlayer{
			Nickname:  *doc.Profile.Nickname,
			Level:     doc.Level,
			IsPremium: doc.Profile.Telegram.IsPremium,
			Points:    doc.Points,
		})
	}

	if err := c.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (db DB) ReadGlobalLeaderboardPlayers(ctx context.Context, excludeUID int64, limit int64, skip int64) ([]domain.LeaderboardPlayer, error) {
	list := make([]domain.LeaderboardPlayer, 0, limit)

	opts := &options.FindOptions{}
	opts.SetProjection(bson.D{
		{Key: "_id", Value: 0},
		{Key: "playedAt", Value: 0},
		{Key: "referralPoints", Value: 0},
	})
	opts.SetLimit(limit)
	opts.SetSkip(skip)
	opts.SetSort(bson.M{"points": -1})

	c, err := db.users.Find(ctx, bson.M{
		"profile.telegram.id": bson.D{{Key: "$ne", Value: excludeUID}},
		"profile.nickname":    bson.D{{Key: "$ne", Value: nil}},
		"profile.isGhost":     false,
	}, opts)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = c.Close(ctx)
	}()

	for c.Next(ctx) {
		var doc leaderboardPlayerBSON

		err := c.Decode(&doc)
		if err != nil {
			return nil, err
		}

		list = append(list, domain.LeaderboardPlayer{
			Nickname:  *doc.Profile.Nickname,
			Level:     doc.Level,
			IsPremium: doc.Profile.Telegram.IsPremium,
			Points:    doc.Points,
		})
	}

	if err := c.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
