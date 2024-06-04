package dto

type (
	// InitDeviceV1Request type definition
	InitDeviceV1Request struct {
		Address                string `form:"address"`
		AppVersionCode         int    `form:"app_version_code"`
		Carrier                string `form:"carrier"`
		DeviceName             string `form:"device_name"`
		IFA                    string `form:"ifa" validate:"required"`
		Resolution             string `form:"resolution"`
		SDKVersion             int    `form:"sdk_version"`
		Sex                    string `form:"sex"`
		Package                string `form:"package"`
		AppIDReq               int64  `form:"app_id"`
		UnitIDReq              int64  `form:"unit_id"`
		UnitDeviceToken        string `form:"unit_device_token" validate:"required"`
		YearOfBirth            int    `form:"year_of_birth"`
		IsInBatteryOpts        *bool  `form:"is_in_battery_optimizations"`
		IsBackgroundRestricted *bool  `form:"is_background_restricted"`
		HasOverlayPermission   *bool  `form:"has_overlay_permission"`

		appID *int64
	}

	// InitDeviceV1Response type definition
	InitDeviceV1Response struct {
		Code    int    `json:"code"`
		Message string `json:"msg"`
		Device  struct {
			ID int64 `json:"id"`
		} `json:"device"`
	}

	// UpdateInstalledAppsV1Request type definition
	UpdateInstalledAppsV1Request struct {
		IFA       string `form:"ifa" query:"ifa" validate:"required"`
		UnitIDReq int64  `form:"unit_id" query:"unit_id"`
		AppIDReq  int64  `form:"app_id" query:"app_id"`
		AppsData  string `form:"installed_apps" query:"installed_apps" validate:"required"`

		appID *int64
	}

	// UpdateInstalledAppsV1Response type definition
	UpdateInstalledAppsV1Response struct {
		UpdatePeriod int `json:"update_period"`
	}
)

type (
	// InitSdkV1Request type definition
	InitSdkV1Request struct {
		AppIDReq  int64  `form:"app_id" query:"app_id"`
		UnitIDReq int64  `form:"unit_id" query:"unit_id"`
		GUID      string `form:"guid" query:"guid"`

		appID *int64
	}
)

//GetAppID returns AppID
func (idr *InitDeviceV1Request) GetAppID() int64 {
	if idr.appID != nil {
		return *idr.appID
	}
	if idr.AppIDReq > 0 {
		idr.appID = &idr.AppIDReq
		return *idr.appID
	}

	idr.appID = &idr.UnitIDReq
	return *idr.appID
}

//GetAppID returns AppID
func (uiar *UpdateInstalledAppsV1Request) GetAppID() int64 {
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

//GetAppID returns AppID
func (isr *InitSdkV1Request) GetAppID() int64 {
	if isr.appID != nil {
		return *isr.appID
	}
	if isr.AppIDReq > 0 {
		isr.appID = &isr.AppIDReq
		return *isr.appID
	}

	isr.appID = &isr.UnitIDReq
	return *isr.appID
}
