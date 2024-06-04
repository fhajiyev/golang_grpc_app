package contentcampaign

import (
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
)

// UseCase type definition
type UseCase interface {
	GetContentCampaignByID(campaignID int64) (*ContentCampaign, error)
	IncreaseClick(campaignID int64, unitID int64) error
	IncreaseImpression(campaignID int64, unitID int64) error
	IsContentCampaignExpired(contentCampaign *ContentCampaign) bool
}

type useCase struct {
	repo Repository
}

// GetContentCampaignByID func definition
func (u *useCase) GetContentCampaignByID(campaignID int64) (*ContentCampaign, error) {
	return u.repo.GetContentCampaignByID(campaignID)
}

// IncreaseClick func definition
func (u *useCase) IncreaseClick(campaignID int64, unitID int64) error {
	return u.repo.IncreaseClick(campaignID, unitID)
}

// IncreaseImpression func definition
func (u *useCase) IncreaseImpression(campaignID int64, unitID int64) error {
	return u.repo.IncreaseImpression(campaignID, unitID)
}

// Period constants
const (
	DAY  int64 = 60 * 60 * 24
	WEEK int64 = DAY * 7
)

// IsContentCampaignExpired func definition
func (u *useCase) IsContentCampaignExpired(contentCampaign *ContentCampaign) bool {
	now := time.Now().Unix()
	endedAt := contentCampaign.EndDate.Unix()

	if endedAt < now {
		core.Logger.Infof("contentCampaign expired %v hours", (now-endedAt)/60/60)
		if endedAt+WEEK < now {
			return true
		}
	}

	return false
}

// NewUseCase func definition
func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}
