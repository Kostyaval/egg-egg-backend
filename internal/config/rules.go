package config

import (
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"time"
)

type Rules struct {
	Referral     ReferralRules    `yaml:"referral"`
	DailyRewards []int            `yaml:"dailyRewards"`
	AutoClicker  AutoClickerRules `yaml:"autoClicker"`
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
