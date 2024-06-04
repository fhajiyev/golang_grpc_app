package trackingdata

import (
	"encoding/json"
	"errors"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/cypher"
)

// UseCase interface definition
type UseCase interface {
	ParseTrackingData(trackingDataString string) (*TrackingData, error)
	BuildTrackingDataString(trackingData *TrackingData) string
}

type useCase struct {
}

// ParseTrackingData func definition
func (u *useCase) ParseTrackingData(trackingDataString string) (*TrackingData, error) {
	var trackingData *TrackingData

	if trackingDataString == "" {
		return nil, errors.New("invalid tracking data")
	}

	decrypted, err := cypher.DecryptAesWithBase64([]byte(trackingDataAESKey), []byte(trackingDataAESKey), trackingDataString, trackingDataURLSafe)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(decrypted, &trackingData)
	return trackingData, nil
}

// BuildTrackingDataString func definition
func (u *useCase) BuildTrackingDataString(trackingData *TrackingData) string {
	return cypher.EncryptAesBase64Dict(trackingData, trackingDataAESKey, trackingDataAESKey, trackingDataURLSafe)
}

// NewUseCase returns UseCase interface
func NewUseCase() UseCase {
	return &useCase{}
}
