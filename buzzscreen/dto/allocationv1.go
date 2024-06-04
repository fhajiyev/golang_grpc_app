package dto

type (
	// AllocV1Request type definition
	AllocV1Request struct {
		ContentAllocV1Request
		IsTest                      bool   `form:"is_test" query:"is_test"`
		UserAgentReq                string `form:"user_agent" query:"user_agent"`
		SettingsPageDisplayRatio    string `form:"settings_page_display_ratio" query:"settings_page_display_ratio"`
		DefaultBrowser              string `form:"default_browser" query:"default_browser"`
		InstalledBrowsers           string `form:"installed_browsers" query:"installed_browsers"`
		IsIFALimitAdTrackingEnabled bool   `form:"is_ifa_limit_ad_tracking_enabled" query:"is_ifa_limit_ad_tracking_enabled"`
		Mcc                         string `form:"mcc" query:"mcc"`
		Mnc                         string `form:"mnc" query:"mnc"`
		TimeZone                    string `form:"timezone" query:"timezone" validate:"required"`
		CreativeSize                int    `form:"creative_size"`

		userAgent string
	}

	// AllocV1Settings type definition
	AllocV1Settings struct {
		AdFilteringWords      *string `binding:"omitempty" json:"ad_filtering_words"`
		BaseHourLimit         int     `json:"base_hour_limit"`
		BaseInitPeriod        int     `json:"base_init_period"`
		ExternalBaseReward    int     `json:"external_base_reward"`
		ExternalCampaignID    int     `json:"external_campaign_id"`
		ExternalImpressionCap int     `json:"external_impression_cap"`
		ExternalAddCallLimit  int     `json:"external_add_call_limit"`
		FirstDisplayRatio     string  `json:"first_display_ratio"`
		HourLimit             int     `json:"hour_limit"`
		PageDisplayRatio      string  `json:"page_display_ratio"`
		PageLimit             int     `json:"page_limit"`
		RequestTrigger        int     `json:"request_trigger"`
		RequestPeriod         int     `json:"request_period"`
		ShuffleOption         int     `json:"shuffle_option"`
		WebUserAgent          *string `json:"web_ua"`
	}

	// AllocV1Response type definition
	AllocV1Response struct {
		ContentAllocV1Response
		NativeCampaigns []*NativeCampaignV1 `json:"native_ads"`
		Settings        *AllocV1Settings    `json:"settings"`
	}
)

// GetUserAgent func definition
func (ar *AllocV1Request) GetUserAgent() string {
	if ar.userAgent == "" {
		if ar.UserAgentReq != "" {
			ar.userAgent = ar.UserAgentReq
		} else {
			ar.userAgent = ar.Request.UserAgent()
		}
	}
	return ar.userAgent
}
