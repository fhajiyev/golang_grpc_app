package dto

import (
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/session"
)

type (
	// GetDeviceRequest struct definition
	GetDeviceRequest struct {
		AppID     int64   `query:"app_id" validate:"required"`
		IFA       string  `query:"ifa"`
		IFV       *string `query:"ifv"`
		PubUserID string  `query:"publisher_user_id"`
	}

	// PostDeviceRequest struct definition
	PostDeviceRequest struct {
		AdID                   string  `form:"ad_id" validate:"required"`
		AndroidID              string  `form:"android_id"`
		AppID                  int64   `form:"app_id" validate:"required"`
		AppVersion             int     `form:"app_version"`
		Birthday               string  `form:"birthday"`
		BirthYear              int     `form:"birth_year"`
		Carrier                string  `form:"carrier"`
		DeviceName             string  `form:"device_name" validate:"required"`
		DeviceID               string  `form:"device_id"`
		Gender                 string  `form:"gender"`
		HMAC                   string  `form:"hmac"`
		Locale                 string  `form:"locale" validate:"required"`
		Resolution             string  `form:"resolution" validate:"required"`
		SDKVersion             int     `form:"sdk_version"`
		SerialNum              string  `form:"serial_num"`
		UserID                 string  `form:"user_id" validate:"required"`
		Package                string  `form:"package"`
		IsInBatteryOpts        *bool   `form:"is_in_battery_optimizations"`
		IsBackgroundRestricted *bool   `form:"is_background_restricted"`
		HasOverlayPermission   *bool   `form:"has_overlay_permission"`
		IFV                    *string `form:"ifv"`
	}

	// CreateDeviceResponse type definition
	CreateDeviceResponse struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Result  map[string]interface{} `json:"result"`
	}

	// Location type definition
	Location struct {
		Country   string  `json:"country"`
		ZipCode   string  `json:"zipCode"`
		State     string  `json:"state"`
		City      string  `json:"city"`
		TimeZone  string  `json:"timeZone"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		IPAddress string  `json:"ipAddress"`
	}
)

type (
	// ConfigRequest type definition
	ConfigRequest struct {
		VideoAutoplay string `form:"video_autoplay" query:"video_autoplay" json:"video_autoplay"`
	}
)

type (
	// PackagesRequest type definition
	PackagesRequest struct {
		SessionKey string `form:"session_key" query:"session_key" validate:"required"`
		Packages   string `form:"packages" query:"packages"`

		Session session.Session
	}

	// AppsResponse type definition
	AppsResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Result  struct {
			IDs string `json:"ids"`
		} `json:"result"`
	}
)

// UnpackSession unpacks session key and assign to Session
func (req *PackagesRequest) UnpackSession() error {
	session, err := buzzscreen.Service.SessionUseCase.GetSessionFromKey(req.SessionKey)
	if err != nil {
		return err
	}

	req.Session = *session
	return nil
}
