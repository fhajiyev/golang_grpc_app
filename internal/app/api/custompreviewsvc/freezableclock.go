package custompreviewsvc

import "time"

func newFreezableClock(freezedTime *time.Time) freezableClock {
	return freezableClock{
		freezedTime: freezedTime,
	}
}

type freezableClock struct {
	freezedTime *time.Time
}

func (c freezableClock) now() time.Time {
	if c.freezedTime == nil {
		return time.Now()
	}
	return *c.freezedTime
}
