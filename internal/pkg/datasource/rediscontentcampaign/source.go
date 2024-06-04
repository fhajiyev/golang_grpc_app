package rediscontentcampaign

// RedisSource interface definition
type RedisSource interface {
	IncreaseImpression(campaignID int64, unitID int64) error
	IncreaseClick(campaignID int64, unitID int64) error
}
