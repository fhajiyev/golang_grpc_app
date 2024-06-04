package repo

import (
	"strconv"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
)

type mapper struct {
}

// ToEntityPoint func definition
func (m *mapper) ToEntityPoint(p DBPoint) (*reward.Point, error) {
	campaignID, err := strconv.ParseInt(p.ReferKey, 10, 64)
	if err != nil {
		return nil, err
	}

	return &reward.Point{
		DeviceID:   p.DeviceID,
		Version:    p.Version,
		CampaignID: campaignID,
		Type:       p.Type,
		CreatedAt:  p.CreatedAt,
	}, nil
}
