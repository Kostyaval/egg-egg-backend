package domain

type LeaderboardPlayer struct {
	Nickname  string `json:"nickname"`
	Level     Level  `json:"level"`
	IsPremium bool   `json:"isPremium"`
	Points    int    `json:"points"`
	Rank      int64  `json:"rank"`
}
