package dbcontentcampaign

import (
	"github.com/jinzhu/gorm"
)

// GormDB structu definition
type GormDB struct {
	db *gorm.DB
}

// GetContentCampaignByID returns content campaign record for the campaignID
func (r *GormDB) GetContentCampaignByID(campaignID int64) (*ContentCampaign, error) {
	contentCampaignForQuery := ContentCampaign{ID: campaignID}
	var contentCampaign ContentCampaign

	err := r.db.Where(&contentCampaignForQuery).Find(&contentCampaign).Error
	return &contentCampaign, err
}

// NewSource returns GormDB struct
func NewSource(db *gorm.DB) *GormDB {
	return &GormDB{db}
}

var _ DBSource = &GormDB{}
