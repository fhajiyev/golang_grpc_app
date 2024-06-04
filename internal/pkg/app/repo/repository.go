package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/rediscache"
	"github.com/jinzhu/gorm"
)

const (
	welcomeRewardConfigCacheKeyFormat = "CACHE_GO_WRCS-%v"
	cacheExpiration                   = time.Hour * 24
)

// Repository type definition
type Repository struct {
	dbApp      dbapp.DBSource
	mapper     EntityMapper
	redisCache rediscache.RedisSource
}

// AppCache is used caching app with created time
type AppCache struct {
	App       *app.App
	CreatedAt time.Time
}

func (c *AppCache) needRefresh() bool {
	return time.Now().After(c.CreatedAt.Add(time.Minute))
}

func newAppCache(app *app.App) AppCache {
	return AppCache{
		App:       app,
		CreatedAt: time.Now(),
	}
}

// GetAppByID func definition
func (r *Repository) GetAppByID(ctx context.Context, appID int64) (*app.App, error) {
	var appCache AppCache

	key := fmt.Sprintf("CACHE_GO_APP-%v", appID)
	err := r.redisCache.GetCache(key, &appCache)
	if err != nil || appCache.needRefresh() {
		dbapp, err := r.dbApp.GetAppByID(ctx, appID)
		if err != nil {
			return nil, err
		}
		// dbapp이 nil일때도 캐싱
		if dbapp != nil {
			appCache = newAppCache(r.mapper.dbAppToApp(dbapp))
		}
		r.redisCache.SetCacheAsync(key, appCache, cacheExpiration)
	}

	return appCache.App, nil
}

// WelcomeRewardConfigCache struct definition
type WelcomeRewardConfigCache struct {
	WelcomeRewardConfigs app.WelcomeRewardConfigs
	CreatedAt            time.Time
}

func (configCache *WelcomeRewardConfigCache) isExpired() bool {
	return !time.Now().Before(configCache.CreatedAt.Add(time.Minute))
}

// GetRewardingWelcomeRewardConfigs func definition
func (r *Repository) GetRewardingWelcomeRewardConfigs(ctx context.Context, unitID int64) (app.WelcomeRewardConfigs, error) {
	configCacheKey := fmt.Sprintf(welcomeRewardConfigCacheKeyFormat, unitID)
	var configCache WelcomeRewardConfigCache
	if err := r.redisCache.GetCache(configCacheKey, &configCache); err == nil && !configCache.isExpired() {
		return configCache.WelcomeRewardConfigs, nil
	}

	wrcs, err := r.dbApp.FindRewardingWelcomeRewardConfigs(ctx, unitID)
	var configs app.WelcomeRewardConfigs

	if err == gorm.ErrRecordNotFound {
		configCache.WelcomeRewardConfigs = configs
		configCache.CreatedAt = time.Now()
		r.redisCache.SetCacheAsync(configCacheKey, configCache, cacheExpiration)
		return nil, nil
	}

	if err != nil {
		configCache.WelcomeRewardConfigs = configs
		configCache.CreatedAt = time.Now()
		r.redisCache.SetCacheAsync(configCacheKey, configCache, cacheExpiration)
		return nil, err
	}

	for _, wrc := range wrcs {
		configs = append(configs, r.mapper.welcomeRewardConfigToEntity(&wrc))
	}
	configCache.WelcomeRewardConfigs = configs
	configCache.CreatedAt = time.Now()
	r.redisCache.SetCacheAsync(configCacheKey, configCache, cacheExpiration)

	return configs, nil

}

// GetReferralRewardConfig func definition
func (r *Repository) GetReferralRewardConfig(ctx context.Context, appID int64) (*app.ReferralRewardConfig, error) {
	rrc, err := r.dbApp.FindReferralRewardConfig(ctx, appID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	camp := r.mapper.referralRewardConfigToEntity(rrc)
	return &camp, nil
}

// UnitCache struct definition
type UnitCache struct {
	Unit    app.Unit
	Created time.Time
}

// isValid returns true if the ID of UnitCache is greater than 0
func (uc *UnitCache) isValid() bool {
	return uc.Unit.ID != 0
}

func (uc *UnitCache) isNew() bool {
	return time.Now().Before(uc.Created.Add(time.Minute))
}

// GetUnitByID func definition
func (r *Repository) GetUnitByID(ctx context.Context, unitID int64) (*app.Unit, error) {
	dbUnit := dbapp.Unit{ID: unitID}
	return r.getUnit(ctx, &dbUnit, fmt.Sprintf("CACHE_GO_UNIT-%v", unitID))
}

// GetUnitByAppID func definition
func (r *Repository) GetUnitByAppID(ctx context.Context, appID int64) (*app.Unit, error) {
	dbUnit := dbapp.Unit{AppID: appID}
	return r.getUnit(ctx, &dbUnit, fmt.Sprintf("CACHE_GO_APP_UNIT-%v", appID))
}

// GetUnitByAppIDAndType func definition
func (r *Repository) GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType app.UnitType) (*app.Unit, error) {
	dbUnit := dbapp.Unit{AppID: appID, UnitType: r.mapper.unitTypeToDBUnitType(unitType)}
	return r.getUnit(ctx, &dbUnit, fmt.Sprintf("CACHE_GO_APP_UNIT-%v-%v", appID, unitType))
}

// TODO add testcases
func (r *Repository) getUnit(ctx context.Context, unitForQuery *dbapp.Unit, unitCacheKey string) (*app.Unit, error) {
	var unitCache UnitCache
	if err := r.redisCache.GetCache(unitCacheKey, &unitCache); err == nil && unitCache.isValid() && unitCache.isNew() {
		return &unitCache.Unit, nil
	}

	dbUnit, err := r.dbApp.GetUnit(ctx, unitForQuery)
	if err == nil { // Cache 업데이트 후 리턴
		unit := r.mapper.UnitToEntity(dbUnit)
		unitCache.Unit = unit
		unitCache.Created = time.Now()

		// TODO unitCache가 nil unit도 캐싱할 수 있는 구조로 변경
		r.redisCache.SetCacheAsync(unitCacheKey, unitCache, cacheExpiration)
		return &unitCache.Unit, nil
	}

	if unitCache.isValid() { // Cache 업데이트는 실패 했지만 값이 있음
		return &unitCache.Unit, nil
	}

	// Cache 와 DB 에서 모두 가져오는 것을 실패함
	return nil, err
}

// New returns an app.Repository implementation
func New(da dbapp.DBSource, ra rediscache.RedisSource) *Repository {
	return &Repository{
		dbApp:      da,
		mapper:     EntityMapper{},
		redisCache: ra,
	}
}
