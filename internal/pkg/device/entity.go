package device

import "time"

// Profile struct definition
type Profile struct {
	ID                           int64
	RegisteredSeconds            *int64
	ScoredCampaigns              *map[int]int
	ScoredCampaignsKey           *int64
	CategoriesScores             *map[string]float64
	EntityScores                 *map[string]float64
	ModelArtifact                *string
	PackageName                  *string
	InstalledPackages            *string
	IsDau                        bool
	IsDebugScore                 bool
	UnitRegisteredSeconds        *map[int64]int64
	UnitRegisteredSecondsChanged *map[int64]bool
}

// ActivityType type definition
type ActivityType string

// ActivityType constants
const (
	ActivityImpression ActivityType = "i"
	ActivityClick      ActivityType = "c"
)

// Activity struct definition
type Activity struct {
	SeenCampaignIDs          map[string]bool
	SeenCampaignCountForDay  map[string]int
	SeenCampaignCountForHour map[string]int
}

// ChangePackageName returns true if package name is changed
func (dp *Profile) ChangePackageName(newPackage string) bool {
	if dp.PackageName != nil && *dp.PackageName == newPackage {
		return false
	}
	dp.PackageName = &newPackage
	return true
}

// SetUnitRegisteredSecondIfEmpty returns false if unit registered second is already set
func (dp *Profile) SetUnitRegisteredSecondIfEmpty(unitID int64) bool {
	if dp.UnitRegisteredSeconds == nil {
		unitRegisteredSeconds := make(map[int64]int64)
		dp.UnitRegisteredSeconds = &unitRegisteredSeconds
	}
	if dp.UnitRegisteredSecondsChanged == nil {
		unitRegisteredSecondsChanged := make(map[int64]bool)
		dp.UnitRegisteredSecondsChanged = &unitRegisteredSecondsChanged
	}

	if _, alreadySet := (*dp.UnitRegisteredSeconds)[unitID]; alreadySet {
		return false
	}
	(*dp.UnitRegisteredSeconds)[unitID] = time.Now().Unix()
	(*dp.UnitRegisteredSecondsChanged)[unitID] = true
	return true
}

// SetSpecificUnitRegisteredSecondIfEmpty returns false if unit registered second is already set,
// set registered second with speicified time.
// This function is for backward compaitbility. May be removed in the future
func (dp *Profile) SetSpecificUnitRegisteredSecondIfEmpty(unitID int64, timestamp int64) bool {
	if dp.UnitRegisteredSeconds == nil {
		unitRegisteredSeconds := make(map[int64]int64)
		dp.UnitRegisteredSeconds = &unitRegisteredSeconds
	}
	if dp.UnitRegisteredSecondsChanged == nil {
		unitRegisteredSecondsChanged := make(map[int64]bool)
		dp.UnitRegisteredSecondsChanged = &unitRegisteredSecondsChanged
	}

	if _, alreadySet := (*dp.UnitRegisteredSeconds)[unitID]; alreadySet {
		return false
	}
	(*dp.UnitRegisteredSeconds)[unitID] = timestamp
	(*dp.UnitRegisteredSecondsChanged)[unitID] = true
	return true
}

// Device struct definition
type Device struct {
	ID              int64      `json:"id"`
	AppID           int64      `json:"app_id"`
	UnitDeviceToken string     `json:"unit_device_token"`
	IFA             string     `json:"ifa"`
	Address         *string    `json:"-"`
	Birthday        *time.Time `json:"-"`
	Carrier         *string    `json:"carrier,omitempty"`
	DeviceName      string     `json:"device_name,omitempty"`
	Resolution      string     `json:"resolution,omitempty"`
	YearOfBirth     *int       `json:"year_of_birth,omitempty"`
	SDKVersion      *int       `json:"sdk_version,omitempty"`
	Sex             *string    `json:"sex,omitempty"`
	Packages        *string    `json:"-"`
	PackageName     *string    `json:"package_name,omitempty"`
	SignupIP        int64      `json:"signup_ip"`
	SerialNumber    *string    `json:"serial_number,omitempty"`
	CreatedAt       time.Time  `json:"-"`
	UpdatedAt       time.Time  `json:"-"`
}

// Params struct definition
type Params struct {
	AppID     int64
	PubUserID string
	IFA       string
}
