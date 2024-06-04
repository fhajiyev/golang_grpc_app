package event

import (
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzlib-go/jwe"
)

// UseCase defines interface for event
type UseCase interface {
	TrackEvent(handler MessageHandler, token Token) error
	GetTokenEncrypter() TokenEncrypter
	GetRewardStatus(token Token, auth header.Auth) (string, error)
	GetEventsMap(resources []Resource, unitID int64, a header.Auth) (map[int64]Events, error)
	SaveTrackingURL(deviceID int64, resource Resource, trackingURL string)
	GetTrackingURL(deviceID int64, resource Resource) (string, error)
}

type useCase struct {
	repo           Repository
	tokenEncrypter TokenEncrypter
	logger         StructuredLogger
}

// TrackEvent func definition
func (u *useCase) TrackEvent(handler MessageHandler, token Token) error {
	return handler.Publish(token)
}

// GetTokenEncrypter return TokenEncrypter
func (u *useCase) GetTokenEncrypter() TokenEncrypter {
	return u.tokenEncrypter
}

// GetRewardStatus func definition
func (u *useCase) GetRewardStatus(token Token, auth header.Auth) (string, error) {
	return u.repo.GetRewardStatus(token, auth)
}

// GetEventsMap func definition
func (u *useCase) GetEventsMap(resources []Resource, unitID int64, a header.Auth) (map[int64]Events, error) {
	return u.repo.GetEventsMap(resources, unitID, a, u.tokenEncrypter)
}

// SaveTrackURL saves trackingURL
func (u *useCase) SaveTrackingURL(deviceID int64, resource Resource, trackingURL string) {
	u.repo.SaveTrackingURL(deviceID, resource, trackingURL)
	u.logTrackingURLActivity(logTrackURLActivityMethodSave, deviceID, resource, trackingURL)
}

// GetTrackingURL retrieves and deletes trackingURL
func (u *useCase) GetTrackingURL(deviceID int64, resource Resource) (string, error) {
	trackingURL, err := u.repo.GetTrackingURL(deviceID, resource)
	u.logTrackingURLActivity(logTrackURLActivityMethodGet, deviceID, resource, trackingURL)
	if err != nil {
		return "", err
	}

	if trackingURL != "" {
		err = u.repo.DeleteTrackingURL(deviceID, resource)
		if err != nil {
			core.Logger.Warnf("Failed to delete tracking url. err: %v", err)
		}
	}
	return trackingURL, nil
}

func (u *useCase) logTrackingURLActivity(method string, deviceID int64, resource Resource, trackingURL string) {
	m := map[string]interface{}{
		"type":          "tracking_url_activity",
		"method":        method,
		"device_id":     deviceID,
		"resource_id":   resource.ID,
		"resource_type": resource.Type,
		"tracking_url":  trackingURL,
		"event_at":      time.Now().Unix(),
	}

	u.logger.Log(m)
}

// NewUseCase returns UseCase
func NewUseCase(repo Repository, manager jwe.Manager, logger StructuredLogger) UseCase {
	return &useCase{
		repo: repo,
		tokenEncrypter: &tokenEncrypter{
			manager: manager,
		},
		logger: logger,
	}
}
