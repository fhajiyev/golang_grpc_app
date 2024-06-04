package dbcontentcampaign

import "time"

// ContentCampaign model definition
type ContentCampaign struct {
	ID             int64  `gorm:"primary_key"`
	Categories     string `gorm:"type:varchar(100)"`
	ChannelID      *int64
	CleanMode      int
	ClickURL       string
	CleanLink      string
	CreatedAt      time.Time
	Country        string
	Description    string
	DisplayType    string
	DisplayWeight  int
	EndDate        time.Time
	Ipu            *int
	IsCtrFilterOff bool
	IsEnabled      bool
	Image          string
	JSON           string
	LandingReward  int
	LandingType    int
	Name           string
	OrganizationID int64
	OwnerID        int64
	ProviderID     *int64
	PublishedAt    *time.Time
	StartDate      time.Time
	Status         int
	Tags           string

	TargetApp                 string `gorm:"type:varchar(100)"`
	TargetAgeMin              *int
	TargetAgeMax              *int
	TargetSdkMin              *int
	TargetSdkMax              *int
	RegisteredDaysMin         *int
	RegisteredDaysMax         *int
	TargetGender              string
	TargetLanguage            string
	TargetCarrier             string
	TargetRegion              string
	CustomTarget1             string
	CustomTarget2             string
	CustomTarget3             string
	TargetUnit                string
	TargetAppID               string
	TargetOrg                 string
	TargetOsMin               *int
	TargetOsMax               *int
	TargetBatteryOptimization bool

	Title     string
	Timezone  string
	Tipu      *int
	Type      string
	UpdatedAt time.Time
	WeekSlot  string `gorm:"type:varchar(1000)"`
}

// TableName returns name of content campaign table
func (ContentCampaign) TableName() string {
	return "content_campaigns"
}
