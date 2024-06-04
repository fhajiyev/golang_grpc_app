package reportsvc

import (
	"context"
	"net/http"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common/code"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/reportsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/report"
)

// Controller type definition
type Controller struct {
	*common.ControllerBase
	reportCase report.UseCase
	unitCase   app.UseCase
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, reportCase report.UseCase, unitCase app.UseCase) Controller {
	con := Controller{reportCase: reportCase, unitCase: unitCase}
	e.POST("/api/report_campaign/", con.PostReportCampaign)
	return con
}

// PostReportCampaign reports both ad and contents
func (con *Controller) PostReportCampaign(c core.Context) error {
	ctx := c.Request().Context()
	var reportReq dto.ReportCampaignRequest
	if err := con.Bind(c, &reportReq); err != nil {
		return err
	}

	if reportReq.IsAd {
		if err := con.reportCase.ReportAd(con.dtoReportRequestToReportRequest(ctx, reportReq)); err != nil {
			return err
		}
	} else {
		if err := con.reportCase.ReportContent(con.dtoReportRequestToReportRequest(ctx, reportReq)); err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code": code.CodeOk,
	})
}

func (con *Controller) dtoReportRequestToReportRequest(ctx context.Context, dtoRCR dto.ReportCampaignRequest) report.Request {
	rcr := report.Request{
		CampaignID:      dtoRCR.CampaignID,
		CampaignName:    dtoRCR.CampaignName,
		Description:     dtoRCR.Description,
		DeviceID:        dtoRCR.DeviceID,
		HTML:            dtoRCR.HTML,
		IconURL:         dtoRCR.IconURL,
		IFA:             dtoRCR.IFA,
		ImageURL:        dtoRCR.ImageURL,
		LandingURL:      dtoRCR.LandingURL,
		ReportReason:    dtoRCR.ReportReason,
		Title:           dtoRCR.Title,
		UnitDeviceToken: dtoRCR.UnitDeviceToken,
		AdReportData:    dtoRCR.AdReportData,
	}

	if rcr.HTML == "" {
		rcr.HTML = dtoRCR.HTMLTag
	}

	if dtoRCR.UnitIDReq > 0 {
		u, err := con.unitCase.GetUnitByID(ctx, dtoRCR.UnitIDReq)
		if err == nil {
			rcr.UnitID = u.ID
			return rcr
		}
		core.Logger.WithError(err).Warnf("dtoReportRequestToReportRequest() failed to get unit with unit_id - %d", dtoRCR.UnitIDReq)
	}

	if dtoRCR.AppIDReq > 0 {
		u, err := con.unitCase.GetUnitByAppIDAndType(ctx, dtoRCR.AppIDReq, app.UnitTypeLockscreen)
		if err == nil {
			rcr.UnitID = u.ID
			return rcr
		}
		core.Logger.WithError(err).Warnf("dtoReportRequestToReportRequest() failed to get unit with app_id - %d", dtoRCR.AppIDReq)
	}

	return rcr
}
