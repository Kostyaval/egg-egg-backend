package domain

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserDocument struct {
	Profile   UserProfile        `bson:"profile"`
	OfflineAt primitive.DateTime `bson:"offlineAt"`
}

type UserProfile struct {
	Nickname  *string             `bson:"nickname"`
	CreatedAt primitive.DateTime  `bson:"createdAt"`
	UpdatedAt primitive.DateTime  `bson:"updatedAt"`
	HasBan    bool                `bson:"hasBan"`
	IsGhost   bool                `bson:"isGhost"`
	Telegram  TelegramUserProfile `bson:"telegram"`
	Reference *int64              `bson:"ref"`
	JTI       *uuid.UUID          `bson:"jti"`
}

type TelegramUserProfile struct {
	ID        int64  `bson:"id"`
	IsPremium bool   `bson:"isPremium"`
	Firstname string `bson:"firstname"`
	Lastname  string `bson:"lastname"`
	Language  string `bson:"language"`
	Username  string `bson:"username"`
}
