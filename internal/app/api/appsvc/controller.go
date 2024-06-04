package appsvc

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/Buzzvil/buzzscreen-api/internal/app/api/appsvc/dto"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
)

// Controller model definition
type Controller struct {
	useCase app.UseCase
	dtoMapper
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, uc app.UseCase) Controller {
	con := Controller{useCase: uc, dtoMapper: dtoMapper{}}
	e.GET("api/apps/:appID", con.GetApp)
	e.GET("api/apps/:appID/promotions", con.GetPromotions)
	e.GET("api/units/:unitID", con.GetUnit)
	e.GET("api/units/:unitID/configs", con.GetUnitConfigs)
	return con
}

// GetUnitConfigs returns postback configs
func (con *Controller) GetUnitConfigs(c core.Context) error {
	ctx := c.Request().Context()
	authValues := c.Request().Header["Authorization"]
	if len(authValues) < 1 || authValues[0] != os.Getenv("BASIC_AUTHORIZATION_VALUE") {
		return c.NoContent(http.StatusForbidden)
	}

	unitID, err := strconv.ParseInt(c.Param("unitID"), 10, 64)
	if err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	u, err := con.useCase.GetUnitByID(ctx, unitID)
	if err != nil {
		core.Logger.Warnf("Failed to get unit(%v) - %s", unitID, err)
		return &core.HttpError{Code: http.StatusInternalServerError, Message: err.Error()}
	} else if u == nil {
		return c.NoContent(http.StatusBadRequest)
	}

	return c.JSON(http.StatusOK, con.dtoMapper.unitToDTOUnitConfig(*u))
}

// GetApp api definition
func (con *Controller) GetApp(c core.Context) error {
	ctx := c.Request().Context()
	appID, err := strconv.ParseInt(c.Param("appID"), 10, 64)
	if err != nil {
		return err
	}

	ap, err := con.useCase.GetAppByID(ctx, appID)
	if err != nil {
		core.Logger.Warnf("Failed to get app(%v) - %s", appID, err)
		return &core.HttpError{Code: http.StatusInternalServerError, Message: err.Error()}
	} else if ap == nil {
		return c.NoContent(http.StatusBadRequest)
	}

	return c.JSON(http.StatusOK, con.dtoMapper.appToDTOApp(*ap))
}

// GetUnit api definition
func (con *Controller) GetUnit(c core.Context) error {
	ctx := c.Request().Context()
	unitID, err := strconv.ParseInt(c.Param("unitID"), 10, 64)
	if err != nil {
		return err
	}

	u, err := con.useCase.GetUnitByID(ctx, unitID)
	if err != nil {
		core.Logger.Warnf("Failed to get unit(%v) - %s", unitID, err)
		return &core.HttpError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	if u == nil {
		return c.NoContent(http.StatusBadRequest)
	}

	return c.JSON(http.StatusOK, con.dtoMapper.unitToDTOUnit(*u))
}

// GetPromotions api definition
// This api only returns one or zero active configuration even if there exists multiple active confirgurations
// The api is only used by OCB organization, and they expect existence of only one active configuration
func (con *Controller) GetPromotions(c core.Context) error {
	ctx := c.Request().Context()
	appID, err := strconv.ParseInt(c.Param("appID"), 10, 64)
	if err != nil {
		return &core.HttpError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	promos, err := con.getActivePromotions(ctx, appID)
	if err != nil {
		core.Logger.Warnf("Failed to get App's WelcomePromotion - %s", err)
		return &core.HttpError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return c.JSON(http.StatusOK, dto.GetCampaignsRes{
		Promotions: promos,
	})
}

func (con *Controller) getActivePromotions(ctx context.Context, appID int64) ([]*dto.Promotion, error) {
	wrcs, err := con.useCase.GetActiveWelcomeRewardConfigs(ctx, appID)
	if err != nil {
		return nil, err
	}

	if len(wrcs) == 0 {
		return nil, nil
	}

	promos := make([]*dto.Promotion, 0)

	wrc := wrcs.FindOngoingWRC()
	promos = append(promos, &dto.Promotion{
		Type:      dto.WelcomePromotion,
		Amount:    wrc.Amount,
		StartTime: wrc.StartTime,
		EndTime:   wrc.EndTime,
	})

	return promos, nil
}

type dtoMapper struct{}

func (m *dtoMapper) appToDTOApp(ap app.App) *dto.App {
	return &dto.App{
		ID: ap.ID,
	}
}

func (m *dtoMapper) unitToDTOUnitConfig(u app.Unit) *dto.UnitConfigs {
	return &dto.UnitConfigs{
		OrganizationID: u.OrganizationID,
		Postback: dto.Postback{
			URL:     u.PostbackURL,
			AESIv:   u.PostbackAESIv,
			AESKey:  u.PostbackAESKey,
			Headers: u.PostbackHeaders,
			HMACKey: u.PostbackHMACKey,
			Params:  u.PostbackParams,
			Class:   u.PostbackClass,
			Config:  u.PostbackConfig,
		},
	}
}

func (m *dtoMapper) unitToDTOUnit(u app.Unit) *dto.Unit {
	settings := dto.Settings{
		BaseHourLimit:    1,
		BaseReward:       u.BaseReward,
		BaseInitPeriod:   u.BaseInitPeriod,
		FeedRatio:        u.FeedRatio,
		FirstScreenRatio: u.FirstScreenRatio,
		HourLimit:        12,
		PagerRatio:       u.PagerRatio,
		PageLimit:        u.PageLimit,
	}

	return &dto.Unit{
		ID:       u.ID,
		Type:     u.UnitType,
		Settings: settings,
	}
}
