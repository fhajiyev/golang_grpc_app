package dto

import "time"

// GetConfigReq struct definition
type GetConfigReq struct {
	UnitID int64 `query:"unit_id" validate:"required"`
}

// GetConfigRes struct definition
type GetConfigRes struct {
	Config
}

// Config struct definition
type Config struct {
	ID         int64  `json:"id"`
	UnitID     int64  `json:"unit_id"`
	Message    string `json:"message"`
	LandingURL string `json:"landing_url"`
	Period
	FrequencyLimit

	Icon *string `json:"icon,omitempty"`
}

// Period struct definition
type Period struct {
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	StartHourMinute string    `json:"start_hour_minute"`
	EndHourMinute   string    `json:"end_hour_minute"`
}

// FrequencyLimit struct definition
type FrequencyLimit struct {
	DIPU int `json:"dipu"`
	TIPU int `json:"tipu"`
	DCPU int `json:"dcpu"`
	TCPU int `json:"tcpu"`
}
