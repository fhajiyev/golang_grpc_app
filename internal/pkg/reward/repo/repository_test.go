package repo_test

import (
	"log"
	"math/rand"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/bxcodec/faker"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward/repo"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/guregu/dynamo"
	"github.com/stretchr/testify/suite"
)

const (
	pointTableName string = "point_test_repo"
)

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

func (ts *RepoTestSuite) Test_GetImpressionPoints() {
	deviceID := rand.Int63n(100000000) + 1
	maxPeriod := reward.Period(rand.Intn(10000) + 3600)
	createdAt := time.Now().Unix() - rand.Int63n(int64(maxPeriod))

	point := ts.createPoint(deviceID, createdAt)
	ts.insertPoint(point)
	entityPoint := ts.MapToEntity(point)

	resultPoints := ts.repo.GetImpressionPoints(deviceID, maxPeriod)

	ts.Equal(1, len(resultPoints))
	ts.Equal(*entityPoint, resultPoints[0])
}

func (ts *RepoTestSuite) Test_GetImpressionPoints_NoPointInRange() {
	deviceID := rand.Int63n(100000000) + 1
	maxPeriod := reward.Period(rand.Intn(10000) + 3600)
	createdAt := time.Now().Unix() - rand.Int63n(int64(maxPeriod)) - int64(maxPeriod)

	point := ts.createPoint(deviceID, createdAt)
	ts.insertPoint(point)

	resultPoints := ts.repo.GetImpressionPoints(deviceID, maxPeriod)

	ts.Equal(0, len(resultPoints))
}

func (ts *RepoTestSuite) MapToEntity(point repo.DBPoint) *reward.Point {
	campaignID, err := strconv.ParseInt(point.ReferKey, 10, 64)
	ts.NoError(err)
	return &reward.Point{
		DeviceID:   point.DeviceID,
		Version:    point.Version,
		CampaignID: campaignID,
		Type:       point.Type,
		CreatedAt:  point.CreatedAt,
	}
}

func (ts *RepoTestSuite) createPoint(deviceID int64, createAt int64) repo.DBPoint {
	p := repo.DBPoint{}
	ts.NoError(faker.FakeData(&p))
	p.ReferKey = strconv.FormatInt(rand.Int63n(1000000)+1, 10)
	p.DeviceID = deviceID
	p.CreatedAt = createAt
	p.Type = "imp"
	return p
}

func (ts *RepoTestSuite) insertPoint(p repo.DBPoint) {
	err := ts.pointTable.Put(p).Run()
	ts.NoError(err)
	log.Printf("Point %+v", p)
	log.Printf("Err %+v", err)
}

type RepoTestSuite struct {
	suite.Suite
	server     *httptest.Server
	pointTable dynamo.Table
	repo       reward.Repository
}

func (ts *RepoTestSuite) SetupSuite() {
	ts.server = tests.GetTestServer(nil)
}

func (ts *RepoTestSuite) TearDownSuite() {
	ts.server.Close()
}

func (ts *RepoTestSuite) SetupTest() {
	// TODO env can occur cycled depedency
	dynamoDB := env.GetDynamoDB()

	// Create the table
	if err := dynamoDB.CreateTable(pointTableName, repo.DBPoint{}).Run(); err != nil {
		dynamoDB.Table(pointTableName).DeleteTable().Run()
		if err := dynamoDB.CreateTable(pointTableName, repo.DBPoint{}).Run(); err != nil {
			core.Logger.Fatalf("SetupTest failed with %v", err)
		}
	}

	ts.pointTable = dynamoDB.Table(pointTableName)
	ts.repo = repo.New(&ts.pointTable)
}

func (ts *RepoTestSuite) TearDownTest() {
	ts.pointTable.DeleteTable().Run()
}
