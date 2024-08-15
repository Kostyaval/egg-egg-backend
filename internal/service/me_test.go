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
			Tap: domain.UserTap{
				Energy: domain.UserTapEnergy{
					RechargedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				},
			},
		}
	)

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertNotCalled(s.T(), "SetDailyReward")
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
			Tap: domain.UserTap{
				Energy: domain.UserTapEnergy{
					RechargedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				},
			},
		}
		dailyRewardPoints = doc.Points + s.cfg.Rules.DailyRewards[1]
		dailyReward       = &domain.DailyReward{Day: 2}
	)

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	s.dbMocks.On("SetDailyReward", ctx, uid, dailyRewardPoints, dailyReward).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, dailyRewardPoints).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)
	s.Equal(u.Points, dailyRewardPoints)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "SetDailyReward", ctx, uid, dailyRewardPoints, dailyReward)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, dailyRewardPoints)
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

func (s *Suite) TestGetMe_AutoClicker() {
	var (
		uid int64 = 1
		doc       = domain.UserDocument{
			Profile: domain.UserProfile{
				Telegram: domain.TelegramUserProfile{
					ID: uid,
				},
			},
			PlayedAt: primitive.NewDateTimeFromTime(time.Now().Truncate(time.Second).UTC().Add(-time.Hour)),
			Points:   1000,
			DailyReward: domain.DailyReward{
				ReceivedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				Day:        1,
			},
			AutoClicker: domain.AutoClicker{
				IsAvailable: true,
				IsEnabled:   true,
			},
			Tap: domain.UserTap{
				Energy: domain.UserTapEnergy{
					RechargedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				},
			},
		}
		autoClickerPoints = doc.Points + int(math.Floor(time.Hour.Seconds()/s.cfg.Rules.AutoClicker.Speed.Seconds()))
	)

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	s.dbMocks.On("SetPoints", ctx, uid, autoClickerPoints).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, autoClickerPoints).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)
	s.Equal(u.Points, autoClickerPoints)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertNotCalled(s.T(), "SetDailyReward")
	s.dbMocks.AssertCalled(s.T(), "SetPoints", ctx, uid, u.Points)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, u.Points)
}

func (s *Suite) TestGetMe_DailyRewardWithAutoClicker() {
	var (
		uid int64 = 1
		now       = time.Now().UTC()
		doc       = domain.UserDocument{
			Profile: domain.UserProfile{
				Telegram: domain.TelegramUserProfile{
					ID: uid,
				},
			},
			PlayedAt: primitive.NewDateTimeFromTime(time.Now().Truncate(time.Second).UTC().Add(-time.Hour)),
			Points:   1000,
			DailyReward: domain.DailyReward{
				ReceivedAt: primitive.NewDateTimeFromTime(time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, time.UTC)),
				Day:        1,
			},
			AutoClicker: domain.AutoClicker{
				IsAvailable: true,
				IsEnabled:   true,
			},
			Tap: domain.UserTap{
				Energy: domain.UserTapEnergy{
					RechargedAt: primitive.NewDateTimeFromTime(time.Now().UTC()),
				},
			},
		}
		dailyReward       = &domain.DailyReward{Day: 2}
		dailyRewardPoints = doc.Points + s.cfg.Rules.DailyRewards[1]
		autoClickerPoints = dailyRewardPoints + int(math.Floor(time.Hour.Seconds()/s.cfg.Rules.AutoClicker.Speed.Seconds()))
	)

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)
	s.dbMocks.On("SetDailyReward", ctx, uid, dailyRewardPoints, dailyReward).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, dailyRewardPoints).Return(nil)
	s.dbMocks.On("SetPoints", ctx, uid, autoClickerPoints).Return(nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, uid, doc.Level, autoClickerPoints).Return(nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)
	s.Equal(u.Points, autoClickerPoints)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "SetDailyReward", ctx, uid, dailyRewardPoints, dailyReward)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, dailyRewardPoints)
	s.dbMocks.AssertCalled(s.T(), "SetPoints", ctx, uid, autoClickerPoints)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, uid, doc.Level, autoClickerPoints)
}

func (s *Suite) TestGetMe_checkAutoClicker() {
	now := time.Now().UTC().Truncate(time.Second)
	startPoints := 1000

	u := &domain.UserDocument{
		Points: startPoints,
		AutoClicker: domain.AutoClicker{
			IsEnabled:   false,
			IsAvailable: false,
		},
	}

	// autoclicker is not available
	s.Equal(s.srv.checkAutoClicker(u), startPoints)

	// autoclicker is available but not enabled
	u.AutoClicker.IsAvailable = true
	s.Equal(s.srv.checkAutoClicker(u), startPoints)

	// autoclicker is available and enabled
	u.AutoClicker.IsEnabled = true

	// played 1 hour ago
	u.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-time.Hour))
	s.Equal(s.srv.checkAutoClicker(u), startPoints+int(math.Floor(time.Hour.Seconds()/s.cfg.Rules.AutoClicker.Speed.Seconds())))

	// played `cfg.Rules.AutoClicker.TTL` ago
	u.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-s.cfg.Rules.AutoClicker.TTL))
	s.Equal(s.srv.checkAutoClicker(u), startPoints+int(math.Floor(s.cfg.Rules.AutoClicker.TTL.Seconds()/s.cfg.Rules.AutoClicker.Speed.Seconds())))

	// played `cfg.Rules.AutoClicker.TTL - 1 hour` ago
	u.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-s.cfg.Rules.AutoClicker.TTL).Add(-time.Hour))
	s.Equal(s.srv.checkAutoClicker(u), startPoints+int(math.Floor(s.cfg.Rules.AutoClicker.TTL.Seconds()/s.cfg.Rules.AutoClicker.Speed.Seconds())))
}

func (s *Suite) TestGetMe_ResetTapEnergyRechargeAvailable() {
	var (
		uid int64 = 1
		now       = time.Now().UTC().Truncate(time.Second)
		doc       = domain.UserDocument{
			Profile: domain.UserProfile{
				Telegram: domain.TelegramUserProfile{
					ID: uid,
				},
			},
			PlayedAt: primitive.NewDateTimeFromTime(now.Add(-time.Hour)),
			Points:   1000,
			DailyReward: domain.DailyReward{
				ReceivedAt: primitive.NewDateTimeFromTime(now),
				Day:        1,
			},
			AutoClicker: domain.AutoClicker{
				IsAvailable: false,
				IsEnabled:   false,
			},
			Tap: domain.UserTap{
				Energy: domain.UserTapEnergy{
					Charge:            s.cfg.Rules.TapsBaseEnergyCharge,
					RechargeAvailable: 0,
					RechargedAt:       primitive.NewDateTimeFromTime(now.Add(-24 * time.Hour)),
				},
			},
		}
	)

	ctx := context.Background()

	s.dbMocks.On("GetUserDocumentWithID", ctx, uid).Return(doc, nil)

	energyRechargeDoc := doc
	energyRechargeDoc.Tap.Energy.RechargeAvailable = s.cfg.Rules.Taps[doc.Level].Energy.RechargeAvailable
	s.dbMocks.On(
		"UpdateUserTapEnergyRecharge",
		ctx,
		uid,
		s.cfg.Rules.Taps[doc.Level].Energy.RechargeAvailable,
		s.cfg.Rules.TapsBaseEnergyCharge,
		doc.Points).Return(energyRechargeDoc, nil)

	u, t, err := s.srv.GetMe(ctx, uid)
	s.Nil(err)

	claims, err := s.cfg.JWT.Decode(t)
	s.NoError(err)
	s.Equal(u.Profile.Telegram.ID, claims.UID)
	s.Equal(u.Profile.Nickname, claims.Nickname)
	s.Equal(u.Tap.Energy.RechargeAvailable, energyRechargeDoc.Tap.Energy.RechargeAvailable)
	s.Equal(u.Tap.Energy.RechargedAt, energyRechargeDoc.Tap.Energy.RechargedAt)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertNotCalled(s.T(), "SetDailyReward")
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
	s.dbMocks.AssertNotCalled(s.T(), "SetPoints")
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
	s.dbMocks.AssertCalled(s.T(), "UpdateUserTapEnergyRecharge", ctx,
		u.Profile.Telegram.ID,
		u.Tap.Energy.RechargeAvailable,
		u.Tap.Energy.Charge,
		u.Points)
}
