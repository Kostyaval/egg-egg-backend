package domain

type LeaderboardPlayer struct {
	Nickname  string `json:"nickname"`
	Level     int    `json:"level"`
	IsPremium bool   `json:"isPremium"`
	Points    int    `json:"points"`
	Rank      int    `json:"rank"`
}
