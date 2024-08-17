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

func (db DB) GetUserDocumentWithID(ctx context.Context, uid int64) (domain.UserDocument, error) {
	var result domain.UserDocument

	opts := &options.FindOneOptions{}
	opts.SetProjection(bson.D{{Key: "_id", Value: 0}})

	r := db.users.FindOne(ctx, bson.M{"profile.telegram.id": uid}, opts)
	if err := r.Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return result, domain.ErrNoUser
		}

		return result, err
	}

	return result, nil
}

func (db DB) GetUserProfileWithID(ctx context.Context, uid int64) (domain.UserProfile, error) {
	var result struct {
		Profile domain.UserProfile `bson:"profile"`
	}

	opts := &options.FindOneOptions{}
	opts.SetProjection(bson.D{
		bson.E{Key: "_id", Value: 0},
		bson.E{Key: "profile", Value: 1},
	})

	r := db.users.FindOne(ctx, bson.M{"profile.telegram.id": uid}, opts)
	if err := r.Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return result.Profile, domain.ErrNoUser
		}

		return result.Profile, err
	}

	return result.Profile, nil
}

func (db DB) CreateUser(ctx context.Context, user *domain.UserDocument) error {
	_, err := db.users.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (db DB) CheckUserNickname(ctx context.Context, nickname string) (bool, error) {
	rx := primitive.Regex{Pattern: fmt.Sprintf("^%s$", nickname), Options: "i"}

	count, err := db.users.CountDocuments(ctx, bson.D{{Key: "profile.nickname", Value: rx}})
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (db DB) UpdateUserNickname(ctx context.Context, uid int64, nickname string, jti uuid.UUID) error {
	res, err := db.users.UpdateOne(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"profile.jti":       jti,
			"profile.nickname":  nickname,
			"profile.updatedAt": primitive.NewDateTimeFromTime(time.Now().UTC()),
			"playedAt":          primitive.NewDateTimeFromTime(time.Now().UTC()),
			"dailyReward": domain.DailyReward{
				ReceivedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				Day:        0,
			},
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

func (db DB) UpdateReferralUserProfile(ctx context.Context, uid int64, ref *domain.ReferralUserProfile) error {
	res, err := db.users.UpdateOne(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"profile.updatedAt": primitive.NewDateTimeFromTime(time.Now().UTC()),
			"profile.ref":       ref,
		}},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return domain.ErrNoUser
	}

	return nil
}

func (db DB) CreateUserAutoClicker(ctx context.Context, uid int64, cost int) (domain.UserDocument, error) {
	var doc domain.UserDocument

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$inc", Value: bson.M{"points": -cost}},
		{Key: "$set", Value: bson.M{
			"playedAt":                primitive.NewDateTimeFromTime(time.Now().UTC()),
			"autoClicker.isAvailable": true,
			"autoClicker.isEnabled":   true,
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

func (db DB) UpdateUserAutoClicker(ctx context.Context, uid int64, isEnabled bool) (domain.UserDocument, error) {
	var doc domain.UserDocument

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"playedAt":              primitive.NewDateTimeFromTime(time.Now().UTC()),
			"autoClicker.isEnabled": isEnabled,
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

func (db DB) SetUserIsTelegramChannelMember(ctx context.Context, uid int64, channelID int64) error {
	var doc domain.UserDocument

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$addToSet", Value: bson.M{
			"tasks.telegram": channelID,
		}},
	}, opt).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.ErrNoUser
		}

		return err
	}

	return nil
}

func (db DB) SetUserIsTelegramChannelLeft(ctx context.Context, uid int64, channelID int64) error {
	var doc domain.UserDocument

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$pull", Value: bson.M{
			"tasks.telegram": channelID,
		}},
	}, opt).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return domain.ErrNoUser
		}

		return err
	}

	return nil
}

func (db DB) ReadTotalUserReferrals(ctx context.Context, uid int64) (int64, error) {
	return db.users.CountDocuments(ctx, bson.M{
		"profile.referral.id": uid,
	})
}

func (db DB) UpdateUserLevel(ctx context.Context, uid int64, level int, cost int) (domain.UserDocument, error) {
	var doc domain.UserDocument

	opt := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := db.users.FindOneAndUpdate(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"playedAt": primitive.NewDateTimeFromTime(time.Now().UTC()),
			"level":    level,
		}},
		{Key: "$inc", Value: bson.M{"points": -cost}},
	}, opt).Decode(&doc)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return doc, domain.ErrNoUser
		}

		return doc, err
	}

	return doc, nil
}

func (db DB) DeleteAllUsers(ctx context.Context) error {
	_, err := db.users.DeleteMany(ctx, bson.D{})
	return err
}

func (db DB) CreateUsers(ctx context.Context, users []domain.UserDocument) error {
	if len(users) == 0 {
		return nil
	}

	docs := make([]interface{}, len(users))
	for i, v := range users {
		docs[i] = v
	}

	_, err := db.users.InsertMany(ctx, docs)

	return err
}

func (db DB) UpdateUserQuests(ctx context.Context, uid int64, quests domain.UserQuests) error {
	res, err := db.users.UpdateOne(ctx, bson.D{
		{Key: "profile.telegram.id", Value: uid},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"quests": quests,
		}},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return domain.ErrNoUser
	}

	return nil
}

func (db DB) UpdateUserDocument(ctx context.Context, u *domain.UserDocument) error {
	primitive.NewDateTimeFromTime(time.Now().UTC())

	res, err := db.users.UpdateOne(ctx, bson.D{
		{Key: "profile.telegram.id", Value: u.Profile.Telegram.ID},
		{Key: "profile.hasBan", Value: false},
		{Key: "profile.isGhost", Value: false},
	}, bson.D{
		{Key: "$set", Value: bson.M{
			"points":      u.Points,
			"profile":     u.Profile,
			"tap":         u.Tap,
			"dailyReward": u.DailyReward,
			"playedAt":    primitive.NewDateTimeFromTime(time.Now().UTC()),
		}},
	})

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return domain.ErrNoUser
	}

	return nil
}
