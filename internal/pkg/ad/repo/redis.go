package repo

import (
	"errors"
	"fmt"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
)

type adDetailCache struct {
	Detail    ad.Detail
	CreatedAt time.Time
}

func (r *Repository) getCacheKey(adID int64) string {
	return fmt.Sprintf("CACHE_GO_AD_DETAIL-%v", adID)
}

func (r *Repository) getAdDetailFromCache(cacheKey string) (*ad.Detail, error) {
	var cache adDetailCache
	err := r.redisCache.GetCache(cacheKey, &cache)
	if err != nil {
		return nil, err
	} else if cache.Detail.ID == 0 {
		return nil, errors.New("failed to parse ad detail")
	} else if cache.CreatedAt.Before(time.Now().Add(-time.Hour)) {
		return nil, errors.New("expired cache")
	}

	return &cache.Detail, nil
}

func (r *Repository) setAdDetailToCacheAsync(cacheKey string, detail ad.Detail) {
	cache := adDetailCache{
		Detail:    detail,
		CreatedAt: time.Now(),
	}

	r.redisCache.SetCacheAsync(cacheKey, &cache, time.Hour)
}
