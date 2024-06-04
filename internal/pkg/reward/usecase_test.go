package reward_test

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (ts *UseCaseTestSuite) Test_ValidateRequest() {
	ts.Run("success", func() {
		req := ts.buildRequestIngredients()
		err := ts.useCase.ValidateRequest(req)
		ts.NoError(err)
	})

	ts.Run("upper-case checksum", func() {
		req := ts.buildRequestIngredients()
		req.Checksum = strings.ToUpper(req.Checksum)
		err := ts.useCase.ValidateRequest(req)
		ts.NoError(err)
	})
}

func (ts *UseCaseTestSuite) buildRequestIngredients() reward.RequestIngredients {
	req := reward.RequestIngredients{}
	ts.NoError(faker.FakeData(&req))
	req.Checksum = ts.buildChecksum(req)
	return req
}

func (ts *UseCaseTestSuite) buildChecksum(req reward.RequestIngredients) string {
	source := fmt.Sprintf("buz:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v:%v",
		req.DeviceID, req.IFA, req.UnitDeviceToken, req.AppID,
		req.CampaignID, req.CampaignType, req.CampaignName, *req.CampaignOwnerID, req.CampaignIsMedia,
		req.Slot, req.Reward, req.BaseReward)
	return reward.GetMD5Hash(source)
}

func (ts *UseCaseTestSuite) Test_ValidateRequest_Invalid() {
	ingr := reward.RequestIngredients{}
	ts.NoError(faker.FakeData(&ingr))

	err := ts.useCase.ValidateRequest(ingr)

	ts.Error(err)
}

func (ts *UseCaseTestSuite) Test_GiveReward() {
	ingr := reward.RequestIngredients{}
	ts.NoError(faker.FakeData(&ingr))
	ingr.Reward += 1
	ts.repo.On("Save", mock.Anything).Return(ingr.Reward, nil).Once()

	rewardReceived, err := ts.useCase.GiveReward(ingr)

	ts.NoError(err)
	ts.Equal(ingr.Reward, rewardReceived)
}

func (ts *UseCaseTestSuite) Test_GiveZeroReward() {
	ingr := reward.RequestIngredients{}
	ts.NoError(faker.FakeData(&ingr))
	ingr.Reward = 0
	ingr.BaseReward = 0

	rewardReceived, err := ts.useCase.GiveReward(ingr)

	ts.NoError(err)
	ts.Equal(ingr.Reward, rewardReceived)
}

func (ts *UseCaseTestSuite) Test_Hours() {
	period := reward.Period(3600)
	ts.Equal(int64(1), period.Hours())

	period = reward.Period(3601)
	ts.Equal(int64(2), period.Hours())

	period = reward.Period(3599)
	ts.Equal(int64(1), period.Hours())
}

func (ts *UseCaseTestSuite) Test_GetReceivedStatusMap() {
	now := time.Now().Unix()
	rand.Seed(now)
	deviceID := rand.Int63n(10000000000) + 1
	hours := 3
	n := 100

	pointsMap := ts.createPoints(deviceID, now, n, hours)
	points := make([]reward.Point, 0)
	for _, ps := range pointsMap {
		points = append(points, ps...)
	}
	periodForCampaign, receivedStatusMap := createdMaps(pointsMap, now, n, hours)
	maxPeriod := periodForCampaign.MaxPeriod()
	ts.repo.On("GetImpressionPoints", deviceID, maxPeriod).Return(points).Once()

	resReceivedStatusMap := ts.useCase.GetReceivedStatusMap(deviceID, periodForCampaign)

	for campaignID, receivedStatus := range resReceivedStatusMap {
		ts.Equal(receivedStatusMap[campaignID], receivedStatus)
	}
}

func createdMaps(pointsMap map[int64][]reward.Point, now int64, n int, hours int) (periodForCampaign reward.PeriodForCampaign, receivedStatusMap map[int64]reward.ReceivedStatus) {
	periodForCampaign = make(reward.PeriodForCampaign)
	receivedStatusMap = make(map[int64]reward.ReceivedStatus)
	for campaignID, points := range pointsMap {
		if rand.Intn(2) == 0 { // 적립내역이있는 캠페인중에 대략 절반을 receivedStatus 조회 요청할 값으로 선택
			if rand.Intn(2) == 0 { // 그중에 절반은 리워드를 받아갈 수 있는 캠페인으로 결정
				periodForCampaign[campaignID] = reward.Period(now - points[len(points)-1].CreatedAt - 300)
				receivedStatusMap[campaignID] = reward.StatusUnknown
			} else {
				periodForCampaign[campaignID] = reward.Period(now - points[len(points)-1].CreatedAt + 300)
				receivedStatusMap[campaignID] = reward.StatusReceived
			}
		}
	}

	// 대략 n/2개의 적립내역이 없는 campaign 생성
	for i := 0; i < n/2; i++ {
		campaignID := rand.Int63n(10000000000) + 1
		if _, ok := periodForCampaign[campaignID]; !ok { // 적립내역 있는 캠페인과 중복되지않도록 함
			periodForCampaign[campaignID] = reward.Period(rand.Int63n(int64(hours)*3600) / 300)
			receivedStatusMap[campaignID] = reward.StatusUnknown
		}
	}

	return periodForCampaign, receivedStatusMap
}

func (ts *UseCaseTestSuite) createPoints(deviceID int64, now int64, n int, hours int) map[int64][]reward.Point {
	pointMap := make(map[int64][]reward.Point)
	core.Logger.Infof("now %v", now)
	for i := 0; i < n; i++ {
		var p reward.Point
		err := faker.FakeData(&p)
		ts.NoError(err)

		p.CampaignID = rand.Int63n(int64(n/4)) + 1 // 한 campaignID에 평균 4개의 적립기록이 들어가도록 함
		p.DeviceID = deviceID
		p.CreatedAt = now - int64((hours*3600)/(n)*(n-i))

		// points는 createdAt의 오름차순으로 정렬되어 들어감
		if points, ok := pointMap[p.CampaignID]; ok {
			points = append(points, p)
			pointMap[p.CampaignID] = points
		} else {
			var points []reward.Point
			points = append(points, p)
			pointMap[p.CampaignID] = points
		}
	}
	return pointMap
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
	_ reward.Repository    = &mockRepo{}
)

type UseCaseTestSuite struct {
	suite.Suite
	repo    *mockRepo
	useCase reward.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = reward.NewUseCase(ts.repo)
}

func (ts *UseCaseTestSuite) AfterTest(_, _ string) {
	ts.repo.AssertExpectations(ts.T())
}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) Save(req reward.RequestIngredients) (int, error) {
	ret := r.Called(req)
	return ret.Get(0).(int), ret.Error(1)
}

func (r *mockRepo) GetImpressionPoints(deviceID int64, maxPeriod reward.Period) []reward.Point {
	ret := r.Called(deviceID, maxPeriod)
	return ret.Get(0).([]reward.Point)
}
