package session

import (
	"time"

	"github.com/Buzzvil/buzzlib-go/datetime"
)

var keyValue []byte = []byte("BuzzscreenServer")

// Session struct definition
type Session struct {
	AppID          int64
	UserID         string
	DeviceID       int64
	AndroidID      string
	CreatedSeconds int64
}

// GetMembershipDays returns the number of days after session created
func (s *Session) GetMembershipDays() int {
	if s.CreatedSeconds > 0 {
		return datetime.DaysDiff(time.Now(), time.Unix(0, s.CreatedSeconds*int64(time.Second))) + 1
	}
	return 0
}

// IsSignedIn returns true if the session has a user information.
func (s *Session) IsSignedIn() bool {
	return s.UserID != "" && s.DeviceID > 0
}
