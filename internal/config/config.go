package config

import (
	"time"
)

type Duration string

func (dur Duration) Parse() time.Duration {
	rd, err := time.ParseDuration(string(dur))
	if err != nil {
		panic(err)
	}
	return rd
}

type PollerConfig struct {
	TimeOut  Duration `json:"timeout"`
	Interval Duration `json:"interval"`
	MaxIters int      `json:"max_iters"`
}

type Config struct {
	Logger struct {
		Enabled  bool     `json:"enabled"`
		Interval Duration `json:"interval"`
	} `json:"logger"`

	Api struct {
		DigPoller          PollerConfig `json:"dig_poller"`
		HealthCheckPoller  PollerConfig `json:"health_check_poller"`
		CashPoller         PollerConfig `json:"cash_poller"`
		ExplorePoller      PollerConfig `json:"explore_poller"`
		BalancePoller      PollerConfig `json:"balance_poller"`
		IssueLicensePoller PollerConfig `json:"issue_license_poller"`
		ListLicensesPoller PollerConfig `json:"list_licenses_poller"`
	} `json:"api"`

	App struct {
		Cashers                      int      `json:"cashers"`
		LicenseIssuers               int      `json:"license_issuers"`
		Explorers                    int      `json:"explorers"`
		Diggers                      int      `json:"diggers"`
		PreExploreWorkers            int      `json:"pre_explore_workers"`
		MaxBlocksPerPreExploreWorker int      `json:"max_block_per_pre_explore_worker"`
		PreExplorationTimeout        Duration `json:"pre_exploration_timeout"`

		MinTreasuresPerBlock 		 int `json:"min_treasures_per_block"`

		License struct {
			PriceList struct {
				Experiments int `json:"experiments"`
				K int `json:"k"`
				G float64 `json:"g"`
			} `json:"price_list"`
			MaxAmount int `json:"max_amount"`
		} `json:"license"`

		World struct {
			SX     int `json:"sx"`
			SY     int `json:"sy"`
			Width  int `json:"width"`
			Height int `json:"height"`
			Depth  int `json:"depth"`
			DepthOptimizer struct {
				K int `json:"k"`
				G int `json:"g"`
			} `json:"depth_optimizer"`
		} `json:"world"`
		Block struct {
			Width  int `json:"width"`
			Height int `json:"height"`
			Auto bool `json:"auto"`
		} `json:"block"`
	} `json:"app"`
	BaseURL string `json:"base_url"`
}
