package dto

import (
	"net/http"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
)

// CampaignType Definition
const (
	CampaignTypeCPM    = "I"
	CampaignTypeCPC    = "J"
	CampaignTypeAction = "B"
)

const (
	// PlaceholderCampaignID is default campaign ID
	PlaceholderCampaignID = int64(1)
)

// PostUnlockRequest is parameters for POST unlock API
type PostUnlockRequest struct {
	AppID           *int64 `form:"app_id"`
	UnitID          *int64 `form:"unit_id"` // New SDK will not send unit_id parameter
	DeviceID        int64  `form:"device_id" validate:"required"`
	IFA             string `form:"ifa" validate:"required"`
	UnitDeviceToken string `form:"unit_device_token" validate:"required"`

	Reward     int `form:"reward"`
	BaseReward int `form:"base_reward"`

	CampaignID      int64   `form:"click_campaign_id"`
	CampaignType    string  `form:"click_campaign_type"`
	CampaignName    string  `form:"click_campaign_name"`
	CampaignOwnerID *string `form:"click_campaign_owner_id"`
	CampaignIsMedia int     `form:"click_campaign_is_media"` // deprecated
	ClickType       string  `form:"click_type"`

	ExternalCampaignID      *string `form:"external_click_campaign_id"`
	ExternalCampaignUnlocks *string `form:"external_click_campaign_imps"`

	Slot       int    `form:"slot"`
	PayloadStr string `form:"click_campaign_payload"`
	Checksum   string `form:"check"`

	Request *http.Request `form:"-"`
	Payload *payload.Payload
}
