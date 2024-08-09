package config

import (
	"time"

	"gitlab.com/egg-be/egg-backend/internal/domain"
)

type Rules struct {
	Referral                   ReferralRules    `yaml:"referral"`
	DailyRewards               []int            `yaml:"dailyRewards"`
	AutoClicker                AutoClickerRules `yaml:"autoClicker"`
	TapsBaseEnergyCharge       int              `yaml:"tapsBaseEnergyCharge"`
	Taps                       TapRules         `yaml:"taps"`
	Tasks                      LevelTasks       `yaml:"tasks"`
	TelegramBotAllowedChannels []int            `yaml:"telegramBotAllowedChannels"`
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
	BoostCost      int `yaml:"boostCost"`
	BoostAvailable int `yaml:"boostAvailable"`
	Energy         struct {
		ChargeTimeSegment      time.Duration `yaml:"chargeTimeSegment"`
		BoostCharge            int           `yaml:"boostCharge"`
		BoostChargeCost        int           `yaml:"boostChargeCost"`
		BoostChargeAvailable   int           `yaml:"boostChargeAvailable"`
		RechargeAvailable      int           `yaml:"rechargeAvailable"`
		RechargeAvailableAfter time.Duration `yaml:"rechargeAvailableAfter"`
	} `yaml:"energy"`
	NextLevel struct {
		Tasks LevelTasks `yaml:"tasks"`
		Cost  int        `yaml:"cost"`
	} `yaml:"nextLevel"`
}

type LevelTasks struct {
	Telegram []int `yaml:"telegram"`
	Referral int   `yaml:"referral"`
}
