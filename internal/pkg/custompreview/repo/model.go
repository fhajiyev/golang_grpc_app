package repo

import "time"

// DBConfig struct definition
type DBConfig struct {
	ID         int64 `gorm:"primary_key"`
	UnitID     int64
	Message    string
	LandingURL string

	// Period
	StartDate       time.Time
	EndDate         time.Time
	StartHourMinute string
	EndHourMinute   string

	// FrequencyLimit
	DIPU *int `gorm:"column:dipu"`
	TIPU *int `gorm:"column:tipu"`
	DCPU *int `gorm:"column:dcpu"`
	TCPU *int `gorm:"column:tcpu"`

	Icon *string

	IsActive bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName is name of DBConfig table
func (DBConfig) TableName() string {
	return "custom_preview_config"
}
