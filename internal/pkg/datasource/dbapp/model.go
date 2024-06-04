package dbapp

import "time"

// App type definition
type App struct {
	ID               int64 `gorm:"primary_key"`
	LatestAppVersion *int
	IsEnabled        bool
}

// TableName func definition
func (App) TableName() string {
	return "apps"
}

// WelcomeRewardConfig type definition
type WelcomeRewardConfig struct {
	ID            int64 `gorm:"primary_key"`
	UnitID        int64
	Country       *string
	StartTime     *time.Time
	EndTime       *time.Time
	Name          string
	Amount        int
	RetentionDays int
	IsExhausted   bool
	IsTerminated  bool
	MaxNumRewards *int

	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

// TableName func definition
func (WelcomeRewardConfig) TableName() string {
	return "welcome_reward_config"
}

// ReferralRewardConfig stores configration of referral reward for App in DB
type ReferralRewardConfig struct {
	AppID               int64 `gorm:"primary_key"`
	Enabled             bool
	Amount              int
	MaxReferral         int
	StartDate           *time.Time
	EndDate             *time.Time
	VerifyURL           string
	TitleForReferee     string
	TitleForReferrer    string
	TitleForMaxReferrer string
	ExpireHours         int
	MinSdkVersion       int
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TableName is table name in db for RewardConfig
func (ReferralRewardConfig) TableName() string {
	return "referral_reward_config"
}

// AdType type definition
type AdType int

// ContentType type definition
type ContentType int

// UnitType type definition
type UnitType string

// IsActive type definition
type IsActive string

// UnitType constants
const (
	UnitTypeUnknown              UnitType = "U"
	UnitTypeLockscreen           UnitType = "L"
	UnitTypeNative               UnitType = "N"
	UnitTypeBenefitNative        UnitType = "BBN"
	UnitTypeBenefitFeed          UnitType = "BBF"
	UnitTypeBenefitInterstitial  UnitType = "BBI"
	UnitTypeBenefitPop           UnitType = "BBP"
	UnitTypeAdapterRewardedVideo UnitType = "ADTR"
)

// IsActive constants
const (
	Inactive IsActive = "N"
	Active   IsActive = "Y"
)

// Unit type definition
type Unit struct {
	ID                   int64 `gorm:"primary_key"`
	AppID                int64
	AdType               *AdType //`gorm:"DEFAULT:1"`
	BaseReward           *int    //`gorm:"base_reward"`
	BaseInitPeriod       *int    //`gorm:"base_init_period"`
	BuzzadUnitID         int64
	BuzzvilLandingReward int          //`gorm:"buzzvil_landing_reward"`
	ContentType          *ContentType //`gorm:"DEFAULT:1"`
	Country              string       //`gorm:"country"`
	FeedRatio            string
	FirstScreenRatio     string
	FilteredProviders    *string
	InitHMACKey          string //`gorm:"init_hmac_key"`
	Platform             string //`gorm:"platform"`
	Timezone             string
	OrganizationID       int64 //`gorm:"organization_id"`
	PageLimit            *int
	PagerRatio           string
	ShuffleOption        *int
	UnitType             UnitType
	PostbackURL          string
	PostbackAESIv        string
	PostbackAESKey       string
	PostbackHeaders      string
	PostbackHMACKey      string
	PostbackParams       string
	PostbackClass        string
	PostbackConfig       string
	IsActive             IsActive
}

// TableName func definition
func (Unit) TableName() string {
	return "unit"
}
