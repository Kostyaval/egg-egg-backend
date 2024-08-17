package domain

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
	"testing"
	"time"
)

func TestUserDocument_calculateAutoClicker(t *testing.T) {
	t.Parallel()

	rules, err := NewRules()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	startPoints := 1000

	u := NewUserDocument(rules)
	u.Points = startPoints

	a := assert.New(t)

	// autoclicker is not available
	u.calculateAutoClicker(rules)
	a.Equal(u.Points, startPoints)

	// autoclicker is available but not enabled
	u.AutoClicker.IsAvailable = true
	u.calculateAutoClicker(rules)
	a.Equal(u.Points, startPoints)

	// --
	// autoclicker is available and enabled
	u.AutoClicker.IsEnabled = true

	// played 1 hour ago
	u.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-time.Hour))
	u.Points = startPoints
	u.calculateAutoClicker(rules)
	a.Equal(u.AutoClicker.Points, int(math.Floor(time.Hour.Seconds()/rules.AutoClicker.Speed.Seconds())))
	a.Equal(u.Points, startPoints+u.AutoClicker.Points)

	// played `rules.AutoClicker.TTL` ago
	u.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-rules.AutoClicker.TTL))
	u.Points = startPoints
	u.calculateAutoClicker(rules)
	a.Equal(u.AutoClicker.Points, int(math.Floor(rules.AutoClicker.TTL.Seconds()/rules.AutoClicker.Speed.Seconds())))
	a.Equal(u.Points, startPoints+u.AutoClicker.Points)

	// played `rules.AutoClicker.TTL - 1 hour` ago
	u.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-rules.AutoClicker.TTL).Add(-time.Hour))
	u.Points = startPoints
	u.calculateAutoClicker(rules)
	a.Equal(u.AutoClicker.Points, int(math.Floor(rules.AutoClicker.TTL.Seconds()/rules.AutoClicker.Speed.Seconds())))
	a.Equal(u.Points, startPoints+u.AutoClicker.Points)
}

func TestUserDocument_calculateDailyReward(t *testing.T) {
	t.Parallel()

	rules, err := NewRules()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().Truncate(time.Second).UTC()
	startPoints := 1000
	a := assert.New(t)

	u := NewUserDocument(rules)
	u.Points = startPoints

	// test initial data
	a.Equal(1, u.DailyReward.Day)
	a.Equal(u.DailyReward.ReceivedAt.Time(), u.Profile.CreatedAt.Time())
	a.True(u.DailyReward.Notify)

	u.DailyReward.Day = 2
	u.DailyReward.Notify = false
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(
		time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC))

	// received at start of day
	u.calculateDailyReward(rules)
	a.Equal(u.Points, startPoints)
	a.Equal(2, u.DailyReward.Day)
	a.False(u.DailyReward.Notify)

	// received at after start of day
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(u.DailyReward.ReceivedAt.Time().Add(time.Hour))
	u.calculateDailyReward(rules)
	a.Equal(u.Points, startPoints)
	a.Equal(2, u.DailyReward.Day)
	a.False(u.DailyReward.Notify)

	// received at before start of day (yesterday)
	u.Points = startPoints
	u.DailyReward.Day = 2
	u.DailyReward.Notify = false
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(
		time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC))
	u.calculateDailyReward(rules)
	a.Equal(u.Points, startPoints+rules.DailyRewards[2])
	a.Equal(3, u.DailyReward.Day)
	a.True(u.DailyReward.Notify)

	u.Points = startPoints
	u.DailyReward.Day = 2
	u.DailyReward.Notify = false
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(
		time.Date(now.Year(), now.Month(), now.Day()-1, 1, 0, 0, 0, time.UTC))
	u.calculateDailyReward(rules)
	a.Equal(u.Points, startPoints+rules.DailyRewards[2])
	a.Equal(3, u.DailyReward.Day)
	a.True(u.DailyReward.Notify)

	// user skip a day
	u.Points = startPoints
	u.DailyReward.Day = 2
	u.DailyReward.Notify = false
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(
		time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, time.UTC))
	u.calculateDailyReward(rules)
	a.Equal(u.Points, startPoints+rules.DailyRewards[0])
	a.Equal(1, u.DailyReward.Day)
	a.True(u.DailyReward.Notify)

	// test reward points
	for i, drPts := range rules.DailyRewards {
		u.Points = startPoints
		u.DailyReward.Day = i
		u.DailyReward.Notify = false
		u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(
			time.Date(now.Year(), now.Month(), now.Day()-1, 1, 0, 0, 0, time.UTC))

		u.calculateDailyReward(rules)
		a.Equal(u.Points, startPoints+drPts)
		a.Equal(u.DailyReward.Day, i+1)
		a.True(u.DailyReward.Notify)
	}

	// test repeat rewards after last day
	u.Points = startPoints
	u.DailyReward.Day = len(rules.DailyRewards)
	u.DailyReward.Notify = false
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(
		time.Date(now.Year(), now.Month(), now.Day()-1, 1, 0, 0, 0, time.UTC))

	u.calculateDailyReward(rules)
	a.Equal(u.Points, startPoints+rules.DailyRewards[0])
	a.Equal(1, u.DailyReward.Day)
	a.True(u.DailyReward.Notify)
}

func TestUserDocument_calculateTapEnergyCharge(t *testing.T) {
	t.Parallel()

	rules, err := NewRules()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	a := assert.New(t)
	u := NewUserDocument(rules)

	// test initial data
	a.Equal(u.Tap.Energy.Charge, rules.TapsBaseEnergyCharge)
	a.Equal(u.Tap.Energy.RechargeAvailable, rules.Taps[u.Level].Energy.RechargeAvailable)
	a.Equal(u.Tap.Energy.RechargedAt, u.Profile.CreatedAt)
	a.Len(u.Tap.Energy.Boost, len(rules.Taps))

	u.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-10 * rules.Taps[1].Energy.ChargeTimeSegment))
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-10 * rules.Taps[1].Energy.ChargeTimeSegment))
	u.Tap.Count = 100
	u.Tap.Energy.Charge = 0
	u.Tap.Energy.Boost[1] = 1

	u.calculateTapEnergyCharge(rules)
	a.Equal(10, u.Tap.Energy.Charge)

	u.Tap.Energy.Charge = 1000
	u.calculateTapEnergyCharge(rules)
	a.Equal(1000, u.Tap.Energy.Charge)

	u.Tap.Energy.Charge = 2000
	u.calculateTapEnergyCharge(rules)
	a.Equal(1000, u.Tap.Energy.Charge)

	u.Tap.Energy.Charge = 100
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-24 * time.Hour))
	u.calculateTapEnergyCharge(rules)
	a.Equal(1000, u.Tap.Energy.Charge)

	u.Tap.Energy.Charge = 100
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-1 * time.Second))
	u.calculateTapEnergyCharge(rules)
	a.Equal(100, u.Tap.Energy.Charge)

	u.Tap.Energy.Charge = 0
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-1 * time.Second))
	u.calculateTapEnergyCharge(rules)
	a.Equal(0, u.Tap.Energy.Charge)

	u.Tap.Energy.Charge = 100
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-10 * rules.Taps[1].Energy.ChargeTimeSegment))
	u.calculateTapEnergyCharge(rules)
	a.Equal(110, u.Tap.Energy.Charge)

	u.Tap.Energy.Charge = 999
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-10 * rules.Taps[1].Energy.ChargeTimeSegment))
	u.calculateTapEnergyCharge(rules)
	a.Equal(1000, u.Tap.Energy.Charge)
}

func TestUserDocument_tapEnergyChargeMax(t *testing.T) {
	t.Parallel()

	rules, err := NewRules()
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	u := NewUserDocument(rules)
	j := rules.TapsBaseEnergyCharge

	for k, v := range rules.Taps {
		u.Tap.Energy.Boost[k] = v.Energy.BoostChargeAvailable
		j += v.Energy.BoostCharge * v.Energy.BoostChargeAvailable
		a.Equal(u.TapEnergyChargeMax(rules), j)
	}
}

func TestUserDocument_calculateIsChannelMember(t *testing.T) {
	t.Parallel()

	rules, err := NewRules()
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	u := NewUserDocument(rules)

	// initial
	a.False(u.Profile.IsChannelMember)

	u.calculateIsChannelMember()
	a.False(u.Profile.IsChannelMember)

	u.Profile.Channel.ID = 1
	u.calculateIsChannelMember()
	a.True(u.Profile.IsChannelMember)
}

func TestUserDocument_CalculateNextLevelAvailability(t *testing.T) {
	t.Parallel()

	rules, err := NewRules()
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	u := NewUserDocument(rules)

	// initial
	a.False(u.IsNextLevelAvailable)
	u.calculateNextLevelAvailability(rules)
	a.False(u.IsNextLevelAvailable)

	u.Profile.Channel.ID = 1
	u.calculateIsChannelMember()
	a.True(u.Profile.IsChannelMember)

	for k, v := range rules.Taps {
		u.Points = v.NextLevel.Cost
		u.ReferralCount = v.NextLevel.Referrals
		u.Level = Level(k)

		u.calculateNextLevelAvailability(rules)

		if k != len(rules.Taps)-1 {
			a.True(u.IsNextLevelAvailable)
		} else {
			a.False(u.IsNextLevelAvailable)
		}
	}

	for k, v := range rules.Taps {
		u.Points = v.NextLevel.Cost + 1
		u.ReferralCount = v.NextLevel.Referrals + 1
		u.Level = Level(k)

		u.calculateNextLevelAvailability(rules)

		if k != len(rules.Taps)-1 {
			a.True(u.IsNextLevelAvailable)
		} else {
			a.False(u.IsNextLevelAvailable)
		}
	}

	for k, v := range rules.Taps {
		u.Points = v.NextLevel.Cost - 1
		u.ReferralCount = v.NextLevel.Referrals - 1
		u.Level = Level(k)

		u.calculateNextLevelAvailability(rules)
		a.False(u.IsNextLevelAvailable)
	}

	u.Profile.Channel.ID = 0
	u.calculateIsChannelMember()

	for k, v := range rules.Taps {
		u.Points = v.NextLevel.Cost
		u.ReferralCount = v.NextLevel.Referrals
		u.Level = Level(k)

		u.calculateNextLevelAvailability(rules)
		a.False(u.IsNextLevelAvailable)
	}
}
