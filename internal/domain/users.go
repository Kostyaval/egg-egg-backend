package domain

import (
	"crypto/rand"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
	"math/big"
	"time"
)

type UserDocument struct {
	Profile        UserProfile        `bson:"profile" json:"profile"`
	PlayedAt       primitive.DateTime `bson:"playedAt" json:"playedAt"`
	Points         int                `bson:"points" json:"points"`
	ReferralPoints int                `bson:"referralPoints" json:"referralPoints"`
	Level          Level              `bson:"level" json:"level"`
	Tap            UserTap            `bson:"tap" json:"tap"`
	DailyReward    DailyReward        `bson:"dailyReward" json:"dailyReward"`
	AutoClicker    AutoClicker        `bson:"autoClicker" json:"autoClicker"`
	Tasks          UserTasks          `bson:"tasks" json:"tasks"`
	Quests         UserQuests         `bson:"quests" json:"quests"`
}

func NewUserDocument(rules *Rules) UserDocument {
	now := time.Now().UTC()

	return UserDocument{
		PlayedAt: primitive.NewDateTimeFromTime(now),
		Level:    Lv0,
		Profile: UserProfile{
			CreatedAt: primitive.NewDateTimeFromTime(now),
			UpdatedAt: primitive.NewDateTimeFromTime(now),
		},
		Tap: UserTap{
			Points: 1,
			Boost:  make([]int, len(rules.Taps)),
			Energy: UserTapEnergy{
				Charge:            rules.TapsBaseEnergyCharge,
				Boost:             make([]int, len(rules.Taps)),
				RechargeAvailable: rules.Taps[Lv0].Energy.RechargeAvailable,
				RechargedAt:       primitive.NewDateTimeFromTime(now),
			},
			PlayedAt: primitive.NewDateTimeFromTime(now),
		},
		// at registration user gets daily reward
		Points: rules.DailyRewards[0],
		DailyReward: DailyReward{
			Day:        1,
			ReceivedAt: primitive.NewDateTimeFromTime(now),
			Notify:     true, // at registration user gets daily reward
		},
		Tasks: UserTasks{
			Telegram: make([]int, 0),
		},
		Quests: UserQuests{
			Telegram: 0,
			Youtube:  0,
			X:        0,
		},
	}
}

func (u *UserDocument) TapEnergyChargeMax(rules *Rules) int {
	maxCharge := rules.TapsBaseEnergyCharge

	for i, v := range u.Tap.Energy.Boost {
		if i < len(rules.Taps) {
			maxCharge += rules.Taps[i].Energy.BoostCharge * v
		}
	}

	return maxCharge
}

func (u *UserDocument) Calculate(rules *Rules) {
	u.calculateTapEnergyCharge(rules)
	u.calculateAutoClicker(rules)
	u.calculateQuests(rules)
	u.calculateDailyReward(rules)
	u.calculateTapEnergyRecharge(rules)
	u.calculateIsChannelMember()
}

func (u *UserDocument) calculateTapEnergyCharge(rules *Rules) {
	maxCharge := u.TapEnergyChargeMax(rules)

	// when used recharge
	if u.Tap.Energy.Charge >= maxCharge {
		u.Tap.Energy.Charge = maxCharge
		return
	}

	// when left 1 day from last tap request
	now := time.Now().UTC()
	ago := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)

	if u.Tap.PlayedAt.Time().UTC().Before(ago) {
		u.Tap.Energy.Charge = maxCharge
		return
	}

	delta := now.Sub(u.Tap.PlayedAt.Time().UTC()).Milliseconds()
	if delta < rules.Taps[u.Level].Energy.ChargeTimeSegment.Milliseconds() {
		return
	}

	charge := int(delta / rules.Taps[u.Level].Energy.ChargeTimeSegment.Milliseconds())
	if charge > maxCharge {
		u.Tap.Energy.Charge = maxCharge
		return
	}

	u.Tap.Energy.Charge += charge
	if u.Tap.Energy.Charge > maxCharge {
		u.Tap.Energy.Charge = maxCharge
	}
}

func (u *UserDocument) calculateAutoClicker(rules *Rules) {
	if !u.AutoClicker.IsAvailable || !u.AutoClicker.IsEnabled {
		return
	}

	delta := time.Now().Truncate(time.Second).UTC().Sub(u.PlayedAt.Time()).Seconds()
	if delta <= 0 {
		return
	}

	if delta >= rules.AutoClicker.TTL.Seconds() {
		u.AutoClicker.Points = int(math.Floor(rules.AutoClicker.TTL.Seconds() / rules.AutoClicker.Speed.Seconds()))
	} else {
		u.AutoClicker.Points = int(math.Floor(delta / rules.AutoClicker.Speed.Seconds()))
	}

	u.Points += u.AutoClicker.Points
}

func (u *UserDocument) calculateQuests(rules *Rules) {
	var (
		now            = time.Now().UTC()
		solvedRandTime = func(startedAt time.Time) bool {
			if startedAt.After(now.Add(-2 * time.Hour)) {
				return false
			}

			if startedAt.Add(24 * time.Hour).Before(now) {
				return true
			}

			d := 24 * time.Hour
			n, err := rand.Int(rand.Reader, big.NewInt(d.Nanoseconds()))
			if err != nil {
				return false
			}

			return startedAt.Add(time.Duration(n.Int64())).Before(now)
		}
	)

	if u.Quests.Telegram == -1 && solvedRandTime(u.Quests.TelegramStartedAt.Time()) {
		u.Points += rules.Quests.Telegram
		u.Quests.Telegram = 1
	}

	if u.Quests.Youtube == -1 && solvedRandTime(u.Quests.YoutubeStartedAt.Time()) {
		u.Points += rules.Quests.Youtube
		u.Quests.Youtube = 1
	}

	if u.Quests.X == -1 && solvedRandTime(u.Quests.XStartedAt.Time()) {
		u.Points += rules.Quests.X
		u.Quests.X = 1
	}
}

func (u *UserDocument) calculateDailyReward(rules *Rules) {
	now := time.Now().Truncate(time.Second).UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfYesterday := startOfToday.AddDate(0, 0, -1)

	if u.DailyReward.ReceivedAt.Time().After(startOfToday) || u.DailyReward.ReceivedAt.Time().Equal(startOfToday) {
		u.DailyReward.Notify = false

		return
	}

	if u.DailyReward.ReceivedAt.Time().After(startOfYesterday) || u.DailyReward.ReceivedAt.Time().Equal(startOfYesterday) {
		if u.DailyReward.Day >= len(rules.DailyRewards) {
			u.DailyReward.Day = 1
		} else {
			u.DailyReward.Day++
		}
	} else {
		u.DailyReward.Day = 1
	}

	u.Points += rules.DailyRewards[u.DailyReward.Day-1]
	u.DailyReward.Notify = true
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(now)
}

func (u *UserDocument) calculateTapEnergyRecharge(rules *Rules) {
	if u.Tap.Energy.RechargedAt.Time().UTC().Day() != time.Now().UTC().Day() {
		u.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(time.Now().Truncate(time.Second).UTC())
		u.Tap.Energy.RechargeAvailable = rules.Taps[u.Level].Energy.RechargeAvailable
	}
}

func (u *UserDocument) calculateIsChannelMember() {
	u.Profile.IsChannelMember = u.Profile.Channel.ID != 0
}

type UserTap struct {
	Count    int                `bson:"count" json:"-"`
	Points   int                `bson:"points" json:"points"`
	Boost    []int              `bson:"boost" json:"boost"`
	Energy   UserTapEnergy      `bson:"energy" json:"energy"`
	PlayedAt primitive.DateTime `bson:"playedAt" json:"playedAt"`
}

type UserTapEnergy struct {
	Charge            int                `bson:"charge" json:"charge"`
	Boost             []int              `bson:"boost" json:"boost"`
	RechargeAvailable int                `bson:"rechargeAvailable" json:"rechargeAvailable"`
	RechargedAt       primitive.DateTime `bson:"rechargedAt" json:"rechargedAt"`
}

type UserProfile struct {
	Nickname        *string              `bson:"nickname" json:"nickname"`
	CreatedAt       primitive.DateTime   `bson:"createdAt" json:"-"`
	UpdatedAt       primitive.DateTime   `bson:"updatedAt" json:"-"`
	HasBan          bool                 `bson:"hasBan" json:"-"`
	IsGhost         bool                 `bson:"isGhost" json:"-"`
	Telegram        TelegramUserProfile  `bson:"telegram" json:"telegram"`
	Referral        *ReferralUserProfile `bson:"ref" json:"referral"`
	Channel         ChannelUserProfile   `bson:"channel" json:"-"`
	IsChannelMember bool                 `json:"isChannelMember"`
}

type ReferralUserProfile struct {
	ID       int64  `bson:"id" json:"id"`
	Nickname string `bson:"nickname" json:"nickname"`
}

type ChannelUserProfile struct {
	ID        int64              `bson:"id" json:"-"`
	UpdatedAt primitive.DateTime `bson:"updatedAt" json:"-"`
	CheckedAt primitive.DateTime `bson:"checkedAt" json:"-"`
}

type TelegramUserProfile struct {
	ID              int64  `bson:"id" json:"id"`
	IsPremium       bool   `bson:"isPremium" json:"isPremium"`
	FirstName       string `bson:"firstname" json:"-"`
	LastName        string `bson:"lastname" json:"-"`
	Language        string `bson:"language" json:"language"`
	Username        string `bson:"username" json:"username"`
	AllowsWriteToPm bool   `bson:"allowsWriteToPm" json:"-"`
}

type DailyReward struct {
	ReceivedAt primitive.DateTime `bson:"receivedAt" json:"receivedAt"`
	Day        int                `bson:"day" json:"day"`
	Notify     bool               `json:"notify"`
}

type AutoClicker struct {
	IsEnabled   bool `bson:"isEnabled" json:"isEnabled"`
	IsAvailable bool `bson:"isAvailable" json:"isAvailable"`
	Points      int  `json:"points"`
}

type UserTasks struct {
	Telegram []int `bson:"telegram" json:"telegram"`
}

type UserQuests struct {
	Telegram          int8               `bson:"telegram" json:"telegram"`
	TelegramStartedAt primitive.DateTime `bson:"telegramStartedAt" json:"-"`
	Youtube           int8               `bson:"youtube" json:"youtube"`
	YoutubeStartedAt  primitive.DateTime `bson:"youtubeStartedAt" json:"-"`
	X                 int8               `bson:"x" json:"x"`
	XStartedAt        primitive.DateTime `bson:"xStartedAt" json:"-"`
}
