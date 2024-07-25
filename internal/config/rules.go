package config

type Rules struct {
	Referral          ReferralRules `yaml:"referral"`
	DailyRewardsRules []int         `yaml:"dailyRewards"`
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
