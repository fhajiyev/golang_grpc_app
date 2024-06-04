package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
)

// CampaignType type definition
const (
	CampaignTypeCPM    = "I"
	CampaignTypeCPC    = "J"
	CampaignTypeAction = "B"
)

var (
	// LandingTypeResponseMapping type definition
	LandingTypeResponseMapping = map[model.LandingType]string{
		model.LandingTypeBrowser: "browser",
		model.LandingTypeOverlay: "overlay",
		model.LandingTypeCard:    "cardview",
	}
)

// PostAllocationV1 func definition
func PostAllocationV1(c core.Context) error {
	ctx := c.Request().Context()
	// 요청을 파싱
	var allocReq dto.AllocV1Request
	if err := bindValue(c, &allocReq); err != nil {
		return err
	}

	ok, err := buzzscreen.Service.DeviceUseCase.ValidateUnitDeviceToken(allocReq.UnitDeviceToken)
	if !ok {
		core.Logger.Warnf("PostAllocationV1() - failed to validate udt. err: %s", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	unit := allocReq.GetUnit(ctx)
	if unit == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "unit can't be nil"})
	} else if !unit.IsActive {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "unit is inactive"})
	}

	if allocReq.SdkVersion == 0 || allocReq.IFA == "" {
		return &core.HttpError{Message: "sdkVersion and ifa cannot be empty"}
	}

	profile := allocReq.GetDynamoProfile()
	if profile != nil {
		service.UpdateProfileUnitRegisterSeconds(profile, unit.ID, allocReq.GetAppID(ctx), c.Path())
		service.GiveWelcomeReward(ctx, profile, allocReq.UnitDeviceToken, unit.ID, allocReq.GetCountry(ctx))
	}

	res, filteredAdIDs, err := getAllocationRes(ctx, &allocReq)
	if err != nil {
		return err
	}

	if unit.UnitType.IsTypeLockscreen() {
		env.SetRedisDeviceDau(allocReq.DeviceID)
	}
	logAllocationV1(ctx, allocReq, *res, filteredAdIDs)

	return c.JSON(http.StatusOK, res)
}

func logAllocationV1(ctx context.Context, req dto.AllocV1Request, res dto.AllocV1Response, filteredAdIDs []int64) {
	var adIDs, nativeAdIDs, contentIDs []int64

	for _, campaign := range res.Campaigns {
		if campaign.ID > dto.BuzzAdCampaignIDOffset {
			adIDs = append(adIDs, campaign.ID-dto.BuzzAdCampaignIDOffset)
		} else {
			contentIDs = append(contentIDs, campaign.ID)
		}
	}

	for _, nativeAd := range res.NativeCampaigns {
		nativeAdIDs = append(nativeAdIDs, nativeAd.ID-dto.BuzzAdCampaignIDOffset)
	}

	core.Logger.Infof("/api/allocation/ - Request{AppID:%d, UnitID:%d, DeviceID:%d, IFA:%s, UnitDeviceToken:%s, SDKVersion:%d, OSVersion:%d, AppVersionCode:%d} Response{Ads:%v, NativeAds:%v, Contents:%v, FilteredAds: %v}",
		req.GetAppID(ctx),
		req.GetUnit(ctx).BuzzadUnitID,
		req.DeviceID,
		req.IFA,
		req.UnitDeviceToken,
		req.SdkVersion,
		req.DeviceOs,
		req.AppVersionCode,
		adIDs,
		nativeAdIDs,
		contentIDs,
		filteredAdIDs, // TODO export filteredAdIDs as structured log
	)
}

func removeBaseReward(camps []*dto.CampaignV1) {
	if len(camps) > 0 {
		for i := range camps {
			camps[i].BaseReward = 0
		}
	}
}

func getV1ContentWithCampaignIDs(ctx context.Context, allocReq *dto.ContentAllocV1Request, campIDs ...int64) ([]*dto.CampaignV1, error) {
	esCamps, err := service.GetContentCampaignByIDs(allocReq.DeviceID, campIDs...)
	camps := getCampaignsFromESContentCampaigns(ctx, esCamps, allocReq)
	return camps, err
}

func getAllocationResContentOnly(ctx context.Context, allocReq *dto.ContentAllocV1Request, campIDs ...int64) ([]*dto.CampaignV1, error) {
	camps, err := getContent(ctx, allocReq)
	if err != nil {
		return camps, err
	}

	if len(campIDs) > 0 {
		esCamps, err := service.GetContentCampaignByIDs(allocReq.DeviceID, campIDs...)
		if err != nil {
			camps = getCampaignsFromESContentCampaigns(ctx, esCamps, allocReq)
		}
		return camps, err
	}
	return camps, nil
}

// ContentAllocV1Result struct definition
type ContentAllocV1Result struct {
	campaigns []*dto.CampaignV1
	err       error
}

func getAllocationRes(ctx context.Context, allocReq *dto.AllocV1Request, additionalCampIDs ...int64) (*dto.AllocV1Response, []int64, error) {
	cContentCamps := make(chan ContentAllocV1Result)
	go func(allocReq dto.ContentAllocV1Request) {
		camps, err := getContent(ctx, &allocReq)
		//Payload / ClickURL 및 2048 길이 제한등은 안에서 전부 처리해버림
		cContentCamps <- ContentAllocV1Result{
			campaigns: camps,
			err:       err,
		}
	}(allocReq.ContentAllocV1Request)

	baAds, baNativeAds, baSettings := service.GetV1AdFromBuzzAd(ctx, allocReq)
	filteredAdIDs := make([]int64, 0)
	core.Logger.Debugf("getAllocationRes() - baAds: %v, baNativeAds: %v, baSettings: %#v", len(baAds), len(baNativeAds), baSettings)
	baCamps := make([]*dto.CampaignV1, 0)
	for _, ad := range baAds {
		camp := parseAdToCampaign(ctx, ad, allocReq)

		for _, c := range camp.Creatives {
			if clickURL, ok := c["click_url"]; ok {
				switch clickURL.(type) {
				case string:
					clickReq := dto.ClickRequest{
						ClickURL:        clickURL.(string),
						ClickURLClean:   camp.ClickURLClean,
						DeviceID:        allocReq.DeviceID,
						ID:              camp.ID,
						IFA:             allocReq.IFA,
						Name:            camp.Name,
						OrganizationID:  camp.OrganizationID,
						OwnerID:         camp.OwnerID,
						Type:            camp.Type,
						Unit:            allocReq.GetUnit(ctx),
						UnitDeviceToken: allocReq.UnitDeviceToken,
					}
					if ma := allocReq.GetModelArtifact(ctx); ma != nil {
						td := &trackingdata.TrackingData{
							ModelArtifact: *ma,
						}
						clickReq.TrackingData = td
					}
					c["click_url"] = clickReq.BuildClickRedirectURL()
				}
			}
		}

		// url length problems in kitkat (api level 19 - 20)
		if allocReq.DeviceOs <= 20 && len(camp.ClickURL) > 2048 {
			// 안드로이드 OS 20 버전 이하에서, url 길이가 길면 잘리는 이슈
			filteredAdIDs = append(filteredAdIDs, ad.ID)
			continue
		}

		camp.SetPayloadWith(ctx, &(allocReq.ContentAllocV1Request))
		baCamps = append(baCamps, &camp)
	}

	bsResult := <-cContentCamps
	bsCamps, err := bsResult.campaigns, bsResult.err

	if err != nil && len(baCamps) == 0 { // BuzzAd 에서도 아무것도 못가져오고 ES 에도 아무것도 못가져 온경우에는 500 을 Return
		core.Logger.WithError(err).Warnf("getAllocationRes() - err: %v", err)
		return nil, nil, &core.HttpError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	camps := append(baCamps, bsCamps...)
	if len(additionalCampIDs) > 0 {
		esCamps, _ := service.GetContentCampaignByIDs(allocReq.DeviceID, additionalCampIDs...)
		camps = append(camps, getCampaignsFromESContentCampaigns(ctx, esCamps, &allocReq.ContentAllocV1Request)...)
	}

	nativeCamps := make([]*dto.NativeCampaignV1, 0)
	for _, nativeAd := range baNativeAds {
		nativeCamp := parseNativeAdToNativeCampaign(ctx, nativeAd, allocReq)
		nativeCamps = append(nativeCamps, &nativeCamp)
	}

	res := &(dto.AllocV1Response{
		ContentAllocV1Response: dto.ContentAllocV1Response{
			Code:      http.StatusOK,
			Message:   "ok",
			Campaigns: camps,
		},
		NativeCampaigns: nativeCamps,
		Settings:        getAllocationV1Settings(ctx, allocReq, baSettings),
	})

	return res, filteredAdIDs, nil
}

func getAllocationV1Settings(ctx context.Context, allocReq *dto.AllocV1Request, buzzAdSettings *ad.AdsV1Settings) *dto.AllocV1Settings {
	fsr := allocReq.GetUnit(ctx).FirstScreenRatio

	if allocReq.SdkVersion < 1820 && fsr["ad"] > 9 {
		fsr["ad"] = 9
	}
	settings := dto.AllocV1Settings{
		BaseInitPeriod:        allocReq.GetUnit(ctx).BaseInitPeriod,
		RequestTrigger:        600,
		RequestPeriod:         3600,
		HourLimit:             12,
		BaseHourLimit:         1,
		ExternalBaseReward:    0,
		ExternalCampaignID:    1,
		ExternalImpressionCap: 500,
		ExternalAddCallLimit:  500,
		FirstDisplayRatio:     fsr.String(),
		PageLimit:             allocReq.GetUnit(ctx).PageLimit,
	}

	if allocReq.GetAppID(ctx) == app.SjAppID || allocReq.GetAppID(ctx) == appIDCjone {
		settings.RequestTrigger = 300
	}

	shuffleOption := allocReq.GetUnit(ctx).ShuffleOption
	if shuffleOption != nil {
		settings.ShuffleOption = *shuffleOption
	} else {
		settings.ShuffleOption = 0
	}

	if buzzAdSettings != nil {
		if buzzAdSettings.FilteringWords != nil {
			settings.AdFilteringWords = buzzAdSettings.FilteringWords
		}
		if buzzAdSettings.WebUa != nil {
			settings.WebUserAgent = buzzAdSettings.WebUa
		}
	}

	if app.IsHoneyscreenAppID(allocReq.GetUnit(ctx).AppID) && time.Since(time.Unix(allocReq.GetRegisteredSeconds(), 0)).Hours() < 168 { // Honeyscreen new user 들은 Init period 를 줄여준다
		settings.BaseInitPeriod = settings.BaseInitPeriod * 2 / 3
	}

	switch allocReq.SettingsPageDisplayRatio {
	case "1:1", "1:2", "1:3":
		settings.PageDisplayRatio = allocReq.SettingsPageDisplayRatio
	default:
		settings.PageDisplayRatio = allocReq.GetUnit(ctx).PagerRatio.String()
	}

	return &settings
}

func parseNativeAdToNativeCampaign(ctx context.Context, nativeAd *ad.NativeAdV1, allocReq *dto.AllocV1Request) dto.NativeCampaignV1 {
	nativeCampaign := dto.NativeCampaignV1{
		CampaignV1: dto.CampaignV1{
			ActionReward:          0,
			AdReportData:          nativeAd.AdReportData,
			BaseReward:            0,
			ClickURL:              nativeAd.ClickURL,
			CleanMode:             model.CleanModeForceDisabled,
			DisplayType:           "A",
			Extra:                 make(map[string]interface{}),
			EndedAt:               int64(nativeAd.EndedAt),
			FirstDisplayPriority:  10,
			FirstDisplayWeight:    nativeAd.FirstDisplayWeight,
			ID:                    nativeAd.ID + dto.BuzzAdCampaignIDOffset,
			Image:                 nativeAd.Image,
			ImpressionURLs:        nativeAd.ImpressionURLs,
			IsAd:                  true,
			Ipu:                   nativeAd.Ipu,
			LandingReward:         0,
			Meta:                  nativeAd.Meta,
			Name:                  nativeAd.Name,
			OwnerID:               nativeAd.OwnerID,
			RemoveAfterImpression: nativeAd.RemoveAfterImpression,
			Slot:                  nativeAd.Slot,
			SupportWebp:           false,
			StartedAt:             int64(nativeAd.StartedAt),
			TargetApp:             nativeAd.TargetApp,
			Timezone:              nativeAd.Timezone,
			Tipu:                  nativeAd.Tipu,
			Type:                  "A",
			UnlockReward:          0,
			UnitPrice:             int(nativeAd.UnitPrice),
			Creatives:             nativeAd.Creatives,
		},
		NativeAd: nativeAd.NativeAd,
		Banner:   nativeAd.Banner,
	}

	if nativeAd.OrganizationID == allocReq.GetUnit(ctx).OrganizationID {
		nativeCampaign.IsMedia = true
	} else {
		nativeCampaign.IsMedia = false
	}

	if nativeAd.ClickBeacons != nil && len(nativeAd.ClickBeacons) > 0 {
		nativeCampaign.ClickBeacons = nativeAd.ClickBeacons
	}

	if nativeCampaign.EndedAt == 0 {
		core.Logger.Warnf("Allocation() - EndDate is 0, nativeCamp: %+v, nativeAd: %+v", nativeCampaign, nativeAd)
	}

	nativeCampaign.SetPayloadWith(ctx, &(allocReq.ContentAllocV1Request))

	return nativeCampaign
}

func parseAdToCampaign(ctx context.Context, ad *ad.AdV1, allocReq *dto.AllocV1Request) dto.CampaignV1 {
	campaign := dto.CampaignV1{
		AdChoiceURL:           ad.AdChoiceURL,
		AdNetworkID:           ad.AdNetworkID,
		AdReportData:          ad.AdReportData,
		ClickURL:              ad.ClickURL,
		DisplayType:           ad.DisplayType,
		Extra:                 ad.Extra,
		EndedAt:               int64(ad.EndedAt),
		FailTrackers:          ad.FailTrackers,
		FirstDisplayPriority:  ad.FirstDisplayPriority,
		FirstDisplayWeight:    ad.FirstDisplayWeight,
		ID:                    ad.ID + dto.BuzzAdCampaignIDOffset,
		Image:                 ad.Image,
		ImpressionURLs:        ad.ImpressionURLs,
		Ipu:                   getIntDefault(ad.Ipu, 9999),
		IsAd:                  true,
		LandingType:           ad.LandingType,
		Meta:                  ad.Meta,
		Name:                  ad.Name,
		OrganizationID:        ad.OrganizationID,
		OwnerID:               ad.OwnerID,
		PreferredBrowser:      ad.PreferredBrowser,
		RemoveAfterImpression: ad.RemoveAfterImpression,
		Slot:                  ad.Slot,
		StartedAt:             int64(ad.StartedAt),
		SupportWebp:           ad.SupportWebp,
		TargetApp:             ad.TargetApp,
		Timezone:              ad.Timezone,
		Tipu:                  ad.Tipu,
		Type:                  ad.Type,
		UnitPrice:             int(ad.UnitPrice),
		Creatives:             ad.Creatives,

		UnlockReward:  ad.UnlockReward,
		ActionReward:  ad.ActionReward,
		LandingReward: ad.LandingReward,
	}

	if ad.Creative != nil {
		campaign.Creative = ad.Creative
	}

	if ad.WebHTML != nil {
		campaign.WebHTML = ad.WebHTML
	}

	if ad.ClickBeacons != nil && len(ad.ClickBeacons) > 0 {
		campaign.ClickBeacons = ad.ClickBeacons
	}

	if ad.WebHTML != nil || allocReq.GetCountry(ctx) == "US" || allocReq.SdkVersion <= 1240 {
		campaign.BaseReward = 0
	} else {
		campaign.BaseReward = allocReq.GetUnit(ctx).BaseReward
	}

	if ad.Type == "cpy" && campaign.Creative != nil && campaign.Creative["type"] == "VIDEO" {
		campaign.BaseReward = 0
	}

	if ad.LandingType != LandingTypeResponseMapping[model.LandingTypeBrowser] && ad.Type == "cpy" {
		campaign.LandingType = LandingTypeResponseMapping[model.LandingTypeOverlay]
	} else if ad.LandingType != LandingTypeResponseMapping[model.LandingTypeCard] {
		// cardview로 설정이 안된 경우는 안전하게 모두 browser 타입으로 변환해준다.
		campaign.LandingType = LandingTypeResponseMapping[model.LandingTypeBrowser]
	}

	if campaign.Creative != nil {
		campaign.Creative["landing_type"] = campaign.LandingType
	}

	// https://buzzvil.atlassian.net/browse/BA-1273
	// https://buzzvil.atlassian.net/browse/BS-1517
	// TODO: 추후 1947 버전 이하가 일정 수준 이하로 떨어지면 해당 유저들에 대해서
	// creative에 short_action_description 대신 call_to_action이
	// 보이게 되는것을 감수하고 아래 코드 제거 가능
	if campaign.Creative != nil && campaign.Creative["type"] == "VIDEO" && allocReq.SdkVersion < 1947 {
		campaign.Creative["call_to_action"] = campaign.Creative["short_action_description"]
	}

	switch ad.Type {
	case "cpc":
		campaign.Type = CampaignTypeCPC
	case "cpm":
		campaign.Type = CampaignTypeCPM
	default:
		campaign.Type = CampaignTypeAction
	}

	if ad.OrganizationID == allocReq.GetUnit(ctx).OrganizationID {
		campaign.IsMedia = true
	} else {
		campaign.IsMedia = false
	}

	/* clickurl생성 로직
	1. 공통으로 들어가는 click request 생성
	2. landed event가 있으면 click request에 추가함
	3. campaign.ClickURL과 campaign.Creative["click_url"]에 각각 click_url 생성
	*/
	clickReq := dto.ClickRequest{
		ClickURLClean:   campaign.ClickURLClean,
		DeviceID:        allocReq.DeviceID,
		ID:              campaign.ID,
		IFA:             allocReq.IFA,
		Name:            campaign.Name,
		OrganizationID:  campaign.OrganizationID,
		OwnerID:         campaign.OwnerID,
		Type:            campaign.Type,
		Unit:            allocReq.GetUnit(ctx),
		UnitDeviceToken: allocReq.UnitDeviceToken,
	}
	if ma := allocReq.GetModelArtifact(ctx); ma != nil {
		td := &trackingdata.TrackingData{
			ModelArtifact: *ma,
		}
		clickReq.TrackingData = td
	}

	// 리워드 서비스에서 리워드를 발급 받은 경우, 하위 호환을 위해 리워드 정보를 기존에 사용하던 필드로 옮긴다.
	landedEvent, ok := ad.Events.LandedEvent()
	if ok && landedEvent.Reward != nil {
		campaign.LandingReward = int(landedEvent.Reward.Amount)

		if len(landedEvent.TrackingURLs) > 0 {
			// TODO 여러개 tracking url이 생긴다면 문제 될 수 있음
			clickReq.TrackingURL = &landedEvent.TrackingURLs[0]
		}

		minimumStayDuration := landedEvent.Reward.MinimumStayDuration()
		if allocReq.SdkVersion >= 1980 && minimumStayDuration > 0 {
			// 광고 설정시 landingType을 CardView로 설정해야만 체류 시간을 설정할 수 있도록 되어있지만
			// 만약의 경우(e.g. DB에서 직접 수정)를 대비하여 체류 시간이 설정되어 있으면 항상 CardView로 열리도록 함
			if campaign.LandingType != LandingTypeResponseMapping[model.LandingTypeCard] {
				core.Logger.Errorf("AllocationV1() - landing_type must be set to CardView for stay over reward. lineitem_id: %d", campaign.ID)
				campaign.LandingType = LandingTypeResponseMapping[model.LandingTypeCard]
			}

			campaign.Meta["landing_page_duration"] = minimumStayDuration

			clickReq.UseRewardAPI = true
		} else {
			if campaign.LandingType != LandingTypeResponseMapping[model.LandingTypeBrowser] {
				core.Logger.Errorf("AllocationV1() - landing_type must be set to Browser for click reward. lineitem_id: %d", campaign.ID)
				campaign.LandingType = LandingTypeResponseMapping[model.LandingTypeBrowser]
			}

			clickReq.UseRewardAPI = false
		}
	}

	if campaign.ClickURL != "" {
		clickReq.ClickURL = campaign.ClickURL
		campaign.ClickURL = clickReq.BuildClickRedirectURL()
	}

	clickURL, exist := campaign.Creative["click_url"]
	clickURLString, ok := clickURL.(string)
	if exist && ok {
		clickReq.ClickURL = clickURLString
		campaign.Creative["click_url"] = clickReq.BuildClickRedirectURL()
	}

	return campaign
}
