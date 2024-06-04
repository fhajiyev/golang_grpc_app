package configsvc

import (
	"errors"
	"net/http"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/configsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/config"
)

// Controller type definition
type Controller struct {
	*common.ControllerBase
	useCase config.UseCase
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, uc config.UseCase) Controller {
	con := Controller{useCase: uc}
	e.GET("/api/v3/configs", con.GetConfigurations)
	return con
}

var unitIDForPackage = map[string]int64{
	"com.buzzvil.adhours":            100000043,
	"com.buzzvil.honeyscreen.jp":     100000045,
	"com.buzzvil.honeyscreen.tw.new": 100000050,
	"com.slidejoy":                   210342277740215,
}

func (con *Controller) initRequest(c core.Context, req interface{}) error {
	r := req.(*dto.GetConfigsRequest)

	if r.UnitID == 0 {
		/*
		   HS/SJ sdk_version 3500,3600 에서 app_id, unit_id를 request param에 넣어주지 않는 문제. HS-704
		   user-agent에 있는 package_name과 sdk_version으로 app_id를 복구함
		*/
		// TODO remove
		unitID, accepted := unitIDForPackage[r.Package]
		if accepted && (r.SDKVersion == 3500 || r.SDKVersion == 3600) {
			r.UnitID = unitID
		} else {
			return errors.New("unit_id not found")
		}
	}

	return nil
}

// GetConfigurations returns configs based on the request parameters
func (con *Controller) GetConfigurations(c core.Context) error {
	var req dto.GetConfigsRequest
	if err := con.Bind(c, &req, con.initRequest); err != nil {
		core.Logger.Warnf("GetConfigurations() - failed to bind. err %v", err)
		return &core.HttpError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	configReq := config.RequestIngredients{
		UnitID:       req.UnitID,
		Manufacturer: req.Manufacturer,
	}

	configsRes := con.useCase.GetConfigs(configReq)
	configs := make([]*dto.Config, 0)
	if configsRes != nil {
		for _, item := range *configsRes {
			configs = append(configs, &dto.Config{
				Key:   item.Key,
				Value: item.Value,
			})
		}
	}

	return c.JSON(http.StatusOK,
		map[string]interface{}{
			"configs": configs,
		})
}
