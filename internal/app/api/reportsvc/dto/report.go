package dto

// ReportCampaignRequest struct definition
type (
	ReportCampaignRequest struct {
		CampaignID      int64  `form:"campaign_id" query:"campaign_id" validate:"required"`
		CampaignName    string `form:"campaign_name" query:"campaign_name"`
		Description     string `form:"description" query:"description"`
		DeviceID        int64  `form:"device_id" query:"device_id"`
		HTML            string `form:"html" query:"html"`
		HTMLTag         string `form:"htmlTag" query:"htmlTag"`
		IconURL         string `form:"icon_url" query:"icon_url"`
		IFA             string `form:"ifa" query:"ifa"`
		ImageURL        string `form:"image_url" query:"image_url"`
		IsAd            bool   `form:"is_ad" query:"is_ad"`
		LandingURL      string `form:"landing_url" query:"landing_url"`
		ReportReason    int    `form:"report_reason" query:"report_reason"`
		Title           string `form:"title" query:"title"`
		UnitDeviceToken string `form:"unit_device_token" query:"unit_device_token"`
		UnitIDReq       int64  `form:"unit_id" query:"unit_id"`
		AppIDReq        int64  `form:"app_id" query:"app_id"`
		AdReportData    string `form:"ad_report_data" query:"ad_report_data"`
	}
)
