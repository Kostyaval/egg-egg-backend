package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TelegramUserProfile struct {
	ID        int64  `bson:"id"`
	IsPremium bool   `bson:"isPremium"`
	Firstname string `bson:"firstname"`
	Lastname  string `bson:"lastname"`
	Language  string `bson:"language"`
	Username  string `bson:"username"`
}

type UserProfile struct {
	Nickname  *string             `bson:"nickname"`
	CreatedAt primitive.DateTime  `bson:"createdAt"`
	UpdatedAt primitive.DateTime  `bson:"updatedAt"`
	HasBan    bool                `bson:"hasBan"`
	IsGhost   bool                `bson:"isGhost"`
	Telegram  TelegramUserProfile `bson:"telegram"`
}
