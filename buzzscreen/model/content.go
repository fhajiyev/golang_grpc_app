package model

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/utils"
	"github.com/jinzhu/gorm"
)

type (
	// ContentCampaign type definition
	// TODO apply cap convention (IPU, DIPU, TIPU, TargetSDK, TargetOS, ...)
	ContentCampaign struct {
		ID             int64       `gorm:"primary_key" json:"id"`
		Categories     string      `gorm:"type:varchar(100)" json:"categories" esNull:"__GLOB__"`
		ChannelID      *int64      `json:"channel_id"`
		CleanMode      int         `json:"clean_mode"` //get
		ClickURL       string      `json:"click_url"`
		CleanLink      string      `json:"clean_link"`
		CreatedAt      string      `json:"created_at"`
		Country        string      `json:"country"`
		Description    string      `json:"description"`
		DisplayType    string      `json:"display_type"`
		DisplayWeight  int         `json:"display_weight"`
		EndDate        string      `json:"end_date"`
		Ipu            *int        `json:"ipu"` //get("ipu", 9999)
		Dipu           *int        `json:"dipu"`
		IsCtrFilterOff bool        `json:"is_ctr_filter_off"`
		IsEnabled      bool        `json:"is_enabled"`
		Image          string      `json:"-"`
		JSON           string      `json:"json"`
		LandingReward  int         `json:"landing_reward"`
		LandingType    LandingType `json:"landing_type"` //LANDING_TYPE_RESPONSE_MAPPING ...
		Name           string      `json:"name"`
		OrganizationID int64       `json:"organization_id"`
		OwnerID        int64       `json:"owner_id"`
		ProviderID     int64       `json:"provider_id"`
		PublishedAt    string      `json:"published_at"`
		StartDate      string      `json:"start_date"`
		Status         Status      `json:"status"`
		Tags           string      `json:"-"  esNull:"__GLOB__"`

		TargetApp                 string `gorm:"type:varchar(100)" json:"target_app" esNull:"__GLOB__"` //get("target_app")
		TargetAgeMin              int    `json:"target_age_min"`
		TargetAgeMax              int    `json:"target_age_max"`
		TargetSdkMin              int    `json:"target_sdk_min"`
		TargetSdkMax              int    `json:"target_sdk_max"`
		RegisteredDaysMin         int    `json:"registered_days_min"`
		RegisteredDaysMax         int    `json:"registered_days_max"`
		TargetGender              string `json:"target_gender" esNull:"__GLOB__"`
		TargetLanguage            string `json:"target_language" esNull:"__GLOB__"`
		TargetCarrier             string `json:"target_carrier" esNull:"__GLOB__"`
		TargetRegion              string `json:"target_region" esNull:"__GLOB__"`
		CustomTarget1             string `json:"custom_target_1" esNull:"__GLOB__"`
		CustomTarget2             string `json:"custom_target_2" esNull:"__GLOB__"`
		CustomTarget3             string `json:"custom_target_3" esNull:"__GLOB__"`
		TargetUnit                string `json:"target_unit" esNull:"__GLOB__"`
		TargetAppID               string `json:"target_app_id" esNull:"__GLOB__"`
		TargetOrg                 string `json:"target_org" esNull:"__GLOB__"`
		TargetOsMin               int    `json:"target_os_min"`
		TargetOsMax               int    `json:"target_os_max"`
		TargetBatteryOptimization bool   `json:"target_battery_optimization"`

		Title     string `json:"title"`
		Timezone  string `json:"timezone"`
		Tipu      int    `json:"tipu"` //get("tipu", 9999)
		Type      string `json:"type"`
		UpdatedAt string `json:"updated_at"`
		WeekSlot  string `gorm:"type:varchar(1000)" json:"week_slot" esNull:"__GLOB__"`

		extraData *map[string]interface{} `gorm:"-"`
	}

	// ContentCategory type definition
	ContentCategory struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		//Deprecated
		Translation string  `json:"translation"`
		IconURL     *string `json:"icon_url,omitempty"`
	}

	// ContentCategories type definition
	ContentCategories []*ContentCategory

	// ContentChannel type definition
	ContentChannel struct {
		ID        int64  `gorm:"primary_key" json:"id"`
		Category  string `json:"category"`
		Logo      string `json:"icon_url"`
		Name      string `json:"name"`
		Publisher string `json:"-"`
	}

	// ContentChannels type definition
	ContentChannels []*ContentChannel

	// ContentProvider type definition
	ContentProvider struct {
		ID         int64 `gorm:"primary_key" json:"id"`
		Publisher  string
		Enabled    string
		Name       string
		Country    string
		Categories string
		ChannelID  int64
	}
)

// GetExtraData func definition
func (cc *ContentCampaign) GetExtraData() *map[string]interface{} {
	if cc.extraData == nil {
		data := make(map[string]interface{})
		json.Unmarshal([]byte(cc.JSON), &data)
		cc.extraData = &data
	}
	return cc.extraData
}

// TableName func definition
func (ContentCampaign) TableName() string {
	return "content_campaigns"
}

// TableName func definition
func (ContentChannel) TableName() string {
	return "content_channels"
}

// TableName func definition
func (ContentProvider) TableName() string {
	return "content_providers"
}

// TableName func definition
func (DeviceContentConfig) TableName() string {
	return "device_content_config"
}

// IsEmpty func definition
func (devicePrefs *DeviceContentConfig) IsEmpty() bool {
	return devicePrefs.Channel == "" && devicePrefs.Campaign == ""
}

// Save func definition
func (devicePrefs *DeviceContentConfig) Save() {
	buzzscreen.Service.DB.Save(devicePrefs)
}

// ContentChannelsQuery struct definition
type ContentChannelsQuery struct {
	query *gorm.DB
}

// NewContentChannelsQuery func definition
func NewContentChannelsQuery() *ContentChannelsQuery {
	return &ContentChannelsQuery{}
}

// WithCountryAndCategoryID func definition
func (c *ContentChannelsQuery) WithCountryAndCategoryID(country, categoryID string) *ContentChannelsQuery {
	c.query = buzzscreen.Service.DB.Select("distinct *").Where(fmt.Sprintf("id in (select channel_id from content_providers where country in ('%s') and enabled = 'Y')", country))
	if categoryID != "" {
		c.query = c.query.Where("category = ?", categoryID).Order("name asc")
	}
	return c
}

// WithIDs func definition
func (c *ContentChannelsQuery) WithIDs(ids string) *ContentChannelsQuery {
	strIDs := strings.Split(ids, ",")
	intIDs, err := utils.SliceAtoi(strIDs)
	if err != nil {
		panic(err)
	}

	if c.query == nil {
		c.query = buzzscreen.Service.DB
	}

	c.query = c.query.Where("id IN (?)", intIDs).Order(fmt.Sprintf("FIELD(id, %s)", strings.Join(strIDs, ",")))
	return c
}

// Get func definition
func (c *ContentChannelsQuery) Get() *ContentChannels {
	var channels ContentChannels
	err := c.query.Find(&channels).Error
	if err != nil {
		core.Logger.WithError(err).Errorf("ContentChannelsQuery.Get() - error: %s", err)
	}
	return &channels
}
