package model

import (
	"strconv"
	"time"

	"github.com/gorhill/cronexpr"
)

const (
	// NotificationTypeFeed const definition
	NotificationTypeFeed NotificationType = 1 + iota
)

// NotificationImportance constants
const (
	NotificationImportanceLow NotificationImportance = 1 + iota
	NotificationImportanceDefault
	NotificationImportanceHigh
)

type (
	// NotificationType indicates the type of notification.
	NotificationType int
	// NotificationImportance type definition
	NotificationImportance int

	// NotificationSchedule defines the schedule of notification.
	NotificationSchedule struct {
		ID               int64                  `gorm:"primary_key" json:"id"`
		UnitID           int64                  `json:"unit_id"`
		Title            string                 `json:"title"`
		Description      string                 `json:"description"`
		Schedule         string                 `json:"schedule"`
		IconURL          string                 `json:"icon_url"`
		InboxSummary     string                 `json:"inbox_summary"`
		NotificationType NotificationType       `json:"notification_type"`
		Importance       NotificationImportance `json:"importance"`
		MinVersionCode   *int                   `json:"min_version_code"`
		MaxVersionCode   *int                   `json:"max_version_code"`

		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

// Contains checks if the time is included in the schedule.
func (schedule NotificationSchedule) Contains(targetTime time.Time) bool {
	nextTime := cronexpr.MustParse(schedule.Schedule).Next(targetTime.Truncate(time.Minute))
	return nextTime.Sub(targetTime) <= time.Second
}

// Link returns URI for landing.
func (schedule NotificationSchedule) Link() string {
	switch schedule.NotificationType {
	case NotificationTypeFeed:
		return "buzzad://benefit/feed?unit_id=" + strconv.FormatInt(schedule.UnitID, 10)
	}
	return ""
}

// TableName indicates the name of db table.
func (NotificationSchedule) TableName() string {
	return "notification_schedules"
}

// String func definition
func (importance NotificationImportance) String() string {
	return []string{"none", "low", "default", "high"}[importance]
}
