package utils

import (
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
)

var (
	defaultDateLayout  = "2006-01-02"
	defaultTimeLayouts = []string{"2006-01-02T15:04:05-07:00", "2006-01-02T15:04:05.000000-07:00", "2006-01-02T15:04:05.000-07:00", "2006-01-02T15:04:05.000Z", "2006-01-02T15:04:05.000000Z", "2006-01-02T15:04:05", "2006-01-02T15:04:05Z", defaultDateLayout}
)

// GetAge func definition
func GetAge(yyyymmdd string) int {
	birthday := ConvertToTime(yyyymmdd, defaultDateLayout)
	now := time.Now()
	years := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		years--
	}
	return years
}

// GetDaysFrom func definition
func GetDaysFrom(fromSeconds int64) int {
	return int((time.Now().Unix() - fromSeconds) / int64(time.Hour/time.Second*24))
}

// ConvertToUnixTime func definition
func ConvertToUnixTime(timeStr string) int64 {
	for _, layout := range defaultTimeLayouts {
		if len(timeStr) != len(layout) {
			continue
		}
		t := ConvertToTime(timeStr, layout)
		return t.Unix()
	}
	panic("Failed to convert time. timeStr: " + timeStr)
}

// ConvertToTime func definition
func ConvertToTime(timeStr, layout string) time.Time {
	t, err := time.Parse(layout, timeStr)
	if err != nil {
		core.Logger.Printf("convertFail - %v", err)
	}
	return t
}

// GetWeekdayStartsFromMonday func definition
func GetWeekdayStartsFromMonday(t *time.Time) int {
	return (int(t.Weekday()) + 6) % 7
}
