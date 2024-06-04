package clickredirectsvc

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/clickredirectsvc/cookiehandler"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/clickredirectsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/profilerequest"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
	"github.com/pkg/errors"
)

// Controller type definition
type Controller struct {
	*common.ControllerBase
	buzzAdURL              string
	RewardUseCase          reward.UseCase
	AppUseCase             app.UseCase
	ContentCampaignUseCase contentcampaign.UseCase
	PayloadUseCase         payload.UseCase
	TrackingDataUseCase    trackingdata.UseCase
	DeviceUseCase          device.UseCase
	ProfileRequestUseCase  profilerequest.UseCase
	EventUseCase           event.UseCase
}

// NewController returns new controller and binds requests to the controller
func NewController(
	e *core.Engine,
	rewardUseCase reward.UseCase,
	appUseCase app.UseCase,
	contentCampaignUseCase contentcampaign.UseCase,
	payloadUseCase payload.UseCase,
	trackingDataUseCase trackingdata.UseCase,
	deviceUseCase device.UseCase,
	eventUseCase event.UseCase,
	profileRequestUseCase profilerequest.UseCase,
	buzzAdURL string,
) Controller {
	con := Controller{
		RewardUseCase:          rewardUseCase,
		AppUseCase:             appUseCase,
		ContentCampaignUseCase: contentCampaignUseCase,
		PayloadUseCase:         payloadUseCase,
		TrackingDataUseCase:    trackingDataUseCase,
		DeviceUseCase:          deviceUseCase,
		EventUseCase:           eventUseCase,
		ProfileRequestUseCase:  profileRequestUseCase,
		buzzAdURL:              buzzAdURL,
	}
	e.GET("api/click_redirect/", con.ClickRedirect)
	e.GET("api/external/click_redirect/", con.ClickRedirect)
	return con
}

func (con *Controller) initRequest(c core.Context, req interface{}) error {
	r := req.(*dto.GetClickRedirectRequest)

	r.Request = c.Request()
	r.IsFalseClick = r.IsFalseClickStr == "1"

	if r.AppID == nil {
		appID := r.UnitID
		r.AppID = &appID
	}
	if r.Position == "__position__" {
		r.Position = ""
	}
	if r.SessionID == "__session_id__" {
		r.SessionID = ""
	}

	r.UseCleanMode = false
	if value, err := strconv.Atoi(r.UseCleanModeStr); err == nil && value > 0 {
		r.UseCleanMode = true
	}

	if r.CampaignID == 0 {
		r.CampaignID = dto.PlaceholderCampaignID
	}

	core.Logger.Infof("ClickRedirect() - appID:%d, unitID:%d, deviceID:%d, ifa:%s, udt:%s, campaignID: %d, reward:%d, baseReward:%d, slot: %d", *r.AppID, r.UnitID, r.DeviceID, r.IFA, r.UnitDeviceToken, r.CampaignID, r.Reward, r.BaseReward, r.Slot)
	return nil
}

// ClickRedirect roles following:
// 1. try to save reward with validation
// 2. increase click for contentcampaign and device
// 3. log click data to "click.log"
// 4. redirect to BuzzAD
func (con *Controller) ClickRedirect(c core.Context) error {
	ctx := c.Request().Context()
	var req dto.GetClickRedirectRequest
	if err := con.Bind(c, &req, con.initRequest); err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	rewardIngredients, err := con.validateAndGetRewardReq(req)
	if err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	ok, err := con.DeviceUseCase.ValidateUnitDeviceToken(rewardIngredients.UnitDeviceToken)
	if !ok {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	u, err := con.AppUseCase.GetUnitByID(ctx, req.UnitID)
	if err != nil || u == nil || !u.IsActive {
		return &core.HttpError{Code: http.StatusBadRequest, Message: "invalid unit_id"}
	}

	if req.ExternalCampaignID != nil && *req.ExternalCampaignID != "" {
		return c.Redirect(302, req.RedirectURL)
	}

	// 할당때만든 payload가 expire되었거나 content_campaign이 expired되었는지 확인, expired되면 리워드를 적립시켜주지않음
	var campaignPayload *payload.Payload
	var expired bool
	if campaignPayload, err = con.PayloadUseCase.ParsePayload(req.PayloadStr); err == nil {
		expired = con.PayloadUseCase.IsPayloadExpired(campaignPayload)
	} else {
		contentCampaign, err := con.ContentCampaignUseCase.GetContentCampaignByID(req.CampaignID)
		if err != nil {
			return &core.HttpError{Code: http.StatusBadRequest, Message: err}
		}
		expired = con.ContentCampaignUseCase.IsContentCampaignExpired(contentCampaign)
	}

	if err := con.increaseClick(req, u, campaignPayload); err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	var redirectURL string
	if redirectURL, err = con.getClickRedirectURL(req); err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	// 리워드 금액 조정
	if req.TrackingURL != nil {
		// Reward == BaseReward + Landing Reward
		// Landing Reward를 BA의 Track Event API를 통해 지급 하려 하므로, 기존의 아래 GiveReward를 통해 보내는 Reward에서 BaseReward만을 남긴다
		rewardIngredients.Reward = rewardIngredients.BaseReward
		url := con.replaceInternalBAURL(*req.TrackingURL)

		if req.UseRewardAPI {
			con.saveTrackingURL(req.DeviceID, req.CampaignID, url)
		} else {
			err := con.callTracker(url)
			if err != nil {
				core.Logger.Warnf("ClickRedirect() - err %+v", err)
			}
		}
	}

	// 리워드 적립
	if !expired && rewardIngredients.Reward > 0 {
		if _, err := con.RewardUseCase.GiveReward(*rewardIngredients); err != nil {
			core.Logger.Warnf("Failed to give reward %v", err)
			switch err.(type) {
			case reward.DuplicatedError, reward.UnprocessableError:
			default:
				return &core.HttpError{Code: http.StatusBadRequest, Message: fmt.Sprintf("%v", err)}
			}
		}
	}

	cookieHandler := cookiehandler.CookieHandler{Ctx: c}
	_, cookieID := cookieHandler.SetCookieIDAndChecksum(req.IFA)

	account := profilerequest.Account{
		IFA:       req.IFA,
		AccountID: req.DeviceID,
		CookieID:  cookieID,
		AppUserID: req.UnitDeviceToken,
	}
	if req.AppID != nil {
		account.AppID = *req.AppID
	}

	// TODO: Make profile population run asynchronously
	if err := con.ProfileRequestUseCase.PopulateProfile(account); err != nil {
		core.Logger.Warnf("PopulateProfile() err: %s", err)
	}

	if req.UseCleanMode && req.RedirectURLClean != nil && *req.RedirectURLClean != "" {
		return c.Redirect(http.StatusFound, *req.RedirectURLClean)
	}

	return c.Redirect(http.StatusFound, redirectURL)
}

// TODO move it to ad domain
func (con *Controller) getClickRedirectURL(req dto.GetClickRedirectRequest) (string, error) {
	if req.CampaignType != dto.CampaignTypeCPM && req.CampaignType != dto.CampaignTypeCPC {
		return req.RedirectURL, nil
	}

	redirectURL := req.RedirectURL

	internalBuzzAdURL := con.replaceInternalBAURL(redirectURL)

	// buzzad processes adequate action for the url and redirects to final ad url
	res, err := (&network.Request{
		URL:              internalBuzzAdURL,
		Method:           http.MethodGet,
		Header:           &req.Request.Header,
		DisallowRedirect: true,
	}).MakeRequest()

	if err != nil {
		return "", err
	} else if res == nil {
		return "", errors.New("empty response returned from buzzad")
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusFound && res.StatusCode != http.StatusMovedPermanently {
		return "", &core.HttpError{Code: http.StatusBadRequest, Message: fmt.Sprintf("BuzzAd responds with bad status code %v", res.StatusCode)}
	}

	// buzzad redirects to final redirect url for ad and client will be redirected to the url
	return res.Header.Get("Location"), nil
}

func (con *Controller) increaseClick(req dto.GetClickRedirectRequest, unit *app.Unit, campaignPayload *payload.Payload) error {
	isDuplicatedClick := false
	if req.CampaignID < dto.BuzzAdCampaignIDOffset {
		// BS SDK 1960 초과 버전에 대해서는 is_false_click 을 client에서 받아서 처리하지만,
		// 1960 이하 버전인데 카드뷰를 사용하도록 이미 나간 허니스크린 unit에 대해서는 부득이하게 아래와 같이 체크한다.
		// TODO : 2019/03/27 아래 코드가 필요한건 BS SDK 1960 이하 버전임.
		// TODO : 허니스크린이 BS SDK 1960 초과 버전으로 migration이 어느정도 진행되었을때 제거 가능.
		if req.UnitID == 100000043 || req.UnitID == 100000045 || req.UnitID == 100000050 {
			// clean_mode 옵션에 따라 original(use_clean_mode = false) / quick(use_clean_mode = true) 페이지를 둘다 호출 가능할 경우
			// 두번 호출되어서 클릭이 중복으로 잡힐 가능성이 있어 이를 막기위한 로직.
			// 옵션 0 : original만 호출 가능 -> 중복체크 필요없음
			// 옵션 1 : quick 먼저 보여주기 -> 유저가 original로 전환시 클릭이 중복으로 잡힐 수 있음
			// 옵션 2 : original 먼저 보여주기 -> 유저가 quick으로 전환시 클릭이 중복으로 잡힐 수 있음
			// 옵션 3 : 둘다 로드 하고 일정 시간 뒤에 original 보여주기(일정 시간 지나기 전 클릭시 quick) -> 무조건 클릭이 중복으로 잡힘
			contentCampaign, err := con.ContentCampaignUseCase.GetContentCampaignByID(req.CampaignID)
			if err != nil {
				return err
			}
			isDuplicatedClick = (!req.UseCleanMode && contentCampaign.CleanMode == 1) ||
				(req.UseCleanMode && contentCampaign.CleanMode == 2) ||
				(req.UseCleanMode && contentCampaign.CleanMode == 3)
		}

		if !req.IsFalseClick && !isDuplicatedClick {
			con.ContentCampaignUseCase.IncreaseClick(req.CampaignID, req.UnitID)
			con.logClick(req, unit, campaignPayload)
			con.DeviceUseCase.SaveActivity(req.DeviceID, req.CampaignID, device.ActivityClick)
		}
	}

	return nil
}

func (con *Controller) logClick(req dto.GetClickRedirectRequest, unit *app.Unit, campaignPayload *payload.Payload) {
	modelArtifact := ""
	trackingData, err := con.TrackingDataUseCase.ParseTrackingData(req.TrackingDataStr)
	if err == nil {
		modelArtifact = trackingData.ModelArtifact
	}

	mapForLog := map[string]interface{}{
		"device_id":         req.DeviceID,
		"unit_id":           req.UnitID,
		"ifa":               req.IFA,
		"campaign_id":       req.CampaignID,
		"ip":                utils.IPToInt64(utils.GetClientIP(req.Request)),
		"unit_device_token": req.GetUDT(),
		"session_id":        req.SessionID,
		"model_artifact":    modelArtifact,
		"position":          req.Position,
		"country":           unit.Country,
		"message":           "click",
	}

	if campaignPayload != nil {
		if campaignPayload.Gender != nil {
			mapForLog["sex"] = *campaignPayload.Gender
		}
		if campaignPayload.YearOfBirth != nil {
			mapForLog["year_of_birth"] = *campaignPayload.YearOfBirth
		}
		if campaignPayload.Country != "" {
			mapForLog["country"] = campaignPayload.Country
		}
	}

	core.Loggers["click"].WithFields(mapForLog).Info("Log")
}

func (con *Controller) validateAndGetRewardReq(req dto.GetClickRedirectRequest) (*reward.RequestIngredients, error) {
	rewardIngredients := reward.RequestIngredients{
		AppID:           *req.AppID,
		UnitID:          req.UnitID,
		DeviceID:        req.DeviceID,
		IFA:             req.IFA,
		UnitDeviceToken: req.GetUDT(),
		CampaignID:      req.CampaignID,
		CampaignName:    req.CampaignName,
		CampaignType:    req.CampaignType,
		CampaignOwnerID: &req.CampaignOwnerID,
		CampaignIsMedia: req.CampaignIsMedia,
		Slot:            req.Slot,
		Reward:          req.Reward,
		BaseReward:      req.BaseReward,
		ClickType:       reward.ClickTypeLanding,
		Checksum:        req.Checksum,
	}

	if con.RewardUseCase.ValidateRequest(rewardIngredients) == nil {
		return &rewardIngredients, nil
	}

	// lockscreen에서 AppID관련 이슈로 인해 UnitID를 강제로 세팅하여 다시 시도
	ingredients := rewardIngredients
	ingredients.AppID = req.UnitID
	if con.RewardUseCase.ValidateRequest(ingredients) == nil {
		return &ingredients, nil
	}

	// ClientUnitDeviceToken이 encoding 되지 않은 상태로 올때 처리하는 로직. IOS SDK Version <= 20108. BS-2860
	// UnitDeviceTokenClient 대신 UnitDeviceToken 사용
	// TODO remove
	if rewardIngredients.UnitDeviceToken != req.UnitDeviceToken {
		rewardIngredients.UnitDeviceToken = req.UnitDeviceToken
		if con.RewardUseCase.ValidateRequest(rewardIngredients) == nil {
			return &ingredients, nil
		}
	}

	return nil, errors.New("failed to ValidateRequest. checksum is invalid")
}

func (con *Controller) callTracker(url string) error {
	core.Logger.Infof("callTracker() url %v", url)
	res, err := (&network.Request{
		URL:    url,
		Method: http.MethodGet,
	}).MakeRequest()
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("buzzad tracker api call with url %+v has error : ", url))
	} else if res.StatusCode/100 != 2 {
		return fmt.Errorf("buzzad tracker api call with url %+v has status code : %+v", url, res.StatusCode)
	}
	return nil
}

func (con *Controller) replaceInternalBAURL(url string) string {
	// BA API 호출 시, Public Domain보다 Internal Domain이 비용을 아낄 수 있으니 Replace 해준다.

	url = strings.Replace(url, "https://api.buzzad.io/", con.buzzAdURL, -1)
	return strings.Replace(url, "https://ad.buzzvil.com", con.buzzAdURL, -1)
}

func (con *Controller) saveTrackingURL(deviceID int64, campaignID int64, trackingURL string) {
	// 컨텐츠는 BA tracking URL을 가지지 않음
	if campaignID < dto.BuzzAdCampaignIDOffset {
		return
	}

	r := event.Resource{
		ID:   campaignID - dto.BuzzAdCampaignIDOffset,
		Type: event.ResourceTypeAd,
	}
	con.EventUseCase.SaveTrackingURL(deviceID, r, trackingURL)
}
