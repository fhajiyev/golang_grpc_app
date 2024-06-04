package repo

import (
	"fmt"
	"strings"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"

	rewardsvc "github.com/Buzzvil/buzzapis/go/reward"
)

type mapper struct {
}

func (m *mapper) mapToProtoResource(resource event.Resource) (*rewardsvc.Resource, error) {
	resourceType, err := m.mapToProtoResourceType(resource.Type)
	if err != nil {
		return nil, err
	}

	return &rewardsvc.Resource{
		Id:   resource.ID,
		Type: resourceType,
	}, nil
}

func (m *mapper) mapToResource(protoResource rewardsvc.Resource) (*event.Resource, error) {
	resourceType, err := m.mapToResourceType(protoResource.Type)
	if err != nil {
		return nil, err
	}

	return &event.Resource{
		ID:   protoResource.Id,
		Type: resourceType,
	}, nil
}

func (m *mapper) mapToProtoResourceType(resourceType string) (rewardsvc.ResourceType, error) {
	typeUpper := strings.ToUpper(resourceType)
	protoTypeInt, ok := rewardsvc.ResourceType_value[typeUpper]
	if !ok {
		return 0, fmt.Errorf("unsupported pb.resourceType. resourceType %s", resourceType)
	}
	return rewardsvc.ResourceType(protoTypeInt), nil
}

func (m *mapper) mapToResourceType(protoResourceType rewardsvc.ResourceType) (string, error) {
	typeUpper, ok := rewardsvc.ResourceType_name[int32(protoResourceType)]
	if !ok {
		return "", fmt.Errorf("unsupported rewardsvc.ResourceType. protoResourceType %d", protoResourceType)
	}
	return strings.ToLower(typeUpper), nil
}

func (m *mapper) mapToRewardStatus(protoRewardStatus rewardsvc.RewardStatus) (string, error) {
	statusUpper, ok := rewardsvc.RewardStatus_name[int32(protoRewardStatus)]
	if !ok {
		return "", fmt.Errorf("unsupported rewardsvc.RewardStatus. protoRewardStatus %d", protoRewardStatus)
	}
	return strings.ToLower(statusUpper), nil
}
