package service

import (
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (s *Suite) TestGetMe_checkDailyReward() {
	now := time.Now().UTC()
	startPoints := 1000

	u := &domain.UserDocument{
		Points: startPoints,
		DailyReward: domain.DailyReward{
			ReceivedAt: primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)),
			Day:        1,
		},
	}

	// received at start of day
	dr, pts := s.srv.checkDailyReward(u)
	s.Nil(dr)
	s.Equal(u.Points, pts)

	// received at after start of day
	u.DailyReward.ReceivedAt.Time().Add(time.Hour)
	dr, pts = s.srv.checkDailyReward(u)
	s.Nil(dr)
	s.Equal(u.Points, pts)

	// received at before start of day (yesterday)
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC))
	dr, pts = s.srv.checkDailyReward(u)
	s.NotNil(dr)
	s.Equal(dr.Day, u.DailyReward.Day+1)

	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-1, 1, 0, 0, 0, time.UTC))
	dr, pts = s.srv.checkDailyReward(u)
	s.NotNil(dr)
	s.Equal(dr.Day, u.DailyReward.Day+1)

	// user skip a day
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, time.UTC))
	dr, pts = s.srv.checkDailyReward(u)
	s.NotNil(dr)
	s.Equal(dr.Day, 0)

	// test reward points
	for i, drp := range s.cfg.Rules.DailyRewards {
		u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-1, 1, 0, 0, 0, time.UTC))
		u.DailyReward.Day = i

		dr, pts = s.srv.checkDailyReward(u)
		s.NotNil(dr)
		s.Equal(dr.Day, u.DailyReward.Day+1)
		s.Equal(pts, startPoints+drp)
	}

	// test repeat rewards after last day
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-1, 1, 0, 0, 0, time.UTC))
	u.DailyReward.Day = len(s.cfg.Rules.DailyRewards)

	dr, pts = s.srv.checkDailyReward(u)
	s.NotNil(dr)
	s.Equal(dr.Day, 1)
	s.Equal(pts, startPoints+s.cfg.Rules.DailyRewards[0])
}
