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

	player := domain.LeaderboardPlayer{
		Nickname:  "me",
		Points:    playerPoints,
		Rank:      0,
		Level:     0,
		IsPremium: false,
	}

	if count > 1 {
		for i := 0; i < count; i++ {
			points := count - i
			if points == playerPoints {
				players[i] = player
			} else {
				players[i] = domain.LeaderboardPlayer{
					Nickname:  fmt.Sprintf("nick%d", i),
					Points:    points,
					Rank:      0,
					Level:     0,
					IsPremium: false,
				}
			}
		}
	} else {
		players[0] = player
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

	if len(l) > 1 {
		for i := 0; i < len(l); i++ {
			s.Equal(l[i].Rank, int64(i+1)+skip)
		}
	}

	return p.Rank
}

func (s *Suite) TestReadLeaderboard_FriendsTab() {
	// me in fetched players list range
	rank := s.readLeaderboardFriendsTab(10, 10, 0, 5)
	s.Equal(rank, int64(6))

	s.SetupTest()

	// me in range of fetched players list (on current page)
	rank = s.readLeaderboardFriendsTab(10, 5, 5, 3)
	s.Equal(rank, int64(8))

	s.SetupTest()

	// me out of fetched players list range (2 page but me on 1)
	rank = s.readLeaderboardFriendsTab(10, 5, 5, 9)
	s.Equal(rank, int64(0))

	s.SetupTest()

	// me out of fetched players list range (1 page but me on 2)
	rank = s.readLeaderboardFriendsTab(10, 5, 0, 3)
	s.Equal(rank, int64(0))

	s.SetupTest()

	// me out of fetched players list range (out of page range)
	rank = s.readLeaderboardFriendsTab(10, 10, 10, 5)
	s.Equal(rank, int64(0))

	s.SetupTest()

	// only me (no friends)
	rank = s.readLeaderboardFriendsTab(1, 5, 0, 10)
	s.Equal(rank, int64(0))

	s.SetupTest()

	// one friend and me
	rank = s.readLeaderboardFriendsTab(2, 5, 0, 1)
	s.Equal(rank, int64(2))

	s.SetupTest()

	// one friend and me
	rank = s.readLeaderboardFriendsTab(2, 5, 0, 2)
	s.Equal(rank, int64(1))

	s.SetupTest()

	// two friends and me
	rank = s.readLeaderboardFriendsTab(3, 5, 0, 3)
	s.Equal(rank, int64(1))
}
