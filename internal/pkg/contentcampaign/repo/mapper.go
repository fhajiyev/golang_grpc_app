package repo

import (
	"encoding/json"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbcontentcampaign"
)

type entityMapper struct {
}

func (m *entityMapper) dbCampaignToCampaign(dbContentCampaign *dbcontentcampaign.ContentCampaign) contentcampaign.ContentCampaign {
	contentCampaign := contentcampaign.ContentCampaign{
		ID:             dbContentCampaign.ID,
		Categories:     dbContentCampaign.Categories,
		ChannelID:      dbContentCampaign.ChannelID,
		CleanMode:      dbContentCampaign.CleanMode,
		ClickURL:       dbContentCampaign.ClickURL,
		CleanLink:      dbContentCampaign.CleanLink,
		CreatedAt:      dbContentCampaign.CreatedAt,
		Country:        dbContentCampaign.Country,
		Description:    dbContentCampaign.Description,
		DisplayType:    dbContentCampaign.DisplayType,
		DisplayWeight:  dbContentCampaign.DisplayWeight,
		EndDate:        dbContentCampaign.EndDate,
		Ipu:            dbContentCampaign.Ipu,
		IsCtrFilterOff: dbContentCampaign.IsCtrFilterOff,
		IsEnabled:      dbContentCampaign.IsEnabled,
		Image:          dbContentCampaign.Image,
		LandingReward:  dbContentCampaign.LandingReward,
		LandingType:    contentcampaign.LandingType(dbContentCampaign.LandingType),
		Name:           dbContentCampaign.Name,
		OrganizationID: dbContentCampaign.OrganizationID,
		OwnerID:        dbContentCampaign.OwnerID,
		ProviderID:     dbContentCampaign.ProviderID,
		PublishedAt:    dbContentCampaign.PublishedAt,
		StartDate:      dbContentCampaign.StartDate,
		Status:         contentcampaign.Status(dbContentCampaign.Status),
		Tags:           dbContentCampaign.Tags,

		TargetApp:                 dbContentCampaign.TargetApp,
		TargetAgeMin:              dbContentCampaign.TargetAgeMin,
		TargetAgeMax:              dbContentCampaign.TargetAgeMax,
		TargetSdkMin:              dbContentCampaign.TargetSdkMin,
		TargetSdkMax:              dbContentCampaign.TargetSdkMax,
		RegisteredDaysMin:         dbContentCampaign.RegisteredDaysMin,
		RegisteredDaysMax:         dbContentCampaign.RegisteredDaysMax,
		TargetGender:              dbContentCampaign.TargetGender,
		TargetLanguage:            dbContentCampaign.TargetLanguage,
		TargetCarrier:             dbContentCampaign.TargetCarrier,
		TargetRegion:              dbContentCampaign.TargetRegion,
		CustomTarget1:             dbContentCampaign.CustomTarget1,
		CustomTarget2:             dbContentCampaign.CustomTarget2,
		CustomTarget3:             dbContentCampaign.CustomTarget3,
		TargetUnit:                dbContentCampaign.TargetUnit,
		TargetAppID:               dbContentCampaign.TargetAppID,
		TargetOrg:                 dbContentCampaign.TargetOrg,
		TargetOsMin:               dbContentCampaign.TargetOsMin,
		TargetOsMax:               dbContentCampaign.TargetOsMax,
		TargetBatteryOptimization: dbContentCampaign.TargetBatteryOptimization,

		Title:     dbContentCampaign.Title,
		Timezone:  dbContentCampaign.Timezone,
		Tipu:      dbContentCampaign.Tipu,
		Type:      dbContentCampaign.Type,
		UpdatedAt: dbContentCampaign.UpdatedAt,
		WeekSlot:  dbContentCampaign.WeekSlot,
	}

	data := make(map[string]interface{})
	json.Unmarshal([]byte(dbContentCampaign.JSON), &data)
	contentCampaign.ExtraData = &data

	return contentCampaign
}
