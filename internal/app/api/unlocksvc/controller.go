package unlocksvc

import (
	"context"
	"errors"
	"net/http"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/unlocksvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
)

// Controller type definition
type Controller struct {
	*common.ControllerBase
	RewardUseCase  reward.UseCase
	AppUseCase     app.UseCase
	PayloadUseCase payload.UseCase
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, rewardUseCase reward.UseCase, appUseCase app.UseCase, payloadUseCase payload.UseCase) Controller {
	con := Controller{
		RewardUseCase:  rewardUseCase,
		AppUseCase:     appUseCase,
		PayloadUseCase: payloadUseCase,
	}
	e.POST("api/impression/", con.Unlock)
	e.POST("api/unlock", con.Unlock)
	return con
}

func (con *Controller) initRequest(c core.Context, req interface{}) error {
	var err error
	r := req.(*dto.PostUnlockRequest)
	r.Request = c.Request()

	r.Payload, err = con.PayloadUseCase.ParsePayload(r.PayloadStr)
	if err != nil {
		return err
	}

	if r.AppID == nil && r.UnitID == nil {
		if r.Payload.UnitID == nil {
			return errors.New("app_id & unit_id not found")
		}
		r.UnitID = r.Payload.UnitID
		r.AppID = r.Payload.UnitID
	} else if r.UnitID == nil {
		unitID := int64(0)
		r.UnitID = &unitID
	} else if r.AppID == nil {
		r.AppID = r.UnitID
	}

	if r.ClickType != string(reward.ClickTypeUnlock) {
		return errors.New("invalid click type")
	}

	if r.CampaignID == 0 {
		r.CampaignID = dto.PlaceholderCampaignID
	}

	return nil
}

// Unlock request reward for the device
func (con *Controller) Unlock(c core.Context) error {
	ctx := c.Request().Context()
	var req dto.PostUnlockRequest
	if err := con.Bind(c, &req, con.initRequest); err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	if req.Reward == 0 {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"reward_received": 0,
		})
	}

	core.Logger.Infof("Unlock() - UnlockReq aid:%v, uid:%v, did: %v ifa: %v, udt: %v, r: %v, br: %v, cid: %v, ct: %v, slot: %v", req.AppID, req.UnitID, req.DeviceID, req.IFA, req.UnitDeviceToken, req.Reward, req.BaseReward, req.CampaignID, req.ClickType, req.Slot)

	rewardAmount, err := con.validateAndGiveReward(ctx, req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"reward_received": rewardAmount,
	})
}

func (con *Controller) validateAndGiveReward(ctx context.Context, req dto.PostUnlockRequest) (rewardAmount int, err error) {
	if con.PayloadUseCase.IsPayloadExpired(req.Payload) {
		return 0, nil
	}

	rewardIngredients := reward.RequestIngredients{
		AppID:           *req.AppID,
		UnitID:          *req.UnitID,
		DeviceID:        req.DeviceID,
		IFA:             req.IFA,
		UnitDeviceToken: req.UnitDeviceToken,
		CampaignID:      req.CampaignID,
		CampaignName:    req.CampaignName,
		CampaignType:    req.CampaignType,
		CampaignOwnerID: req.CampaignOwnerID,
		CampaignIsMedia: req.CampaignIsMedia,
		Slot:            req.Slot,
		Reward:          req.Reward,
		BaseReward:      req.BaseReward,
		ClickType:       reward.ClickTypeUnlock,
		Checksum:        req.Checksum,
	}

	if err := con.RewardUseCase.ValidateRequest(rewardIngredients); err != nil {
		rewardIngredients.AppID = *req.UnitID
		if err := con.RewardUseCase.ValidateRequest(rewardIngredients); err != nil {
			return 0, &core.HttpError{Code: http.StatusBadRequest}
		}
	}

	u, err := con.AppUseCase.GetUnitByID(ctx, *req.UnitID)
	if err != nil {
		return 0, &core.HttpError{Code: http.StatusBadRequest}
	} else if u == nil || !u.IsActive {
		return 0, &core.HttpError{Code: http.StatusBadRequest}
	}

	// [5a59616] sdk 1240버전에서 unlock시 base_reward세팅 안되는 버그로 인해 한동안은 강제로 세팅
	// 현재에도 계속 버그가 있는 것으로 보임, sdkversion 20088에서도 발생확인
	if req.Reward != 0 && req.BaseReward == 0 {
		rewardIngredients.BaseReward = req.Reward
	}

	rewardReceived, err := con.RewardUseCase.GiveReward(rewardIngredients)
	if err == nil {
		return rewardReceived, nil
	}

	switch err.(type) {
	case reward.DuplicatedError, reward.UnprocessableError:
		core.Logger.Warnf("Unlock() - Failed to give reward. err: %s", err)
		return rewardReceived, nil
	case reward.RemoteError:
		return 0, common.NewInternalServerError(err)
	default:
		return 0, common.NewInternalServerError(err)
	}
}
