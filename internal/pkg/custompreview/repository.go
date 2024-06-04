package custompreview

import "time"

// Repository interface definition
type Repository interface {
	GetConfigByUnitID(unitID int64, isActive bool, targetTime time.Time) (*Config, error)
}
