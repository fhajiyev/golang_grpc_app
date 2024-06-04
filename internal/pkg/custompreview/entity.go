package custompreview

import "time"

// Config struct definition
type Config struct {
	ID         int64
	UnitID     int64
	Message    string
	LandingURL string
	Period
	FrequencyLimit

	Icon *string
}

// Period struct definition
type Period struct {
	StartDate       time.Time
	EndDate         time.Time
	StartHourMinute string
	EndHourMinute   string
}

// FrequencyLimit srtuct definition
type FrequencyLimit struct {
	DIPU *int
	TIPU *int
	DCPU *int
	TCPU *int
}
