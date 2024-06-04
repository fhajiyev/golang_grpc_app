package repo

import (
	"strconv"
	"strings"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"
)

// EntityMapper type definition
type EntityMapper struct {
}

func (m *EntityMapper) dbAppToApp(dbApp *dbapp.App) *app.App {
	return &app.App{
		ID:               dbApp.ID,
		LatestAppVersion: dbApp.LatestAppVersion,
		IsEnabled:        dbApp.IsEnabled,
	}
}

func (m *EntityMapper) welcomeRewardConfigToEntity(wrc *dbapp.WelcomeRewardConfig) app.WelcomeRewardConfig {
	return app.WelcomeRewardConfig{
		ID:            wrc.ID,
		UnitID:        wrc.UnitID,
		Amount:        wrc.Amount,
		Country:       wrc.Country,
		Name:          wrc.Name,
		RetentionDays: wrc.RetentionDays,
		StartTime:     wrc.StartTime,
		EndTime:       wrc.EndTime,
	}
}

func (m *EntityMapper) referralRewardConfigToEntity(rrc *dbapp.ReferralRewardConfig) app.ReferralRewardConfig {
	return app.ReferralRewardConfig{
		AppID:               rrc.AppID,
		Enabled:             rrc.Enabled,
		Amount:              rrc.Amount,
		MaxReferral:         rrc.MaxReferral,
		StartDate:           rrc.StartDate,
		EndDate:             rrc.EndDate,
		VerifyURL:           rrc.VerifyURL,
		TitleForReferee:     rrc.TitleForReferee,
		TitleForReferrer:    rrc.TitleForReferrer,
		TitleForMaxReferrer: rrc.TitleForMaxReferrer,
		ExpireHours:         rrc.ExpireHours,
		MinSdkVersion:       rrc.MinSdkVersion,
	}
}

func (m *EntityMapper) unitTypeToDBUnitType(t app.UnitType) dbapp.UnitType {
	switch t {
	case app.UnitTypeLockscreen:
		return dbapp.UnitTypeLockscreen
	case app.UnitTypeNative:
		return dbapp.UnitTypeNative
	case app.UnitTypeBenefitNative:
		return dbapp.UnitTypeBenefitNative
	case app.UnitTypeBenefitFeed:
		return dbapp.UnitTypeBenefitFeed
	case app.UnitTypeBenefitInterstitial:
		return dbapp.UnitTypeBenefitInterstitial
	case app.UnitTypeBenefitPop:
		return dbapp.UnitTypeBenefitPop
	case app.UnitTypeAdapterRewardedVideo:
		return dbapp.UnitTypeAdapterRewardedVideo
	default:
		return dbapp.UnitTypeUnknown
	}
}

func (m *EntityMapper) dbUnitTypeToUnitType(dbt dbapp.UnitType) app.UnitType {
	switch dbt {
	case dbapp.UnitTypeLockscreen:
		return app.UnitTypeLockscreen
	case dbapp.UnitTypeNative:
		return app.UnitTypeNative
	case dbapp.UnitTypeBenefitNative:
		return app.UnitTypeBenefitNative
	case dbapp.UnitTypeBenefitFeed:
		return app.UnitTypeBenefitFeed
	case dbapp.UnitTypeBenefitInterstitial:
		return app.UnitTypeBenefitInterstitial
	case dbapp.UnitTypeBenefitPop:
		return app.UnitTypeBenefitPop
	case dbapp.UnitTypeAdapterRewardedVideo:
		return app.UnitTypeAdapterRewardedVideo
	default:
		return app.UnitTypeUnknown
	}
}

func (m *EntityMapper) dbUnitIsActiveToUnitIsActive(da dbapp.IsActive) bool {
	switch da {
	case dbapp.Active:
		return true
	case dbapp.Inactive:
		return false
	default:
		return true
	}
}

func (m *EntityMapper) ratioStrToMap(ratioStr string, defRatio app.AdContentRatio) app.AdContentRatio {
	if ratioStr != "" {
		ratio := strings.Split(ratioStr, ":")
		ad, _ := strconv.Atoi(ratio[0])
		content, _ := strconv.Atoi(ratio[1])
		return app.NewAdContentRatio(ad, content)
	}
	return defRatio
}

// UnitToEntity func definition
func (m *EntityMapper) UnitToEntity(dbUnit *dbapp.Unit) app.Unit {
	u := app.Unit{
		ID:                   dbUnit.ID,
		AppID:                dbUnit.AppID,
		BuzzadUnitID:         dbUnit.BuzzadUnitID,
		BuzzvilLandingReward: dbUnit.BuzzvilLandingReward,
		Country:              dbUnit.Country,
		FeedRatio:            m.ratioStrToMap(dbUnit.FeedRatio, app.AdContentRatio{"ad": 1, "content": 5}),
		FirstScreenRatio:     m.ratioStrToMap(dbUnit.FirstScreenRatio, app.AdContentRatio{"ad": 9, "content": 1}),
		FilteredProviders:    dbUnit.FilteredProviders,
		InitHMACKey:          dbUnit.InitHMACKey,
		Platform:             dbUnit.Platform,
		Timezone:             dbUnit.Timezone,
		OrganizationID:       dbUnit.OrganizationID,
		PagerRatio:           m.ratioStrToMap(dbUnit.PagerRatio, app.AdContentRatio{"ad": 3, "content": 1}),
		ShuffleOption:        dbUnit.ShuffleOption,
		UnitType:             m.dbUnitTypeToUnitType(dbUnit.UnitType),
		PostbackURL:          dbUnit.PostbackURL,
		PostbackAESIv:        dbUnit.PostbackAESIv,
		PostbackAESKey:       dbUnit.PostbackAESKey,
		PostbackHeaders:      dbUnit.PostbackHeaders,
		PostbackHMACKey:      dbUnit.PostbackHMACKey,
		PostbackParams:       dbUnit.PostbackParams,
		PostbackClass:        dbUnit.PostbackClass,
		PostbackConfig:       dbUnit.PostbackConfig,
		IsActive:             m.dbUnitIsActiveToUnitIsActive(dbUnit.IsActive),
	}

	if dbUnit.BaseInitPeriod != nil {
		u.BaseInitPeriod = *dbUnit.BaseInitPeriod
	} else {
		u.BaseInitPeriod = 3600
	}

	if dbUnit.BaseReward != nil {
		u.BaseReward = *dbUnit.BaseReward
	} else {
		u.BaseReward = 2
	}

	if dbUnit.PageLimit != nil {
		u.PageLimit = *dbUnit.PageLimit
	} else {
		u.PageLimit = 20
	}

	if dbUnit.AdType != nil {
		u.AdType = app.AdType(*dbUnit.AdType)
	} else {
		u.AdType = app.AdTypeAll
	}

	if dbUnit.ContentType != nil {
		u.ContentType = app.ContentType(*dbUnit.ContentType)
	} else {
		u.ContentType = app.ContentTypeAll
	}

	return u
}
