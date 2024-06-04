package service

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
)

//AdBuilder should be passed when getting AdV2s from BuzzAd
type AdBuilder interface {
	Build(ctx context.Context, baAd ad.AdV2, rewardStatus reward.ReceivedStatus) *dto.Ad
	BuildRewardServiceResponse(ctx context.Context, ad *dto.Ad, baAd ad.AdV2, canHandleRewardServiceResponse bool)
}

//GetV1AdFromBuzzAd fetches ad list from buzzad server and returns Ad, NativeAd & Settings.
func GetV1AdFromBuzzAd(ctx context.Context, allocReq *dto.AllocV1Request) ([]*ad.AdV1, []*ad.NativeAdV1, *ad.AdsV1Settings) {
	appUseCase := buzzscreen.Service.AppUseCase
	if allocReq.SdkVersion < 1501 {
		core.Logger.Debugf("GetV1AdFromBuzzAd() - skip to get ads. SDKVersion: %v", allocReq.SdkVersion)
		return nil, nil, &ad.AdsV1Settings{}
	}

	switch allocReq.GetUnit(ctx).AdType {
	case app.AdTypeNone:
		core.Logger.Debugf("GetV1AdFromBuzzAd() - skip to get ads. AdType: %v", allocReq.GetUnit(ctx).AdType)
		return nil, nil, &ad.AdsV1Settings{}
	case app.AdTypeOldVerOnly:
		ap, err := appUseCase.GetAppByID(ctx, allocReq.GetAppID(ctx))
		if err != nil || ap == nil {
			core.Logger.Errorf("GetV1AdFromBuzzAd() - skip to get ads. err: %v, app_id: %v", err, allocReq.GetAppID(ctx))
			return nil, nil, &ad.AdsV1Settings{}
		}
		if ap.LatestAppVersion == nil {
			core.Logger.Warnf("GetV1AdFromBuzzAd() - skip to get ads. App's LatestAppVersion is not set.")
			return nil, nil, &ad.AdsV1Settings{}
		}
		if allocReq.AppVersionCode >= *ap.LatestAppVersion && allocReq.AppVersionCode != 0 { // 0 이면 완전 구버전
			core.Logger.Debugf("GetV1AdFromBuzzAd() - skip to get ads. AppVersionCode(%v) >= LatestAppVersion(%v)", allocReq.AppVersionCode, *ap.LatestAppVersion)
			return nil, nil, &ad.AdsV1Settings{}
		}
	}

	v1Req := ad.V1AdsRequest{
		AccountID:                    allocReq.DeviceID,
		Carrier:                      allocReq.Carrier,
		ClientIP:                     allocReq.GetClientIP(),
		CustomTarget1:                allocReq.CustomTarget1,
		CustomTarget2:                allocReq.CustomTarget2,
		CustomTarget3:                allocReq.CustomTarget3,
		DeviceOS:                     allocReq.DeviceOs,
		DeviceName:                   allocReq.DeviceName,
		Region:                       allocReq.Region,
		IFA:                          allocReq.IFA,
		IsAllocationTest:             allocReq.IsAllocationTest,
		MembershipDays:               utils.GetDaysFrom(allocReq.GetRegisteredSeconds()) + 1,
		NetworkType:                  allocReq.NetworkType,
		PackageName:                  allocReq.Package,
		UnitID:                       allocReq.GetUnit(ctx).BuzzadUnitID,
		UserAgent:                    allocReq.GetUserAgent(),
		UnitDeviceToken:              allocReq.UnitDeviceToken,
		SDKVersion:                   allocReq.SdkVersion,
		Sex:                          allocReq.Gender,
		MCC:                          allocReq.Mcc,
		MNC:                          allocReq.Mnc,
		TimeZone:                     allocReq.TimeZone,
		RevenueTypes:                 `["-cpi"]`,
		CreativeSize:                 allocReq.CreativeSize,
		SupportRemoveAfterImpression: 1, // SDKVersion 1500이상(1501미만 할당하지 않으므로 언제나)일때 true
	}

	if allocReq.GetYearOfBirth() > 0 {
		yob := allocReq.GetYearOfBirth()
		v1Req.YearOfBirth = &yob
	}

	if allocReq.IsTest {
		isTest := 1
		v1Req.IsTest = &isTest
	}

	if allocReq.IsIFALimitAdTrackingEnabled {
		isIFALimitAdTrackingEnabled := 1
		v1Req.IsIFALimitAdTrackingEnabled = &isIFALimitAdTrackingEnabled
	}

	if allocReq.InstalledBrowsers != "" {
		v1Req.InstalledBrowsers = &allocReq.InstalledBrowsers
	}

	if allocReq.DefaultBrowser != "" {
		v1Req.DefaultBrowser = &allocReq.DefaultBrowser
	}

	adsResponse, err := buzzscreen.Service.AdUseCase.GetAdsV1(v1Req)
	if err != nil {
		core.Logger.Warnf("error occured on GetAdsV1. %v", err)
		return nil, nil, &ad.AdsV1Settings{}
	}

	buzzscreen.Service.AdUseCase.LogAdAllocationRequestV1(allocReq.AppIDReq, v1Req)

	return adsResponse.Ads, adsResponse.NativeAds, adsResponse.Settings
}

// GetAdsByStatus fetches ad list from buzzad server and remove reward from buzzad ads using campaign statuses.
func GetAdsByStatus(ctx context.Context, adsReq *dto.AdsRequest, loc *location.Location, adBuilder AdBuilder, deviceID int64, unitID int64, fill int, auth *header.Auth) (*dto.AdsResponse, error) {
	u := adsReq.GetUnit(ctx)
	// https://buzzvil.atlassian.net/browse/BZZRWRDD-369
	// TODO replace u.IsIOS()
	if u.Platform == app.PlatformIOS && u.Country != "" && u.Country != loc.Country {
		return &dto.AdsResponse{
			Target:   getTarget(adsReq),
			Settings: getSettings(nil, u),
		}, nil
	}

	v2Req := getV2Request(adsReq, loc, fill, adsReq.GetUnit(ctx).BuzzadUnitID)

	buzzAdResponse, err := buzzscreen.Service.AdUseCase.GetAdsV2(v2Req)
	if err != nil {
		return nil, ad.RemoteError{Err: err}
	}

	ads := transformToAds(ctx, buzzAdResponse.Ads, adBuilder, deviceID, canHandleRewardServiceResponse(adsReq.GetUnit(ctx), adsReq.SdkVersion))
	ads = filterOutNotAllowedAds(*ads, adsReq.GetUnit(ctx), adsReq.SdkVersion, adsReq.OsVersion)

	buzzscreen.Service.AdUseCase.LogAdAllocationRequestV2(adsReq.Session.AppID, v2Req)

	return &dto.AdsResponse{
		Ads:      *ads,
		Cursor:   buzzAdResponse.Cursor,
		Target:   getTarget(adsReq),
		Settings: getSettings(buzzAdResponse.Settings, u),
	}, nil
}

func getV2Request(adsReq *dto.AdsRequest, loc *location.Location, fill int, unitID int64) ad.V2AdsRequest {
	v2Req := ad.V2AdsRequest{
		AccountID:      adsReq.Session.DeviceID,
		ClientIP:       loc.IPAddress,
		Country:        loc.Country,
		Cursor:         adsReq.Cursor,
		CustomTarget1:  adsReq.CustomTarget1,
		CustomTarget2:  adsReq.CustomTarget2,
		CustomTarget3:  adsReq.CustomTarget3,
		DeviceName:     adsReq.DeviceName,
		DeviceID:       adsReq.DeviceID,
		Gender:         adsReq.Gender,
		IFA:            adsReq.AdID,
		IsMockResponse: adsReq.IsMockResponse,
		Language:       adsReq.GetLanguage(),
		MembershipDays: adsReq.Session.GetMembershipDays(),
		NetworkType:    adsReq.NetworkType,
		OsVersion:      adsReq.OsVersion,
		Platform:       "A",
		Relationship:   adsReq.Relationship,
		SdkVersion:     adsReq.SdkVersion,
		Timezone:       adsReq.TimeZone,
		Types:          adsReq.TypesString,
		LineitemIds:    adsReq.LineitemIDsString,
		UserAgent:      adsReq.UserAgent,
		UnitID:         unitID,
		CreativeSize:   adsReq.CreativeSize,
	}

	if fill > 0 {
		v2Req.TargetFill = fill
	}

	if adsReq.Session.UserID != "" {
		v2Req.UnitDeviceToken = &adsReq.Session.UserID
	}

	if adsReq.Session.AndroidID != "" {
		v2Req.AndroidID = &adsReq.Session.AndroidID
	}

	if adsReq.Birthday != "" && adsReq.Birthday != "null" {
		v2Req.Birthday = &adsReq.Birthday
	} else if adsReq.BirthYear > 0 {
		birthday := fmt.Sprintf("%v-12-31", adsReq.BirthYear)
		v2Req.Birthday = &birthday
	}

	if adsReq.MccMnc != "" && len(adsReq.MccMnc) > 3 {
		mcc := adsReq.MccMnc[:3]
		v2Req.MCC = &mcc

		mnc := adsReq.MccMnc[3:len(adsReq.MccMnc)]
		v2Req.MNC = &mnc
	}

	if adsReq.Latitude != "" && adsReq.Longitude != "" {
		v2Req.Latitude = &adsReq.Latitude
		v2Req.Longitude = &adsReq.Longitude
	}

	if adsReq.RevenueTypes != "" {
		v2Req.RevenueTypes = &adsReq.RevenueTypes
	}

	if adsReq.CPSCategory != "" {
		v2Req.CPSCategory = &adsReq.CPSCategory
	}

	return v2Req
}

func getSettings(adsSettings *ad.AdsV2Settings, unit *app.Unit) *map[string]interface{} {
	settings := map[string]interface{}{
		"ad_request_interval":      0,
		"content_request_interval": 600,
		"first_screen_ratio":       unit.FirstScreenRatio.String(),
		"pager_ratio":              unit.PagerRatio.String(),
		"feed_ratio":               unit.FeedRatio.String(),
		"reload_after_interaction": true,
	}
	if val, _ := settings["first_screen_ratio"]; val == "9:1" {
		settings["first_screen_ratio"] = "1:0"
	}

	if adsSettings != nil {
		for k, v := range *adsSettings {
			settings[utils.CamelToUnderscore(k)] = v
		}
	}

	return &settings
}

func transformToAds(ctx context.Context, baAds []ad.AdV2, builder AdBuilder, deviceID int64, canHandleRewardServiceResponse bool) *dto.Ads {
	var receivedStatusMap map[int64]reward.ReceivedStatus
	if deviceID > 0 { // anonymous user는 리워드 적립 기록을 확인하지 않음
		periodForAd := make(reward.PeriodForCampaign)
		for _, buzzAd := range baAds {
			periodForAd[buzzAd.ID+dto.BuzzAdCampaignIDOffset] = reward.Period(buzzAd.RewardPeriod)
		}
		receivedStatusMap = buzzscreen.Service.RewardUseCase.GetReceivedStatusMap(deviceID, periodForAd)
	}

	ads := make(dto.Ads, 0)

	for _, baAd := range baAds {
		rewardStatus, ok := receivedStatusMap[baAd.ID+dto.BuzzAdCampaignIDOffset]
		if !ok {
			rewardStatus = reward.StatusUnknown
		}

		ad := builder.Build(ctx, baAd, rewardStatus)
		builder.BuildRewardServiceResponse(ctx, ad, baAd, canHandleRewardServiceResponse)

		ads = append(ads, ad)

		core.Logger.Debugf("getV3AdFromBAAd() - BAAd: %+v, ad: %+v", baAd, ad)
	}

	return &ads
}

func filterOutNotAllowedAds(ads dto.Ads, unit *app.Unit, sdkVersion int, osVersion string) *dto.Ads {
	filteredAds := make(dto.Ads, 0)

	for _, ad := range ads {
		if shouldIgnoreAdWithNoReward(unit, sdkVersion) && *(ad.RewardStatus) != reward.StatusUnknown {
			continue
		}

		if shouldIgnoreAdWithClickURLLength(*ad, *unit, osVersion) {
			continue
		}

		filteredAds = append(filteredAds, ad)
	}

	return &filteredAds
}

func getTarget(adReq *dto.AdsRequest) *map[string]interface{} {
	targeting := make(map[string]interface{})

	ref := reflect.ValueOf(*adReq)
	//var ok bool
	for i := 0; i < ref.NumField(); i++ {
		valueField := ref.Field(i)
		typeField := ref.Type().Field(i)
		tag := typeField.Tag
		//case reflect.Int:
		if tag.Get("targeting") != "" && valueField.String() != "" && valueField.Interface() != nil {
			switch valueField.Kind() {
			case reflect.String:
				targeting[tag.Get("targeting")] = valueField.String()
			case reflect.Int:
				targeting[tag.Get("targeting")] = valueField.Int()
			default:
				targeting[tag.Get("targeting")] = valueField.Interface()
			}
		}
	}

	now := time.Now()
	age := -1

	birthday := adReq.Birthday
	if birthday == "" && adReq.BirthYear > 0 {
		birthday = fmt.Sprintf("%d-12-31", adReq.BirthYear)
	}

	if birthday != "" {
		t, _ := time.Parse("2006-01-02", birthday)
		age = now.Year() - t.Year()
		if now.YearDay() < t.YearDay() {
			age--
		}
	}
	if age > 0 {
		targeting["age"] = age
	}
	return &targeting
}

func canHandleRewardServiceResponse(unit *app.Unit, sdkVersion int) bool {
	disallows := []bool{
		sdkVersion < 30004,                                       // Benefit 1.X
		30201 <= sdkVersion && sdkVersion <= 30302,               // 체류리워드 지급되지 않는 버그
		unit.ID == 137906464725548 || unit.ID == 326799703219226, // 노티플러스 유닛에서 events 객체가 있으면 크래쉬
	}

	for _, disallow := range disallows {
		if disallow {
			return false
		}
	}

	return true
}

// 초기 Benefit SDK 와 OLock 의 버그로 인한 하드코딩
func shouldIgnoreAdWithNoReward(unit *app.Unit, sdkVersion int) bool {
	return unit.ID == 485347864697339 || (unit.UnitType.IsTypeBenefit() && sdkVersion < 20042)
}

func shouldIgnoreAdWithClickURLLength(ad dto.Ad, unit app.Unit, osVersionString string) bool {
	if unit.IsAndroid() {
		osVersion, _ := strconv.ParseInt(osVersionString, 0, 64)

		// https://buzzvil.atlassian.net/browse/PO-613 url length problems in kitkat (api level 19 - 20)
		if osVersion <= 20 {
			clickURL, ok := ad.Creative["click_url"]
			if !ok {
				return false
			}

			if len(clickURL.(string)) > 2048 {
				return true
			}
		}
	}

	return false
}

// GetShoppingCategoriesFromBuzzAd returns shopping categories from buzzad
func GetShoppingCategoriesFromBuzzAd() ([]string, error) {
	var response []string

	request := &network.Request{
		Method: "GET",
		URL:    buzzscreen.Service.BuzzAdURL + "/shopping/categories/",
	}
	statusCode, err := request.GetResponse(&response)

	if err != nil {
		return nil, err
	} else if statusCode != http.StatusOK {
		return nil, fmt.Errorf("status code is %v", statusCode)
	}
	return response, nil
}
