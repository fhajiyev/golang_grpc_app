package custompreviewsvc

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/custompreviewsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview"
)

// TODO Solve daylight saving time (summer time) case
// 1. Support above cases
// 2. Add test cases for edge cases

// Controller struct definition
type Controller struct {
	*common.ControllerBase
	useCase custompreview.UseCase
	mapper  dto.Mapper
	freezableClock
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, uc custompreview.UseCase, freezedTime *time.Time) Controller {
	c := Controller{useCase: uc, freezableClock: newFreezableClock(freezedTime)}
	e.GET("/api/v3/custom-preview-message/config", c.GetConfig)
	e.GET("/api/v3/custom-preview/config", c.DeprecatedGetConfig)
	return c
}

// GetConfig returns latest config with given unit id
func (c *Controller) GetConfig(ctx core.Context) error {
	var req dto.GetConfigReq
	if err := c.Bind(ctx, &req); err != nil {
		return err
	}

	timezoneHeader := ctx.Request().Header["Time-Zone"]
	if len(timezoneHeader) != 1 {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{"error": fmt.Sprintf("timezone is invalid %+v", timezoneHeader)})
	}
	timezone := timezoneHeader[0]

	config, err := c.useCase.GetConfigByUnitID(req.UnitID, timezone, c.freezableClock.now())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{"error": err.Error()})
	} else if config == nil {
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{"error": fmt.Sprintf("config not found with unit id %v", req.UnitID)})
	}

	return ctx.JSON(http.StatusOK, dto.GetConfigRes{
		Config: *c.mapper.ConfigToDTOConfig(*config),
	})
}

// DeprecatedGetConfig returns status code 410
func (c *Controller) DeprecatedGetConfig(ctx core.Context) error {
	return ctx.JSON(http.StatusGone, map[string]interface{}{"error": fmt.Sprintf("this api is deprecated")})
}
