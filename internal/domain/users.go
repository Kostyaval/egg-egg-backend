package domain

import (
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserDocument struct {
	Profile        UserProfile        `bson:"profile" json:"profile"`
	PlayedAt       primitive.DateTime `bson:"playedAt" json:"playedAt"`
	Points         int                `bson:"points" json:"points"`
	ReferralPoints int                `bson:"referralPoints" json:"referralPoints"`
	Level          Level              `bson:"level" json:"level"`
	Taps           Taps               `bson:"taps" json:"taps"`
}

type Taps struct {
	Energy int `bson:"energy" json:"energy"`
	Count  int `bson:"count" json:"count"`
}

type Level struct {
	Level        int `bson:"level" json:"level"`
	Improvements int `bson:"improvements" json:"improvements"`
}

type UserProfile struct {
	Nickname  *string             `bson:"nickname" json:"nickname"`
	CreatedAt primitive.DateTime  `bson:"createdAt" json:"-"`
	UpdatedAt primitive.DateTime  `bson:"updatedAt" json:"-"`
	HasBan    bool                `bson:"hasBan" json:"-"`
	IsGhost   bool                `bson:"isGhost" json:"-"`
	Telegram  TelegramUserProfile `bson:"telegram" json:"telegram"`
	Referral  *int64              `bson:"ref" json:"referral"`
	JTI       *uuid.UUID          `bson:"jti" json:"-"`
}

type TelegramUserProfile struct {
	ID        int64  `bson:"id" json:"id"`
	IsPremium bool   `bson:"isPremium" json:"isPremium"`
	Firstname string `bson:"firstname" json:"-"`
	Lastname  string `bson:"lastname" json:"-"`
	Language  string `bson:"language" json:"language"`
	Username  string `bson:"username" json:"-"`
}
