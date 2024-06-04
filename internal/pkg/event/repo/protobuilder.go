package repo

import (
	rewardsvc "github.com/Buzzvil/buzzapis/go/reward"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
)

type protoBuilder struct {
	m *mapper
}

func (b *protoBuilder) buildProtoResource(token event.Token) (*rewardsvc.Resource, error) {
	resourceType, err := b.m.mapToProtoResourceType(token.Resource.Type)
	if err != nil {
		return nil, err
	}

	return &rewardsvc.Resource{
		Id:   token.Resource.ID,
		Type: resourceType,
	}, nil
}

func (b *protoBuilder) buildCheckRewardStatusRequest(token event.Token) (*rewardsvc.CheckRewardStatusRequest, error) {
	protoResource, err := b.buildProtoResource(token)
	if err != nil {
		return nil, err
	}

	return &rewardsvc.CheckRewardStatusRequest{
		Resource:      protoResource,
		EventType:     token.EventType,
		TransactionId: token.TransactionID,
	}, nil
}

func (b *protoBuilder) buildIssueRewardsRequest(resources []event.Resource) (*rewardsvc.IssueRewardsRequest, error) {
	req := &rewardsvc.IssueRewardsRequest{}
	req.Resources = make([]*rewardsvc.Resource, 0)
	for _, resource := range resources {
		protoResource, err := b.m.mapToProtoResource(resource)
		if err != nil {
			return nil, err
		}

		req.Resources = append(req.Resources, protoResource)
	}

	return req, nil
}

func (b *protoBuilder) buildRewardsForResourceMap(resources []event.Resource, protoMap map[int64]*rewardsvc.IssueRewardsResponse_RewardsForResourceTypeMap) (map[event.Resource]*rewardsvc.Rewards, error) {
	rewardsForResourceMap := make(map[event.Resource]*rewardsvc.Rewards)

	for _, resource := range resources {
		rewardsForResourceTypeMap, ok := protoMap[resource.ID]
		if !ok {
			continue
		}

		resourceType, err := b.m.mapToProtoResourceType(resource.Type)
		if err != nil {
			return nil, err
		}

		rewards, ok := rewardsForResourceTypeMap.RewardsForResourceTypeMap[int32(resourceType)]
		if !ok {
			continue
		}

		rewardsForResourceMap[resource] = rewards
	}

	return rewardsForResourceMap, nil
}
