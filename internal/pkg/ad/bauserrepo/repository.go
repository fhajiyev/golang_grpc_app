package bauserrepo

import (
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	"github.com/jinzhu/gorm"
)

// Repository struct definition
type Repository struct {
	db         *gorm.DB
	redisCache rediscache.RedisSource
	mapper     *entityMapper
}

// GetBAUserByID returns BAUser entity by id
func (r *Repository) GetBAUserByID(id int64) (*ad.BAUser, error) {
	cacheKey := r.getCacheKey(id)
	baUser, err := r.getBAUserFromCache(cacheKey)
	if err != nil {
		baUser, err = r.getByIDFromDB(id)
		if err != nil {
			return nil, err
		}
		r.setBAUserToCacheAsync(cacheKey, baUser)
	}

	return baUser, nil
}

func (r *Repository) getByIDFromDB(id int64) (*ad.BAUser, error) {
	var record BAUser
	if err := r.db.Where("id = ?", id).Find(&record).Error; err != nil {
		return nil, err
	}

	return r.mapper.dbBAUserToBAUser(&record), nil
}

// New returns Repository struct
func New(db *gorm.DB, redisCache rediscache.RedisSource) *Repository {
	return &Repository{db, redisCache, &entityMapper{}}
}
