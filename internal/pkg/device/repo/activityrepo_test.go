package repo_test

import (
	"math/rand"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/tests"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device/repo"
	"github.com/guregu/dynamo"
	"github.com/stretchr/testify/suite"
)

const ActivityTableName = "device-activity_test"

func (ts *ActivityRepoTestSuite) TestRepo_GetByID() {
	now := time.Now()
	deviceID := rand.Int63n(10000000) + 1

	campaignID := int64(100)
	ts.createActivityRecord(deviceID, campaignID, now.Add(-(time.Minute * 1)).Unix())            // 59분 전
	ts.createActivityRecord(deviceID, campaignID, now.Add(-(time.Minute * 2)).Unix())            // 59분 전
	ts.createActivityRecord(deviceID, campaignID, now.Add(-(time.Minute * 59)).Unix())           // 59분 전
	ts.createActivityRecord(deviceID, campaignID, now.Add(-(time.Minute * 61)).Unix())           // 61분 전
	ts.createActivityRecord(deviceID, campaignID, now.Add(-(time.Hour*24 + time.Minute)).Unix()) // 24시간 1분 전
	ts.createActivityRecord(deviceID, campaignID, now.Add(-(time.Hour*48 + time.Minute)).Unix()) // 48시간 1분 전

	expectedActivity := &device.Activity{
		SeenCampaignIDs:          map[string]bool{"100": true},
		SeenCampaignCountForDay:  map[string]int{"100": 4},
		SeenCampaignCountForHour: map[string]int{"100": 3},
	}

	resultActivity, err := ts.repo.GetByID(deviceID)
	ts.NoError(err)

	ts.Equal(expectedActivity, resultActivity)
}

func (ts *ActivityRepoTestSuite) createActivityRecord(deviceID int64, campaignID int64, createdAt int64) *repo.Activity {
	activity := &repo.Activity{
		DeviceID:   deviceID,
		CreatedAt:  float64(createdAt),
		ActionType: string(device.ActivityImpression),
		CampaignID: campaignID,
		TTL:        time.Now().AddDate(0, 0, 1).Unix(),
	}

	err := ts.activityTable.Put(activity).Run()
	ts.NoError(err)
	return activity
}

func TestActivityRepoSuite(t *testing.T) {
	suite.Run(t, new(ActivityRepoTestSuite))
}

type ActivityRepoTestSuite struct {
	suite.Suite
	activityTable dynamo.Table
	server        *httptest.Server
	repo          device.ActivityRepository
}

func (ts *ActivityRepoTestSuite) SetupTest() {
	ts.server = tests.GetTestServer(nil)
	dyDB := env.GetDynamoDB()
	// Create the table
	if err := dyDB.CreateTable(ActivityTableName, repo.Activity{}).Run(); err != nil {
		dyDB.Table(ActivityTableName).DeleteTable().Run()
		if err := dyDB.CreateTable(ActivityTableName, repo.Activity{}).Run(); err != nil {
			core.Logger.Fatalf("SetupTest failed with %v", err)
		}
	}

	ts.activityTable = dyDB.Table(ActivityTableName)
	ts.repo = repo.NewActivityRepo(&ts.activityTable)
}

func (ts *ActivityRepoTestSuite) TearDownTest() {
	ts.activityTable.DeleteTable().Run()
	ts.server.Close()
}
