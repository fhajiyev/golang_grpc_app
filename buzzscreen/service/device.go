package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/datetime"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbdevice"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
)

// DebugScoreForLog type definition
type DebugScoreForLog struct {
	dto.ESContentCampaign
	LogType       string `json:"log_type"`
	DeviceID      int64  `json:"device_id"`
	TimeAllocated string `json:"time_allocated"`
}

// DeviceForLog type definition
type DeviceForLog struct {
	device.Device
	UnitID                 int64   `json:"unit_id"`
	Country                *string `json:"country,omitempty"`
	Message                string  `json:"message"`
	DateJoined             string  `json:"date_joined"`
	AppVersion             *int    `json:"app_version,omitempty"`
	IsInBatteryOpts        *bool   `json:"is_in_battery_optimizations,omitempty"`
	IsBackgroundRestricted *bool   `json:"is_background_restricted,omitempty"`
	HasOverlayPermission   *bool   `json:"has_overlay_permission,omitempty"`
}

// BuildMap func definition
func (dsfl *DebugScoreForLog) BuildMap() (map[string]interface{}, error) {
	bytes, err := json.Marshal(dsfl)
	if err != nil {
		return nil, err
	}
	var keyValues map[string]interface{}
	json.Unmarshal(bytes, &keyValues)
	return keyValues, err
}

// BuildMap func definition
func (dfl *DeviceForLog) BuildMap() (map[string]interface{}, error) {
	bytes, err := json.Marshal(dfl)
	if err != nil {
		return nil, err
	}
	var keyValues map[string]interface{}
	json.Unmarshal(bytes, &keyValues)
	return keyValues, err
}

// HandleNewDevice func definition
func HandleNewDevice(d device.Device) (*device.Device, error) {
	du := buzzscreen.Service.DeviceUseCase

	deviceUpdated, err := du.UpsertDevice(d)
	if err != nil {
		return nil, err
	}

	isNewDevice := false
	if deviceProfile, _ := du.GetProfile(deviceUpdated.ID); deviceProfile == nil || deviceProfile.RegisteredSeconds == nil {
		isNewDevice = true
		seconds := deviceUpdated.CreatedAt.Unix()
		if err := du.SaveProfile(device.Profile{
			ID:                deviceUpdated.ID,
			RegisteredSeconds: &seconds,
		}); err != nil {
			return nil, err
		}
	}

	if app.IsHoneyscreenAppID(deviceUpdated.AppID) == false && isNewDevice {
		pipeline := buzzscreen.Service.Redis.Pipeline()
		dateString := datetime.GetDate("2006-01-02", "Asia/Tokyo")
		pipeline.Incr(fmt.Sprintf("stat:device:nru:total:%s", dateString))
		pipeline.SetBit(fmt.Sprintf("stat:device:profile:unit:%d", deviceUpdated.AppID), deviceUpdated.ID, 1) //TODO: newDevice 일때만 하도록 추후 변경해야 함
		pipeline.Exec()
	}

	return deviceUpdated, nil
}

// LogDevice func definition
func LogDevice(device *device.Device, country *string, appVersion *int, isInBatteryOpts *bool, IsBackgroundRestricted *bool, HasOverlayPermission *bool) {
	deviceForLog := DeviceForLog{
		Device:                 *device,
		AppVersion:             appVersion,
		DateJoined:             device.CreatedAt.Format("2006-01-02T15:04:05"),
		UnitID:                 device.AppID,
		Message:                "device",
		Country:                country,
		IsInBatteryOpts:        isInBatteryOpts,
		IsBackgroundRestricted: IsBackgroundRestricted,
		HasOverlayPermission:   HasOverlayPermission,
	}
	// Log to device logger
	if mapForLog, err := deviceForLog.BuildMap(); err == nil {
		core.Loggers["device"].WithFields(mapForLog).Info("Log")
	}
}

func logDeviceUnitRegisterSeconds(unitID int64, appID int64, apiPath string, profile *device.Profile) {
	registerSecondsLogObj := map[string]interface{}{
		"log_type":         "unit_register_seconds",
		"unit_id":          unitID,
		"app_id":           appID,
		"api_path":         apiPath,
		"message":          "general",
		"devicd_id":        profile.ID,
		"register_seconds": (*profile.UnitRegisteredSeconds)[unitID],
	}
	core.Loggers["general"].WithFields(registerSecondsLogObj).Info("Log")
}

// CheckHMAC func definition
func CheckHMAC(key []byte, plain, encrypt string) bool {
	hmacBytes := getHac(key, plain)
	return hex.EncodeToString(hmacBytes) == encrypt || base64.StdEncoding.EncodeToString(hmacBytes) == encrypt
}

func getHac(key []byte, plain string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(plain))
	return h.Sum(nil)
}

// GetDeviceByID returns the device by ID
func GetDeviceByID(id int64) *dbdevice.Device {
	device := &dbdevice.Device{}
	buzzscreen.Service.DB.Where(&dbdevice.Device{ID: id}).First(device)
	return device
}

// UpdateProfileUnitRegisterSeconds update profile's unit register seconds if this is the first time the device is using the unit
func UpdateProfileUnitRegisterSeconds(profile *device.Profile, unitID int64, appID int64, apiPath string) {
	isNewUnitRegisterSecondsSet := false
	// This if statement is for backward compatiblity, which is for giving accurate unit register seconds
	// to devices that have actually used the units before adoption of unit register second record,
	// on the unit and its app has the same id value
	// Use of an app imples use of a unit with the same id value if such unit exists.
	// This if statement will be removed if this backward compatiblity is no longer needed.

	urAppliedSince, _ := time.Parse("2006-Jan-02", "2020-Oct-01") // Actual time is earlier, but let's be safe
	if appID == unitID && *profile.RegisteredSeconds < urAppliedSince.Unix() {
		isNewUnitRegisterSecondsSet = profile.SetSpecificUnitRegisteredSecondIfEmpty(unitID, *profile.RegisteredSeconds)
	} else {
		isNewUnitRegisterSecondsSet = profile.SetUnitRegisteredSecondIfEmpty(unitID)
	}

	if isNewUnitRegisterSecondsSet {
		err := buzzscreen.Service.DeviceUseCase.SaveProfileUnitRegisteredSeconds(*profile)
		if err != nil {
			core.Logger.WithError(err).Errorf("%s - Unit Register Seconds Save err: %s", apiPath, err)
		} else {
			logDeviceUnitRegisterSeconds(unitID, appID, apiPath, profile)
		}

		// Old returning device, who has been registered before there were unit level registration time recording on DB
		// The device cannot be counted as NRU
		if appID == unitID && *profile.RegisteredSeconds < urAppliedSince.Unix() {
			return
		}
		pipeline := buzzscreen.Service.Redis.Pipeline()
		dateString := datetime.GetDate("2006-01-02", "Asia/Tokyo")
		pipeline.HIncrBy(fmt.Sprintf("stat:device:nru:%d", unitID), dateString, 1)
		pipeline.Exec()
	}

}
