package eventsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzlib-go/mq"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/eventsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/eventsvc/publisher"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/authmw"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Controller type definition
type Controller struct {
	*common.ControllerBase
	appUseCase             app.UseCase
	adUseCase              ad.UseCase
	deviceUseCase          device.UseCase
	contentCampaignUseCase contentcampaign.UseCase
	eventUseCase           event.UseCase

	publisher *publisher.Wrapper
}

const (
	temporaryLineItemOffset = 50000000
	staffOrganizationID     = 1
)

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, appUseCase app.UseCase, authUseCase auth.UseCase, deviceUseCase device.UseCase, eventUseCase event.UseCase, contentCampaignUseCase contentcampaign.UseCase, adUseCase ad.UseCase, p mq.Publisher) Controller {
	con := Controller{
		appUseCase:             appUseCase,
		eventUseCase:           eventUseCase,
		deviceUseCase:          deviceUseCase,
		contentCampaignUseCase: contentCampaignUseCase,
		adUseCase:              adUseCase,

		publisher: publisher.New(p),
	}

	e.POST("/api/event/", con.PostEventV1)
	e.GET("/api/track-event", con.TrackEvent, authmw.Authenticate(authUseCase))
	e.GET("/api/reward-status", con.GetRewardStatus, authmw.Authenticate(authUseCase))
	return con
}

// GetRewardStatus returns reward status for a token
func (con *Controller) GetRewardStatus(c core.Context) error {
	init := func(c core.Context, req interface{}) error {
		r := req.(*dto.RewardStatusRequest)

		te := con.eventUseCase.GetTokenEncrypter()
		token, err := te.Parse(r.TokenStr)
		if err != nil {
			return err
		}

		r.Token = *token
		return nil
	}

	var rewardStatusReq dto.RewardStatusRequest
	if err := con.Bind(c, &rewardStatusReq, init); err != nil {
		core.Logger.Warnf("GetRewardStatus() Bind error: %v", err)
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	parser := header.NewHTTPParser(c.Request().Header)
	auth, err := parser.Auth()
	if err != nil {
		return &core.HttpError{Code: http.StatusForbidden, Message: err}
	}

	status, err := con.eventUseCase.GetRewardStatus(rewardStatusReq.Token, *auth)
	if err != nil {
		core.Logger.Warnf("GetRewardStatus() error: %v", err)
		return err
	}

	return c.JSON(http.StatusOK, dto.RewardStatusResponse{Status: status})
}

// TrackEvent publish a create event message
func (con *Controller) TrackEvent(c core.Context) error {
	init := func(c core.Context, req interface{}) error {
		r := req.(*dto.TrackEventRequest)

		te := con.eventUseCase.GetTokenEncrypter()
		token, err := te.Parse(r.TokenStr)
		if err != nil {
			return err
		}

		r.Token = *token
		return nil
	}

	var trackEventReq dto.TrackEventRequest
	if err := con.Bind(c, &trackEventReq, init); err != nil {
		core.Logger.Warnf("TrackEvent() Bind error: %v", err)
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	parser := header.NewHTTPParser(c.Request().Header)
	auth, err := parser.Auth()
	if err != nil {
		return &core.HttpError{Code: http.StatusForbidden, Message: err}
	}

	resourceData, err := con.getResourceData(trackEventReq.Token.Resource)
	if err != nil {
		return err
	}

	handler := con.publisher.Handler(*resourceData, *auth)

	err = con.eventUseCase.TrackEvent(handler, trackEventReq.Token)
	if err != nil {
		core.Logger.Warnf("TrackEvent() error: %v", err)
		return &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{})
}

func (con *Controller) getResourceData(resource event.Resource) (*publisher.ResourceData, error) {
	var resourceData *publisher.ResourceData
	var err error

	switch resource.Type {
	case event.ResourceTypeAd:
		resourceData, err = con.getResourceDataFromAd(resource)
	case event.ResourceTypeArticle:
		resourceData, err = con.getResourceDataFromArticle(resource)
	default:
		err = fmt.Errorf("Undefined resource type. %s", resource.Type)
	}
	if err != nil {
		errmsg := fmt.Errorf("Failed to get resourceData. err: %s", err)
		core.Logger.Warn(errmsg)
		return nil, &core.HttpError{Code: http.StatusBadRequest, Message: err}
	}

	return resourceData, nil
}

func (con *Controller) getResourceDataFromAd(resource event.Resource) (*publisher.ResourceData, error) {
	detail, err := con.adUseCase.GetAdDetail(resource.ID)
	if err != nil {
		return nil, err
	}

	extra := make(map[string]interface{})
	extraData, ok := detail.Extra["extra_data"]
	if ok {
		extra = extraData.(map[string]interface{})
	}

	name := detail.Name
	if resource.Name != nil {
		name = *resource.Name // use name from allocation response instead of lineitem name
	}

	resourceData := &publisher.ResourceData{
		ID:             resource.ID,
		Name:           name,
		OrganizationID: detail.OrganizationID,
		RevenueType:    detail.RevenueType,
		Extra:          extra,
	}

	return resourceData, nil
}

func (con *Controller) getResourceDataFromArticle(resource event.Resource) (*publisher.ResourceData, error) {
	cc, err := con.contentCampaignUseCase.GetContentCampaignByID(resource.ID)
	if err != nil {
		return nil, err
	}

	extra := make(map[string]interface{})
	if cc.ExtraData != nil {
		extraUnit, ok := (*cc.ExtraData)["unit"]
		if ok {
			extra = extraUnit.(map[string]interface{})
		}
	}

	return &publisher.ResourceData{
		ID:             resource.ID,
		Name:           cc.Name,
		OrganizationID: cc.OrganizationID,
		RevenueType:    "",
		Extra:          extra,
	}, nil
}

// PostEventV1 logs event request to "general.log"
func (con *Controller) PostEventV1(c core.Context) error {
	ctx := c.Request().Context()
	var eventReq dto.EventV1Request
	if err := con.Bind(c, &eventReq); err != nil {
		return err
	}

	if !con.validateUnitID(ctx, &eventReq) {
		return errors.New("unit_id or app_id should be valid")
	}

	if eventReq.UnitDeviceToken != nil {
		ok, err := con.deviceUseCase.ValidateUnitDeviceToken(*eventReq.UnitDeviceToken)
		if !ok {
			core.Logger.Infof("PostEventV1() - failed to validate unit_device_token. set unit_device_token to nil. err: %s req: %+v", err.Error(), eventReq)
			eventReq.UnitDeviceToken = nil
		}
	}

	if eventReq.Gudid == "" {
		eventReq.Gudid = uuid.NewV4().String()
	}
	jsonStr, err := json.Marshal(eventReq)
	if err != nil {
		return err
	}
	jsonData := map[string]interface{}{
		"log_type":  "Event",
		"json_data": string(jsonStr),
		"message":   "general",
	}
	if logger := core.Loggers["general"]; logger != nil {
		logger.WithFields(jsonData).Info("Log")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"gudid":       eventReq.Gudid,
		"period":      600,
		"type_filter": "all",
	})
}

func (con *Controller) validateUnitID(ctx context.Context, er *dto.EventV1Request) bool {
	if er.UnitID > 0 {
		return true
	}
	if er.UnitIDReq > 0 {
		er.UnitID = er.UnitIDReq
		return true
	}
	if er.AppIDReq > 0 {
		u, err := con.appUseCase.GetUnitByAppIDAndType(ctx, er.AppIDReq, app.UnitTypeLockscreen)
		if err != nil {
			return false
		}
		er.UnitID = u.ID
		return true
	}
	return false
}
