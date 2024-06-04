package controller

import (
	"net/http"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
)

// appEventKey is used for BI event
const appEventKey = "WZPe8YIbYE8JX2S1WokPpajp05mRRbR23LH9qo71"

// InitDeviceV1 api definition
func InitDeviceV1(c core.Context) error {
	ctx := c.Request().Context()
	var req dto.InitDeviceV1Request
	if err := bindValue(c, &req); err != nil {
		return err
	}

	if req.GetAppID() == 0 {
		return errors.New("app_id is 0")
	}

	app, err := buzzscreen.Service.AppUseCase.GetAppByID(ctx, req.GetAppID())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": err.Error()})
	} else if app == nil || app.IsDeactivated() {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "app is not found"})
	}

	ok, err := buzzscreen.Service.DeviceUseCase.ValidateUnitDeviceToken(req.UnitDeviceToken)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	clientIP := utils.GetClientIP(c.Request())

	requestedDevice := device.Device{
		DeviceName:      req.DeviceName,
		IFA:             req.IFA,
		Resolution:      req.Resolution,
		SignupIP:        utils.IPToInt64(clientIP),
		AppID:           req.GetAppID(),
		UnitDeviceToken: req.UnitDeviceToken,
		PackageName:     &req.Package,
	}

	if req.Address != "" {
		requestedDevice.Address = &req.Address
	}
	if req.Carrier != "" {
		requestedDevice.Carrier = &req.Carrier
	}
	if req.Sex != "" {
		requestedDevice.Sex = &req.Sex
	}
	if req.YearOfBirth > 0 {
		requestedDevice.YearOfBirth = &req.YearOfBirth
	}
	if req.SDKVersion > 0 {
		requestedDevice.SDKVersion = &req.SDKVersion
	}

	device, err := service.HandleNewDevice(requestedDevice)
	if err != nil {
		return err
	}

	if device != nil {
		country := &(buzzscreen.Service.LocationUseCase.GetClientLocation(c.Request(), "").Country)
		var appVersion *int
		if req.AppVersionCode > 0 {
			appVersion = &req.AppVersionCode
		}

		service.LogDevice(device, country, appVersion, req.IsInBatteryOpts, req.IsBackgroundRestricted, req.HasOverlayPermission)
	}

	res := dto.InitDeviceV1Response{
		Code:    200,
		Message: "ok",
	}
	res.Device.ID = device.ID

	return c.JSON(http.StatusOK, res)
}

// InitSdkV1 api definition
func InitSdkV1(c core.Context) error {
	ctx := c.Request().Context()
	var initReq dto.InitSdkV1Request
	if err := bindValue(c, &initReq); err != nil {
		return err
	} else if initReq.GetAppID() == 0 {
		return c.NoContent(http.StatusBadRequest)
	}

	appUseCase := buzzscreen.Service.AppUseCase
	unit, err := appUseCase.GetUnitByID(ctx, initReq.UnitIDReq)
	if err != nil {
		unit, err = appUseCase.GetUnitByAppID(ctx, initReq.GetAppID())
		if err != nil {
			return &core.HttpError{Code: http.StatusNotFound, Message: "invalid app_id"}
		}
	}

	res := map[string]interface{}{
		"code":              200,
		"msg":               "ok",
		"event_app_id":      unit.AppID,
		"event_api_key":     appEventKey,
		"event_period":      600,
		"event_type_filter": make([]string, 0),
	}

	if initReq.GUID == "" {
		res["event_guid"] = uuid.NewV4().String()
	} else {
		res["event_guid"] = initReq.GUID
	}

	return c.JSON(http.StatusOK, res)
}
