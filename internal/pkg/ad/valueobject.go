package ad

// V1AdsRequest definition
// When this struct is modified, check LogAdAllocationRequestV1 method has been modified accordingly
type V1AdsRequest struct {
	AccountID       int64  `url:"account_id"`
	UnitID          int64  `url:"unit_id"`
	IFA             string `url:"ifa"`
	UnitDeviceToken string `url:"user_id"`

	DeviceOS       int    `url:"device_os"`
	DeviceName     string `url:"device_name"`
	MembershipDays int    `url:"membership_days"`
	Carrier        string `url:"carrier"`
	PackageName    string `url:"package_name"`
	UserAgent      string `url:"user_agent"`
	SDKVersion     int    `url:"sdk_version"`
	Sex            string `url:"sex"`
	MCC            string `url:"mcc"`
	MNC            string `url:"mnc"`

	ClientIP    string `url:"client_ip"`
	Region      string `url:"region"`
	TimeZone    string `url:"timezone"`
	NetworkType string `url:"network_type"`

	CustomTarget1 string `url:"custom_target_1"`
	CustomTarget2 string `url:"custom_target_2"`
	CustomTarget3 string `url:"custom_target_3"`

	RevenueTypes                 string `url:"revenue_types"`
	CreativeSize                 int    `url:"creative_size"`
	SupportRemoveAfterImpression int    `url:"support_remove_after_impression"`

	YearOfBirth                 *int    `url:"year_of_birth,omitempty"`
	IsIFALimitAdTrackingEnabled *int    `url:"is_ifa_limit_ad_tracking_enabled,omitempty"`
	InstalledBrowsers           *string `url:"installed_browsers,omitempty"`
	DefaultBrowser              *string `url:"default_browser,omitempty"`

	IsAllocationTest int  `url:"is_allocation_test"`
	IsTest           *int `url:"is_test,omitempty"`
}

// V1AdsResponse definition
type V1AdsResponse struct {
	Ads       []*AdV1        `json:"ads"`
	NativeAds []*NativeAdV1  `json:"native_ads"`
	Code      int            `json:"code"`
	Msg       string         `json:"msg"`
	Settings  *AdsV1Settings `json:"settings"`
}

// V2AdsRequest definition
// When this struct is modified, check LogAdAllocationRequestV2 method has been modified accordingly
type V2AdsRequest struct {
	AccountID int64  `url:"accountId"`
	UnitID    int64  `url:"unitId"`
	IFA       string `url:"ifa"`

	DeviceID       string `url:"deviceId"`
	DeviceName     string `url:"deviceName"`
	SdkVersion     int    `url:"sdkVersion"`
	OsVersion      string `url:"osVersion"`
	UserAgent      string `url:"userAgent"`
	Relationship   string `url:"relationship"`
	MembershipDays int    `url:"membershipDays"`
	Country        string `url:"country"`
	Gender         string `url:"gender"`

	Timezone string `url:"timezone"`
	ClientIP string `url:"clientIp"`

	Language string `url:"language"`

	NetworkType  string `url:"networkType"`
	CreativeSize int    `url:"creativeSize"`

	Platform string `url:"platform"`

	CustomTarget1 string `url:"customTarget1"`
	CustomTarget2 string `url:"customTarget2"`
	CustomTarget3 string `url:"customTarget3"`

	Cursor      string `url:"cursor"`
	Types       string `url:"types"`
	LineitemIds string `url:"lineitemIds"`

	IsMockResponse bool `url:"isMockResponse"`

	TargetFill      int     `url:"targetFill,omitempty"`
	UnitDeviceToken *string `url:"userId,omitempty"`
	AndroidID       *string `url:"androidId,omitempty"`
	Birthday        *string `url:"birthday,omitempty"`
	MCC             *string `url:"mcc,omitempty"`
	MNC             *string `url:"mnc,omitempty"`

	Latitude  *string `url:"latitude,omitempty"`
	Longitude *string `url:"longitude,omitempty"`

	RevenueTypes *string `url:"revenueTypes,omitempty"`
	CPSCategory  *string `url:"cpsCategory,omitempty"`
}

// V2AdsResponse definition
type V2AdsResponse struct {
	Ads      []AdV2         `json:"ads"`
	Cursor   string         `json:"cursor"`
	Settings *AdsV2Settings `json:"settings"`
}
