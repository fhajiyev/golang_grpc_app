package dto

import (
	"net/http"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
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

	// BuzzAdCampaignIDOffset is used for separate contentCampaign(under 1000000000) and AD(over 1000000000)
	BuzzAdCampaignIDOffset = int64(1000000000)
)

// GetClickRedirectRequest struct definition
type GetClickRedirectRequest struct {
	AppID    *int64 `query:"app_id"`
	UnitID   int64  `query:"unit_id" validate:"required"`
	DeviceID int64  `query:"device_id" validate:"required"`
	IFA      string `query:"ifa" validate:"required"`

	/*
		UnitDeviceToken: 할당시점의 값, (click_redirect url을 서버가 만들어줌)
		UnitDeviceTokenClient: 클릭 시점의 값

		할당과 클릭 사이에 UnitDeviceToken이 변경될경우 변경된 UnitDeviceToken을 사용하기 위해
		추가로 client_unit_device_token을 파라미터로 받음

		네이티브 앱 할당은 앱이 켜져있는 동안 할당을 받기 때문에 UnitDeviceToken이 바뀔 일은 없으나
		락스크린 할당은 UnitDeviceToken이 바뀔 수 있음
	*/
	UnitDeviceToken       string  `query:"unit_device_token"`
	UnitDeviceTokenClient *string `query:"client_unit_device_token"`

	Reward     int `query:"reward"`
	BaseReward int `query:"base_reward"`

	CampaignID         int64   `query:"campaign_id"`
	CampaignType       string  `query:"campaign_type"`
	CampaignName       string  `query:"campaign_name" validate:"required"`
	CampaignOwnerID    string  `query:"campaign_owner_id"`
	CampaignIsMedia    int     `query:"campaign_is_media"` // deprecated
	ExternalCampaignID *string `query:"external_campaign_id"`

	RedirectURL      string  `query:"redirect_url"`
	RedirectURLClean *string `query:"redirect_url_clean"`
	Slot             int     `query:"slot"`
	Position         string  `query:"position"`
	SessionID        string  `query:"session_id"`

	PayloadStr      string `query:"campaign_payload"`
	UseCleanModeStr string `query:"use_clean_mode"`
	UseCleanMode    bool
	IsFalseClickStr string `query:"false_click"`
	IsFalseClick    bool

	TrackingDataStr string `query:"tracking_data"`

	TrackingURL  *string `query:"tracking_url"`
	UseRewardAPI bool    `query:"use_reward_api"`

	Request *http.Request `form:"-" query:"-"`

	Checksum string `query:"check"`
}

// GetUDT returns UnitDeviceTokenClient or UnitDeviceToken
// UnitDeviceTokenClient is set before send request by client and UnitDeviceToken is set by allocation logic
// To address the case that UnitDeviceToken is changed by publisher after allocation, GetUDT returns UnitDeviceTokenClient in first
func (r *GetClickRedirectRequest) GetUDT() string {
	if r.UnitDeviceTokenClient == nil || *r.UnitDeviceTokenClient == "__unit_device_token__" || *r.UnitDeviceTokenClient == r.UnitDeviceToken {
		return r.UnitDeviceToken
	} else if strings.Replace(*r.UnitDeviceTokenClient, " ", "+", -1) == r.UnitDeviceToken {
		// IOS에서 잘못 decoding된 unit_device_token에대해 대응하는 로직
		return r.UnitDeviceToken
	}

	// 할당이후, 리워드 적립 사이에 client_unit_device_token이 변경되면 기록
	core.Logger.Warnf("udt is changed. %s to %s unitID: %d, deviceID: %d", r.UnitDeviceToken, *r.UnitDeviceTokenClient, r.UnitID, r.DeviceID)
	return *r.UnitDeviceTokenClient
}
