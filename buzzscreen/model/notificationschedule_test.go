package model

import (
	"strconv"
	"testing"
	"time"
)

func TestNotificationSchedule(t *testing.T) {
	now := time.Now()
	t.Run("Time included", func(t *testing.T) {
		schedule := NotificationSchedule{
			ID:               100001,
			UnitID:           10101,
			Title:            "Example Title",
			Description:      "Example Description",
			Schedule:         "* * * * * * *",
			NotificationType: NotificationTypeFeed,
		}

		if !schedule.Contains(now) {
			t.Fatal("Current time should be included.", schedule, now)
		}
	})

	t.Run("Time not included", func(t *testing.T) {
		schedule := NotificationSchedule{
			ID:               100001,
			UnitID:           10101,
			Title:            "Example Title",
			Description:      "Example Description",
			Schedule:         "* " + strconv.Itoa(now.Minute()) + " * * * * *",
			NotificationType: NotificationTypeFeed,
		}

		if schedule.Contains(now.Add(time.Minute)) {
			t.Fatal("Current time should not be included.", schedule, now)
		}
	})
}

func TestNotificationImportance(t *testing.T) {
	tests := map[string]NotificationImportance{
		"low":     NotificationImportanceLow,
		"default": NotificationImportanceDefault,
		"high":    NotificationImportanceHigh,
	}

	for name, importance := range tests {
		t.Run(name, func(t *testing.T) {
			if name != importance.String() {
				t.Fatalf("%v string should be %v", importance, name)
			}
		})
	}
}
