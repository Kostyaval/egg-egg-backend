package service

import (
	"context"
	"fmt"
	"gitlab.com/egg-be/egg-backend/internal/domain"
)

func (s *Suite) TestReadLeaderboard_UnknownTab() {
	var uid int64 = 1

	ctx := context.Background()
	s.dbMocks.On("ReadLeaderboardPlayer", ctx, uid).Return(domain.LeaderboardPlayer{}, nil)

	_, _, _, err := s.srv.ReadLeaderboard(ctx, uid, "spam", 50, 0)
	s.ErrorContains(err, "invalid tab")

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "ReadLeaderboardPlayer", ctx, uid)
}

func (s *Suite) TestReadLeaderboard_UnknownPlayer() {
	var uid int64 = 1

	ctx := context.Background()
	s.dbMocks.On("ReadLeaderboardPlayer", ctx, uid).Return(domain.LeaderboardPlayer{}, domain.ErrNoUser)

	_, _, _, err := s.srv.ReadLeaderboard(ctx, uid, "spam", 50, 0)
	s.ErrorIs(err, domain.ErrNoUser)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "ReadLeaderboardPlayer", ctx, uid)
}

func (s *Suite) readLeaderboardFriendsTab(count int, limit int64, skip int64, playerPoints int) int64 {
	var (
		uid     int64 = 1
		players       = make([]domain.LeaderboardPlayer, count)
	)

	for i := 0; i < count; i++ {
		players[i] = domain.LeaderboardPlayer{
			Nickname:  fmt.Sprintf("nick%d", i),
			Points:    count - i,
			Rank:      0,
			Level:     0,
			IsPremium: false,
		}
	}

	player := domain.LeaderboardPlayer{
		Nickname:  "me",
		Points:    playerPoints,
		Rank:      0,
		Level:     0,
		IsPremium: false,
	}

	ctx := context.Background()
	s.dbMocks.On("ReadLeaderboardPlayer", ctx, uid).Return(player, nil)

	if int64(count) >= skip+limit {
		s.dbMocks.On("ReadFriendsLeaderboardPlayers", ctx, uid, limit, skip).Return(players[skip:skip+limit], nil)
	} else {
		if count == 0 {
			s.dbMocks.On("ReadFriendsLeaderboardPlayers", ctx, uid, limit, skip).Return(players, nil)
		} else {
			s.dbMocks.On("ReadFriendsLeaderboardPlayers", ctx, uid, limit, skip).Return(players[skip:count], nil)
		}
	}

	s.dbMocks.On("ReadFriendsLeaderboardTotalPlayers", ctx, uid).Return(int64(count), nil)

	p, l, c, err := s.srv.ReadLeaderboard(ctx, uid, "friends", limit, skip)
	s.NoError(err)

	s.dbMocks.AssertExpectations(s.T())
	s.dbMocks.AssertCalled(s.T(), "ReadLeaderboardPlayer", ctx, uid)
	s.dbMocks.AssertCalled(s.T(), "ReadFriendsLeaderboardPlayers", ctx, uid, limit, skip)
	s.dbMocks.AssertCalled(s.T(), "ReadFriendsLeaderboardTotalPlayers", ctx, uid)

	s.Equal(int64(count), c)

	if int64(count) > limit {
		s.Len(l, int(limit))
	}

	if p.Rank != 0 {
		for i := 0; i < len(l); i++ {
			if p.Points >= l[i].Points {
				s.Less(p.Rank, l[i].Rank)
			} else {
				s.Greater(p.Rank, l[i].Rank)
			}
		}
	}

	return p.Rank
}

func (s *Suite) TestReadLeaderboard_FriendsTab() {
	rank := s.readLeaderboardFriendsTab(10, 10, 0, 5)
	s.Equal(rank, int64(6))

	s.SetupTest()

	rank = s.readLeaderboardFriendsTab(0, 10, 0, 5)
	s.Equal(rank, int64(0))

	s.SetupTest()

	rank = s.readLeaderboardFriendsTab(10, 10, 10, 5)
	s.Equal(rank, int64(0))

	s.SetupTest()

	rank = s.readLeaderboardFriendsTab(10, 5, 0, 3)
	s.Equal(rank, int64(0))

	s.SetupTest()

	rank = s.readLeaderboardFriendsTab(10, 5, 5, 3)
	s.Equal(rank, int64(8))
}
