package repo

import (
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbnotiplus"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/notiplus"
	"github.com/jinzhu/gorm"
)

// Repository struct definition
type Repository struct {
	db     *gorm.DB
	mapper entityMapper
}

// GetConfigsByUnitID returns config entities for the unit id
func (r *Repository) GetConfigsByUnitID(unitID int64) ([]notiplus.Config, error) {
	var records []dbnotiplus.Config
	if err := r.db.Where("unit_id = ?", unitID).Order("schedule_hour_minute ASC").Find(&records).Error; err != nil {
		return nil, err
	}

	configs := make([]notiplus.Config, 0)
	for _, dbConfig := range records {
		config := r.mapper.dbConfigToConfig(&dbConfig)
		configs = append(configs, config)
	}

	return configs, nil
}

// New returns Repository struct
func New(db *gorm.DB) *Repository {
	return &Repository{db: db, mapper: entityMapper{}}
}
