package controller

import (
	"net/http"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/pkg/errors"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
)

// GetDeviceConfig func definition
func GetDeviceConfig(c core.Context) error {
	var contentReq dto.GetDeviceConfigRequest
	if err := bindRequestSupport(c, &contentReq, &GetDeviceConfigV2Request{}); err != nil {
		return err
	}

	contentReq.Method = c.Param("method")
	err := contentReq.UnpackSession()
	if err != nil {
		return common.NewSessionError(err)
	}

	config, err := getDeviceContentConfig(contentReq.Session.DeviceID)
	if err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	res := map[string]interface{}{
		"code": dto.CodeOk,
	}

	switch contentReq.Method {
	case "article":
		res["result"] = map[string]interface{}{
			"data": config.Campaign,
		}
	case "channel":
		res["result"] = map[string]interface{}{
			"data": config.Channel,
		}
	default:
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    dto.CodeBadRequest,
			"message": "Method is not right.",
		})
	}

	return c.JSON(http.StatusOK, res)
}

// PutDeviceConfig func definition
func PutDeviceConfig(c core.Context) error {
	var contentReq dto.PutDeviceConfigRequest
	if err := bindRequestSupport(c, &contentReq, &PutDeviceConfigV2Request{}); err != nil {
		return err
	}

	contentReq.Method = c.Param("method")
	err := contentReq.UnpackSession()
	if err != nil {
		return common.NewSessionError(err)
	}

	deviceConfig, err := getDeviceContentConfig(contentReq.Session.DeviceID)
	if err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	res := map[string]interface{}{
		"code": dto.CodeOk,
	}

	switch contentReq.Method {
	case "article":
		deviceConfig.Campaign = contentReq.Data
		deviceConfig.Save()
	case "channel":
		deviceConfig.Channel = contentReq.Data
		deviceConfig.Save()
	default:
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    dto.CodeBadRequest,
			"message": "Method is not right.",
		})
	}

	logObj := map[string]interface{}{
		"log_type":  "DeviceContentConfig",
		"json_data": deviceConfig.GetJSONStr(),
		"message":   "general",
	}

	// Log to general logger
	core.Loggers["general"].WithFields(logObj).Info("Log")

	return c.JSON(http.StatusOK, res)
}

func getDeviceContentConfig(deviceID int64) (*model.DeviceContentConfig, error) {
	if deviceID == 0 {
		return nil, errors.New("invalid deviceID")
	}

	var config model.DeviceContentConfig
	buzzscreen.Service.DB.Where(&model.DeviceContentConfig{DeviceID: deviceID}).FirstOrInit(&config)
	return &config, nil
}

type (
	// GetDeviceConfigV2Request type definition
	GetDeviceConfigV2Request struct {
		ContentV2BaseRequest
		Method string
	}

	// PutDeviceConfigV2Request type definition
	PutDeviceConfigV2Request struct {
		ContentV2BaseRequest
		Method string
		Data   string `form:"data" query:"data" validate:"required"`
	}
)
