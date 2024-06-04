package payload

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/cypher"
)

// UseCase interface definition
type UseCase interface {
	ParsePayload(payloadString string) (*Payload, error)
	BuildPayloadString(payload *Payload) string
	IsPayloadExpired(payload *Payload) bool
}

type useCase struct {
}

// Day is day length as second
const (
	DAY  int64 = 60 * 60 * 24
	WEEK int64 = DAY * 7
)

// IsPayloadExpired func definition
func (u *useCase) IsPayloadExpired(payload *Payload) bool {
	isExpired := false
	now := time.Now().Unix()

	/*
		"t"에는 할당된 시점, "endedAt"에는 광고가 만료되는 시각이 저장되어있음
		"t"를 이용해서 할당된지 하루가 지난 광고에 대해 expired를 주려 했지만,
		꽤나 자주 expired가 나오고 만료 시간도 몇천시간 이전에 발급되었다고 나오는경우도 빈번함.
		실제 expired를 리턴 할 경우 문제가 발생될 것으로 추정되어 "t"에 대해선 로그만 찍고 있으며, "endedAt"으로 만료확인중
	*/
	if payload.Time+DAY < now {
		core.Logger.Infof("payload expired t %v %v", payload, (now-payload.Time)/60/60)
	}

	if payload.EndedAt < now {
		core.Logger.Infof("payload expired EndedAt %v %v", payload, (now-payload.EndedAt)/60/60)
		if payload.EndedAt+WEEK < now {
			isExpired = true
		}
	}

	return isExpired
}

// ParsePayload func definition
func (u *useCase) ParsePayload(payloadString string) (*Payload, error) {
	var payload *Payload

	if payloadString == "__campaign_payload__" || payloadString == "" {
		return nil, errors.New("invalid payload")
	}

	decrypted, err := cypher.DecryptAesWithBase64([]byte(payloadAESKey), []byte(payloadAESKey), payloadString, payloadURLSafe)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(decrypted, &payload)
	return payload, nil
}

// BuildPayloadString func definition
func (u *useCase) BuildPayloadString(payload *Payload) string {
	return cypher.EncryptAesBase64Dict(payload, payloadAESKey, payloadAESKey, payloadURLSafe)
}

// NewUseCase returns UseCase interface
func NewUseCase() UseCase {
	return &useCase{}
}
