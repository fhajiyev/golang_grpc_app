package ad

import (
	"strconv"

	"github.com/Buzzvil/buzzlib-go/core"
)

// BAUser entity
type BAUser struct {
	ID          int64
	AccessToken string
	Name        string
	IsMedia     bool
}

const staffBAUserID = 1

// Detail represents specific or secret data for Ad
type Detail struct {
	ID             int64
	Name           string
	OrganizationID int64
	RevenueType    string
	Extra          map[string]interface{}
}

// AdV1 definition
type AdV1 struct {
	AdChoiceURL           string                   `json:"adchoice_url"`
	AdNetworkID           int                      `json:"adnetwork_id"`
	AgeFrom               int                      `json:"age_from"`
	AgeTo                 int                      `json:"age_to"`
	Carrier               string                   `json:"carrier"`
	ClickBeacons          []string                 `json:"click_beacons"`
	ClickURL              string                   `json:"click_url"`
	DeviceName            string                   `json:"device_name"`
	Dipu                  float64                  `json:"dipu"`
	DisplayType           string                   `json:"display_type"`
	EndedAt               float64                  `json:"ended_at"`
	Extra                 map[string]interface{}   `json:"extra"`
	FirstDisplayPriority  float64                  `json:"first_display_priority"`
	FirstDisplayWeight    int                      `json:"first_display_weight"`
	FailTrackers          []string                 `json:"fail_trackers"`
	Icon                  string                   `json:"icon"`
	ID                    int64                    `json:"id"`
	Image                 string                   `json:"image"`
	ImageIos              string                   `json:"image_ios"`
	ImpressionURLs        []string                 `json:"impression_urls"`
	Ipu                   int                      `json:"ipu"`
	IsRtb                 bool                     `json:"is_rtb"`
	LandingType           string                   `json:"landing_type"`
	Meta                  map[string]interface{}   `json:"meta"`
	Name                  string                   `json:"name"`
	OrganizationID        int64                    `json:"organization_id"`
	OwnerID               int64                    `json:"owner_id"`
	PreferredBrowser      *string                  `json:"preferred_browser,omitempty"`
	Region                string                   `json:"region"`
	RemoveAfterImpression bool                     `json:"remove_after_impression"`
	Sex                   string                   `json:"sex"`
	Slot                  string                   `json:"slot"`
	StartedAt             float64                  `json:"started_at"`
	SupportWebp           bool                     `json:"support_webp"`
	TargetApp             string                   `json:"target_app"`
	Timezone              string                   `json:"timezone"`
	Tipu                  int                      `json:"tipu"`
	Type                  string                   `json:"type"`
	UnitPrice             float64                  `json:"unit_price"`
	UseWebUa              bool                     `json:"use_web_ua"`
	BannerAd              *BannerAdV1Settings      `json:"banner_ad"`
	WebHTML               *map[string]interface{}  `json:"web_html,omitempty"`
	Creative              map[string]interface{}   `json:"creative"`
	IsIncentive           bool                     `json:"is_incentive"`
	AdReportData          string                   `json:"ad_report_data"`
	Creatives             []map[string]interface{} `json:"creatives,omitempty"`

	// TODO remove
	UnlockReward  int `json:"unlock_reward"`
	ActionReward  int `json:"action_reward"`
	LandingReward int `json:"landing_reward"`

	// TODO remove omitempty option
	Events Events `json:"events,omitempty"`
}

// AdV2 definition
type AdV2 struct {
	ID                 int64                    `json:"id"`
	ActionReward       int                      `json:"actionReward"`
	Creative           map[string]interface{}   `json:"creative"`
	CallToAction       string                   `json:"callToAction"`
	ClickTrackers      []string                 `json:"clickTrackers"`
	FailTrackers       []string                 `json:"failTrackers"`
	ImpressionTrackers []string                 `json:"impressionTrackers"`
	ConversionCheckURL *string                  `json:"conversionCheckUrl"`
	SkStoreProductURL  *string                  `json:"skStoreProductUrl"`
	LandingReward      int                      `json:"landingReward"`
	Meta               map[string]interface{}   `json:"meta"`
	Name               string                   `json:"name"`
	Network            *string                  `json:"network"`
	OrganizationID     int64                    `json:"organizationId"`
	OwnerID            int64                    `json:"owner_id"`
	TTL                *int                     `json:"ttl"`
	RevenueType        string                   `json:"revenueType"`
	UnlockReward       int                      `json:"unlockReward"`
	AdReportData       *string                  `json:"adReportData"`
	Extra              map[string]interface{}   `json:"extra,omitempty"`
	PreferredBrowser   *string                  `json:"preferredBrowser,omitempty"`
	RewardPeriod       int64                    `json:"rewardPeriod"`
	Creatives          []map[string]interface{} `json:"creatives,omitempty"`
	Product            map[string]interface{}   `json:"product,omitempty"`

	// TODO remove omitempty option
	Events Events `json:"events,omitempty"`
}

const (
	eventTypeLanded = "landed"
	eventTypeAction = "action" // TODO replace
)

// Event definition
type Event struct {
	Type         string            `json:"event_type"`
	TrackingURLs []string          `json:"tracking_urls"`
	Reward       *Reward           `json:"reward,omitempty"`
	Extra        map[string]string `json:"extra"`
}

// Events definition
type Events []Event

// LandedEvent finds landed event
func (events Events) LandedEvent() (Event, bool) {
	for _, e := range events {
		if e.Type == eventTypeLanded {
			return e, true
		}
	}
	return Event{}, false
}

// ActionEvent finds action event
func (events Events) ActionEvent() (Event, bool) {
	// TODO multi reward 대응 필요
	for _, e := range events {
		if e.Type == eventTypeAction {
			return e, true
		}
	}
	return Event{}, false
}

// Reward definition
type Reward struct {
	Amount         int64             `json:"amount"`
	Status         string            `json:"status"`
	IssueMethod    string            `json:"issue_method"`
	StatusCheckURL string            `json:"status_check_url"`
	TTL            int64             `json:"ttl"`
	Extra          map[string]string `json:"extra"`
}

// MinimumStayDuration returns minimum stay duration from extra
// It returns 0 if minimum stay duration is not exist
func (reward Reward) MinimumStayDuration() int {
	minimumStayDurationStr, ok := reward.Extra["minimum_stay_duration"]
	if !ok {
		return 0
	}

	minimumStayDuration, err := strconv.Atoi(minimumStayDurationStr)
	if err != nil {
		core.Logger.Warnf("minimum stay druation is not number format.")
		return 0
	}

	return minimumStayDuration
}

// BannerAdV1Settings definition
type BannerAdV1Settings struct {
	HTML       string  `json:"html"`
	Width      float64 `json:"width"`
	Height     float64 `json:"height"`
	Background string  `json:"background"`
}

// NativeAdV1Settings definition
type NativeAdV1Settings struct {
	LifeTime    float64 `json:"lifetime"`
	Period      float64 `json:"period"`
	ID          string  `json:"id"`
	Background  string  `json:"background"`
	Network     string  `json:"network"`
	Name        string  `json:"name"`
	PlacementID string  `json:"placement_id"`
	PublisherID string  `json:"publisher_id"`
	ReferrerURL string  `json:"referrer_url"`
}

// NativeAdV1 definition
type NativeAdV1 struct {
	AdV1
	NativeAd *NativeAdV1Settings `json:"native_ad"`
	Banner   *NativeAdV1Settings `json:"banner"`
}

// AdsV1Settings definition
type AdsV1Settings struct {
	FilteringWords *string `json:"filtering_words"`
	WebUa          *string `json:"web_ua"`
}

// AdsV2Settings definition
type AdsV2Settings map[string]interface{}
