package dto

import (
	"context"
	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
)

type (
	// CampaignV1 type definition
	CampaignV1 struct {
		ActionReward          int                      `json:"action_reward"`
		AdChoiceURL           string                   `json:"adchoice_url"`
		BaseReward            int                      `json:"base_reward"`
		Category              string                   `json:"category"`
		ChannelID             *int64                   `json:"channel_id,omitempty"`
		Channel               *ESContentChannel        `json:"channel"`
		ClickURLClean         string                   `json:"-"`
		CleanMode             int                      `json:"clean_mode"`
		ClickURL              string                   `json:"click_url"`
		ClickBeacons          []string                 `json:"click_beacons,omitempty"`
		DisplayType           string                   `json:"display_type"`
		EndedAt               int64                    `json:"ended_at"`
		Extra                 map[string]interface{}   `json:"extra"` // 구버전에서 빈 값이라도 보내줘야함
		FailTrackers          []string                 `json:"fail_trackers,omitempty"`
		FirstDisplayPriority  float64                  `json:"first_display_priority"`
		FirstDisplayWeight    int                      `json:"first_display_weight"`
		ID                    int64                    `json:"id"`
		Image                 string                   `json:"image"` // TODO: Optional (sdk_version >= 1440)
		ImpressionURLs        []string                 `json:"impression_urls,omitempty"`
		Ipu                   int                      `json:"ipu"`
		IsAd                  bool                     `json:"is_ad"`
		IsMedia               bool                     `json:"is_media"`
		LandingReward         int                      `json:"landing_reward"`
		LandingType           string                   `json:"landing_type"`
		Meta                  map[string]interface{}   `json:"meta"`
		Name                  string                   `json:"name"`
		OrganizationID        int64                    `json:"-"`
		OwnerID               int64                    `json:"owner_id"`
		Payload               string                   `json:"payload"`
		ProviderID            int64                    `json:"-"`
		PreferredBrowser      *string                  `json:"preferred_browser,omitempty"`
		RemoveAfterImpression bool                     `json:"remove_after_impression"`
		Slot                  string                   `json:"slot"`
		StartedAt             int64                    `json:"started_at"`
		SourceURL             string                   `json:"source_url"`
		SupportWebp           bool                     `json:"support_webp"`
		TargetApp             string                   `json:"target_app"`
		Timezone              string                   `json:"timezone"`
		Tipu                  int                      `json:"tipu"`
		Type                  string                   `json:"type"`
		UnitPrice             int                      `json:"unit_price"` // TODO: Deprecate (sdk_version >= 1440)
		UnlockReward          int                      `json:"unlock_reward"`
		WebHTML               *map[string]interface{}  `json:"web_html,omitempty"`
		AdNetworkID           int                      `json:"-"`
		Creative              map[string]interface{}   `json:"creative"`
		AdReportData          string                   `json:"ad_report_data"`
		Creatives             []map[string]interface{} `json:"creatives,omitempty"`
	}

	// NativeCampaignV1 type definition
	NativeCampaignV1 struct {
		CampaignV1
		NativeAd *ad.NativeAdV1Settings `json:"native_ad,omitempty"`
		Banner   *ad.NativeAdV1Settings `json:"banner,omitempty"`
	}
)

// ContentByCtr type definition
type ContentByCtr []*ESContentCampaign

// Len func definition
func (s ContentByCtr) Len() int {
	return len(s)
}

// Swap func definition
func (s ContentByCtr) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ContentByCtr) Less(x, y int) bool {
	xClicks, yClicks := s[x].Clicks, s[y].Clicks
	xImps, yImps := s[x].Impressions+1, s[y].Impressions+1

	if xImps == 1 {
		xImps = 100000000
	}

	if yImps == 1 {
		yImps = 100000000
	}

	xCtr, yCtr := float64(xClicks)/float64(xImps), float64(yClicks)/float64(yImps)

	return xCtr < yCtr
}

// SetPayloadWith func definition
func (camp *CampaignV1) SetPayloadWith(ctx context.Context, allocReq *ContentAllocV1Request) {
	p := &payload.Payload{
		Country:  allocReq.GetCountry(ctx),
		EndedAt:  camp.EndedAt,
		OrgID:    camp.OrganizationID,
		Time:     time.Now().Unix(),
		Timezone: camp.Timezone,
	}

	if allocReq.Gender != "" {
		p.Gender = &allocReq.Gender
	}

	yob := allocReq.GetYearOfBirth()
	if yob > 0 {
		p.YearOfBirth = &yob
	}

	p.SetUnitID(allocReq.UnitIDReq, allocReq.SdkVersion)

	camp.Payload = buzzscreen.Service.PayloadUseCase.BuildPayloadString(p)
}
