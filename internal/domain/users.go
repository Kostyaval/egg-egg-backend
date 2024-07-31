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
	DailyReward    DailyReward        `bson:"dailyReward" json:"dailyReward"`
	AutoClicker    AutoClicker        `bson:"autoClicker" json:"autoClicker"`
}

type Taps struct {
	TapCount            int                `bson:"tapCount" json:"tapCount"`
	TotalTapBoosts      int                `bson:"tapBoosts" json:"tapBoosts"`
	TotalEnergyBoosts   []int              `bson:"energyBoosts" json:"energyBoosts"`
	LevelTapBoosts      int                `bson:"levelTapBoosts" json:"levelTapBoosts"`
	EnergyCount         int                `bson:"energyCount" json:"energyCount"`
	PlayedAt            primitive.DateTime `bson:"playedAt" json:"playedAt"`
	EnergyRechargeCount int                `bson:"energyRechargeCount" json:"energyRechargeCount"`
	EnergyRechargedAt   primitive.DateTime `bson:"energyRechargedAt" json:"energyRechargedAt"`
}

type UserProfile struct {
	Nickname  *string              `bson:"nickname" json:"nickname"`
	CreatedAt primitive.DateTime   `bson:"createdAt" json:"-"`
	UpdatedAt primitive.DateTime   `bson:"updatedAt" json:"-"`
	HasBan    bool                 `bson:"hasBan" json:"-"`
	IsGhost   bool                 `bson:"isGhost" json:"-"`
	Telegram  TelegramUserProfile  `bson:"telegram" json:"telegram"`
	Referral  *ReferralUserProfile `bson:"ref" json:"referral"`
	JTI       *uuid.UUID           `bson:"jti" json:"-"`
}

type ReferralUserProfile struct {
	ID       int64  `bson:"id" json:"id"`
	Nickname string `bson:"nickname" json:"nickname"`
}

type TelegramUserProfile struct {
	ID        int64  `bson:"id" json:"id"`
	IsPremium bool   `bson:"isPremium" json:"isPremium"`
	Firstname string `bson:"firstname" json:"-"`
	Lastname  string `bson:"lastname" json:"-"`
	Language  string `bson:"language" json:"language"`
	Username  string `bson:"username" json:"-"`
}

type DailyReward struct {
	ReceivedAt primitive.DateTime `bson:"receivedAt" json:"receivedAt"`
	Day        int                `bson:"day" json:"day"`
}

type AutoClicker struct {
	IsEnabled   bool `bson:"isEnabled" json:"isEnabled"`
	IsAvailable bool `bson:"isAvailable" json:"isAvailable"`
}
