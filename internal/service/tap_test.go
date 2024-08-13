package service

import (
	"context"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (s *Suite) TestUserTapEnergy() {
	now := time.Now().UTC().Truncate(time.Second)
	boost := make([]int, len(s.cfg.Rules.Taps))
	boost[1] = 1

	u := &domain.UserDocument{
		PlayedAt: primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[1].Energy.ChargeTimeSegment)),
		Tap: domain.UserTap{
			Count:    100,
			PlayedAt: primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[1].Energy.ChargeTimeSegment)),
			Energy:   domain.UserTapEnergy{Boost: boost},
		},
	}

	a, m := s.srv.userTapEnergy(u)
	s.Equal(a, 10)
	s.Equal(m, 1000)

	u.Tap.Energy.Charge = 1000
	a, m = s.srv.userTapEnergy(u)
	s.Equal(a, 1000)
	s.Equal(m, 1000)

	u.Tap.Energy.Charge = 2000
	a, m = s.srv.userTapEnergy(u)
	s.Equal(a, 1000)
	s.Equal(m, 1000)

	u.Tap.Energy.Charge = 100
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-24 * time.Hour))
	a, m = s.srv.userTapEnergy(u)
	s.Equal(a, 1000)
	s.Equal(m, 1000)

	u.Tap.Energy.Charge = 100
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-1 * time.Second))
	a, m = s.srv.userTapEnergy(u)
	s.Equal(a, 100)
	s.Equal(m, 1000)

	u.Tap.Energy.Charge = 0
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-1 * time.Second))
	a, m = s.srv.userTapEnergy(u)
	s.Equal(a, 0)
	s.Equal(m, 1000)

	u.Tap.Energy.Charge = 100
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[1].Energy.ChargeTimeSegment))
	a, m = s.srv.userTapEnergy(u)
	s.Equal(a, 110)
	s.Equal(m, 1000)
	s.Equal(u.Tap.Energy.Charge, 100) // Charge should not be reset

	u.Tap.Energy.Charge = 999
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[1].Energy.ChargeTimeSegment))
	a, m = s.srv.userTapEnergy(u)
	s.Equal(a, 1000)
	s.Equal(m, 1000)
}

func (s *Suite) TestAddTap_GhostUser() {
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			IsGhost: true,
			HasBan:  true,
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
	}
	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	_, err := s.srv.AddTap(ctx, doc.Profile.Telegram.ID, 1)
	s.ErrorIs(err, domain.ErrGhostUser)

	s.dbMocks.AssertNotCalled(s.T(), "UpdateUserTap")
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
}

func (s *Suite) TestAddTap_BannedUser() {
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			IsGhost: false,
			HasBan:  true,
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
	}
	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	_, err := s.srv.AddTap(ctx, doc.Profile.Telegram.ID, 1)
	s.ErrorIs(err, domain.ErrBannedUser)

	s.dbMocks.AssertNotCalled(s.T(), "UpdateUserTap")
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
}

func (s *Suite) TestAddTap_NoEnergy() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Tap: domain.UserTap{
			PlayedAt: primitive.NewDateTimeFromTime(now.Add(-1 * time.Second)),
			Energy:   domain.UserTapEnergy{Charge: 0},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	_, err := s.srv.AddTap(ctx, doc.Profile.Telegram.ID, 1)
	s.ErrorIs(err, domain.ErrNoTapEnergy)

	s.dbMocks.AssertNotCalled(s.T(), "UpdateUserTap")
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
}

func (s *Suite) TestAddTap_NoEnergy_TapPoints() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Tap: domain.UserTap{
			Points:   10,
			PlayedAt: primitive.NewDateTimeFromTime(now.Add(-10 * time.Second)),
			Energy:   domain.UserTapEnergy{Charge: 2},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	_, err := s.srv.AddTap(ctx, doc.Profile.Telegram.ID, 1)
	s.ErrorIs(err, domain.ErrNoTapEnergy)

	s.dbMocks.AssertNotCalled(s.T(), "UpdateUserTap")
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
}

func (s *Suite) TestAddTap() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Level:  0,
		Points: 100,
		Tap: domain.UserTap{
			Count:    200,
			Points:   2,
			PlayedAt: primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[0].Energy.ChargeTimeSegment)),
			Energy:   domain.UserTapEnergy{Charge: 4},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	doc.Tap = domain.UserTap{
		Count:    207,
		Points:   2,
		PlayedAt: primitive.NewDateTimeFromTime(now),
		Energy: domain.UserTapEnergy{
			Charge: 0,
		},
	}
	doc.Points += 14
	s.dbMocks.On("UpdateUserTap", ctx, doc.Profile.Telegram.ID, doc.Tap, doc.Points).Return(doc, nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, doc.Profile.Telegram.ID, doc.Level, doc.Points).Return(nil)

	u, err := s.srv.AddTap(ctx, doc.Profile.Telegram.ID, 7)
	s.NoError(err)
	s.Equal(u.Points, doc.Points)
	s.Equal(u.Tap.Count, doc.Tap.Count)
	s.Equal(u.Tap.Energy.Charge, doc.Tap.Energy.Charge)

	s.dbMocks.AssertExpectations(s.T())
	s.rdbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "UpdateUserTap", ctx, doc.Profile.Telegram.ID, doc.Tap, u.Points)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, doc.Profile.Telegram.ID, u.Level, u.Points)
}

func (s *Suite) TestAddTap_MoreThanEnergyCharge() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Level:  0,
		Points: 100,
		Tap: domain.UserTap{
			Count:    200,
			Points:   2,
			PlayedAt: primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[0].Energy.ChargeTimeSegment)),
			Energy:   domain.UserTapEnergy{Charge: 4},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	doc.Tap = domain.UserTap{
		Count:    207,
		Points:   2,
		PlayedAt: primitive.NewDateTimeFromTime(now),
		Energy: domain.UserTapEnergy{
			Charge: 0,
		},
	}
	doc.Points += 14
	s.dbMocks.On("UpdateUserTap", ctx, doc.Profile.Telegram.ID, doc.Tap, doc.Points).Return(doc, nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, doc.Profile.Telegram.ID, doc.Level, doc.Points).Return(nil)

	u, err := s.srv.AddTap(ctx, doc.Profile.Telegram.ID, 70)
	s.NoError(err)
	s.Equal(u.Points, doc.Points)
	s.Equal(u.Tap.Count, doc.Tap.Count)
	s.Equal(u.Tap.Energy.Charge, doc.Tap.Energy.Charge)

	s.dbMocks.AssertExpectations(s.T())
	s.rdbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "UpdateUserTap", ctx, doc.Profile.Telegram.ID, doc.Tap, u.Points)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, doc.Profile.Telegram.ID, u.Level, u.Points)
}

func (s *Suite) TestRechargeTapEnergy_GhostUser() {
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			IsGhost: true,
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	_, err := s.srv.RechargeTapEnergy(ctx, doc.Profile.Telegram.ID)
	s.ErrorIs(err, domain.ErrGhostUser)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
	s.dbMocks.AssertNotCalled(s.T(), "UpdateUserTapEnergyRecharge")
}

func (s *Suite) TestRechargeTapEnergy_BannedUser() {
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			HasBan: true,
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	_, err := s.srv.RechargeTapEnergy(ctx, doc.Profile.Telegram.ID)
	s.ErrorIs(err, domain.ErrBannedUser)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
	s.dbMocks.AssertNotCalled(s.T(), "UpdateUserTapEnergyRecharge")
}

func (s *Suite) TestRechargeTapEnergy_NotAvailable() {
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Tap: domain.UserTap{
			Energy: domain.UserTapEnergy{
				RechargeAvailable: 0,
			},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	_, err := s.srv.RechargeTapEnergy(ctx, doc.Profile.Telegram.ID)
	s.ErrorIs(err, domain.ErrNoEnergyRecharge)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
	s.dbMocks.AssertNotCalled(s.T(), "UpdateUserTapEnergyRecharge")
}

func (s *Suite) TestRechargeTapEnergy_NotAvailableAfter() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Level: 0,
		Tap: domain.UserTap{
			Energy: domain.UserTapEnergy{
				RechargeAvailable: s.cfg.Rules.Taps[0].Energy.RechargeAvailable - 1,
				RechargedAt:       primitive.NewDateTimeFromTime(now),
			},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	_, err := s.srv.RechargeTapEnergy(ctx, doc.Profile.Telegram.ID)
	s.ErrorIs(err, domain.ErrNoEnergyRecharge)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
	s.dbMocks.AssertNotCalled(s.T(), "UpdateUserTapEnergyRecharge")
}

func (s *Suite) TestRechargeTapEnergy_AvailableNoDelay() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Points: 10,
		Level:  0,
		Tap: domain.UserTap{
			Energy: domain.UserTapEnergy{
				Charge:            4,
				RechargeAvailable: s.cfg.Rules.Taps[0].Energy.RechargeAvailable,
				RechargedAt:       primitive.NewDateTimeFromTime(now),
			},
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	doc.Tap.Energy.RechargeAvailable--
	doc.Tap.Energy.Charge = s.cfg.Rules.TapsBaseEnergyCharge
	s.dbMocks.On("UpdateUserTapEnergyRecharge", ctx, doc.Profile.Telegram.ID, doc.Tap.Energy.RechargeAvailable, s.cfg.Rules.TapsBaseEnergyCharge, 10).Return(doc, nil)

	u, err := s.srv.RechargeTapEnergy(ctx, doc.Profile.Telegram.ID)
	s.NoError(err)

	s.Equal(u.Points, 10)
	s.Equal(u.Tap.Energy.Charge, s.cfg.Rules.TapsBaseEnergyCharge)
	s.Equal(u.Tap.Energy.RechargeAvailable, doc.Tap.Energy.RechargeAvailable)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
	s.dbMocks.AssertCalled(s.T(), "UpdateUserTapEnergyRecharge", ctx, doc.Profile.Telegram.ID, doc.Tap.Energy.RechargeAvailable, s.cfg.Rules.TapsBaseEnergyCharge, 10)
}

func (s *Suite) TestRechargeTapEnergy() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Points: 100,
		Level:  0,
		Tap: domain.UserTap{
			Energy: domain.UserTapEnergy{
				Charge:            0,
				Boost:             make([]int, len(s.cfg.Rules.Taps)),
				RechargeAvailable: 4,
				RechargedAt:       primitive.NewDateTimeFromTime(now.Add(-1 * s.cfg.Rules.Taps[0].Energy.RechargeAvailableAfter).Add(-5 * time.Second)),
			},
		},
	}

	ctx := context.Background()
	d := doc
	d.Tap.Energy.RechargeAvailable--
	d.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(now)
	d.Tap.Energy.Charge = s.cfg.Rules.TapsBaseEnergyCharge
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)
	s.dbMocks.On("UpdateUserTapEnergyRecharge", ctx, doc.Profile.Telegram.ID, d.Tap.Energy.RechargeAvailable, s.cfg.Rules.TapsBaseEnergyCharge, 100).Return(d, nil)

	u, err := s.srv.RechargeTapEnergy(ctx, doc.Profile.Telegram.ID)
	s.NoError(err)
	s.Equal(u.Tap.Energy.RechargeAvailable, doc.Tap.Energy.RechargeAvailable-1)
	s.Equal(u.Tap.Energy.Charge, s.cfg.Rules.TapsBaseEnergyCharge)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	// TODO fix error Should have called with given arguments
	//s.dbMocks.AssertCalled(s.T(), "UpdateUserTapEnergyRecharge", ctx, u.Profile.Telegram.ID, u.Tap.Energy.RechargeAvailable, u.Tap.Energy.Charge, u.Points)
	s.rdbMocks.AssertNotCalled(s.T(), "SetLeaderboardPlayerPoints")
}

func (s *Suite) TestRechargeTapEnergy_AutoClicker() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Points:   100,
		Level:    0,
		PlayedAt: primitive.NewDateTimeFromTime(now.Add(-1 * s.cfg.Rules.Taps[0].Energy.RechargeAvailableAfter).Add(-time.Second)),
		Tap: domain.UserTap{
			Energy: domain.UserTapEnergy{
				Charge:            0,
				Boost:             make([]int, len(s.cfg.Rules.Taps)),
				RechargeAvailable: 5,
				RechargedAt:       primitive.NewDateTimeFromTime(now.Add(-1 * s.cfg.Rules.Taps[0].Energy.RechargeAvailableAfter).Add(-time.Second)),
			},
		},
		AutoClicker: domain.AutoClicker{
			IsAvailable: true,
			IsEnabled:   true,
		},
	}
	uDoc := domain.UserDocument{
		Profile: domain.UserProfile{
			Telegram: domain.TelegramUserProfile{
				ID: 1,
			},
		},
		Points:   101,
		Level:    0,
		PlayedAt: primitive.NewDateTimeFromTime(now.Add(-1 * s.cfg.Rules.Taps[0].Energy.RechargeAvailableAfter).Add(-time.Second)),
		Tap: domain.UserTap{
			Energy: domain.UserTapEnergy{
				Charge:            s.cfg.Rules.TapsBaseEnergyCharge,
				Boost:             make([]int, len(s.cfg.Rules.Taps)),
				RechargeAvailable: 4,
				RechargedAt:       primitive.NewDateTimeFromTime(now),
			},
		},
		AutoClicker: domain.AutoClicker{
			IsAvailable: true,
			IsEnabled:   true,
		},
	}

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, doc.Profile.Telegram.ID, doc.Level, uDoc.Points).Return(nil)
	s.dbMocks.On(
		"UpdateUserTapEnergyRecharge",
		ctx,
		uDoc.Profile.Telegram.ID,
		uDoc.Tap.Energy.RechargeAvailable,
		s.cfg.Rules.TapsBaseEnergyCharge,
		101,
	).Return(uDoc, nil)

	u, err := s.srv.RechargeTapEnergy(ctx, doc.Profile.Telegram.ID)
	s.NoError(err)
	s.Equal(u.Tap.Energy.RechargeAvailable, uDoc.Tap.Energy.RechargeAvailable)
	s.Equal(u.Tap.Energy.Charge, s.cfg.Rules.TapsBaseEnergyCharge)
	s.Equal(u.Points, 101)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	// TODO fix error Should have called with given arguments
	//s.dbMocks.AssertCalled(s.T(), "UpdateUserTapEnergyRecharge", ctx, u.Profile.Telegram.ID, u.Tap.Energy.RechargeAvailable, u.Tap.Energy.Charge, u.Points)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, doc.Profile.Telegram.ID, u.Level, u.Points)
}
