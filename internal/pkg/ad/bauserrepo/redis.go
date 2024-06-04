package bauserrepo

import (
	"errors"
	"fmt"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
)

type baUserCache struct {
	BAUser    ad.BAUser
	CreatedAt time.Time
}

func (r *Repository) getCacheKey(id int64) string {
	return fmt.Sprintf("CACHE_GO_BAUSER-%v", id)
}

func (r *Repository) getBAUserFromCache(cacheKey string) (*ad.BAUser, error) {
	var cache baUserCache
	err := r.redisCache.GetCache(cacheKey, &cache)
	if err != nil {
		return nil, err
	} else if cache.BAUser.ID == 0 {
		return nil, errors.New("unparsed cache")
	} else if cache.CreatedAt.Before(time.Now().Add(-time.Hour)) {
		return nil, errors.New("expired cache")
	}

	return &cache.BAUser, nil
}

func (r *Repository) setBAUserToCacheAsync(cacheKey string, baUser *ad.BAUser) {
	baUserCache := baUserCache{
		BAUser:    *baUser,
		CreatedAt: time.Now(),
	}

	r.redisCache.SetCacheAsync(cacheKey, &baUserCache, time.Hour*24)
}
