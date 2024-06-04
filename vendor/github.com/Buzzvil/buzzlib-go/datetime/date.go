package datetime

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func GetTime(timeZone string) *time.Time {
	now := time.Now()
	if timeZone != "" {
		now = now.UTC()
		timeLocal, err := time.LoadLocation(timeZone)
		if err != nil {
			log.Error(err)
		}
		now = now.In(timeLocal)
	}
	return &now
}

func GetDate(format, timeZone string) (result string) {
	result = GetTime(timeZone).Format(format)
	return
}

// ref) https://play.golang.org/p/nTcjGZQKAa
func lastDayOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 0, 0, 0, 0, t.Location())
}

func firstDayOfNextYear(t time.Time) time.Time {
	return time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, t.Location())
}

func DaysDiff(a, b time.Time) (days int) {
	cur := b
	for cur.Year() < a.Year() {
		// add 1 to count the last day of the year too.
		days += lastDayOfYear(cur).YearDay() - cur.YearDay() + 1
		cur = firstDayOfNextYear(cur)
	}
	days += a.YearDay() - cur.YearDay()
	if b.AddDate(0, 0, days).After(a) {
		days -= 1
	}
	return days
}
