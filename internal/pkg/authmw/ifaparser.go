package authmw

import (
	"errors"

	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	"github.com/labstack/echo"
)

type ifaReq struct {
	IFA    string `form:"ifa" query:"ifa" validate:"required"`
	AppID  int64  `form:"app_id" query:"app_id"`
	UnitID int64  `form:"unit_id" query:"unit_id"`
}

// IFAParser parse v1 requests. It assumes that request has `ifa` and `app_id` fields as query for form data.
func IFAParser(c echo.Context) (*auth.Identifier, error) {
	req := new(ifaReq)
	if err := c.Bind(req); err != nil {
		return nil, common.NewBindError(err)
	}

	if err := c.Validate(req); err != nil {
		return nil, common.NewBindError(err)
	}

	appID := req.AppID

	// In case of AppID is empty, use UnitID as AppID to support backward compatibility.
	if appID == 0 {
		appID = req.UnitID
	}

	if appID == 0 {
		return nil, errors.New("`app_id` is required")
	}

	return &auth.Identifier{AppID: appID, IFA: req.IFA}, nil
}
