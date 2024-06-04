package dbapp

import (
	"time"

	"context"

	"github.com/jinzhu/gorm"
	gormtrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/jinzhu/gorm"
)

// GormDB type definition
type GormDB struct {
	db *gorm.DB
}

// GetAppByID func definition
func (r *GormDB) GetAppByID(ctx context.Context, appID int64) (*App, error) {
	var app App
	db := gormtrace.WithContext(ctx, r.db)
	err := db.Where(&App{ID: appID}).First(&app).Error
	if err != nil {
		return nil, err
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return &app, nil
}

// FindRewardingWelcomeRewardConfigs func definition
func (r *GormDB) FindRewardingWelcomeRewardConfigs(ctx context.Context, unitID int64) ([]WelcomeRewardConfig, error) {
	var configs []WelcomeRewardConfig
	timeNow := time.Now()
	db := gormtrace.WithContext(ctx, r.db)
	query := db.Where("is_exhausted = ? and unit_id = ? and start_time < ? and (max_num_rewards is not null or ? < DATE_ADD(end_time, INTERVAL retention_days DAY)) and is_terminated = ?", false, unitID, timeNow, timeNow, false)
	err := query.Find(&configs).Error
	return configs, err
}

// FindReferralRewardConfig func definition
func (r *GormDB) FindReferralRewardConfig(ctx context.Context, appID int64) (*ReferralRewardConfig, error) {
	rrc := ReferralRewardConfig{}
	db := gormtrace.WithContext(ctx, r.db)
	err := db.Where(&ReferralRewardConfig{AppID: appID}).First(&rrc).Error
	return &rrc, err
}

// GetUnit func definition
// TODO support to return nil unit on record not found
func (r *GormDB) GetUnit(ctx context.Context, unitForQuery *Unit) (*Unit, error) {
	var unit Unit
	db := gormtrace.WithContext(ctx, r.db)
	err := db.Where(&unitForQuery).First(&unit).Error
	return &unit, err
}

// NewSource func definition
func NewSource(db *gorm.DB) *GormDB {
	return &GormDB{db}
}
