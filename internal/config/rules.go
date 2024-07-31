package config

import (
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"time"
)

type Rules struct {
	Referral     ReferralRules    `yaml:"referral"`
	DailyRewards []int            `yaml:"dailyRewards"`
	AutoClicker  AutoClickerRules `yaml:"autoClicker"`
	Taps         TapRules         `yaml:"taps"`
}

// ReferralRules has values of bonus points and index is an egg level.
type ReferralRules []struct {
	Sender struct {
		Plain   int `yaml:"plain"`
		Premium int `yaml:"premium"`
	} `yaml:"sender"`
	Recipient struct {
		Plain   int `yaml:"plain"`
		Premium int `yaml:"premium"`
	} `yaml:"recipient"`
}

type AutoClickerRules struct {
	Speed    time.Duration `yaml:"speed"`
	TTL      time.Duration `yaml:"ttl"`
	Cost     int           `yaml:"cost"`
	MinLevel domain.Level  `yaml:"minLevel"`
}

type TapRules []struct {
	Points int `yaml:"points"`
	Energy struct {
		Max                      int     `yaml:"max"`
		RecoverySeconds          int     `yaml:"recoverySeconds"`
		BoostLimit               int     `yaml:"boostLimit"`
		BoostPackage             int     `yaml:"boostPackage"`
		BoostCost                int     `yaml:"boostCost"`
		RechargeSeconds          float64 `yaml:"rechargeSeconds"`
		FullRechargeCount        int     `yaml:"fullRechargeCount"`
		FullRechargeDelaySeconds int     `yaml:"fullRechargeDelaySeconds"`
	} `yaml:"energy"`
}
