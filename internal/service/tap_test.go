package service

import (
	"context"
	"fmt"
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

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
	u := domain.NewUserDocument(s.cfg.Rules)
	u.Profile.Telegram.ID = 1
	u.Points = 100
	u.Tap.Count = 200
	u.Tap.Points = 2
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[0].Energy.ChargeTimeSegment))
	u.Tap.Energy.Charge = 4
	u.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[0].Energy.ChargeTimeSegment))

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, u.Profile.Telegram.ID).Return(u, nil)

	u.Tap.Count = 207
	u.Tap.PlayedAt = primitive.NewDateTimeFromTime(now)
	u.Tap.Energy.Charge = 0
	u.Points += 14
	s.dbMocks.On("UpdateUserTap", ctx, u.Profile.Telegram.ID, u.Tap, u.Points).Return(u, nil)
	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, u.Profile.Telegram.ID, u.Level, u.Points).Return(nil)

	uu, err := s.srv.AddTap(ctx, u.Profile.Telegram.ID, 7)

	s.NoError(err)
	fmt.Printf("========= %d === %d", uu.Points, u.Points)
	s.Equal(uu.Points, u.Points)
	s.Equal(uu.Tap.Count, u.Tap.Count)
	s.Equal(uu.Tap.Energy.Charge, u.Tap.Energy.Charge)

	s.dbMocks.AssertExpectations(s.T())
	s.rdbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "UpdateUserTap", ctx, u.Profile.Telegram.ID, u.Tap, u.Points)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, u.Profile.Telegram.ID, u.Level, u.Points)
}

func (s *Suite) TestAddTap_MoreThanEnergyCharge() {
	now := time.Now().UTC().Truncate(time.Second)
	doc := domain.NewUserDocument(s.cfg.Rules)
	doc.Profile.Telegram.ID = 1
	doc.Points = 100
	doc.Tap.Count = 200
	doc.Tap.Points = 2
	doc.Tap.PlayedAt = primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[0].Energy.ChargeTimeSegment))
	doc.Tap.Energy = domain.UserTapEnergy{Charge: 4}
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(now.Add(-10 * s.cfg.Rules.Taps[0].Energy.ChargeTimeSegment))

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	doc.Tap.Count = 207
	doc.Tap.Points = 2
	doc.Tap.PlayedAt = primitive.NewDateTimeFromTime(now)
	doc.Tap.Energy.Charge = 0
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
	doc := domain.NewUserDocument(s.cfg.Rules)
	doc.Profile.Telegram.ID = 1
	doc.Tap.Energy.RechargeAvailable = 0

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
	doc := domain.NewUserDocument(s.cfg.Rules)

	doc.Profile.Telegram.ID = 1
	doc.Points = 10
	doc.Tap.Energy.Charge = 4
	doc.Tap.Energy.RechargeAvailable = s.cfg.Rules.Taps[0].Energy.RechargeAvailable
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(now)

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
	doc := domain.NewUserDocument(s.cfg.Rules)
	doc.Profile.Telegram.ID = 1
	doc.Points = 100
	doc.Tap.Energy.Charge = 0
	doc.Tap.Energy.RechargeAvailable = 4
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(
		now.Add(-1 * s.cfg.Rules.Taps[0].Energy.RechargeAvailableAfter).Add(-5 * time.Second))

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
	doc := domain.NewUserDocument(s.cfg.Rules)
	doc.Profile.Telegram.ID = 1
	doc.Points = 100
	doc.PlayedAt = primitive.NewDateTimeFromTime(
		now.Add(-1 * s.cfg.Rules.Taps[0].Energy.RechargeAvailableAfter).Add(-time.Second))
	doc.Tap.Energy.Charge = 0
	doc.Tap.Energy.RechargeAvailable = 5
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(
		now.Add(-1 * s.cfg.Rules.Taps[0].Energy.RechargeAvailableAfter).Add(-time.Second))
	doc.AutoClicker.IsAvailable = true
	doc.AutoClicker.IsEnabled = true

	ctx := context.Background()
	s.dbMocks.On("GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID).Return(doc, nil)

	doc.Points = 101
	doc.PlayedAt = primitive.NewDateTimeFromTime(
		now.Add(-1 * s.cfg.Rules.Taps[0].Energy.RechargeAvailableAfter).Add(-time.Second))
	doc.Tap.Energy.Charge = s.cfg.Rules.TapsBaseEnergyCharge
	doc.Tap.Energy.RechargeAvailable = 4
	doc.Tap.Energy.RechargedAt = primitive.NewDateTimeFromTime(now)

	s.rdbMocks.On("SetLeaderboardPlayerPoints", ctx, doc.Profile.Telegram.ID, doc.Level, doc.Points).Return(nil)
	s.dbMocks.On(
		"UpdateUserTapEnergyRecharge",
		ctx,
		doc.Profile.Telegram.ID,
		doc.Tap.Energy.RechargeAvailable,
		s.cfg.Rules.TapsBaseEnergyCharge,
		101,
	).Return(doc, nil)

	u, err := s.srv.RechargeTapEnergy(ctx, doc.Profile.Telegram.ID)
	s.NoError(err)
	s.Equal(u.Tap.Energy.RechargeAvailable, doc.Tap.Energy.RechargeAvailable)
	s.Equal(u.Tap.Energy.Charge, s.cfg.Rules.TapsBaseEnergyCharge)
	s.Equal(u.Points, 101)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "GetUserDocumentWithID", ctx, doc.Profile.Telegram.ID)
	s.dbMocks.AssertCalled(s.T(), "UpdateUserTapEnergyRecharge", ctx, u.Profile.Telegram.ID, u.Tap.Energy.RechargeAvailable, u.Tap.Energy.Charge, u.Points)
	s.rdbMocks.AssertCalled(s.T(), "SetLeaderboardPlayerPoints", ctx, doc.Profile.Telegram.ID, u.Level, u.Points)
}
