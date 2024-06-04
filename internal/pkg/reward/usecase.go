package reward

import (
	"fmt"
	"time"
)

// UseCase reward
type UseCase interface {
	ValidateRequest(ingredients RequestIngredients) error
	GiveReward(ingredients RequestIngredients) (int, error)
	GetReceivedStatusMap(deviceID int64, periodForCampaign PeriodForCampaign) map[int64]ReceivedStatus
}

type useCase struct {
	repo      Repository
	validator *validator
}

// GetReceivedStatusMap func definition
func (u *useCase) GetReceivedStatusMap(deviceID int64, periodForCampaign PeriodForCampaign) map[int64]ReceivedStatus {
	now := time.Now().Unix()
	maxPeriod := periodForCampaign.MaxPeriod()
	points := u.repo.GetImpressionPoints(deviceID, maxPeriod)

	receivedStatusMap := make(map[int64]ReceivedStatus)
	for _, point := range points {
		// dynamoDB에서 가져온 campaignID가 요청값에 없을경우 검사 안함
		period, ok := periodForCampaign[point.CampaignID]
		if !ok {
			continue
		}

		// 이전 iteration에서 이미 received 표시가 되어있을경우 검사 안함
		receivedStatus, ok := receivedStatusMap[point.CampaignID]
		if ok && receivedStatus == StatusReceived {
			continue
		}

		// period 내에 받은 내역이 있으면 received
		if now-int64(period) <= point.CreatedAt {
			receivedStatusMap[point.CampaignID] = StatusReceived
		}
	}

	// received로 되어있지 않은 모든 campaign에 대해 unknown으로 설정
	for campaignID := range periodForCampaign {
		_, ok := receivedStatusMap[campaignID]
		if !ok {
			receivedStatusMap[campaignID] = StatusUnknown
		}
	}

	return receivedStatusMap
}

// Checksum 검사
func (u *useCase) ValidateRequest(is RequestIngredients) error {
	if is.CampaignName == "" {
		return NewBadRequestError(fmt.Sprintf("campaign_name is empty. did: %v, cid: %v", is.DeviceID, is.CampaignID))
	}
	if !u.validator.validateChecksum(is) {
		return NewBadRequestError("checksum is invalid")
	}

	return nil
}

// Reward 지급 요청
func (u *useCase) GiveReward(ingredients RequestIngredients) (int, error) {
	if ingredients.Reward == 0 {
		return 0, nil
	}

	return u.repo.Save(ingredients)
}

// NewUseCase reward 생성
func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo, validator: &validator{}}
}
