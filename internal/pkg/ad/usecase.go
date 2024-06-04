package ad

import "time"

// UseCase interface definition
type UseCase interface {
	GetAdDetail(id int64) (*Detail, error)
	GetAdsV1(v1Req V1AdsRequest) (*V1AdsResponse, error)
	GetAdsV2(v2Req V2AdsRequest) (*V2AdsResponse, error)
	LogAdAllocationRequestV1(appID int64, v1Req V1AdsRequest)
	LogAdAllocationRequestV2(appID int64, v2Req V2AdsRequest)
}

type useCase struct {
	baUserRepo BAUserRepository
	adRepo     Repository
	logger     StructuredLogger
}

// GetAdDetail returns detail of ad if it's being requested by staff user
func (u *useCase) GetAdDetail(id int64) (*Detail, error) {
	baUser, err := u.baUserRepo.GetBAUserByID(staffBAUserID)
	if err != nil {
		return nil, err
	}

	return u.adRepo.GetAdDetail(id, baUser.AccessToken)
}

// GetAdsV1 request to ad server
func (u *useCase) GetAdsV1(v1Req V1AdsRequest) (*V1AdsResponse, error) {
	return u.adRepo.GetAdsV1(v1Req)
}

// GetAdsV2 request to ad server
func (u *useCase) GetAdsV2(v2Req V2AdsRequest) (*V2AdsResponse, error) {
	return u.adRepo.GetAdsV2(v2Req)
}

func (u *useCase) LogAdAllocationRequestV1(appID int64, v1Req V1AdsRequest) {
	m := map[string]interface{}{
		"type":               "allocation_request",
		"api_version":        "v1",
		"app_id":             appID,
		"unit_id":            v1Req.UnitID,
		"device_id":          v1Req.AccountID,
		"event_at":           time.Now().Unix(),
		"ifa":                v1Req.IFA,
		"is_allocation_test": v1Req.IsAllocationTest,
		"user_id":            v1Req.UnitDeviceToken,

		"os_version":      v1Req.DeviceOS,
		"device_name":     v1Req.DeviceName,
		"membership_days": v1Req.MembershipDays,
		"carrier":         v1Req.Carrier,
		"package_name":    v1Req.PackageName,
		"user_agent":      v1Req.UserAgent,
		"sdk_version":     v1Req.SDKVersion,
		"gender":          v1Req.Sex,
		"mcc":             v1Req.MCC,
		"mnc":             v1Req.MNC,

		"client_ip": v1Req.ClientIP,
		"region":    v1Req.Region,
		"timezone":  v1Req.TimeZone,

		"network_type":                    v1Req.NetworkType,
		"revenue_types":                   v1Req.RevenueTypes,
		"creative_size":                   v1Req.CreativeSize,
		"support_remove_after_impression": v1Req.SupportRemoveAfterImpression,
	}

	if v1Req.YearOfBirth != nil {
		m["year_of_birth"] = *v1Req.YearOfBirth
	}
	if v1Req.IsIFALimitAdTrackingEnabled != nil {
		m["is_ifa_limit_ad_tracking_enabled"] = *v1Req.IsIFALimitAdTrackingEnabled
	}
	if v1Req.InstalledBrowsers != nil {
		m["installed_browsers"] = *v1Req.InstalledBrowsers
	}
	if v1Req.DefaultBrowser != nil {
		m["default_browser"] = *v1Req.DefaultBrowser
	}
	if v1Req.IsTest != nil {
		m["is_test"] = *v1Req.IsTest
	}

	u.logger.Log(m)
}

func (u *useCase) LogAdAllocationRequestV2(appID int64, v2Req V2AdsRequest) {
	m := map[string]interface{}{
		"type":        "allocation_request",
		"api_version": "v2",
		"app_id":      appID,
		"unit_id":     v2Req.UnitID,
		"device_id":   v2Req.AccountID,
		"event_at":    time.Now().Unix(),
		"ifa":         v2Req.IFA,

		"device_name":     v2Req.DeviceName,
		"sdk_version":     v2Req.SdkVersion,
		"os_version":      v2Req.OsVersion,
		"user_agent":      v2Req.UserAgent,
		"relationship":    v2Req.Relationship,
		"membership_days": v2Req.MembershipDays,
		"gender":          v2Req.Gender,
		"language":        v2Req.Language,

		"country":      v2Req.Country,
		"timezone":     v2Req.Timezone,
		"client_ip":    v2Req.ClientIP,
		"network_type": v2Req.NetworkType,

		"creative_size":    v2Req.CreativeSize,
		"platform":         v2Req.Platform,
		"creative_types":   v2Req.Types,
		"lineitem_ids":     v2Req.LineitemIds,
		"is_mock_response": v2Req.IsMockResponse,
		"target_fill":      v2Req.TargetFill,
	}

	if v2Req.UnitDeviceToken != nil {
		m["user_id"] = *v2Req.UnitDeviceToken
	}
	if v2Req.AndroidID != nil {
		m["android_id"] = *v2Req.AndroidID
	}
	if v2Req.Birthday != nil {
		m["birthday"] = *v2Req.Birthday
	}
	if v2Req.MCC != nil {
		m["mcc"] = *v2Req.MCC
	}
	if v2Req.MNC != nil {
		m["mnc"] = *v2Req.MNC
	}
	if v2Req.Latitude != nil {
		m["latitude"] = *v2Req.Latitude
	}
	if v2Req.Longitude != nil {
		m["longitude"] = *v2Req.Longitude
	}
	if v2Req.RevenueTypes != nil {
		m["revenue_types"] = *v2Req.RevenueTypes
	}
	if v2Req.CPSCategory != nil {
		m["cps_category"] = *v2Req.CPSCategory
	}

	u.logger.Log(m)
}

// NewUseCase returns UseCase interface
func NewUseCase(baUserRepo BAUserRepository, adRepo Repository, logger StructuredLogger) UseCase {
	return &useCase{baUserRepo, adRepo, logger}
}
