package domain

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserDocument struct {
	Profile  UserProfile  `bson:"profile"`
	Activity UserActivity `bson:"activity"`
}

type UserProfile struct {
	Nickname  *string             `bson:"nickname"`
	CreatedAt primitive.DateTime  `bson:"createdAt"`
	UpdatedAt primitive.DateTime  `bson:"updatedAt"`
	HasBan    bool                `bson:"hasBan"`
	IsGhost   bool                `bson:"isGhost"`
	Telegram  TelegramUserProfile `bson:"telegram"`
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

type UserActivity struct {
	OnlineAt  primitive.DateTime `bson:"onlineAt"`
	OfflineAt primitive.DateTime `bson:"offlineAt"`
}
