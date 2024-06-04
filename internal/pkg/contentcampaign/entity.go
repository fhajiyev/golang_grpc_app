package contentcampaign

import "time"

// LandingType type definition
type LandingType int

const (
	_ = iota
	// LandingTypeBrowser const definition
	LandingTypeBrowser LandingType = 1
	// LandingTypeOverlay const definition
	LandingTypeOverlay LandingType = 2
	// LandingTypeCard const definition
	LandingTypeCard LandingType = 3
	// LandingTypeYoutube const definition
	LandingTypeYoutube LandingType = 8 | LandingTypeOverlay
)

// Status type definition
type Status int

// Status constants
const (
	_                      = iota
	StatusManual    Status = 1
	StatusCuratable Status = 2
	StatusSelected  Status = 3
	StatusAccepted  Status = 4
	StatusRejected  Status = 5
	StatusComplete  Status = 6
	StatusFeedOnly  Status = 7
)

// StatusesForLockscreen var definition
var StatusesForLockscreen = []Status{StatusManual, StatusComplete}

// StatusesForFeed var definition
var StatusesForFeed = []Status{StatusCuratable, StatusSelected, StatusAccepted, StatusFeedOnly}

// ContentCampaign type definition
type ContentCampaign struct {
	ID             int64
	Categories     string
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
	LandingReward  int
	LandingType    LandingType
	Name           string
	OrganizationID int64
	OwnerID        int64
	ProviderID     *int64
	PublishedAt    *time.Time
	StartDate      time.Time
	Status         Status
	Tags           string

	TargetApp                 string
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
	WeekSlot  string

	ExtraData *map[string]interface{} `gorm:"-"`
}
