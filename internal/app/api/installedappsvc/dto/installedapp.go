package dto

import "net/http"

// UpdateInstalledAppsRequest type definition
type UpdateInstalledAppsRequest struct {
	IFA       string `form:"ifa" query:"ifa" validate:"required"`
	UnitIDReq int64  `form:"unit_id" query:"unit_id"`
	AppIDReq  int64  `form:"app_id" query:"app_id"`
	AppsData  string `form:"installed_apps" query:"installed_apps" validate:"required"`

	appID   *int64
	Request *http.Request `form:"-"`
}

// UpdateInstalledAppsResponse type definition
type UpdateInstalledAppsResponse struct {
	UpdatePeriod int `json:"update_period"`
}

// GetAppID returns AppID
func (uiar *UpdateInstalledAppsRequest) GetAppID() int64 {
	if uiar.appID != nil {
		return *uiar.appID
	}
	if uiar.AppIDReq > 0 {
		uiar.appID = &uiar.AppIDReq
		return *uiar.appID
	}

	uiar.appID = &uiar.UnitIDReq
	return *uiar.appID
}
