package repo

import (
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
)

type mapper struct {
}

func (m *mapper) toEntityDetail(detail adDetail) *ad.Detail {
	return &ad.Detail{
		ID:             detail.ID,
		Name:           detail.ItemName,
		OrganizationID: detail.OrganizationID,
		RevenueType:    detail.RevenueType,
		Extra:          detail.ExtraData,
	}
}
