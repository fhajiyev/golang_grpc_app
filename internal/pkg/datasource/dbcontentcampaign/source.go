package dbcontentcampaign

// DBSource interface definition
type DBSource interface {
	GetContentCampaignByID(campaignID int64) (*ContentCampaign, error)
}
