package ad_test

import (
	"math/rand"
	"testing"

	"github.com/bxcodec/faker"

	"github.com/stretchr/testify/mock"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/stretchr/testify/suite"
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

func (ts *UseCaseTestSuite) TestLogAdAllocationRequestV1() {
	appID := rand.Int63n(10000000) + 1
	unitID := rand.Int63n(10000000) + 1
	deviceID := rand.Int63n(10000000) + 1
	installedBrwosers := "internet_explorer,mosaic"

	adsReqV1 := &ad.V1AdsRequest{AccountID: deviceID, UnitID: unitID, InstalledBrowsers: &installedBrwosers}

	logValidator := ts.validateLogAdAllocationRequestV1(appID, adsReqV1)
	ts.structuredLogger.On("Log", mock.MatchedBy(logValidator)).Once()

	ts.useCase.LogAdAllocationRequestV1(appID, *adsReqV1)
}

func (ts *UseCaseTestSuite) TestLogAdAllocationRequestV2() {
	appID := rand.Int63n(10000000) + 1
	unitID := rand.Int63n(10000000) + 1
	deviceID := rand.Int63n(10000000) + 1
	androidID := "Jellybean"

	adsReqV2 := &ad.V2AdsRequest{AccountID: deviceID, UnitID: unitID, AndroidID: &androidID}

	logValidator := ts.validateLogAdAllocationRequestV2(appID, adsReqV2)
	ts.structuredLogger.On("Log", mock.MatchedBy(logValidator)).Once()

	ts.useCase.LogAdAllocationRequestV2(appID, *adsReqV2)
}

func (ts *UseCaseTestSuite) Test_GetBAUserByID() {
	baUser := &ad.BAUser{}
	ts.NoError(faker.FakeData(&baUser))
	baUser.ID = 1
	detail := ts.createAdDetail()

	ts.baUserRepo.On("GetBAUserByID", baUser.ID).Return(baUser, nil).Once()
	ts.repo.On("GetAdDetail", detail.ID, baUser.AccessToken).Return(detail, nil).Once()

	resultDetail, err := ts.useCase.GetAdDetail(detail.ID)
	ts.NoError(err)
	ts.equalAdDetail(detail, resultDetail)
	ts.baUserRepo.AssertExpectations(ts.T())
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) TestGetAdsV1() {
	adsReq := &ad.V1AdsRequest{}
	ts.NoError(faker.FakeData(&adsReq))

	adsRes := ad.V1AdsResponse{}

	ts.repo.On("GetAdsV1", *adsReq).Return(&adsRes, nil).Once()

	result, err := ts.useCase.GetAdsV1(*adsReq)
	ts.NoError(err)
	ts.Equal(adsRes, *result)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) TestGetAdsV2() {
	adsReq := &ad.V2AdsRequest{}
	ts.NoError(faker.FakeData(&adsReq))

	adsRes := ad.V2AdsResponse{}

	ts.repo.On("GetAdsV2", *adsReq).Return(&adsRes, nil).Once()

	result, err := ts.useCase.GetAdsV2(*adsReq)
	ts.NoError(err)
	ts.Equal(adsRes, *result)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) validateLogAdAllocationRequestV1(appID int64, adsReqV1 *ad.V1AdsRequest) func(map[string]interface{}) bool {
	return func(m map[string]interface{}) bool {
		ts.Equal("allocation_request", m["type"])
		ts.Equal("v1", m["api_version"])
		ts.Equal(appID, m["app_id"])
		ts.Equal(adsReqV1.UnitID, m["unit_id"])
		ts.Equal(adsReqV1.AccountID, m["device_id"])
		if adsReqV1.DefaultBrowser != nil {
			ts.Equal(*adsReqV1.DefaultBrowser, m["default_browser"])
		} else {
			ts.NotContains(m, "default_browser")
		}
		if adsReqV1.InstalledBrowsers != nil {
			ts.Equal(*adsReqV1.InstalledBrowsers, m["installed_browsers"])
		} else {
			ts.NotContains(m, "installed_browsers")
		}

		return true
	}
}

func (ts *UseCaseTestSuite) validateLogAdAllocationRequestV2(appID int64, adsReqV2 *ad.V2AdsRequest) func(map[string]interface{}) bool {
	return func(m map[string]interface{}) bool {
		ts.Equal("allocation_request", m["type"])
		ts.Equal("v2", m["api_version"])
		ts.Equal(appID, m["app_id"])
		ts.Equal(adsReqV2.UnitID, m["unit_id"])
		ts.Equal(adsReqV2.AccountID, m["device_id"])
		ts.Equal(adsReqV2.TargetFill, m["target_fill"])
		if adsReqV2.AndroidID != nil {
			ts.Equal(*adsReqV2.AndroidID, m["android_id"])
		} else {
			ts.NotContains(m, "android_id")
		}
		if adsReqV2.CPSCategory != nil {
			ts.Equal(*adsReqV2.CPSCategory, m["cps_category"])
		} else {
			ts.NotContains(m, "cps_category")
		}

		return true
	}
}

func (ts *UseCaseTestSuite) createAdDetail() *ad.Detail {
	return &ad.Detail{
		ID:             rand.Int63n(100000) + 1,
		RevenueType:    "cpc",
		OrganizationID: rand.Int63n(100000) + 1,
		Name:           "TEST_AD_NAME",
		Extra: map[string]interface{}{
			"unit": rand.Int63n(100000) + 1,
		},
	}
}

func (ts *UseCaseTestSuite) equalAdDetail(expected *ad.Detail, actual *ad.Detail) {
	ts.Equal(expected.ID, actual.ID)
	ts.Equal(expected.Name, actual.Name)
	ts.Equal(expected.OrganizationID, actual.OrganizationID)
	ts.Equal(expected.RevenueType, actual.RevenueType)
}

type UseCaseTestSuite struct {
	suite.Suite
	baUserRepo       *mockBAUserRepo
	repo             *mockRepo
	structuredLogger *mockStructuredLogger
	useCase          ad.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.baUserRepo = new(mockBAUserRepo)
	ts.repo = new(mockRepo)
	ts.structuredLogger = new(mockStructuredLogger)
	ts.useCase = ad.NewUseCase(ts.baUserRepo, ts.repo, ts.structuredLogger)
}

var _ ad.BAUserRepository = &mockBAUserRepo{}
var _ ad.Repository = &mockRepo{}

type mockBAUserRepo struct {
	mock.Mock
}

func (r *mockBAUserRepo) GetBAUserByID(id int64) (*ad.BAUser, error) {
	ret := r.Called(id)
	return ret.Get(0).(*ad.BAUser), ret.Error(1)
}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) GetAdDetail(id int64, accessToken string) (*ad.Detail, error) {
	ret := r.Called(id, accessToken)
	return ret.Get(0).(*ad.Detail), ret.Error(1)
}

func (r *mockRepo) GetAdsV1(v1Req ad.V1AdsRequest) (*ad.V1AdsResponse, error) {
	ret := r.Called(v1Req)
	return ret.Get(0).(*ad.V1AdsResponse), ret.Error(1)
}

func (r *mockRepo) GetAdsV2(v2Req ad.V2AdsRequest) (*ad.V2AdsResponse, error) {
	ret := r.Called(v2Req)
	return ret.Get(0).(*ad.V2AdsResponse), ret.Error(1)
}

type mockStructuredLogger struct {
	mock.Mock
}

func (l *mockStructuredLogger) Log(m map[string]interface{}) {
	l.Called(m)
}
