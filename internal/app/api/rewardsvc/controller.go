package rewardsvc

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/rewardsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
)

// Controller struct definition
type Controller struct {
	*common.ControllerBase
	appUseCase    app.UseCase
	eventUseCase  event.UseCase
	rewardUseCase reward.UseCase
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, appUseCase app.UseCase, eventUseCase event.UseCase, rewardUseCase reward.UseCase) Controller {
	con := Controller{
		appUseCase:    appUseCase,
		eventUseCase:  eventUseCase,
		rewardUseCase: rewardUseCase,
	}
	e.POST("/api/rewards", con.PostReward)
	return con
}

// PostReward request reward for the device
func (con *Controller) PostReward(c core.Context) error {
	var req dto.PostRewardReq
	if err := con.Bind(c, &req); err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	} else if req.Type != dto.TypeImpression {
		return &core.HttpError{Code: http.StatusBadRequest, Message: errors.New("unsupported type")}
	}

	core.Logger.Infof("PostReward() - appID:%d, unitID:%d, deviceID:%d, ifa:%s, udt:%s, campaignID: %d reward:%d, baseReward:%d, slot: %d", req.AppID, req.UnitID, req.DeviceID, req.IFA, req.UnitDeviceToken, req.CampaignID, req.Reward, req.BaseReward, req.Slot)

	rewardIngredients, err := con.validateAndGetRewardReq(req)
	if err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	ok := con.callTrackingURL(req.DeviceID, req.CampaignID)
	if ok {
		rewardIngredients.Reward = rewardIngredients.BaseReward // trackingURL 호출성공했을땐 baseReward만 지급
	}

	if rewardIngredients.Reward > 0 {
		code, err := con.giveReward(c, rewardIngredients)
		if code != http.StatusOK {
			return &core.HttpError{Code: code, Message: err}
		}
	}

	return c.NoContent(http.StatusOK)
}

func (con *Controller) giveReward(c core.Context, rewardIngredients *reward.RequestIngredients) (int, error) {
	ctx := c.Request().Context()

	unit, err := con.appUseCase.GetUnitByID(ctx, rewardIngredients.UnitID)
	if err != nil {
		return http.StatusInternalServerError, err
	} else if unit == nil || !unit.IsActive {
		return http.StatusBadRequest, errors.New("empty unit")
	}

	// IOS에서 " "가 들어간 Campaign Name이 "+"로 들어오는 문제. BS-3049, validateAndGetRewardReq이후에 변경해야 validation이 잘 동작함
	// TODO remove
	if unit.IsIOS() && strings.Contains(rewardIngredients.CampaignName, "+") {
		rewardIngredients.CampaignName = strings.Replace(rewardIngredients.CampaignName, "+", " ", -1)
	}

	if _, err := con.rewardUseCase.GiveReward(*rewardIngredients); err != nil {
		switch err.(type) {
		case reward.DuplicatedError:
			return http.StatusConflict, nil
		case reward.UnprocessableError:
			return http.StatusUnprocessableEntity, nil
		case reward.RemoteError:
			return http.StatusInternalServerError, err
		default:
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusOK, nil
}

func (con *Controller) validateAndGetRewardReq(req dto.PostRewardReq) (*reward.RequestIngredients, error) {
	rewardIngredients := reward.RequestIngredients{
		AppID:           req.AppID,
		UnitID:          req.UnitID,
		DeviceID:        req.DeviceID,
		IFA:             req.IFA,
		UnitDeviceToken: req.GetUDT(),
		CampaignID:      req.CampaignID,
		CampaignName:    req.CampaignName,
		CampaignType:    req.CampaignType,
		CampaignOwnerID: req.CampaignOwnerID,
		CampaignIsMedia: req.CampaignIsMedia,
		Slot:            req.Slot,
		Reward:          req.Reward,
		BaseReward:      req.BaseReward,
		ClickType:       reward.ClickTypeLanding,
		Checksum:        req.Checksum,
	}

	if con.rewardUseCase.ValidateRequest(rewardIngredients) == nil {
		return &rewardIngredients, nil
	}

	// lockscreen에서 AppID관련 이슈로 인해 UnitID를 강제로 세팅하여 다시 시도
	ingredients := rewardIngredients
	ingredients.AppID = req.UnitID
	if con.rewardUseCase.ValidateRequest(ingredients) == nil {
		return &ingredients, nil
	}

	// ClientUnitDeviceToken이 encoding 되지 않은 상태로 올때 처리하는 로직. IOS SDK Version <= 20108. BS-2860
	// 위의 lockscreen이슈와 중복으로 일어날 수 없음
	// TODO remove
	ingredients = rewardIngredients
	ingredients.UnitDeviceToken = req.UnitDeviceToken
	if con.rewardUseCase.ValidateRequest(ingredients) == nil {
		return &ingredients, nil
	}

	return nil, errors.New("failed to ValidateRequest. checksum is invalid")
}

func (con *Controller) callTrackingURL(deviceID int64, campaignID int64) bool {
	// 컨텐츠는 BA trackingURL없음
	if campaignID < dto.BuzzAdCampaignIDOffset {
		return false
	}

	resource := event.Resource{
		ID:   campaignID - dto.BuzzAdCampaignIDOffset,
		Type: event.ResourceTypeAd,
	}

	url, err := con.eventUseCase.GetTrackingURL(deviceID, resource)
	if err != nil || url == "" {
		return false
	}

	// bs-point를 통한 중복 적립 방지를 위해 trackingURL호출에 성공한것으로 간주
	req := &network.Request{URL: url, Method: http.MethodGet}
	res, err := req.MakeRequest()
	if err != nil {
		return true
	}
	defer res.Body.Close()

	return true
}
