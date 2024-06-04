package repo

import (
	rewardsvc "github.com/Buzzvil/buzzapis/go/reward"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
)

type entityBuilder struct {
	m *mapper
}

func (b *entityBuilder) buildToken(pr *rewardsvc.Reward, resource event.Resource, unitID int64) (*event.Token, error) {
	return &event.Token{
		Resource:      resource,
		EventType:     pr.EventType,
		UnitID:        unitID,
		TransactionID: pr.TransactionId,
	}, nil
}

func (b *entityBuilder) buildEvent(pr *rewardsvc.Reward, trackEventURL string, statusCheckURL string) (*event.Event, error) {
	rewardStatus, err := b.m.mapToRewardStatus(pr.Status)
	if err != nil {
		return nil, err
	}

	return &event.Event{
		Type:         pr.EventType,
		TrackingURLs: []string{trackEventURL},
		Reward: &event.Reward{
			Amount:         pr.Amount,
			Status:         rewardStatus,
			StatusCheckURL: statusCheckURL,
			IssueMethod:    pr.IssueMethod,
			TTL:            pr.Ttl,
			Extra:          pr.Extra,
		},
	}, nil
}
