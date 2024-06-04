package dto

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/pkg/errors"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/session"
)

// BuzzAdCampaignIDOffset const definition
const BuzzAdCampaignIDOffset = 1000000000

// PlaceType type definition
// Deprecated: PlaceType shouldn't be used anymore, instead use unit's type.
type PlaceType int

// Place type definition
const (
	PlaceLockScreen PlaceType = iota
	PlaceRolling
	PlaceFeed
)

type (
	// AdsRequest type definition
	AdsRequest struct {
		AppID          string  `header:"Buzz-App-ID"`
		AdID           string  `query:"ad_id" validate:"required"`
		IFV            *string `query:"ifv"`
		BirthYear      int     `query:"birth_year"`
		Birthday       string  `query:"birthday" targeting:"birthday"`
		Cursor         string  `query:"cursor"`
		CustomTarget1  string  `query:"custom_target_1" targeting:"customTarget1"`
		CustomTarget2  string  `query:"custom_target_2" targeting:"customTarget2"`
		CustomTarget3  string  `query:"custom_target_3" targeting:"customTarget3"`
		DeviceName     string  `query:"device_name" validate:"required"`
		DeviceID       string  `query:"device_id"`
		OsVersion      string  `query:"os_version"` // Android version Rename: osVersion
		IsMockResponse bool    `query:"is_mock_response"`
		Gender         string  `query:"gender" targeting:"gender"`
		Latitude       string  `query:"latitude" targeting:"lat"`
		Longitude      string  `query:"longitude" targeting:"lon"`
		Locale         string  `query:"locale" validate:"required"`
		MccMnc         string  `query:"mcc_mnc"`
		MembershipDays int     `targeting:"membership_days"`
		NetworkType    string  `query:"network_type"`
		Platform       string  `query:"platform"`
		//Deprecated
		PlaceType    PlaceType `query:"place_type" targeting:"placeType"` //0:Lock 1:Rolling
		Relationship string    `query:"relationship" targeting:"relationship"`
		SdkVersion   int       `query:"sdk_version" targeting:"ver"` //20000 이상
		SessionKey   string    `query:"session_key"`
		// Anonymous is true only if the ad request is for anonymous users. SessionKey can be null only if the Anonymous value is true.
		Anonymous         bool   `query:"anonymous"`
		SizeReq           int    `query:"size"`
		TargetFill        int    `query:"target_fill"`
		TimeZone          string `query:"timezone" validate:"required"`
		TypesString       string `query:"types"  validate:"required"` //eg. {"IMAGE":["INTERSTITIAL","FULLSCREEN"],"SDK":["FAN","OUTBRAIN"]}
		LineitemIDsString string `query:"lineitem_ids"`               //eg. [10001, 10002]
		UserAgent         string `query:"user_agent" validate:"required"`
		UnitID            int64  `query:"unit_id"`
		CreativeSize      int    `query:"creative_size"`
		RevenueTypes      string `query:"revenue_types"`
		CPSCategory       string `query:"cps_category"`

		Country       string `targeting:"country"`
		types         map[string][]string
		language      string
		Session       session.Session
		unit          *app.Unit
		dynamoProfile *device.Profile
	}

	// AdsResponse type definition
	AdsResponse struct {
		Ads      Ads                     `json:"ads"`
		Cursor   string                  `json:"cursor"`
		Target   *map[string]interface{} `json:"target"`
		Settings *map[string]interface{} `json:"settings"`
	}

	// Ad type definition
	Ad struct {
		Common

		ActionReward       int                      `json:"action_reward"`
		LandingReward      int                      `json:"landing_reward"`
		Meta               map[string]interface{}   `json:"meta"`
		NetworkName        *string                  `json:"network,omitempty"`
		TTL                *int                     `json:"ttl,omitempty"`
		UnlockReward       int                      `json:"unlock_reward"`
		AdReportData       *string                  `json:"ad_report_data,omitempty"`
		RewardStatus       *reward.ReceivedStatus   `json:"reward_status,omitempty"`
		ConversionCheckURL *string                  `json:"conversion_check_url,omitempty"`
		SkStoreProductURL  *string                  `json:"sk_store_product_url,omitempty"`
		Name               string                   `json:"name"`
		PreferredBrowser   *string                  `json:"preferred_browser,omitempty"`
		Creatives          []map[string]interface{} `json:"creatives,omitempty"`
		Product            map[string]interface{}   `json:"product,omitempty"`

		//Deprecated
		IntegrationType *string `json:"integration_type"`
	}

	// Ads type definition
	Ads []*Ad
)

// GetName returns resource name
func (a *Ad) GetName() *string {
	if a.Creative == nil {
		return &a.Name
	}

	// if creative has title field, use title as name
	title, ok := a.Creative["title"]
	if ok {
		name := title.(string)
		return &name
	}

	return &a.Name
}

// GetRewardSum func definition
func (a *Ad) GetRewardSum() int {
	return a.LandingReward + a.ActionReward + a.UnlockReward
}

// Len func definition
func (a Ads) Len() int {
	return len(a)
}

// Swap func definition
func (a Ads) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Ads) Less(x, y int) bool {
	if a[x].GetRewardSum() != 0 && a[y].GetRewardSum() == 0 {
		return true
	}
	return false
}

// GetTargetFill func definition
func (adsReq *AdsRequest) GetTargetFill() int {
	if adsReq.TargetFill > 0 {
		return adsReq.TargetFill
	}
	return adsReq.SizeReq
}

// GetTypes func definition
func (adsReq *AdsRequest) GetTypes() map[string][]string {
	if adsReq.types != nil {
		return adsReq.types
	}
	if adsReq.TypesString == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(adsReq.TypesString), &adsReq.types); err != nil {
		return nil
	}

	return adsReq.types
}

// Build creates ad from BAAd
func (adsReq *AdsRequest) Build(ctx context.Context, baAd ad.AdV2, rewardStatus reward.ReceivedStatus) *Ad {
	ad := Ad{
		Common: Common{
			ID:                 baAd.ID + BuzzAdCampaignIDOffset,
			ClickTrackers:      baAd.ClickTrackers,
			FailTrackers:       baAd.FailTrackers,
			ImpressionTrackers: baAd.ImpressionTrackers,
			RevenueType:        &baAd.RevenueType,
			Extra:              baAd.Extra,
		},
		Meta:               baAd.Meta,
		NetworkName:        baAd.Network,
		LandingReward:      baAd.LandingReward,
		ActionReward:       baAd.ActionReward,
		UnlockReward:       baAd.UnlockReward,
		TTL:                baAd.TTL,
		AdReportData:       baAd.AdReportData,
		Name:               baAd.Name,
		PreferredBrowser:   baAd.PreferredBrowser,
		RewardStatus:       &rewardStatus,
		ConversionCheckURL: baAd.ConversionCheckURL,
		SkStoreProductURL:  baAd.SkStoreProductURL,
		Product:            utils.TransformMapFromCamelToUnderScore(baAd.Product),
	}

	if val, ok := baAd.Creative["type"]; ok && val != nil {
		creativeType := val.(string)
		ad.IntegrationType = &creativeType
	}

	if val, ok := baAd.Creative["landingType"]; ok && val != nil { //https://sentry.io/buzzvilcom/buzzscreen-api/issues/350326904/#
		landingType := int(val.(float64))
		ad.LandingType = &landingType
	}

	if baAd.CallToAction != "" {
		ad.CallToAction = &baAd.CallToAction
	}

	payload := adsReq.buildPayload(baAd)
	ad.Payload = &payload

	ad.RewardStatus = &rewardStatus
	if rewardStatus == reward.StatusReceived {
		ad.LandingReward = 0
	}

	adsReq.buildCreatives(ctx, &ad, baAd, false, nil)

	return &ad
}

func (adsReq *AdsRequest) buildClickURL(ctx context.Context, ad ad.AdV2, clickURL string, useRewardAPI bool, trackingURLs []string) string {
	clickReq := ClickRequest{
		ClickURL:        clickURL,
		DeviceID:        adsReq.Session.DeviceID,
		ID:              ad.ID + BuzzAdCampaignIDOffset,
		IFA:             adsReq.AdID,
		Name:            ad.Name,
		OrganizationID:  adsReq.Session.AppID,
		OwnerID:         ad.OwnerID,
		Type:            ad.RevenueType,
		Unit:            adsReq.GetUnit(ctx),
		UnitDeviceToken: adsReq.Session.UserID,
	}

	if len(trackingURLs) > 0 {
		clickReq.UseRewardAPI = useRewardAPI
		clickReq.TrackingURL = &(trackingURLs[0])
	}

	return clickReq.BuildClickRedirectURL()
}

func (adsReq *AdsRequest) buildCreatives(ctx context.Context, ad *Ad, baAd ad.AdV2, useRewardAPI bool, trackingURLs []string) {
	ad.Creative = utils.TransformMapFromCamelToUnderScore(baAd.Creative)
	if clickURL, ok := ad.Creative["click_url"]; ok {
		ad.Creative["click_url"] = adsReq.buildClickURL(ctx, baAd, clickURL.(string), useRewardAPI, trackingURLs)
	}

	ad.Creatives = []map[string]interface{}{}
	for _, baCreative := range baAd.Creatives {
		creative := utils.TransformMapFromCamelToUnderScore(baCreative)

		if clickURL, ok := creative["click_url"]; ok {
			creative["click_url"] = adsReq.buildClickURL(ctx, baAd, clickURL.(string), useRewardAPI, trackingURLs)
		}

		ad.Creatives = append(ad.Creatives, creative)
	}
}

func (adsReq *AdsRequest) buildPayload(ad ad.AdV2) string {
	p := &payload.Payload{
		Country: adsReq.GetCountry(),
		EndedAt: time.Now().Unix() + int64(*ad.TTL),
		OrgID:   ad.OrganizationID,
		Time:    time.Now().Unix(),
	}

	if adsReq.Gender != "" {
		p.Gender = &(adsReq.Gender)
	}

	if adsReq.BirthYear > 0 {
		p.YearOfBirth = &(adsReq.BirthYear)
	}

	p.SetUnitID(adsReq.UnitID, adsReq.SdkVersion)

	return buzzscreen.Service.PayloadUseCase.BuildPayloadString(p)
}

// BuildRewardServiceResponse augments ad with reward service response
func (adsReq *AdsRequest) BuildRewardServiceResponse(ctx context.Context, ad *Ad, baAd ad.AdV2, canHandleRewardServiceResponse bool) {
	// Build response for landed event reward
	landedEvent, ok := baAd.Events.LandedEvent()
	if !ok || landedEvent.Reward == nil {
		return
	}

	// SDK will request the reward by calling trackingURLs when the event happened
	if canHandleRewardServiceResponse &&
		landedEvent.Reward.MinimumStayDuration() > 0 &&
		ad.LandingType != nil &&
		model.LandingType(*ad.LandingType) == model.LandingTypeCard {

		r := &Reward{
			Amount:         landedEvent.Reward.Amount,
			Status:         landedEvent.Reward.Status,
			IssueMethod:    landedEvent.Reward.IssueMethod,
			StatusCheckURL: landedEvent.Reward.StatusCheckURL,
			TTL:            landedEvent.Reward.TTL,
			Extra:          landedEvent.Reward.Extra,
		}
		e := Event{
			Reward:       r,
			Type:         landedEvent.Type,
			TrackingURLs: landedEvent.TrackingURLs,
			Extra:        landedEvent.Extra,
		}

		ad.Events = append(ad.Events, e)
		ad.LandingReward = 0
	} else { // SDK will call reward API and trackingURLs will be called in the reward API
		// Overwrite creative/creatives to update click_url
		adsReq.buildCreatives(ctx, ad, baAd, true, landedEvent.TrackingURLs)

		ad.LandingReward = int(landedEvent.Reward.Amount)
	}

	if landedEvent.Reward.Status == event.StatusReceivable {
		rewardStatus := reward.StatusUnknown
		ad.RewardStatus = &rewardStatus
	} else if landedEvent.Reward.Status == event.StatusAlreadyReceived {
		rewardStatus := reward.StatusReceived
		ad.RewardStatus = &rewardStatus
	}

	// TODO: Build response for other reward events
}

// GetUnit func definition
func (adsReq *AdsRequest) GetUnit(ctx context.Context) *app.Unit {
	if adsReq.unit == nil {
		appUseCase := buzzscreen.Service.AppUseCase
		if adsReq.UnitID > 0 {
			adsReq.unit, _ = appUseCase.GetUnitByID(ctx, adsReq.UnitID)
		} else if adsReq.Session.AppID > 0 {
			switch adsReq.PlaceType {
			case PlaceFeed:
				adsReq.unit, _ = appUseCase.GetUnitByAppIDAndType(ctx, adsReq.Session.AppID, app.UnitTypeNative)
			default:
				adsReq.unit, _ = appUseCase.GetUnitByAppID(ctx, adsReq.Session.AppID)
			}
		}
	}
	return adsReq.unit
}

// UnpackSession unpacks Sessionkey and assign to Session
func (adsReq *AdsRequest) UnpackSession() error {
	var session *session.Session

	if adsReq.SessionKey != "" {
		var err error
		session, err = buzzscreen.Service.SessionUseCase.GetSessionFromKey(adsReq.SessionKey)
		if err != nil {
			return err
		}
	} else if adsReq.AppID != "" {
		appID, err := strconv.ParseInt(adsReq.AppID, 10, 64)
		if err != nil {
			return errors.Wrap(err, "appID is not numeric")
		}
		session = buzzscreen.Service.SessionUseCase.NewAnonymousSession(appID)
	} else {
		core.Logger.Infof("empty session_key & app_id. AppID: %+v, UnitID: %+v, DeviceID: %+v, IFA: %+v, UnitDeviceToken: %+v, SDKVersion: %+v, DeviceOS: %+v", adsReq.AppID, adsReq.UnitID, adsReq.DeviceID, adsReq.AdID, adsReq.Session.UserID, adsReq.SdkVersion, adsReq.OsVersion)
		// Deprecated: Anonymous user request without app ID information is deprecated.
		session = buzzscreen.Service.SessionUseCase.NewAnonymousSession(0)
	}

	adsReq.Session = *session
	return nil
}

// GetLanguage func definition
func (adsReq *AdsRequest) GetLanguage() string {
	if adsReq.language == "" {
		adsReq.language, adsReq.Country = utils.SplitLocale(adsReq.Locale)
	}
	return adsReq.language
}

// GetCountry returns country
func (adsReq *AdsRequest) GetCountry() string {
	adsReq.GetLanguage()
	return adsReq.Country
}

// GetDynamoProfile func definition
func (adsReq *AdsRequest) GetDynamoProfile() *device.Profile {
	if adsReq.dynamoProfile == nil {
		var err error
		adsReq.dynamoProfile, err = buzzscreen.Service.DeviceUseCase.GetProfile(adsReq.Session.DeviceID)
		if err != nil {
			core.Logger.Errorf("AdsRequest - Device %d GetProfile error %v", adsReq.Session.DeviceID, err)
		}

	}
	return adsReq.dynamoProfile
}
