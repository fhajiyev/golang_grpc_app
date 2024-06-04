package event

import (
	"github.com/Buzzvil/buzzlib-go/header"
)

// Repository defines interface to communitcate with rewardsvc
type Repository interface {
	GetEventsMap(resources []Resource, unitID int64, a header.Auth, tokenEncrypter TokenEncrypter) (map[int64]Events, error)
	GetRewardStatus(token Token, auth header.Auth) (string, error)
	SaveTrackingURL(deviceID int64, resource Resource, trackURL string)
	GetTrackingURL(deviceID int64, resource Resource) (string, error)
	DeleteTrackingURL(deviceID int64, resource Resource) error
}
