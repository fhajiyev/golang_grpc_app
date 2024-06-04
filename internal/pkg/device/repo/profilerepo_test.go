package repo_test

import (
	"encoding/json"
	"math/rand"
	"net/http/httptest"
	"strconv"
	"time"

	// "testing"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device/repo"
	"github.com/guregu/dynamo"
	"github.com/stretchr/testify/suite"
)

const ProfileTableName = "device-profile_test"

// copied from "internal/pkg/device/profilerepo.go"
const (
	keyID                = "did"
	keyRegisteredSeconds = "rd"
	keyScoredCampaigns   = "sc"
	keyCategoriesScores  = "cs"
	keyEntityScores      = "es"
	keyModelArtifact     = "ma"
	keyPackageName       = "pn"
	keyInstalledPackages = "ip"
	keyDailyActiveUser   = "dau"
	keyIsDebugScore      = "ids"
)

func (ts *ProfileRepoTestSuite) TestRepo_GetByID() {
	now := time.Now()
	deviceID := rand.Int63n(10000000) + 1
	deviceRegisteredSecond := now.Add(-(time.Minute * 61)).Unix()
	deviceModelArtifact := "9_z"
	devicePackageName := "package2"
	deviceInstalledPackages := "bs.com"
	expectedProfile := &device.Profile{
		ID:                deviceID,
		RegisteredSeconds: &deviceRegisteredSecond,
		ScoredCampaigns:   &map[int]int{3444: 222, 95736: 0, 894783: 45},
		CategoriesScores:  &map[string]float64{"aa": 17.3, "dseww": 1.0},
		EntityScores:      &map[string]float64{"john": 0.91, "bob": 2.3},
		ModelArtifact:     &deviceModelArtifact,
		PackageName:       &devicePackageName,
		InstalledPackages: &deviceInstalledPackages,
		IsDau:             true,
		IsDebugScore:      true,
	}

	ts.createProfileRecord(deviceID, keyRegisteredSeconds, strconv.FormatInt(*(expectedProfile.RegisteredSeconds), 10))
	scoredCampaignsPv, err := json.Marshal(expectedProfile.ScoredCampaigns)
	ts.NoError(err)
	ts.createProfileRecord(deviceID, keyScoredCampaigns, string(scoredCampaignsPv))
	categorieScoresPv, err := json.Marshal(expectedProfile.CategoriesScores)
	ts.NoError(err)
	ts.createProfileRecord(deviceID, keyCategoriesScores, string(categorieScoresPv))
	entityScoresPv, err := json.Marshal(expectedProfile.EntityScores)
	ts.NoError(err)
	ts.createProfileRecord(deviceID, keyEntityScores, string(entityScoresPv))
	ts.createProfileRecord(deviceID, keyModelArtifact, *expectedProfile.ModelArtifact)
	ts.createProfileRecord(deviceID, keyPackageName, *expectedProfile.PackageName)
	ts.createProfileRecord(deviceID, keyInstalledPackages, *expectedProfile.InstalledPackages)
	isDauPv, err := json.Marshal(expectedProfile.IsDau)
	ts.NoError(err)
	ts.createProfileRecord(deviceID, keyDailyActiveUser, string(isDauPv))
	isDebugScorePv, err := json.Marshal(expectedProfile.IsDebugScore)
	ts.NoError(err)
	ts.createProfileRecord(deviceID, keyIsDebugScore, string(isDebugScorePv))

	resultProfile, err := ts.repo.GetByID(deviceID)
	ts.NoError(err)

	ts.Equal(expectedProfile.ID, resultProfile.ID)
	ts.Equal(*expectedProfile.RegisteredSeconds, *resultProfile.RegisteredSeconds)
	ts.Equal(*expectedProfile.ScoredCampaigns, *resultProfile.ScoredCampaigns)
	ts.Equal(*expectedProfile.CategoriesScores, *resultProfile.CategoriesScores)
	ts.Equal(*expectedProfile.EntityScores, *resultProfile.EntityScores)
	ts.Equal(*expectedProfile.ModelArtifact, *resultProfile.ModelArtifact)
	ts.Equal(*expectedProfile.InstalledPackages, *resultProfile.InstalledPackages)
	ts.Equal(*expectedProfile.PackageName, *resultProfile.PackageName)
	ts.Equal(*expectedProfile.PackageName, *resultProfile.PackageName)
}

func (ts *ProfileRepoTestSuite) TestRepo_UnitRegisteredSeconds() {
	now := time.Now()
	deviceID := rand.Int63n(10000000) + 1
	deviceRegisteredSecond := now.Add(-(time.Minute * 61)).Unix()

	ts.createProfileRecord(deviceID, keyRegisteredSeconds, strconv.FormatInt(deviceRegisteredSecond, 10))
	testProfile, err := ts.repo.GetByID(deviceID)
	ts.NoError(err)

	ts.Empty(*testProfile.UnitRegisteredSeconds)

	unitID := rand.Int63n(10000000) + 1
	testProfile.SetUnitRegisteredSecondIfEmpty(unitID)
	expectedUnitRegisteredSeconds := (*testProfile.UnitRegisteredSeconds)[unitID]
	time.Sleep(1 * time.Second)
	testProfile.SetUnitRegisteredSecondIfEmpty(unitID)
	ts.Equal(expectedUnitRegisteredSeconds, (*testProfile.UnitRegisteredSeconds)[unitID])

	ts.repo.Save(*testProfile)
	updatedProfile, err := ts.repo.GetByID(deviceID)
	ts.NoError(err)
	ts.Equal(expectedUnitRegisteredSeconds, (*updatedProfile.UnitRegisteredSeconds)[unitID])
}

func (ts *ProfileRepoTestSuite) TestRepo_MultipleUnitRegisteredSeconds() {
	now := time.Now()
	deviceID := rand.Int63n(10000000) + 1
	firstUnitID := rand.Int63n(10000000) + 1
	secondUnitID := firstUnitID + 53
	deviceRegisteredSecond := now.Add(-(time.Minute * 61)).Unix()

	ts.createProfileRecord(deviceID, keyRegisteredSeconds, strconv.FormatInt(deviceRegisteredSecond, 10))
	testProfile, err := ts.repo.GetByID(deviceID)
	ts.NoError(err)

	testProfile.SetUnitRegisteredSecondIfEmpty(firstUnitID)
	firstUnitRegisteredSeconds := (*testProfile.UnitRegisteredSeconds)[firstUnitID]
	ts.repo.Save(*testProfile)
	testProfile, err = ts.repo.GetByID(deviceID)
	ts.NoError(err)

	time.Sleep(1 * time.Second)

	testProfile.SetUnitRegisteredSecondIfEmpty(secondUnitID)
	secondUnitRegisteredSeconds := (*testProfile.UnitRegisteredSeconds)[secondUnitID]
	ts.NotEqual(firstUnitRegisteredSeconds, secondUnitRegisteredSeconds)
	ts.repo.Save(*testProfile)

	updatedProfile, err := ts.repo.GetByID(deviceID)
	ts.NoError(err)
	ts.Equal(firstUnitRegisteredSeconds, (*updatedProfile.UnitRegisteredSeconds)[firstUnitID])
	ts.Equal(secondUnitRegisteredSeconds, (*updatedProfile.UnitRegisteredSeconds)[secondUnitID])
}

func (ts *ProfileRepoTestSuite) createProfileRecord(deviceID int64, profile string, profileValue string) *repo.DynamoDP {
	profileRecord := repo.DynamoDP{
		DeviceID:     deviceID,
		Profile:      profile,
		ProfileValue: profileValue,
		Timestamp:    time.Now().AddDate(0, 0, 1).Unix(),
	}

	err := ts.profileTable.Put(&profileRecord).Run()
	ts.NoError(err)
	return &profileRecord
}

//Error occurs if this test runs with activityrepo_test.go. Need to make separate packages for profilerepo.go and acitivityrepo.go
/*
func TestProfileRepoSuite(t *testing.T) {
	ts := new(ProfileRepoTestSuite)
	ts.server = tests.GetTestServer(nil)
	suite.Run(t, ts)
	ts.server.Close()
}
*/

type ProfileRepoTestSuite struct {
	suite.Suite
	profileTable dynamo.Table
	server       *httptest.Server
	repo         device.ProfileRepository
}

func (ts *ProfileRepoTestSuite) SetupTest() {
	dyDB := env.GetDynamoDB()
	// Create the table
	if err := dyDB.CreateTable(ProfileTableName, repo.DynamoDP{}).Run(); err != nil {
		dyDB.Table(ProfileTableName).DeleteTable().Run()
		if err := dyDB.CreateTable(ProfileTableName, repo.DynamoDP{}).Run(); err != nil {
			core.Logger.Fatalf("SetupTest failed with %v", err)
		}
	}

	ts.profileTable = dyDB.Table(ProfileTableName)
	ts.repo = repo.NewProfileRepo(&ts.profileTable)
}

func (ts *ProfileRepoTestSuite) TearDownTest() {
	ts.profileTable.DeleteTable().Run()
}
