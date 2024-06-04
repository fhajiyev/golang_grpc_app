package contentimpressionsvc

import (
	"net/http"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/contentimpressionsvc/dto"
)

// Controller type definition
type Controller struct {
	*common.ControllerBase
	TrackingDataUseCase    trackingdata.UseCase
	ImpressionDataUseCase  impressiondata.UseCase
	ContentCampaignUseCase contentcampaign.UseCase
	DeviceUseCase          device.UseCase
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, trackingDataUseCase trackingdata.UseCase, impressionDataUseCase impressiondata.UseCase, contentCampaignUseCase contentcampaign.UseCase, deviceUseCase device.UseCase) Controller {
	con := Controller{
		TrackingDataUseCase:    trackingDataUseCase,
		ImpressionDataUseCase:  impressionDataUseCase,
		ContentCampaignUseCase: contentCampaignUseCase,
		DeviceUseCase:          deviceUseCase,
	}
	e.GET("api/content_impression/", con.ContentImpression)
	return con
}

func (con *Controller) initRequest(c core.Context, req interface{}) error {
	r := req.(*dto.GetContentImpressionRequest)
	r.Request = c.Request()

	var err error
	r.ImpressionData, err = con.ImpressionDataUseCase.ParseImpressionData(r.ImpressionDataStr)
	if err != nil {
		return err
	}

	ok, err := con.DeviceUseCase.ValidateUnitDeviceToken(r.ImpressionData.UnitDeviceToken)
	if !ok {
		return err
	}

	if r.Place != nil && *r.Place == "__place__" {
		r.Place = nil
	}
	if r.Position != nil && *r.Position == "__position__" {
		r.Position = nil
	}
	if r.SessionID != nil && *r.SessionID == "__session_id__" {
		r.SessionID = nil
	}
	if r.TrackingDataStr != nil {
		r.TrackingData, _ = con.TrackingDataUseCase.ParseTrackingData(*r.TrackingDataStr)
	}

	return nil
}

// ContentImpression roles following
// 1. increase impression for contentCampaign and device
// 2. log impression data to "impression.log"
func (con *Controller) ContentImpression(c core.Context) error {
	var req dto.GetContentImpressionRequest
	if err := con.Bind(c, &req, con.initRequest); err != nil {
		return &core.HttpError{Code: http.StatusBadRequest}
	}

	con.ContentCampaignUseCase.IncreaseImpression(req.ImpressionData.CampaignID, req.ImpressionData.UnitID)
	con.logImpression(req)
	con.DeviceUseCase.SaveActivity(req.ImpressionData.DeviceID, req.ImpressionData.CampaignID, device.ActivityImpression)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"msg":    "ok",
		"status": 200,
	})
}

func (con *Controller) logImpression(req dto.GetContentImpressionRequest) {
	mapForLog := map[string]interface{}{
		"device_id":         req.ImpressionData.DeviceID,
		"unit_id":           req.ImpressionData.UnitID,
		"ifa":               req.ImpressionData.IFA,
		"campaign_id":       req.ImpressionData.CampaignID,
		"count":             1,
		"unit_device_token": req.ImpressionData.UnitDeviceToken,
		"ip":                utils.IPToInt64(utils.GetClientIP(req.Request)),
		"country":           req.ImpressionData.Country,
		"message":           "impression",
	}

	if req.ImpressionData.YearOfBirth != nil {
		mapForLog["year_of_birth"] = *req.ImpressionData.YearOfBirth
	}
	if req.ImpressionData.Gender != nil {
		mapForLog["sex"] = *req.ImpressionData.Gender
	}
	if req.Place != nil {
		mapForLog["place"] = *req.Place
	}
	if req.Position != nil {
		mapForLog["position"] = *req.Position
	}
	if req.SessionID != nil {
		mapForLog["session_id"] = *req.SessionID
	}
	if req.TrackingData != nil {
		mapForLog["model_artifact"] = req.TrackingData.ModelArtifact
	}

	core.Loggers["impression"].WithFields(mapForLog).Info("Log")
}
