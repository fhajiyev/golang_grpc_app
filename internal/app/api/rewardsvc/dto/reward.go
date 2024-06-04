package dto

import (
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
)

// RewardType is type of reward
// In buzzscreen, more types are defined
type RewardType string

const (
	// TypeImpression is reward type "impression"
	TypeImpression RewardType = "imp"
	// BuzzAdCampaignIDOffset is used to split ad & contents from campaign id
	BuzzAdCampaignIDOffset int64 = 1000000000
)

// PostRewardReq is parameters for POST reward API
type PostRewardReq struct {
	AppID                 int64      `form:"app_id" validate:"required"`
	UnitID                int64      `form:"unit_id"`
	DeviceID              int64      `form:"device_id" validate:"required"`
	IFA                   string     `form:"ifa" validate:"required"`
	IFV                   *string    `form:"ifv"`
	UnitDeviceToken       string     `form:"unit_device_token" validate:"required"`
	UnitDeviceTokenClient *string    `form:"client_unit_device_token"`
	Type                  RewardType `form:"type" validate:"required"`

	CampaignID      int64   `form:"campaign_id" validate:"required"`
	CampaignType    string  `form:"campaign_type"`
	CampaignName    string  `form:"campaign_name" validate:"required"`
	CampaignOwnerID *string `form:"campaign_owner_id"`
	CampaignIsMedia int     `form:"campaign_is_media"`
	Slot            int     `form:"slot"`

	Reward     int    `form:"reward"`
	BaseReward int    `form:"base_reward"`
	Checksum   string `form:"check"`
}

// GetUDT returns UnitDeviceTokenClient or UnitDeviceToken
// UnitDeviceTokenClient is set before send request by client and UnitDeviceToken is set by allocation logic
// To address the case that UnitDeviceToken is changed by publisher after allocation, GetUDT returns UnitDeviceTokenClient in first
func (r *PostRewardReq) GetUDT() string {
	if r.UnitDeviceTokenClient == nil || *r.UnitDeviceTokenClient == "__unit_device_token__" || *r.UnitDeviceTokenClient == r.UnitDeviceToken {
		return r.UnitDeviceToken
	} else if strings.Replace(*r.UnitDeviceTokenClient, " ", "+", -1) == r.UnitDeviceToken {
		// IOS에서 잘못 decoding된 unit_device_token에대해 대응하는 로직
		return r.UnitDeviceToken
	}

	// 할당이후, 리워드 적립 사이에 client_unit_device_token이 변경되면 기록
	core.Logger.Warnf("unitDeviceToken is changed after allocation. %s to %s. unitID: %d, deviceID: %d", r.UnitDeviceToken, *r.UnitDeviceTokenClient, r.UnitID, r.DeviceID)
	return *r.UnitDeviceTokenClient
}

// PostRewardRes is response of POST reward API
type PostRewardRes struct {
	Code int `json:"code"`
}
