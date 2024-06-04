package repo

import (
	"fmt"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview"
	"github.com/jinzhu/gorm"
)

// Repository struct definition
type Repository struct {
	db     *gorm.DB
	mapper *entityMapper
}

// New creates user repository
func New(db *gorm.DB) *Repository {
	repo := Repository{
		db:     db,
		mapper: &entityMapper{},
	}
	return &repo
}

// GetConfigByUnitID returns lastest configs which has period including target time with given unit id
func (r *Repository) GetConfigByUnitID(unitID int64, isActive bool, targetTime time.Time) (*custompreview.Config, error) {
	if unitID == 0 {
		return nil, fmt.Errorf("unit id shouldn't be 0 to get config")
	}

	whereStatement := "unit_id = ? AND " +
		"is_active = ? AND " +
		"start_date <= ? AND " +
		"end_date >= ? "

	var dbconfig DBConfig
	err := r.db.Where(whereStatement, unitID, isActive, targetTime, targetTime).Last(&dbconfig).Error
	if err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			return nil, nil
		default:
			return nil, err
		}
	}

	return r.mapper.dbConfigToConfig(dbconfig), nil
}
