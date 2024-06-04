package userreferralsvc

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/userreferralsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/jwt"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/userreferral"
)

// Controller struct definition
type Controller struct {
	*common.ControllerBase
	userReferralUseCase userreferral.UseCase
	deviceUseCase       device.UseCase
	appUseCase          app.UseCase
	dtoMapper           dtoMapper
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, userReferralUseCase userreferral.UseCase, deviceUseCase device.UseCase, appUseCase app.UseCase) Controller {
	con := Controller{userReferralUseCase: userReferralUseCase, deviceUseCase: deviceUseCase, appUseCase: appUseCase}
	e.GET("/api/users/:device", con.GetUser)
	e.POST("/api/users/:device/referral", con.PostUserReferral)
	return con
}

// GetUser returns user and config
func (con *Controller) GetUser(c core.Context) error {
	ctx := c.Request().Context()
	deviceID, err := strconv.ParseInt(c.Param("device"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	device, err := con.deviceUseCase.GetByID(deviceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	} else if device == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "device not found"})
	}

	config, err := con.appUseCase.GetReferralRewardConfig(ctx, device.AppID)
	if err != nil {
		core.Logger.WithError(err).Warnf("GetUser() - Failed to GetReferralConfigByApp err: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	} else if config == nil {
		core.Logger.WithError(err).Warnf("GetUser() - Failed to GetReferralConfigByApp : config not found")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "config not found"})
	}

	user, err := con.userReferralUseCase.GetOrCreateUserByDevice(device.ID, device.AppID, device.UnitDeviceToken, config.VerifyURL)
	if err != nil {
		core.Logger.WithError(err).Warnf("GetUser() - Failed to GetOrCreateUserByDevice err: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}

	// Disable referral option for old devices and invalid devices
	if user.ReferrerID == 0 && !con.validateRefereeDevice(device, config) {
		user.ReferrerID = 1
	}

	return c.JSON(http.StatusOK, dto.GetUserResponse{
		User:   con.dtoMapper.userToDTOUser(*user),
		Config: con.dtoMapper.configToDTOConfig(*config),
	})
}

// PostUserReferral gives reward to device
func (con *Controller) PostUserReferral(c core.Context) error {
	ctx := c.Request().Context()
	code := c.FormValue("code")
	refereeDeviceID, err := strconv.ParseInt(c.Param("device"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	refereeDevice, err := con.deviceUseCase.GetByID(refereeDeviceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	} else if refereeDevice == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": "refereeDevice not found"})
	}

	referrer, err := con.userReferralUseCase.GetUserByCode(code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
	}

	referrerDevice, err := con.deviceUseCase.GetByID(referrer.DeviceID)
	if err != nil {
		core.Logger.WithError(err).Warnf("PostUserReferral() - Failed to GetByID err: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	} else if referrerDevice == nil {
		core.Logger.WithError(err).Warnf("PostUserReferral() - Failed to GetByID : referrer device not found")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "referrer device not found"})
	}

	config, err := con.appUseCase.GetReferralRewardConfig(ctx, refereeDevice.AppID)
	if err != nil {
		core.Logger.WithError(err).Warnf("PostUserReferral() - Failed to GetReferralConfigByApp err: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	} else if config == nil {
		core.Logger.WithError(err).Warnf("PostUserReferral() - Failed to GetReferralConfigByApp : config not found")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "config not found"})
	}

	// Device validation
	if !con.validateRefereeDevice(refereeDevice, config) {
		core.Logger.Warnf("PostUserReferral() - refereeDevice.ID %v failed device validation", refereeDevice.ID)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": DeviceValidationError{Message: "referee device"}.Error()})
	}
	if !con.validateReferrerDevice(referrerDevice, config) {
		core.Logger.Warnf("PostUserReferral() - referrerDevice.ID %v failed device validation", referrerDevice.ID)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": DeviceValidationError{Message: "referrer device"}.Error()})
	}

	// Set ingredients for calling CreateReferral
	ingr := con.getCreateReferralIngredients(refereeDevice, code, config)

	// Create referral
	ok, err := con.userReferralUseCase.CreateReferral(*ingr)
	if err != nil {
		core.Logger.WithError(err).Warnf("PostUserReferral() - Failed to CreateReferral err: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"success": ok})
}

func (con *Controller) validateRefereeDevice(device *device.Device, config *app.ReferralRewardConfig) bool {
	oldDevice := config.StartDate != nil && config.StartDate.After(device.CreatedAt)
	invalidDevice := config.ExpireHours > 0 && time.Now().Sub(device.CreatedAt).Hours() > float64(config.ExpireHours)
	invalidDevice = invalidDevice || config.MinSdkVersion > 0 && *device.SDKVersion < config.MinSdkVersion

	return !(!config.Enabled || oldDevice || invalidDevice)
}

func (con *Controller) validateReferrerDevice(device *device.Device, config *app.ReferralRewardConfig) bool {
	return !((config.AppID != device.AppID) || (config.EndDate != nil && config.EndDate.Before(device.CreatedAt)))
}

func (con *Controller) getCreateReferralIngredients(refereeDevice *device.Device, code string, config *app.ReferralRewardConfig) *userreferral.CreateReferralIngredients {
	return &userreferral.CreateReferralIngredients{
		DeviceID:        refereeDevice.ID,
		AppID:           refereeDevice.AppID,
		UnitDeviceToken: refereeDevice.UnitDeviceToken,
		Code:            code,
		JWT:             jwt.GetServiceToken(),
		VerifyURL:       config.VerifyURL,
		RewardAmount:    config.Amount,
		MaxReferral:     config.MaxReferral,
		TitleForReferral: userreferral.TitleForReferral{
			TitleForReferee:     config.TitleForReferee,
			TitleForReferrer:    config.TitleForReferrer,
			TitleForMaxReferrer: config.TitleForMaxReferrer,
		},
	}
}

type dtoMapper struct {
}

func (m dtoMapper) userToDTOUser(user userreferral.DeviceUser) *dto.DeviceUser {
	return &dto.DeviceUser{
		ID:         user.ID,
		DeviceID:   user.DeviceID,
		Code:       user.Code,
		ReferrerID: user.ReferrerID,
		IsVerified: user.IsVerified,
	}
}

func (m dtoMapper) configToDTOConfig(config app.ReferralRewardConfig) *dto.RewardConfig {
	dtoConfig := dto.RewardConfig{
		AppID:   config.AppID,
		Enabled: config.Enabled,
		Amount:  config.Amount,
		Ended:   config.IsEnded(),
	}
	return &dtoConfig
}
