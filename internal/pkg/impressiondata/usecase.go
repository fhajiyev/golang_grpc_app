package impressiondata

import (
	"encoding/json"
	"errors"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/cypher"
)

// UseCase interface definition
type UseCase interface {
	ParseImpressionData(impressionDataString string) (*ImpressionData, error)
	BuildImpressionDataString(impressionData ImpressionData) string
}

type useCase struct {
}

// ParseImpressionData func definition
func (u *useCase) ParseImpressionData(impressionDataString string) (*ImpressionData, error) {
	var impressionData ImpressionData

	if impressionDataString == "" {
		return nil, errors.New("invalid tracking data")
	}

	decrypted, err := cypher.DecryptAesWithBase64([]byte(impressionDataAESKey), []byte(impressionDataAESKey), impressionDataString, impressionDataURLSafe)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(decrypted, &impressionData)
	return &impressionData, nil
}

// BuildImpressionDataString func definition
func (u *useCase) BuildImpressionDataString(impressionData ImpressionData) string {
	return cypher.EncryptAesBase64Dict(&impressionData, impressionDataAESKey, impressionDataAESKey, impressionDataURLSafe)
}

// NewUseCase returns UseCase interface
func NewUseCase() UseCase {
	return &useCase{}
}
