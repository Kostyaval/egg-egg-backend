package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (s *Suite) TestGetMe_GhostUser() {
	var (
		uid int64 = 1
		doc       = domain.UserDocument{
			Profile: domain.UserProfile{
				Telegram: domain.TelegramUserProfile{
					ID: uid,
				},
				IsGhost: true,
			},
			Points: 1000,
			DailyReward: domain.DailyReward{
				ReceivedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				Day:        1,
			},
		}
	)

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)

	_, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(t)
	s.ErrorIs(err, domain.ErrGhostUser)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
}

func (s *Suite) TestGetMe_BannedUser() {
	var (
		uid int64 = 1
		doc       = domain.UserDocument{
			Profile: domain.UserProfile{
				Telegram: domain.TelegramUserProfile{
					ID: uid,
				},
				HasBan: true,
			},
			Points: 1000,
			DailyReward: domain.DailyReward{
				ReceivedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				Day:        1,
			},
		}
	)

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)

	_, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(t)
	s.ErrorIs(err, domain.ErrBannedUser)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
}

func (s *Suite) TestGetMe() {
	var (
		uid int64 = 1
		doc       = domain.UserDocument{
			Profile: domain.UserProfile{
				Telegram: domain.TelegramUserProfile{
					ID: uid,
				},
			},
			Points: 1000,
			DailyReward: domain.DailyReward{
				ReceivedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				Day:        1,
			},
		}
	)

	ctx := context.Background()
	jti := mock.MatchedBy(func(_ uuid.UUID) bool { return true })

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	s.dbMocks.On("UpdateUserJWT", ctx, uid, jti).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	if err != nil {
		s.Fail(err.Error())
	}

	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertNotCalled(s.T(), "SetDailyReward")
	s.dbMocks.AssertCalled(s.T(), "UpdateUserJWT", ctx, doc.Profile.Telegram.ID, jti)
}

func (s *Suite) TestGetMe_DailyReward() {
	var (
		now       = time.Now().UTC()
		uid int64 = 1
		doc       = domain.UserDocument{
			Profile: domain.UserProfile{
				Telegram: domain.TelegramUserProfile{
					ID: uid,
				},
			},
			Points: 1000,
			DailyReward: domain.DailyReward{
				ReceivedAt: primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)),
				Day:        1,
			},
		}
		dailyRewardPoints = doc.Points + s.cfg.Rules.DailyRewards[1]
		dailyReward       = &domain.DailyReward{Day: 2}
	)

	ctx := context.Background()
	jti := mock.MatchedBy(func(_ uuid.UUID) bool { return true })

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	s.dbMocks.On("SetDailyReward", ctx, uid, dailyRewardPoints, dailyReward).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, dailyRewardPoints).Return(nil)
	s.dbMocks.On("UpdateUserJWT", ctx, uid, jti).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	if err != nil {
		s.Fail(err.Error())
	}

	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "SetDailyReward", ctx, uid, dailyRewardPoints, dailyReward)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, dailyRewardPoints)
	s.dbMocks.AssertCalled(s.T(), "UpdateUserJWT", ctx, doc.Profile.Telegram.ID, jti)
}

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
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(u.DailyReward.ReceivedAt.Time().Add(time.Hour))
	dr, pts = s.srv.checkDailyReward(u)
	s.Nil(dr)
	s.Equal(u.Points, pts)

	// received at before start of day (yesterday)
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC))
	dr, _ = s.srv.checkDailyReward(u)
	s.NotNil(dr)
	s.Equal(dr.Day, u.DailyReward.Day+1)

	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-1, 1, 0, 0, 0, time.UTC))
	dr, _ = s.srv.checkDailyReward(u)
	s.NotNil(dr)
	s.Equal(dr.Day, u.DailyReward.Day+1)

	// user skip a day
	u.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-2, 0, 0, 0, 0, time.UTC))
	dr, _ = s.srv.checkDailyReward(u)
	s.NotNil(dr)
	s.Equal(dr.Day, 1)

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
