package dto

// GetConfigsRequest struct definition
type GetConfigsRequest struct {
	UnitID int64 `query:"unit_id" validate:"required"`
}

// Config struct definition
type Config struct {
	ID                 int64  `json:"id"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	Icon               string `json:"icon"`
	ScheduleHourMinute string `json:"schedule_hour_minute"`
}

// GetConfigsResponse struct definition
type GetConfigsResponse struct {
	Configs []Config `json:"configs"`
}
