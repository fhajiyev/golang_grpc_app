package repo

import (
	"time"
)

// ContentReported struct definition
type (
	ContentReported struct {
		ID              int64 `gorm:"primary_key"`
		CampaignID      int64
		CampaignName    string
		Description     string
		DeviceID        int64
		HTML            string
		IconURL         string
		IFA             string
		ImageURL        string
		LandingURL      string
		ReportReason    int
		Title           string
		UnitID          int64
		UnitDeviceToken string

		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

// TableName returns name of content repoted table
func (ContentReported) TableName() string {
	return "content_reported"
}
