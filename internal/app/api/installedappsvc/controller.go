package installedappsvc

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/installedappsvc/dto"
)

// Controller contains handlers for installed app service.
type Controller struct {
	*common.ControllerBase
	buzzAdURL     string
	deviceUseCase device.UseCase
}

const (
	updateInstalledAppsPeriod = 3600 * 24
)

// NewController creates a controller for handling installed app service.
func NewController(e *core.Engine, deviceUseCase device.UseCase, buzzAdURL string) Controller {
	con := Controller{
		deviceUseCase: deviceUseCase,
		buzzAdURL:     buzzAdURL,
	}

	e.POST("/api/update_installed_apps/", con.UpdateInstalledApps)

	return con
}

func (con *Controller) initRequest(c core.Context, req interface{}) error {
	r := req.(*dto.UpdateInstalledAppsRequest)

	if r.GetAppID() == 0 || r.IFA == "" {
		return errors.New("app_id or ifa is invalid")
	}

	r.Request = c.Request()

	return nil
}

// UpdateInstalledApps updates the user's installed app package names.
func (con *Controller) UpdateInstalledApps(c core.Context) error {
	var req dto.UpdateInstalledAppsRequest
	if err := con.Bind(c, &req, con.initRequest); err != nil {
		return err
	}

	packages, err := con.decodePackages(req.AppsData)
	if err != nil {
		return err
	}

	if err := con.v1UpdateInstalledApps(req.GetAppID(), req.IFA, packages); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.UpdateInstalledAppsResponse{
		UpdatePeriod: updateInstalledAppsPeriod,
	})
}

func (con *Controller) decodePackages(data string) (string, error) {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	zipReader, _ := zlib.NewReader(bytes.NewReader(decodedData))
	//noinspection GoUnhandledErrorResult
	defer zipReader.Close()

	if err != nil {
		return "", err
	}
	unzipped, err := ioutil.ReadAll(zipReader)
	if err != nil {
		return "", err
	}
	return string(unzipped), nil
}

// v1UpdateInstalledApps updates installed app package information in database.
//
// Deprecated: All the installed app data should be move to installed app service.
func (con *Controller) v1UpdateInstalledApps(appID int64, ifa string, packages string) error {
	buzzAdParams := &url.Values{
		"ifa":           []string{ifa},
		"unit_id":       []string{strconv.FormatInt(appID, 10)},
		"package_names": []string{packages},
	}

	go func(buzzAdParams *url.Values) {
		var buzzAdRes interface{}

		request := &network.Request{
			Method: "POST",
			Params: buzzAdParams,
			URL:    con.buzzAdURL + "/api/v1/installed_app",
		}
		request.GetResponse(&buzzAdRes)
	}(buzzAdParams)
	d, err := con.deviceUseCase.GetByParams(device.Params{
		AppID: appID,
		IFA:   ifa,
	})
	if err != nil {
		return err
	} else if d == nil {
		return nil
	}

	profile, err := con.deviceUseCase.GetProfile(d.ID)
	if profile == nil {
		return nil
	}

	oldPackages := profile.InstalledPackages
	if oldPackages != nil && *oldPackages == packages { // 업데이트가 필요 없는 경우
		return nil
	}

	dp := device.Profile{
		ID:                d.ID,
		InstalledPackages: &packages,
	}

	if err := con.deviceUseCase.SaveProfilePackage(dp); err != nil {
		switch err.(type) {
		case device.RemoteProfileError:
			core.Logger.Warnf("v1UpdateInstalledApps() - saving device profile is failed %d-%s, err: %s", appID, ifa, err)
			return nil
		default:
			return common.NewInternalServerError(err)
		}
	}

	return nil
}
