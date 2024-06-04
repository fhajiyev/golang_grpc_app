package dto

import "github.com/Buzzvil/buzzscreen-api/internal/pkg/app"

// App model definition
type App struct {
	ID int64 `json:"id"`
}

// Unit model definition
type Unit struct {
	ID       int64        `json:"id"`
	Type     app.UnitType `json:"type"`
	Settings Settings     `json:"settings"`
}

// Settings of unit model definition
type Settings struct {
	BaseHourLimit    int                `json:"base_hour_limit"`
	BaseInitPeriod   int                `json:"base_init_period"`
	BaseReward       int                `json:"base_reward"`
	FeedRatio        app.AdContentRatio `json:"feed_ratio"`
	FirstScreenRatio app.AdContentRatio `json:"first_screen_ratio"`
	HourLimit        int                `json:"hour_limit"`
	PagerRatio       app.AdContentRatio `json:"pager_ratio"`
	PageLimit        int                `json:"page_limit"`
}

// UnitConfigs definition
type UnitConfigs struct {
	OrganizationID int64    `json:"organization_id"`
	Postback       Postback `json:"postback"`
}

// Postback definition
type Postback struct {
	URL     string `json:"url"`
	AESIv   string `json:"aes_iv"`
	AESKey  string `json:"aes_key"`
	Headers string `json:"headers"`
	HMACKey string `json:"hmac_key"`
	Params  string `json:"params"`
	Class   string `json:"class"`
	Config  string `json:"config"`
}
