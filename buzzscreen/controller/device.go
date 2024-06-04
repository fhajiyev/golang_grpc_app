package controller

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/recovery"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/service"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/ifa"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/session"
)

const appIDCjone = 49220464620468

// DeviceController handles device apis.
type DeviceController struct {
	mapper          deviceMapper
	appUseCase      app.UseCase
	deviceUseCase   device.UseCase
	locationUseCase location.UseCase
	sessionUseCase  session.UseCase
}

// NewDeviceController registers router to pathes and return the controller instance.
func NewDeviceController(router *core.RouterGroup) DeviceController {
	con := DeviceController{
		mapper:          deviceMapper{},
		appUseCase:      buzzscreen.Service.AppUseCase,
		deviceUseCase:   buzzscreen.Service.DeviceUseCase,
		locationUseCase: buzzscreen.Service.LocationUseCase,
		sessionUseCase:  buzzscreen.Service.SessionUseCase,
	}

	router.GET("/devices", con.GetDevice)
	router.POST("/devices", con.PostDevice)

	router.PUT("/devices/:id/config", con.PutConfig)
	router.PUT("/devices/:id/packages", con.PutPackages)

	// Deprecated
	router.POST("/device", con.PostDevice)
	// Deprecated
	router.PUT("/device/packages/replace", con.PutPackages)
	return con
}

// GetDevice returns a device of given id.
func (con DeviceController) GetDevice(c core.Context) error {
	authValues := c.Request().Header["Authorization"]
	if len(authValues) < 1 || authValues[0] != os.Getenv("BASIC_AUTHORIZATION_VALUE") {
		return c.NoContent(http.StatusForbidden)
	}

	var req dto.GetDeviceRequest
	if err := bindValue(c, &req); err != nil {
		return err
	}

	if req.IFA == "" && req.PubUserID == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	d, err := con.deviceUseCase.GetByParams(device.Params{
		AppID:     req.AppID,
		PubUserID: req.PubUserID,
		IFA:       req.IFA,
	})
	if err != nil {
		return err
	} else if d == nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":                d.ID,
		"ifa":               d.IFA,
		"publisher_user_id": d.UnitDeviceToken,
		"app_id":            d.AppID,
	})
}

// PostDevice handles creating or updating device requests.
func (con DeviceController) PostDevice(c core.Context) error {
	ctx := c.Request().Context()
	var deviceReq dto.PostDeviceRequest
	if err := bindRequestSupport(c, &deviceReq, &DeviceV2Request{}); err != nil {
		return err
	}

	if deviceReq.AdID == ifa.EmptyIFA && (deviceReq.IFV == nil || *deviceReq.IFV == "") {
		formParams, _ := c.FormParams()
		core.Logger.Infof("PostDevice() empty ifa header: %+v params %+v", c.Request().Header, formParams)
	}

	app, err := con.appUseCase.GetAppByID(ctx, deviceReq.AppID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"message": err.Error()})
	} else if app == nil || app.IsDeactivated() {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "app is not found"})
	}

	ok, err := con.deviceUseCase.ValidateUnitDeviceToken(deviceReq.UserID)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"message": err.Error()})
	}

	if deviceReq.HMAC != "" {
		unit, err := con.appUseCase.GetUnitByAppID(ctx, deviceReq.AppID)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		if service.CheckHMAC([]byte(unit.InitHMACKey), deviceReq.UserID, deviceReq.HMAC) == false {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"code":    dto.CodeBadRequest,
				"message": "invalid hmac",
			})
		}
	}

	clientIP := utils.GetClientIP(c.Request())
	device, err := service.HandleNewDevice(con.mapper.dtoToDevice(deviceReq, clientIP))
	if err != nil {
		return err
	}

	if device != nil {
		country := &(con.locationUseCase.GetClientLocation(c.Request(), "").Country)
		var appVersion *int
		if deviceReq.AppVersion > 0 {
			appVersion = &deviceReq.AppVersion
		}
		service.LogDevice(device, country, appVersion, deviceReq.IsInBatteryOpts, deviceReq.IsBackgroundRestricted, deviceReq.HasOverlayPermission)
	}

	sessionKey := con.sessionUseCase.GetNewSessionKey(deviceReq.AppID, deviceReq.UserID, device.ID, deviceReq.AndroidID, device.CreatedAt.Unix())

	return c.JSON(http.StatusOK, getResponseSupport(c, dto.CreateDeviceResponse{
		Code: 0,
		Result: map[string]interface{}{
			"session_key": sessionKey,
			"device_id":   device.ID,
		},
	}))
}

type deviceMapper struct{}

func (dm deviceMapper) dtoToDevice(deviceReq dto.PostDeviceRequest, ip string) device.Device {
	d := device.Device{
		DeviceName:      deviceReq.DeviceName,
		IFA:             deviceReq.AdID,
		Resolution:      deviceReq.Resolution,
		SignupIP:        utils.IPToInt64(ip),
		AppID:           deviceReq.AppID,
		UnitDeviceToken: deviceReq.UserID,
		PackageName:     &deviceReq.Package,
	}

	if ifa.ShouldReplaceIFAWithIFV(deviceReq.AdID, deviceReq.IFV) {
		d.IFA = ifa.GetDeviceIFV(*deviceReq.IFV)
	}

	if deviceReq.SerialNum != "" {
		d.SerialNumber = &deviceReq.SerialNum
	}

	if deviceReq.Gender != "" {
		d.Sex = &deviceReq.Gender
	}
	if deviceReq.Birthday != "" {
		t, err := time.Parse("2006-01-02", deviceReq.Birthday)
		if err == nil {
			d.Birthday = &t
		} else {
			core.Logger.WithError(err).Errorf("PostDevice() - parse error: %s", deviceReq.Birthday)
		}
	}
	if deviceReq.BirthYear > 0 {
		d.YearOfBirth = &deviceReq.BirthYear
	}
	if deviceReq.SDKVersion > 0 {
		d.SDKVersion = &deviceReq.SDKVersion
	}
	if deviceReq.Carrier != "" {
		d.Carrier = &deviceReq.Carrier
	}
	return d
}

// PutConfig handles saving device's configurations.
func (con DeviceController) PutConfig(c core.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}
	var req dto.ConfigRequest
	if err := bindValue(c, &req); err != nil {
		return err
	}
	core.Logger.Infof("PutConfig() - [%d] %v", id, req)

	if bytes, err := json.Marshal(req); err == nil {
		var mapForLog map[string]interface{}
		json.Unmarshal(bytes, &mapForLog)
		mapForLog["id"] = id
		core.Loggers["device_config"].WithFields(mapForLog).Info("Log")
	} else {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"code": dto.CodeOk,
	})
}

// PutPackages handles updating device's package list.
func (con DeviceController) PutPackages(c core.Context) error {
	var packageReq dto.PackagesRequest
	if err := bindRequestSupport(c, &packageReq, &PackagesV2Request{}); err != nil {
		return err
	}

	err := packageReq.UnpackSession()
	if err != nil {
		return common.NewSessionError(err)
	}

	if idStr := c.Param("id"); idStr != "" {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if packageReq.Session.DeviceID != id {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"message": "Device ID is wrong"})
		}
	}

	device, err := buzzscreen.Service.DeviceUseCase.GetByID(packageReq.Session.DeviceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"message": err.Error()})
	}

	if device.Packages == nil || *device.Packages != packageReq.Packages {
		device.Packages = &packageReq.Packages
		_, err := buzzscreen.Service.DeviceUseCase.UpsertDevice(*device)
		if err != nil {
			return err
		}

		defer con.getIDsFromInsight(c, packageReq.Packages)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code": dto.CodeOk,
	})
}
func (con DeviceController) getIDsFromInsight(c core.Context, packageStrings string) string {
	defer recovery.MutedRecovery(c)
	var appsResponse dto.AppsResponse

	pkgs := make([]string, 0)

	getIDs := func(pkgs []string) (string, error) {
		params := &url.Values{
			"appIDs": {strings.Join(pkgs, ",")},
			"os":     {"Android"},
		}
		statusCode, err := (&network.Request{
			URL:    env.Config.InsightURL + "/app/ids",
			Params: params,
			Method: "GET",
		}).GetResponse(&appsResponse)

		if statusCode/100 == 2 && err == nil {
			return appsResponse.Result.IDs, err
		}
		return "", err
	}

	ids := ","

	for index, pkg := range strings.Split(packageStrings, ",") {
		pkgs = append(pkgs, pkg)
		if index%10 == 9 {
			subIDs, err := getIDs(pkgs)
			if err != nil {
				core.Logger.WithError(err).Errorf("getIDsFromInsight() - Failed %s", pkgs)
			} else {
				ids = ids[:len(ids)-1] + subIDs + ","
			}
			pkgs = pkgs[:0]
		}
	}
	subIDs, err := getIDs(pkgs)
	if err != nil {
		core.Logger.WithError(err).Errorf("getIDsFromInsight() - Failed %s", pkgs)
	} else {
		ids = ids[:len(ids)-1] + subIDs + ","
	}

	ids = ids[:len(ids)-1]
	core.Logger.Infof("getIDsFromInsight() - %v", ids)
	return ids
}

type (
	// DeviceV2Request type definition
	DeviceV2Request struct {
		AdID                   string `form:"adId" query:"adId" validate:"required"`
		AndroidID              string `form:"androidId" query:"androidId"`
		AppID                  int64  `form:"appId" query:"appId" validate:"required"`
		AppVersion             int    `form:"appVersion" query:"appVersion"`
		Birthday               string `form:"birthday" query:"birthday"`
		BirthYear              int    `form:"birthYear" query:"birthYear"`
		Carrier                string `form:"carrier" query:"carrier"`
		DeviceName             string `form:"deviceName" query:"deviceName" validate:"required"`
		DeviceID               string `form:"deviceId" query:"deviceId"`
		Gender                 string `form:"gender" query:"gender"`
		HMAC                   string `form:"hmac" query:"hmac"`
		Locale                 string `form:"locale" query:"locale" validate:"required"`
		Resolution             string `form:"resolution" query:"resolution" validate:"required"`
		SdkVersion             int    `form:"sdkVersion" query:"sdkVersion"`
		SerialNum              string `form:"serialNum" query:"serialNum"`
		UserID                 string `form:"userId" query:"userId" validate:"required"`
		Package                string `form:"package" query:"package"`
		IsInBatteryOpts        *bool  `form:"is_in_battery_optimizations" query:"is_in_battery_optimizations"`
		IsBackgroundRestricted *bool  `form:"is_background_restricted" query:"is_background_restricted"`
	}

	// PackagesV2Request type definition
	PackagesV2Request struct {
		SessionKey string `form:"sessionKey" query:"sessionKey" validate:"required"`
		Packages   string `form:"packages" query:"packages"`
	}
)
