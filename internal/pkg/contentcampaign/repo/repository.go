package repo

import (
	"fmt"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbcontentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscontentcampaign"
)

// Repository type definition
type Repository struct {
	mapper               entityMapper
	dbContentCampaign    dbcontentcampaign.DBSource
	redisCache           rediscache.RedisSource
	redisContentCampaign rediscontentcampaign.RedisSource
}

type contentCampaignCache struct {
	ContentCampaign contentcampaign.ContentCampaign
	Created         time.Time
}

// GetContentCampaignByID func definition
func (r *Repository) GetContentCampaignByID(ccID int64) (*contentcampaign.ContentCampaign, error) {
	ccCacheKey := fmt.Sprintf("CACHE_GO_CONTENTCAMPAIGN-%v", ccID)
	var err error
	var contentCampaign *contentcampaign.ContentCampaign

	if contentCampaign = r.getContentCampaignFromCache(ccCacheKey); contentCampaign != nil {
		return contentCampaign, nil
	}

	if contentCampaign, err = r.getContentCampaignFromDB(ccID); err != nil {
		return nil, err
	}

	r.setContentCampaignToCache(ccCacheKey, contentCampaign)

	return contentCampaign, nil
}

func (r *Repository) setContentCampaignToCache(ccCacheKey string, contentCampaign *contentcampaign.ContentCampaign) {
	ccCache := contentCampaignCache{
		ContentCampaign: *contentCampaign,
		Created:         time.Now(),
	}

	r.redisCache.SetCacheAsync(ccCacheKey, ccCache, time.Hour*24)
}

func (r *Repository) getContentCampaignFromCache(ccCacheKey string) *contentcampaign.ContentCampaign {
	var ccCache contentCampaignCache

	if err := r.redisCache.GetCache(ccCacheKey, &ccCache); ccCache.ContentCampaign.ID == 0 ||
		err != nil || time.Now().After(ccCache.Created.Add(time.Minute)) {
		return nil
	}

	return &ccCache.ContentCampaign
}

func (r *Repository) getContentCampaignFromDB(ccID int64) (*contentcampaign.ContentCampaign, error) {
	dbContentCampaign, err := r.dbContentCampaign.GetContentCampaignByID(ccID)
	contentCampaign := r.mapper.dbCampaignToCampaign(dbContentCampaign)
	if err != nil {
		return nil, err
	}

	return &contentCampaign, nil
}

// IncreaseImpression func definition
func (r *Repository) IncreaseImpression(campaignID int64, unitID int64) error {
	return r.redisContentCampaign.IncreaseImpression(campaignID, unitID)
}

// IncreaseClick func definition
func (r *Repository) IncreaseClick(campaignID int64, unitID int64) error {
	return r.redisContentCampaign.IncreaseClick(campaignID, unitID)
}

// New func definition
func New(dbContentCampaign dbcontentcampaign.DBSource, redisCache rediscache.RedisSource, redisContentCampaign rediscontentcampaign.RedisSource) *Repository {
	return &Repository{
		mapper:               entityMapper{},
		dbContentCampaign:    dbContentCampaign,
		redisCache:           redisCache,
		redisContentCampaign: redisContentCampaign,
	}
}
