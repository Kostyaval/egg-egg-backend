package config

import (
	"gitlab.com/egg-be/egg-backend/internal/domain"
	"time"
)

type Rules struct {
	Referral             ReferralRules    `yaml:"referral" json:"referral"`
	DailyRewards         []int            `yaml:"dailyRewards" json:"dailyRewards"`
	AutoClicker          AutoClickerRules `yaml:"autoClicker" json:"autoClicker"`
	TapsBaseEnergyCharge int              `yaml:"tapsBaseEnergyCharge" json:"tapsBaseEnergyCharge"`
	Taps                 TapRules         `yaml:"taps" json:"taps"`
	Tasks                      LevelTasks       `yaml:"tasks" json:"tasks"`
	TelegramBotAllowedChannels []int            `yaml:"telegramBotAllowedChannels" json:"telegramBotAllowedChannels"`
}

// ReferralRules has values of bonus points and index is an egg level.
type ReferralRules []struct {
	Sender struct {
		Plain   int `yaml:"plain" json:"plain"`
		Premium int `yaml:"premium" json:"premium"`
	} `yaml:"sender" json:"sender"`
	Recipient struct {
		Plain   int `yaml:"plain" json:"plain"`
		Premium int `yaml:"premium" json:"premium"`
	} `yaml:"recipient" json:"recipient"`
}

type AutoClickerRules struct {
	Speed    time.Duration `yaml:"speed" json:"speed"`
	TTL      time.Duration `yaml:"ttl" json:"ttl"`
	Cost     int           `yaml:"cost" json:"cost"`
	MinLevel domain.Level  `yaml:"minLevel" json:"minLevel"`
}

type TapRules []struct {
	BoostCost      int `yaml:"boostCost" json:"boostCost"`
	BoostAvailable int `yaml:"boostAvailable" json:"boostAvailable"`
	Energy         struct {
		ChargeTimeSegment      time.Duration `yaml:"chargeTimeSegment" json:"chargeTimeSegment"`
		BoostCharge            int           `yaml:"boostCharge" json:"boostCharge"`
		BoostChargeCost        int           `yaml:"boostChargeCost" json:"boostChargeCost"`
		BoostChargeAvailable   int           `yaml:"boostChargeAvailable" json:"boostChargeAvailable"`
		RechargeAvailable      int           `yaml:"rechargeAvailable" json:"rechargeAvailable"`
		RechargeAvailableAfter time.Duration `yaml:"rechargeAvailableAfter" json:"rechargeAvailableAfter"`
	} `yaml:"energy" json:"energy"`
	NextLevel struct {
		Tasks LevelTasks `yaml:"tasks" json:"tasks"`
		Cost  int        `yaml:"cost" json:"cost"`
	} `yaml:"nextLevel" json:"nextLevel"`
}
