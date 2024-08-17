package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
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
			Tap: domain.UserTap{
				Energy: domain.UserTapEnergy{
					RechargedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				},
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
			Tap: domain.UserTap{
				Energy: domain.UserTapEnergy{
					RechargedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				},
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
		doc       = domain.NewUserDocument(s.cfg.Rules)
	)

	doc.Profile.Telegram.ID = uid
	doc.Points = 1000
	doc.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	doc.DailyReward.Day = 1
	doc.DailyReward.Notify = false
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	s.dbMocks.On("UpdateUserDocument", ctx, &doc).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, doc.Points).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "UpdateUserDocument", ctx, &doc)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, doc.Points)
}

func (s *Suite) TestGetMe_DailyReward() {
	var (
		now       = time.Now().Truncate(time.Second).UTC()
		uid int64 = 1
		doc       = domain.NewUserDocument(s.cfg.Rules)
	)

	doc.Profile.Telegram.ID = uid
	doc.Points = 1000
	doc.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(
		time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC))
	doc.DailyReward.Day = 1
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	doc.DailyReward.Day = 2
	doc.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(now)
	doc.Points += s.cfg.Rules.DailyRewards[1]
	s.dbMocks.On("UpdateUserDocument", ctx, &doc).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, doc.Points).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)
	s.Equal(u.Points, doc.Points)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "UpdateUserDocument", ctx, &doc)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, u.Points)
}

func (s *Suite) TestGetMe_AutoClicker() {
	var (
		uid int64 = 1
		doc       = domain.NewUserDocument(s.cfg.Rules)
	)

	doc.Profile.Telegram.ID = uid
	doc.PlayedAt = primitive.NewDateTimeFromTime(time.Now().Truncate(time.Second).UTC().Add(-time.Hour))
	doc.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(time.Now().UTC())
	doc.DailyReward.Day = 1
	doc.DailyReward.Notify = false
	doc.AutoClicker.IsAvailable = true
	doc.AutoClicker.IsEnabled = true
	doc.Points = 1000
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	doc.AutoClicker.Points = int(math.Floor(time.Hour.Seconds() / s.cfg.Rules.AutoClicker.Speed.Seconds())) // 2000
	doc.Points += 2000
	s.dbMocks.On("UpdateUserDocument", ctx, &doc).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, 3000).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)
	s.Equal(u.Points, 3000)
	s.Equal(u.AutoClicker.Points, 2000)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "UpdateUserDocument", ctx, &doc)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, u.Points)
}

func (s *Suite) TestGetMe_DailyRewardWithAutoClicker() {
	var (
		uid int64 = 1
		now       = time.Now().Truncate(time.Second).UTC()
		doc       = domain.NewUserDocument(s.cfg.Rules)
	)

	doc.Profile.Telegram.ID = uid
	doc.PlayedAt = primitive.NewDateTimeFromTime(time.Now().Truncate(time.Second).UTC().Add(-time.Hour))
	doc.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(
		time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC))
	doc.DailyReward.Day = 1
	doc.AutoClicker.IsAvailable = true
	doc.AutoClicker.IsEnabled = true
	doc.Points = 1000
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(time.Now().UTC())

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	doc.AutoClicker.Points = int(math.Floor(time.Hour.Seconds() / s.cfg.Rules.AutoClicker.Speed.Seconds()))
	doc.Points += 2000

	doc.DailyReward.Day = 2
	doc.DailyReward.ReceivedAt = primitive.NewDateTimeFromTime(now)
	doc.Points += s.cfg.Rules.DailyRewards[1]
	s.dbMocks.On("UpdateUserDocument", ctx, &doc).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, doc.Points).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)
	s.Equal(u.Points, doc.Points)
	s.True(u.DailyReward.Notify)
	s.Equal(u.AutoClicker.Points, 2000)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "UpdateUserDocument", ctx, &doc)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, u.Points)
}

func (s *Suite) TestGetMe_ResetTapEnergyRechargeAvailable() {
	var (
		uid int64 = 1
		now       = time.Now().UTC().Truncate(time.Second)
		doc       = domain.NewUserDocument(s.cfg.Rules)
	)

	doc.Profile.Telegram.ID = uid
	doc.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-time.Hour))
	doc.Points = 1000
	doc.Tap.Energy.RechargeAvailable = 0
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(now.Add(-24 * time.Hour))
	doc.DailyReward.Notify = false

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)

	doc.Tap.Energy.RechargeAvailable = s.cfg.Rules.Taps[doc.Level].Energy.RechargeAvailable
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(now)
	s.dbMocks.On("UpdateUserDocument", ctx, &doc).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, doc.Points).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)
	s.Equal(u.Tap.Energy.RechargeAvailable, doc.Tap.Energy.RechargeAvailable)
	s.Equal(u.Tap.Energy.RechargedAt, doc.Tap.Energy.RechargedAt)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "UpdateUserDocument", ctx, &doc)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, u.Points)
}
