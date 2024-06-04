package notiplussvc

import (
	"net/http"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/notiplussvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/notiplus"
)

// Controller type definition
type Controller struct {
	*common.ControllerBase
	useCase notiplus.UseCase
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, uc notiplus.UseCase) Controller {
	con := Controller{useCase: uc}
	e.GET("/api/v3/notiplus/configs", con.GetConfiguration)
	e.GET("/api/v3/push/configs", con.GetConfiguration)
	return con
}

// GetConfiguration returns configurations based on the unit configuration
func (con *Controller) GetConfiguration(c core.Context) error {
	var req dto.GetConfigsRequest
	if err := con.Bind(c, &req); err != nil {
		return err
	}

	configs, err := con.useCase.GetConfigsByUnitID(req.UnitID)
	if err != nil {
		core.Logger.Warnf("Failed to get notiplusconfig(%v) - %s", req.UnitID, err)
		return &core.HttpError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	configResponses := make([]dto.Config, 0)

	for _, config := range configs {
		configResponses = append(configResponses, dto.Config{
			Title:              config.Title,
			Description:        config.Description,
			Icon:               config.Icon,
			ScheduleHourMinute: config.ScheduleHourMinute,
		})
	}

	return c.JSON(
		http.StatusOK,
		dto.GetConfigsResponse{
			Configs: configResponses,
		},
	)
}
