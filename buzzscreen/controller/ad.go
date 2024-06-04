package controller

import (
	"net/http"
	"strconv"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/ifa"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/pkg/errors"
)

// GetAds fetches ads from buzz ad server and returns them. Ads with invalid landing reward status should be filtered out.
func GetAds(c core.Context) error {
	ctx := c.Request().Context()
	var adsReq dto.AdsRequest
	if err := bindValue(c, &adsReq); err != nil {
		return err
	}

	if ifa.ShouldReplaceIFAWithIFV(adsReq.AdID, adsReq.IFV) {
		adsReq.AdID = ifa.GetDeviceIFV(*adsReq.IFV)
	}

	unit := adsReq.GetUnit(ctx)
	if unit == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "unit can't be nil"})
	} else if !unit.IsActive {
		return c.JSON(http.StatusForbidden, map[string]interface{}{"error": "unit is inactive"})
	}

	if adsReq.Platform == "" {
		adsReq.Platform = "A"
	}

	authParser := header.NewHTTPParser(c.Request().Header)
	auth, err := authParser.Auth()
	if err != nil {
		// ignore Auth() error due to allow nil auth
		// NOTE: expired token을 가지고 있는 유저 존재. REWARD-231
		auth = nil
	}

	err = adsReq.UnpackSession()
	if err != nil {
		return common.NewSessionError(err)
	}

	if adsReq.Session.IsSignedIn() {
		ok, err := buzzscreen.Service.DeviceUseCase.ValidateUnitDeviceToken(adsReq.Session.UserID)
		if !ok {
			core.Logger.Warnf("GetAds() - failed to validate unit_device_token. err: %s. AppID: %+v, UnitID: %+v, DeviceID: %+v, IFA: %+v, UnitDeviceToken: %+v, SDKVersion: %+v, DeviceOS: %+v", err.Error(), adsReq.AppID, adsReq.UnitID, adsReq.DeviceID, adsReq.AdID, adsReq.Session.UserID, adsReq.SdkVersion, adsReq.OsVersion)
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		}

		if unit.UnitType.IsTypeLockscreen() {
			env.SetRedisDeviceDau(adsReq.Session.DeviceID)
		}
	} else if !adsReq.Anonymous {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{"error": "invalid session"})
	}

	profile := adsReq.GetDynamoProfile()
	if profile != nil {
		service.UpdateProfileUnitRegisterSeconds(profile, unit.ID, adsReq.Session.AppID, c.Path())
		service.GiveWelcomeReward(ctx, profile, adsReq.Session.UserID, unit.ID, adsReq.GetCountry())
	}

	if adsReq.GetTypes() == nil {
		return common.NewBindError(errors.New("types is empty"))
	}

	fixParamsBySDKVersion(c, &adsReq)

	if adsReq.UserAgent == "" {
		adsReq.UserAgent = c.Request().UserAgent()
	}

	location := buzzscreen.Service.LocationUseCase.GetClientLocation(c.Request(), adsReq.GetCountry())
	response, err := service.GetAdsByStatus(ctx, &adsReq, location, &adsReq, adsReq.Session.DeviceID, unit.ID, adsReq.GetTargetFill(), auth)
	if err != nil {
		switch err.(type) {
		case ad.RemoteError, reward.RemoteError:
			core.Logger.WithError(err).Warnf("controller.GetAds() - err: %s", err)
		default:
			core.Logger.WithError(err).Errorf("controller.GetAds() - err: %s", err)
		}
		return common.NewInternalServerError(err)
	}

	logGetAds(adsReq, *response)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":    dto.CodeOk,
		"message": nil,
		"result":  response,
	})
}

func logGetAds(req dto.AdsRequest, res dto.AdsResponse) {
	var adIDs []int64

	for _, ad := range res.Ads {
		adIDs = append(adIDs, ad.ID-dto.BuzzAdCampaignIDOffset)
	}

	core.Logger.Infof("/api/v3/ads - Request{AppID:%d, UnitID:%d, DeviceID:%d, IFA:%s, UnitDeviceToken:%s, SDKVersion:%d, OSVersion:%s} Response{Ads:%v}",
		req.Session.AppID,
		req.UnitID,
		req.Session.DeviceID,
		req.AdID,
		req.Session.UserID,
		req.SdkVersion,
		req.OsVersion,
		adIDs,
	)
}

func fixParamsBySDKVersion(c core.Context, req *dto.AdsRequest) {
	ctx := c.Request().Context()
	if req.GetUnit(ctx).UnitType.IsTypeBenefit() {
		// TODO: This block should be removed on 2019-06-30
		// BAB 에서 Parameter 잘못 보내주는 부분 임시 처리 https://buzzvil.atlassian.net/browse/BZZRWRDD-187
		{
			if yob := c.QueryParam("year_of_birth"); yob != "" {
				req.BirthYear, _ = strconv.Atoi(yob)
			}
			if sex := c.QueryParam("sex"); sex != "" {
				req.Gender = sex
			}
		}
		// TODO: This block should be removed on 2019-06-30
		// https://buzzvil.atlassian.net/browse/BZZRWRDD-263
		{
			if req.GetUnit(ctx).UnitType.IsTypeBenefit() && req.SdkVersion < 20042 {
				req.TargetFill = 20
			}
		}
		// TODO: This block should be removed on 2020-01-01
		// https://buzzvil.atlassian.net/browse/BZZRWRDD-441
		{
			if deviceOS := c.QueryParam("device_os"); deviceOS != "" {
				req.OsVersion = deviceOS
			}
		}
	}
}

func GetShoppingCategories(c core.Context) error {

	response, err := service.GetShoppingCategoriesFromBuzzAd()
	if err != nil {
		return common.NewInternalServerError(err)
	}

	return c.JSON(http.StatusOK, response)
}
