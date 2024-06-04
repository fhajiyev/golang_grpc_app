package dto

import (
	"context"
	"net/http"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
)

type (
	// ContentAllocV1Request type definition
	/* ifa:069e6a97-b341-43a0-b9fc-df2556f06a25
	unit_id:419318955785795
	unit_device_token:1b66a654cd0d4c4fb471f4fb02b65015
	device_id:1092
	device_name:SHV-E210S
	year_of_birth:1985
	sex:M
	carrier:KT
	region:서울시 관악구
	sdk_version:1000
	*/
	ContentAllocV1Request struct {
		AppVersionCode       int           `form:"app_version_code" query:"app_version_code"`
		AppVersionName       string        `form:"app_version_name" query:"app_version_name"`
		Carrier              string        `form:"carrier" query:"carrier"`
		Categories           string        `form:"categories" query:"categories"`
		ClientIPReq          string        `form:"client_ip" query:"client_ip"`
		CountryReq           string        `form:"country" query:"country"`
		CustomTarget1        string        `form:"custom_target_1" query:"custom_target_1"`
		CustomTarget2        string        `form:"custom_target_2" query:"custom_target_2"`
		CustomTarget3        string        `form:"custom_target_3" query:"custom_target_3"`
		DeviceID             int64         `form:"device_id" query:"device_id" validate:"required"`
		DeviceName           string        `form:"device_name" query:"device_name"`
		DeviceOs             int           `form:"device_os" query:"device_os"`
		FilterCategories     string        `form:"filter_categories" query:"filter_categories"`
		FilterChannelIDs     string        `form:"filter_channel_ids" query:"filter_channel_ids"`
		IFA                  string        `form:"ifa" query:"ifa"`
		Language             string        `form:"language" query:"language"`
		NetworkType          string        `form:"network_type" query:"network_type"`
		Package              string        `form:"package" query:"package"`
		Region               string        `form:"region" query:"region"`
		RegisteredSecondsReq int64         `form:"registered_date" query:"registered_date"`
		Request              *http.Request `form:"-"`
		AppIDReq             int64         `form:"app_id" query:"app_id"`
		UnitIDReq            int64         `form:"unit_id" query:"unit_id"`
		UnitDeviceToken      string        `form:"unit_device_token" query:"unit_device_token" validate:"required"`
		SdkVersion           int           `form:"sdk_version" query:"sdk_version"`
		Gender               string        `form:"sex" query:"sex"`
		YearOfBirthReq       int           `form:"year_of_birth" query:"year_of_birth"`
		IsAllocationTest     int           `form:"is_allocation_test" query:"is_allocation_test"`
		IsInBatteryOpts      bool          `form:"is_in_battery_optimizations" query:"is_in_battery_optimizations"`

		registeredSeconds *int64
		creativeType      *string
		localTime         *time.Time
		yearOfBirth       *int
		targetAge         *int
		clientIP          string
		country           string
		unit              *app.Unit
		appID             *int64

		deviceCategoriesScores *map[string]float64
		deviceEntityScores     *map[string]float64

		modelArtifact *string `form:"model_artifact" query:"model_artifact"`

		isDebugScore bool

		dynamoProfile  *device.Profile
		dynamoActivity *device.Activity
	}

	// ContentAllocV1Response type definition
	ContentAllocV1Response struct {
		Campaigns []*CampaignV1 `json:"campaigns"`
		Message   string        `json:"msg"`
		Code      int           `json:"code"`
	}

	// ContentChannelsV1Request type definition
	ContentChannelsV1Request struct {
		IDs string `form:"ids" query:"ids" validate:"required"`
	}

	// ContentChannelsV1Response type definition
	ContentChannelsV1Response struct {
		Status   string                           `json:"status"`
		Channels map[string]*model.ContentChannel `json:"channels"`
	}

	// GetDeviceConfigV1Request type definition
	GetDeviceConfigV1Request struct {
		DeviceID int64 `form:"device_id" query:"device_id" validate:"required"`
	}

	// PostDeviceConfigV1Request type definition
	PostDeviceConfigV1Request struct {
		DeviceID int64  `form:"device_id" query:"device_id" validate:"required"`
		Type     string `form:"type" query:"type" validate:"required"`
		Config   string `form:"config" query:"config" validate:"required"`
	}

	// DeviceConfigV1Response type definition
	DeviceConfigV1Response struct {
		Status   string `json:"status"`
		DeviceID int64  `json:"device_id"`
		Channel  string `json:"channel"`
		Category string `json:"category"`
		Campaign string `json:"campaign"`
	}
)

// GetRegisteredSeconds func definition
func (car *ContentAllocV1Request) GetRegisteredSeconds() int64 {
	if car.registeredSeconds == nil {
		if car.RegisteredSecondsReq > 0 {
			car.registeredSeconds = &car.RegisteredSecondsReq
		} else if profile := car.GetDynamoProfile(); profile != nil && profile.RegisteredSeconds != nil {
			car.registeredSeconds = profile.RegisteredSeconds
		} else {
			now := time.Now().Unix()
			car.registeredSeconds = &now
		}
	}
	return *car.registeredSeconds
}

// GetCreativeType func definition
func (car *ContentAllocV1Request) GetCreativeType(ctx context.Context) string {
	if car.creativeType != nil {
		return *car.creativeType
	}

	creativeType := app.PlatformAndroid
	if u := car.GetUnit(ctx); u != nil {
		// TODO replace u.IsMobile()
		switch u.Platform {
		case app.PlatformAndroid, app.PlatformIOS:
			creativeType = u.Platform
		default:
			core.Logger.Errorf("initMembers() - Platform should be defined. %v %v", car.DeviceID, u.ID)
		}
	}

	car.creativeType = &creativeType
	return *car.creativeType
}

// GetLocalTime func definition
func (car *ContentAllocV1Request) GetLocalTime(ctx context.Context) *time.Time {
	if car.localTime != nil {
		return car.localTime
	}

	if u := car.GetUnit(ctx); u != nil {
		loc, err := time.LoadLocation(u.Timezone)
		localTime := time.Now()
		if err == nil {
			localTime = localTime.In(loc)
		}
		car.localTime = &localTime
	}

	return car.localTime
}

// GetYearOfBirth func definition
func (car *ContentAllocV1Request) GetYearOfBirth() int {
	if car.yearOfBirth == nil {
		yob := car.YearOfBirthReq
		if yob > 0 {
			if 1 < yob && yob < 100 {
				yob += 1900 //악시아타에서 year_of_birth를 1985와 같은 형식이 아닌 85와 같은 형식으로 준다
			}
		}
		car.yearOfBirth = &yob
	}
	return *car.yearOfBirth
}

// GetTargetAge func definition
func (car *ContentAllocV1Request) GetTargetAge() int {
	if car.targetAge == nil {
		targetAge := time.Now().Year() - car.GetYearOfBirth() - 1
		if targetAge < 0 {
			targetAge = 0
		}
		car.targetAge = &targetAge
	}
	return *car.targetAge
}

// GetClientIP func definition
func (car *ContentAllocV1Request) GetClientIP() string {
	if car.clientIP == "" {
		if car.ClientIPReq != "" {
			car.clientIP = car.ClientIPReq
		} else if car.Request != nil {
			car.clientIP = utils.GetClientIP(car.Request)
		}
	}
	return car.clientIP
}

// GetCountry func definition
func (car *ContentAllocV1Request) GetCountry(ctx context.Context) string {
	if car.country != "" {
		return car.country
	}

	if car.CountryReq != "" {
		car.country = car.CountryReq
		return car.country
	}

	if u := car.GetUnit(ctx); u != nil {
		car.country = u.Country
		return car.country
	}

	// 아직도 없으면 IP 에서
	location := buzzscreen.Service.LocationUseCase.GetClientLocation(car.Request, "ZZ")
	if location != nil {
		car.country = location.Country
	}
	return car.country
}

// GetDynamoProfile func definition
func (car *ContentAllocV1Request) GetDynamoProfile() *device.Profile {
	if car.dynamoProfile == nil {
		var err error
		car.dynamoProfile, err = buzzscreen.Service.DeviceUseCase.GetProfile(car.DeviceID)
		if err != nil {
			core.Logger.Errorf("ContentAllocV1Request - Device %d GetProfile error %v", car.DeviceID, err)
		}
	}
	return car.dynamoProfile
}

// GetDynamoActivity func definition
func (car *ContentAllocV1Request) GetDynamoActivity() *device.Activity {
	if car.dynamoActivity == nil {
		car.dynamoActivity, _ = buzzscreen.Service.DeviceUseCase.GetActivity(car.DeviceID)
	}
	return car.dynamoActivity
}

// GetAppID func definition
func (car *ContentAllocV1Request) GetAppID(ctx context.Context) int64 {
	if car.appID != nil {
		return *car.appID
	}
	if car.AppIDReq > 0 {
		car.appID = &car.AppIDReq
		return car.AppIDReq
	}
	car.appID = &car.GetUnit(ctx).AppID
	return *car.appID
}

// GetUnit func definition
func (car *ContentAllocV1Request) GetUnit(ctx context.Context) *app.Unit {
	if car.unit != nil {
		return car.unit
	}
	appUseCase := buzzscreen.Service.AppUseCase
	if car.UnitIDReq > 0 {
		u, err := appUseCase.GetUnitByID(ctx, car.UnitIDReq)
		if err != nil {
			core.Logger.Errorf("GetUnit() - failed to get unit with unt_id: %d", car.UnitIDReq)
		}
		if u != nil {
			car.unit = u
			return car.unit
		}
	}
	if car.AppIDReq > 0 {
		u, err := appUseCase.GetUnitByAppIDAndType(ctx, car.AppIDReq, app.UnitTypeLockscreen)
		if err != nil {
			core.Logger.Errorf("GetUnit() - failed to get unit with app_id: %d", car.AppIDReq)
		}
		if u != nil {
			car.unit = u
			return car.unit
		}
	}
	return nil
}

// GetOrganizationID func definition
func (car *ContentAllocV1Request) GetOrganizationID(ctx context.Context) int64 {
	return car.GetUnit(ctx).OrganizationID
}

// GetCategoriesScores func definition
func (car *ContentAllocV1Request) GetCategoriesScores() *map[string]float64 {
	if car.deviceCategoriesScores == nil {
		dp := car.GetDynamoProfile()
		var scores *map[string]float64
		if dp == nil || dp.CategoriesScores == nil {
			return nil
		}
		scores = dp.CategoriesScores
		car.deviceCategoriesScores = scores
	}
	return car.deviceCategoriesScores
}

// GetEntityScores func definition
func (car *ContentAllocV1Request) GetEntityScores() *map[string]float64 {
	if car.deviceEntityScores == nil {
		dp := car.GetDynamoProfile()
		var scores *map[string]float64
		if dp == nil || dp.EntityScores == nil {
			return nil
		}
		scores = dp.EntityScores
		car.deviceEntityScores = scores
	}
	return car.deviceEntityScores
}

// GetModelArtifact func definition
func (car *ContentAllocV1Request) GetModelArtifact(ctx context.Context) *string {
	if car.modelArtifact == nil {
		defaultModelArtifact := "v4_a"
		if car.GetOrganizationID(ctx) == 148 {
			defaultModelArtifact = "v1"
		}
		dp := car.GetDynamoProfile()
		if dp == nil || dp.ModelArtifact == nil {
			car.modelArtifact = &defaultModelArtifact
		} else {
			car.modelArtifact = dp.ModelArtifact
		}
	}
	return car.modelArtifact
}

// GetIsDebugScore func definition
func (car *ContentAllocV1Request) GetIsDebugScore() bool {
	dp := car.GetDynamoProfile()
	if dp == nil {
		return false
	}
	car.isDebugScore = dp.IsDebugScore
	return car.isDebugScore
}
