package config

type Rules struct {
	Referral ReferralRules `yaml:"referral"`
	Taps     TapRules      `yaml:"taps"`
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

type TapRules []struct {
	Points                int   `yaml:"points"`
	Energy                int   `yaml:"energy"`
	EnergyRecovery        int   `yaml:"energyRecovery"`
	EnergyBoosts          []int `yaml:"energyBoosts"`
	EnergyBoostCost       int   `yaml:"energyBoostCost"`
	EnergyRechargeSeconds int   `yaml:"energyRechargeSeconds"`
}
