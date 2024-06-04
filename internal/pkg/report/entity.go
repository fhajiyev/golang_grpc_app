package report

// Request struct definition
type Request struct {
	CampaignID      int64  `url:"campaign_id"`
	CampaignName    string `url:"campaign_name"`
	Description     string `url:"description"`
	DeviceID        int64  `url:"device_id"`
	HTML            string `url:"html"`
	IconURL         string `url:"icon_url"`
	IFA             string `url:"ifa"`
	ImageURL        string `url:"image_url"`
	LandingURL      string `url:"landing_url"`
	ReportReason    int    `url:"report_reason"`
	Title           string `url:"title"`
	UnitDeviceToken string `url:"unit_device_token"`
	UnitID          int64  `url:"unit_id"`
	AdReportData    string `url:"ad_report_data"`
}
