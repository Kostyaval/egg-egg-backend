package domain

type ReferralBonus struct {
	UserID     int64
	UserPoints int

	ReferralUserID     int64
	ReferralUserPoints int
}
