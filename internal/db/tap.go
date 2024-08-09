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

func (db DB) UpdateUserTap(ctx context.Context, uid int64, tap domain.UserTap, points int) (domain.UserDocument, error) {
	var (
		doc domain.UserDocument
		now = primitive.NewDateTimeFromTime(time.Now().UTC().Truncate(time.Second))
		opt = options.FindOneAndUpdate().SetReturnDocument(options.After)
	)

	tap.PlayedAt = now

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"playedAt": now,
			"tap":      tap,
			"points":   points,
		}},
	}, opt).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return doc, domain.ErrNoUser
		}

		return doc, err
	}

	return doc, nil
}

func (db DB) UpdateUserTapBoost(ctx context.Context, uid int64, boost []int, points int) (domain.UserDocument, error) {
	var (
		doc domain.UserDocument
		now = primitive.NewDateTimeFromTime(time.Now().UTC())
		opt = options.FindOneAndUpdate().SetReturnDocument(options.After)
	)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$inc", Value: bson.M{
			"tap.points": 1,
		}},
		{Key: "$set", Value: bson.M{
			"playedAt":  now,
			"points":    points,
			"tap.boost": boost,
		}},
	}, opt).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return doc, domain.ErrNoUser
		}

		return doc, err
	}

	return doc, nil
}

func (db DB) UpdateUserTapEnergyBoost(ctx context.Context, uid int64, boost []int, charge int, points int) (domain.UserDocument, error) {
	var (
		doc domain.UserDocument
		now = primitive.NewDateTimeFromTime(time.Now().UTC())
		opt = options.FindOneAndUpdate().SetReturnDocument(options.After)
	)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"playedAt":          now,
			"tap.energy.boost":  boost,
			"tap.energy.charge": charge,
			"points":            points,
		}},
	}, opt).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return doc, domain.ErrNoUser
		}

		return doc, err
	}

	return doc, nil
}

func (db DB) UpdateUserTapEnergyRecharge(ctx context.Context, uid int64, available int, charge int, points int) (domain.UserDocument, error) {
	var (
		doc domain.UserDocument
		now = primitive.NewDateTimeFromTime(time.Now().UTC())
		opt = options.FindOneAndUpdate().SetReturnDocument(options.After)
	)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"playedAt":                     now,
			"tap.playedAt":                 now,
			"tap.energy.charge":            charge,
			"tap.energy.rechargeAvailable": available,
			"tap.energy.rechargedAt":       now,
			"points":                       points,
		}},
	}, opt).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return doc, domain.ErrNoUser
		}

		return doc, err
	}

	return doc, nil
}
