package contentcampaign

// Repository type definition
type Repository interface {
	GetContentCampaignByID(campaignID int64) (*ContentCampaign, error)
	IncreaseClick(campaignID int64, unitID int64) error
	IncreaseImpression(campaignID int64, unitID int64) error
}
