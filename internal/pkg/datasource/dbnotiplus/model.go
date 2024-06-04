package dbnotiplus

import "time"

// Config record definition
type Config struct {
	ID                 int64 `gorm:"primary_key"`
	UnitID             int64
	Title              string
	Description        string
	Icon               string
	ScheduleHourMinute string

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns "noti_plus_configs"
func (Config) TableName() string {
	return "noti_plus_configs"
}
