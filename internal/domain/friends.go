package domain

type Friend struct {
	Nickname  *string `json:"nickname"`
	Level     int     `json:"level"`
	IsPremium bool    `json:"isPremium"`
	Points    int     `json:"points"`
}
