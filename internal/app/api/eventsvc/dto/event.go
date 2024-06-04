package dto

import "github.com/Buzzvil/buzzscreen-api/internal/pkg/event"

// EventV1Request is request of POST event API
type EventV1Request struct {
	UnitID          int64   `json:"unit_id"`
	UnitIDReq       int64   `form:"unit_id" query:"unit_id" json:"-"`
	AppIDReq        int64   `form:"app_id" query:"app_id" json:"-"`
	Carrier         string  `form:"carrier" query:"carrier" validate:"required" json:"carrier"`
	DeviceName      string  `form:"device_name" query:"device_name" validate:"required" json:"device_name"`
	DeviceOs        int     `form:"device_os" query:"device_os" validate:"required" json:"device_os"`
	DeviceTimeStamp int64   `form:"device_timestamp" query:"device_timestamp" validate:"required" json:"device_timestamp"`
	EventName       string  `form:"event_name" query:"event_name" validate:"required" json:"event_name"`
	Gudid           string  `form:"gudid" query:"gudid" json:"guid"`
	IFA             string  `form:"ifa" query:"ifa" validate:"required" json:"ifa"`
	Package         string  `form:"package" query:"package" validate:"required" json:"package"`
	Resolution      string  `form:"resolution" query:"resolution" validate:"required" json:"resolution"`
	SdkVersion      int     `form:"sdk_version" query:"sdk_version" validate:"required" json:"sdk_version"`
	DeviceID        *int64  `form:"device_id" query:"device_id" binding:"omitempty" json:"device_id,omitempty"`
	Region          *string `form:"region" query:"region" binding:"omitempty" json:"region,omitempty"`
	Sex             *string `form:"sex" query:"sex" binding:"omitempty" json:"sex,omitempty"`
	UnitDeviceToken *string `form:"unit_device_token" query:"unit_device_token" binding:"omitempty" json:"unit_device_token,omitempty"`
	YearOfBirth     *int    `form:"year_of_birth" query:"year_of_birth" binding:"omitempty" json:"year_of_birth,omitempty"`
}

// TrackEventRequest is request of GET track-event API
type TrackEventRequest struct {
	TokenStr string      `query:"token" validate:"required"`
	Token    event.Token `query:"-"`
}

// RewardStatusRequest is request of GET reward-status API
type RewardStatusRequest struct {
	TokenStr string      `query:"token" validate:"required"`
	Token    event.Token `query:"-"`
}

// RewardStatusResponse is response of GET reward-status API
type RewardStatusResponse struct {
	Status string `json:"reward_status"`
}
