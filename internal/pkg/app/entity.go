package app

import (
	"fmt"
	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
)

const countryGlobal string = ""

// WelcomeRewardConfig const definition
type WelcomeRewardConfig struct {
	ID            int64
	UnitID        int64
	Country       *string
	StartTime     *time.Time
	EndTime       *time.Time
	Name          string
	Amount        int
	RetentionDays int
}

// WelcomeRewardConfigs is slice of WelcomeRewardConfig
type WelcomeRewardConfigs []WelcomeRewardConfig

func (wrc *WelcomeRewardConfig) IsEndDateInfinite() bool {
	return wrc.EndTime == (*time.Time)(nil)
}

// IsActive returns true if current time is within start time to end time interval
func (wrc *WelcomeRewardConfig) IsActive() bool {
	now := time.Now()
	if wrc.IsEndDateInfinite(){
		return wrc.StartTime.Before(now)
	}

	return wrc.StartTime.Before(now) && wrc.EndTime.After(now)
}

// IsRewarding returns true if WelcomeRewardConfig may give reward to devices now
func (wrc *WelcomeRewardConfig) IsRewarding() bool {
	now := time.Now()
	if wrc.IsEndDateInfinite(){
		return wrc.StartTime.Before(now)
	}

	retentionDuration := time.Duration(wrc.RetentionDays) * 24 * time.Hour
	return wrc.StartTime.Before(now) && wrc.EndTime.Add(retentionDuration).After(now)
}

// IsRegisteredWithinCampaignInterval function tells if the input register seconds is within the config's period
func (wrc *WelcomeRewardConfig) IsRegisteredWithinCampaignInterval(registeredSeconds int64) bool {
	registeredTime := time.Unix(registeredSeconds, 0)
	if wrc.IsEndDateInfinite(){
		return !registeredTime.Before(*wrc.StartTime)
	}

	return !(registeredTime.Before(*wrc.StartTime) || registeredTime.After(*wrc.EndTime))
}

// IsRegisterSecondsRewardable function tells if the register second has satisfied reward conditions now
func (wrc *WelcomeRewardConfig) IsRegisterSecondsRewardable(registeredSeconds int64) bool {
	if !(wrc.IsRewarding() && wrc.IsRegisteredWithinCampaignInterval(registeredSeconds)) {
		return false
	}

	if deltaDays := utils.GetDaysFrom(registeredSeconds); deltaDays < wrc.RetentionDays {
		return false
	}

	return true
}

// FilterWithCountry filter configs with given country
func (wrcs WelcomeRewardConfigs) FilterWithCountry(country string) WelcomeRewardConfigs {
	var filteredConfigs WelcomeRewardConfigs

	for _, wrc := range wrcs {
		// if input country is global, filter global WRCS
		// else filter WRCS with matching country value
		if wrc.Country == nil && country == countryGlobal {
			filteredConfigs = append(filteredConfigs, wrc)
		} else if wrc.Country != nil && *wrc.Country == country {
			filteredConfigs = append(filteredConfigs, wrc)
		}
	}

	return filteredConfigs
}

// FilterActive filter active configs
func (wrcs WelcomeRewardConfigs) FilterActive() WelcomeRewardConfigs {
	var filteredConfigs WelcomeRewardConfigs

	for _, wrc := range wrcs {
		if wrc.IsActive() {
			filteredConfigs = append(filteredConfigs, wrc)
		}
	}
	return filteredConfigs
}

// FilterRewardable filter configs that can give reward to device registered at given time
func (wrcs WelcomeRewardConfigs) FilterRewardable(unitRegisterSeconds int64) WelcomeRewardConfigs {
	var filteredConfigs WelcomeRewardConfigs

	for _, wrc := range wrcs {
		if wrc.IsRegisterSecondsRewardable(unitRegisterSeconds){
			filteredConfigs = append(filteredConfigs, wrc)
		}
	}
	return filteredConfigs
}

// FindOngoingWRC definition
func (wrcs WelcomeRewardConfigs) FindOngoingWRC() *WelcomeRewardConfig {
	if len(wrcs) == 0 {
		return nil
	}

	ret := &wrcs[0]
	for _, wrc := range wrcs {
		if wrc.IsEndDateInfinite() {
			return &wrc
		}

		if wrc.EndTime.After(*ret.EndTime) {
			ret = &wrc
		}
	}
	return ret
}

// ReferralRewardConfig stores configration of referral reward for App
type ReferralRewardConfig struct {
	AppID               int64
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
}

// IsEnded checks reward config is ended
func (rc *ReferralRewardConfig) IsEnded() bool {
	return rc.EndDate != nil && rc.EndDate.Before(time.Now())
}

// Platform constants
const (
	PlatformAndroid = "A"
	PlatformIOS     = "I"
	PlatformWeb     = "W"
)

// AdType constants
const (
	AdTypeNone AdType = iota
	AdTypeAll
	AdTypeOldVerOnly
)

// ContentType constants
const (
	ContentTypeNone ContentType = iota
	ContentTypeAll
	ContentTypeUnitOnly
)

// UnitType constants
const (
	UnitTypeLockscreen           UnitType = "lockscreen"
	UnitTypeNative               UnitType = "native"
	UnitTypeBenefitNative        UnitType = "benefit-native"
	UnitTypeBenefitFeed          UnitType = "benefit-feed"
	UnitTypeBenefitInterstitial  UnitType = "benefit-interstitial"
	UnitTypeBenefitPop           UnitType = "benefit-pop"
	UnitTypeAdapterRewardedVideo UnitType = "adapter-rewarded-video"
	UnitTypeUnknown              UnitType = "unknown"
)

// AdType type definition
type AdType int

// ContentType type definition
type ContentType int

// UnitType type definition
type UnitType string

// AdContentRatio type definition
type AdContentRatio map[string]int

// IsTypeBenefit returns true if the unit is benefit type
func (ut UnitType) IsTypeBenefit() bool {
	switch ut {
	case UnitTypeBenefitFeed, UnitTypeBenefitInterstitial, UnitTypeBenefitNative, UnitTypeBenefitPop:
		return true
	}
	return false
}

// IsTypeLockscreen returns true if the unit is lockscreen type
func (ut UnitType) IsTypeLockscreen() bool {
	return ut == UnitTypeLockscreen
}

// NewAdContentRatio returns AdContentRatio using ad, content value
func NewAdContentRatio(ad, content int) AdContentRatio {
	return AdContentRatio{"ad": ad, "content": content}
}

// String func definition
func (r AdContentRatio) String() string {
	return fmt.Sprintf("%d:%d", r["ad"], r["content"])
}

// App type definition
type App struct {
	ID               int64
	LatestAppVersion *int
	IsEnabled        bool
}

// IsDeactivated return true when app.IsEnabled == false
func (a *App) IsDeactivated() bool {
	return !a.IsEnabled
}

// Unit type definition
type Unit struct {
	ID                   int64
	AppID                int64
	AdType               AdType
	BaseReward           int
	BaseInitPeriod       int
	BuzzadUnitID         int64
	BuzzvilLandingReward int
	ContentType          ContentType
	Country              string
	FeedRatio            AdContentRatio
	FirstScreenRatio     AdContentRatio
	FilteredProviders    *string
	InitHMACKey          string
	Platform             string
	Timezone             string
	OrganizationID       int64
	PageLimit            int
	PagerRatio           AdContentRatio
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
	IsActive             bool
}

var (
	// HsAppIDs map definition
	HsAppIDs = map[int64]struct{}{
		100000043: {},
		100000044: {},
		100000045: {},
		100000046: {},
		100000050: {},
		100000051: {},
		100000060: {},
	}

	//SjAppID variable definition
	SjAppID = int64(210342277740215)
)

// IsMobile returns true on mobile platform
func (u *Unit) IsMobile() bool {
	return u.IsAndroid() || u.IsIOS()
}

// IsAndroid returns true on Android platform
func (u *Unit) IsAndroid() bool {
	return u.Platform == PlatformAndroid
}

// IsIOS returns true on IOS platform
func (u *Unit) IsIOS() bool {
	return u.Platform == PlatformIOS
}

// IsHoneyscreenOrSlidejoyAppID returns true if the unit is honeyscreen or slidejoy unit id
func IsHoneyscreenOrSlidejoyAppID(appID int64) bool {
	if IsHoneyscreenAppID(appID) {
		return true
	}
	return appID == SjAppID
}

// IsHoneyscreenAppID returns true if the unit is honeyscreen unit id
func IsHoneyscreenAppID(appID int64) bool {
	_, ok := HsAppIDs[appID]
	return ok
}
