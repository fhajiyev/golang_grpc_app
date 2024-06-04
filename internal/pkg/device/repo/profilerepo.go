package repo

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/guregu/dynamo"
)

const (
	keyID                          = "did"
	keyRange                       = "profile"
	keyRegisteredSeconds           = "rd"
	keyScoredCampaigns             = "sc"
	keyCategoriesScores            = "cs"
	keyEntityScores                = "es"
	keyModelArtifact               = "ma"
	keyPackageName                 = "pn"
	keyInstalledPackages           = "ip"
	keyDailyActiveUser             = "dau"
	keyIsDebugScore                = "ids"
	keyUnitRegisteredSecondsPrefix = "urd:"
)

// DynamoDP type definition
type DynamoDP struct {
	DeviceID     int64  `dynamo:"did,hash"`
	Profile      string `dynamo:"profile,range"`
	ProfileValue string `dynamo:"pv"`
	Timestamp    int64  `dynamo:"ts"`
}

// ProfileRepo type definition
type ProfileRepo struct {
	dynamoTable *dynamo.Table
}

// GetByID func definition
func (r *ProfileRepo) GetByID(deviceID int64) (*device.Profile, error) {
	unitRegisteredSecondsChanged := make(map[int64]bool)
	unitRegisteredSeconds := make(map[int64]int64)
	deviceProfile := device.Profile{
		ID:                           deviceID,
		UnitRegisteredSecondsChanged: &unitRegisteredSecondsChanged,
		UnitRegisteredSeconds:        &unitRegisteredSeconds,
	}

	var profiles []DynamoDP
	var err = r.dynamoTable.Get(keyID, deviceID).All(&profiles)
	if err != nil || len(profiles) == 0 {
		return nil, err
	}

	for idx := range profiles {
		profile := profiles[idx]
		switch profile.Profile {
		case keyRegisteredSeconds:
			rs, err := strconv.ParseInt(profile.ProfileValue, 10, 64)
			if err != nil {
				return nil, err
			} else {
				deviceProfile.RegisteredSeconds = &rs
			}
		case keyScoredCampaigns:
			scoredCamps := make(map[int]int)
			deviceProfile.ScoredCampaigns = &scoredCamps
			json.Unmarshal([]byte(profile.ProfileValue), deviceProfile.ScoredCampaigns)
			deviceProfile.ScoredCampaignsKey = &profile.Timestamp
		case keyCategoriesScores:
			catScores := make(map[string]float64)
			deviceProfile.CategoriesScores = &catScores
			json.Unmarshal([]byte(profile.ProfileValue), deviceProfile.CategoriesScores)
		case keyEntityScores:
			entScores := make(map[string]float64)
			deviceProfile.EntityScores = &entScores
			json.Unmarshal([]byte(profile.ProfileValue), deviceProfile.EntityScores)
		case keyModelArtifact:
			deviceProfile.ModelArtifact = &profile.ProfileValue
		case keyPackageName:
			deviceProfile.PackageName = &profile.ProfileValue
		case keyInstalledPackages:
			deviceProfile.InstalledPackages = &profile.ProfileValue
		case keyIsDebugScore:
			if profile.ProfileValue == "true" {
				deviceProfile.IsDebugScore = true
			}
		case keyDailyActiveUser:
			if profile.ProfileValue == "true" {
				deviceProfile.IsDau = true
			}
		}

		if strings.HasPrefix(profile.Profile, keyUnitRegisteredSecondsPrefix) {
			unitIDStr := strings.TrimPrefix(profile.Profile, keyUnitRegisteredSecondsPrefix)
			unitID, err := strconv.ParseInt(unitIDStr, 10, 64)
			if err != nil {
				return nil, err
			}
			unitRegisteredSeconds, err := strconv.ParseInt(profile.ProfileValue, 10, 64)
			if err != nil {
				return nil, err
			}
			(*deviceProfile.UnitRegisteredSeconds)[unitID] = unitRegisteredSeconds

		}
	}

	return &deviceProfile, nil
}

// Save func definition
func (r *ProfileRepo) Save(dp device.Profile) error {
	var err error
	if dp.RegisteredSeconds != nil {
		err = r.saveDynamoDP(DynamoDP{
			DeviceID:     dp.ID,
			Profile:      keyRegisteredSeconds,
			ProfileValue: strconv.FormatInt(*(dp.RegisteredSeconds), 10),
		})
	}
	if err != nil {
		return err
	}

	err = r.SaveUnitRegisteredSeconds(dp)
	if err != nil {
		return err
	}

	if dp.PackageName != nil {
		err = r.saveDynamoDP(DynamoDP{
			DeviceID:     dp.ID,
			Profile:      keyPackageName,
			ProfileValue: *dp.PackageName,
		})
	}
	if err != nil {
		return err
	}

	err = r.SavePackage(dp)
	if err != nil {
		return err
	}

	if dp.EntityScores != nil {
		jsonBytes, err := json.Marshal(&dp.EntityScores)
		jsonString := string(jsonBytes)
		if err != nil {
			return err
		}
		err = r.saveDynamoDP(DynamoDP{
			DeviceID:     dp.ID,
			Profile:      keyEntityScores,
			ProfileValue: jsonString,
		})
	}

	return err
}

// SavePackage only saves package name
func (r *ProfileRepo) SavePackage(dp device.Profile) error {
	var err error
	if dp.InstalledPackages != nil {
		err = r.saveDynamoDP(DynamoDP{
			DeviceID:     dp.ID,
			Profile:      keyInstalledPackages,
			ProfileValue: *dp.InstalledPackages,
		})
	}
	return err
}

// SaveUnitRegisteredSeconds only saves changed unit registered seconds
func (r *ProfileRepo) SaveUnitRegisteredSeconds(dp device.Profile) error {
	var err error
	if dp.UnitRegisteredSecondsChanged != nil {
		for unitID, isNew := range *dp.UnitRegisteredSecondsChanged {
			if unitRegisteredSeconds := (*dp.UnitRegisteredSeconds)[unitID]; isNew && unitRegisteredSeconds > 0 {
				err = r.saveDynamoDP(DynamoDP{
					DeviceID:     dp.ID,
					Profile:      keyUnitRegisteredSecondsPrefix + strconv.FormatInt(unitID, 10),
					ProfileValue: strconv.FormatInt(unitRegisteredSeconds, 10),
				})

				if err != nil {
					return err
				}
			}
		}
	}
	return err
}

// Delete func definition
func (r *ProfileRepo) Delete(dp device.Profile) error {
	return r.dynamoTable.Delete(keyID, dp.ID).If("$ >= 0", keyRegisteredSeconds).Run()
}

func (r *ProfileRepo) saveDynamoDP(dp DynamoDP) error {
	if err := r.dynamoTable.Put(&dp).Run(); err != nil {
		return device.RemoteProfileError{Err: err}
	}

	return nil
}

// NewProfileRepo func definition
func NewProfileRepo(dynamoTable *dynamo.Table) *ProfileRepo {
	return &ProfileRepo{dynamoTable}
}
