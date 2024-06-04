package dto

type (
	// ActivityRequest struct definition
	ActivityRequest struct {
		Activities      []*map[string]interface{} `form:"activities" query:"activities"  validate:"required" json:"activities"`
		Carrier         string                    `form:"carrier" query:"carrier"  validate:"required" json:"carrier"`
		DeviceName      string                    `form:"device_name" query:"device_name" validate:"required" json:"device_name"`
		DeviceOs        string                    `form:"device_os" query:"device_os"  validate:"required" json:"device_os"`
		DeviceTimestamp string                    `form:"device_timestamp" query:"device_timestamp"  validate:"required" json:"device_timestamp"`
		Gudid           string                    `form:"gudid" query:"gudid" json:"gudid,omitempty"`
		IFA             string                    `form:"ifa" query:"ifa"  validate:"required" json:"ifa"`
		Package         string                    `form:"package" query:"package"  validate:"required" json:"package"`
		Resolution      string                    `form:"resolution" query:"resolution"  validate:"required" json:"resolution"`
		SdkVersion      string                    `form:"sdk_version" query:"sdk_version" validate:"required" json:"sdk_version"`
		UnitID          string                    `form:"unit_id" query:"unit_id"  validate:"required" json:"unit_id"`

		UnitDeviceToken *string `form:"unit_device_token" query:"unit_device_token" binding:"omitempty" json:"unit_device_token,omitempty"`
		DeviceID        *string `form:"device_id" query:"device_id" binding:"omitempty" json:"device_id,omitempty"`
		YearOfBirth     *string `form:"year_of_birth" query:"year_of_birth" binding:"omitempty" json:"year_of_birth,omitempty"`
		Sex             *string `form:"sex" query:"sex" binding:"omitempty" json:"sex,omitempty"`
		Region          *string `form:"region" query:"region" binding:"omitempty" json:"region,omitempty"`

		// Additional parameters need by bi_events
		AppID      string `json:"app_id,omitempty"`
		ServerTime int64  `json:"server_time,omitempty"`
		ClientTime int64  `json:"client_time,omitempty"`
		IP         string `json:"ip,omitempty"`
		IPCountry  string `json:"ip_country,omitempty"`
	}

	// ActivityResponse struct definition
	ActivityResponse struct {
		Gudid      string `json:"gudid"`
		Period     int    `json:"period"`
		TypeFilter string `json:"type_filter"`
	}
)
