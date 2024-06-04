package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/bxcodec/faker"

	"github.com/stretchr/testify/mock"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/stretchr/testify/suite"
)

func (ts *UseCaseTestSuite) Test_GetRewardableWelcomeRewardConfig() {
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now().AddDate(0, 0, +1)
	var mockWRC *app.WelcomeRewardConfig
	err := faker.FakeData(&mockWRC)
	ts.NoError(err)
	mockWRC.StartTime = &startTime
	mockWRC.EndTime = &endTime
	mockWRC.Amount = 1
	country := ""
	mockWRC.Country = &country
	mockWRC.RetentionDays = 0

	mockWRCs := app.WelcomeRewardConfigs{*mockWRC}

	ts.repo.On("GetRewardingWelcomeRewardConfigs", mock.Anything).Return(mockWRCs, nil).Once()
	result, err := ts.useCase.GetRewardableWelcomeRewardConfig(context.Background(), mockWRC.UnitID, "", time.Now().Unix())
	ts.NoError(err)
	ts.Equal(mockWRC.ID, result.ID)
	ts.Equal(mockWRC.Amount, result.Amount)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_GetRewardableWelcomeRewardConfig_InfiniteEnddate(){
	startTime := time.Now().AddDate(0, 0, -1)
	var mockWRC *app.WelcomeRewardConfig
	err := faker.FakeData(&mockWRC)
	ts.NoError(err)
	mockWRC.StartTime = &startTime
	mockWRC.EndTime = nil
	mockWRC.Amount = 2
	country := ""
	mockWRC.Country = &country
	mockWRC.RetentionDays = 0

	mockWRCs := app.WelcomeRewardConfigs{*mockWRC}

	ts.repo.On("GetRewardingWelcomeRewardConfigs", mock.Anything).Return(mockWRCs, nil).Once()
	result, err := ts.useCase.GetRewardableWelcomeRewardConfig(context.Background(), mockWRC.UnitID, "", time.Now().Unix())
	ts.NoError(err)
	ts.Equal(mockWRC.ID, result.ID)
	ts.Equal(mockWRC.Amount, result.Amount)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_GetActiveWelcomeRewardConfigs() {
	var mockWRC *app.WelcomeRewardConfig
	err := faker.FakeData(&mockWRC)
	ts.NoError(err)
	startTime := time.Now().AddDate(0, 0, -1)
	endTime := time.Now().AddDate(0, 0, +1)
	mockWRC.StartTime = &startTime
	mockWRC.EndTime = &endTime
	mockWRC.Amount = 1
	country := ""
	mockWRC.Country = &country
	mockWRC.RetentionDays = 0

	mockWRCs := app.WelcomeRewardConfigs{*mockWRC}

	ts.repo.On("GetRewardingWelcomeRewardConfigs", mock.Anything).Return(mockWRCs, nil).Once()
	result, err := ts.useCase.GetActiveWelcomeRewardConfigs(context.Background(), mockWRC.UnitID)
	ts.NoError(err)
	ts.Equal(mockWRC.ID, result[0].ID)
	ts.Equal(mockWRC.Amount, result[0].Amount)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_GetReferralRewardConfig() {
	var mockRRC *app.ReferralRewardConfig
	ts.NoError(faker.FakeData(&mockRRC))
	mockRRC.AppID++

	ts.repo.On("GetReferralRewardConfig", mockRRC.AppID).Return(mockRRC, nil).Once()

	result, err := ts.useCase.GetReferralRewardConfig(context.Background(), mockRRC.AppID)
	ts.NoError(err)
	ts.equalRRC(mockRRC, result)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_GetUnitByID() {
	mockUnit := &app.Unit{
		ID: 1,
	}
	ts.repo.On("GetUnitByID", mock.Anything).Return(mockUnit, nil).Once()

	result, _ := ts.useCase.GetUnitByID(context.Background(), 1234)
	ts.Equal(mockUnit.ID, result.ID)
	ts.repo.AssertExpectations(ts.T())
}
func (ts *UseCaseTestSuite) Test_GetUnitByAppID() {
	mockUnit := &app.Unit{
		ID:    1,
		AppID: 2,
	}
	ts.repo.On("GetUnitByAppID", mock.Anything).Return(mockUnit, nil).Once()

	result, _ := ts.useCase.GetUnitByAppID(context.Background(), 1234)

	ts.Equal(mockUnit.AppID, result.AppID)
	ts.repo.AssertExpectations(ts.T())
}
func (ts *UseCaseTestSuite) Test_GetUnitByAppIDAndType() {
	mockUnit := &app.Unit{
		ID:       1,
		AppID:    2,
		UnitType: app.UnitTypeLockscreen,
	}
	ts.repo.On("GetUnitByAppIDAndType", mock.Anything, mock.Anything).Return(mockUnit, nil).Once()

	result, _ := ts.useCase.GetUnitByAppIDAndType(context.Background(), 1234, app.UnitTypeLockscreen)
	ts.Equal(mockUnit.AppID, result.AppID)
	ts.Equal(mockUnit.UnitType, result.UnitType)
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) equalRRC(config1 *app.ReferralRewardConfig, config2 *app.ReferralRewardConfig) {
	ts.True(config1.AppID == config2.AppID)
	ts.True(config1.Enabled == config2.Enabled)
	ts.True(config1.Amount == config2.Amount)
	ts.True(config1.MaxReferral == config2.MaxReferral)
	ts.True(*config1.StartDate == *config2.StartDate)
	ts.True(*config1.EndDate == *config2.EndDate)
	ts.True(config1.VerifyURL == config2.VerifyURL)
	ts.True(config1.TitleForReferee == config2.TitleForReferee)
	ts.True(config1.TitleForReferrer == config2.TitleForReferrer)
	ts.True(config1.TitleForMaxReferrer == config2.TitleForMaxReferrer)
	ts.True(config1.ExpireHours == config2.ExpireHours)
	ts.True(config1.MinSdkVersion == config2.MinSdkVersion)
}

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

var (
	_ suite.SetupTestSuite = &UseCaseTestSuite{}
)

type UseCaseTestSuite struct {
	suite.Suite
	repo    *mockRepo
	useCase app.UseCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = app.NewUseCase(ts.repo)
}

var _ app.Repository = &mockRepo{}

type mockRepo struct {
	mock.Mock
}

func (ds *mockRepo) GetAppByID(ctx context.Context, appID int64) (*app.App, error) {
	ret := ds.Called(appID)
	return ret.Get(0).(*app.App), ret.Error(1)
}

func (r *mockRepo) GetRewardingWelcomeRewardConfigs(ctx context.Context, unitID int64) (app.WelcomeRewardConfigs, error) {
	ret := r.Called(unitID)
	return ret.Get(0).(app.WelcomeRewardConfigs), ret.Error(1)
}

func (r *mockRepo) GetReferralRewardConfig(ctx context.Context, appID int64) (*app.ReferralRewardConfig, error) {
	ret := r.Called(appID)
	return ret.Get(0).(*app.ReferralRewardConfig), ret.Error(1)
}

func (r *mockRepo) GetUnitByID(ctx context.Context, unitId int64) (*app.Unit, error) {
	ret := r.Called(unitId)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}
func (r *mockRepo) GetUnitByAppID(ctx context.Context, appID int64) (*app.Unit, error) {
	ret := r.Called(appID)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}
func (r *mockRepo) GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType app.UnitType) (*app.Unit, error) {
	ret := r.Called(appID, unitType)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}
