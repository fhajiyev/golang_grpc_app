package rewardsvc_test

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/bxcodec/faker"

	"github.com/Buzzvil/buzzscreen-api/internal/app/api/rewardsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/rewardsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (ts *ControllerTestSuite) Test_PostReward() {
	reqIngr := ts.buildRequestIngredients()
	reqIngr.Checksum = reward.TestCheckSum
	ts.eventUseCase.On("GetTrackingURL", reqIngr.DeviceID, ts.buildResource(reqIngr)).Return("", nil).Once()
	ts.rewardUseCase.On("ValidateRequest", reqIngr).Return(nil).Once()
	ts.appUseCase.On("GetUnitByID", reqIngr.UnitID).Return(ts.createUnit(reqIngr.UnitID), nil).Once()
	ts.rewardUseCase.On("GiveReward", reqIngr).Return(reqIngr.Reward, nil).Once()
	req := ts.buildRequest(reqIngr)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())

	err := ts.controller.PostReward(ctx)
	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
}

func (ts *ControllerTestSuite) Test_PostRewardCallTrackingURL() {
	reqIngr := ts.buildRequestIngredients()
	reqIngr.Checksum = reward.TestCheckSum
	reqIngr.BaseReward = 0
	ts.rewardUseCase.On("ValidateRequest", reqIngr).Return(nil).Once()
	ts.eventUseCase.On("GetTrackingURL", reqIngr.DeviceID, ts.buildResource(reqIngr)).Return("test.url", nil).Once()
	req := ts.buildRequest(reqIngr)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())

	err := ts.controller.PostReward(ctx)
	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
}

func (ts *ControllerTestSuite) Test_PostReward_DeactivatedUnit() {
	reqIngr := ts.buildRequestIngredients()
	reqIngr.Checksum = reward.TestCheckSum
	unit := ts.createUnit(reqIngr.UnitID)
	unit.IsActive = false
	ts.eventUseCase.On("GetTrackingURL", reqIngr.DeviceID, ts.buildResource(reqIngr)).Return("", nil).Once()
	ts.rewardUseCase.On("ValidateRequest", reqIngr).Return(nil).Once()
	ts.appUseCase.On("GetUnitByID", reqIngr.UnitID).Return(unit, nil).Once()
	req := ts.buildRequest(reqIngr)
	ctx, _ := ts.buildContextAndRecorder(req.GetHTTPRequest())

	err := ts.controller.PostReward(ctx)
	ts.Error(err)
	ts.True(strings.Contains(err.Error(), fmt.Sprintf("code=%d", http.StatusBadRequest)))
}

func (ts *ControllerTestSuite) Test_PostReward_Conflict() {
	reqIngr := ts.buildRequestIngredients()
	reqIngr.Checksum = reward.TestCheckSum
	ts.eventUseCase.On("GetTrackingURL", reqIngr.DeviceID, ts.buildResource(reqIngr)).Return("", nil).Once()
	ts.rewardUseCase.On("ValidateRequest", reqIngr).Return(nil).Once()
	ts.appUseCase.On("GetUnitByID", reqIngr.UnitID).Return(ts.createUnit(reqIngr.UnitID), nil).Once()
	ts.rewardUseCase.On("GiveReward", reqIngr).Return(0, reward.DuplicatedError{}).Once()
	req := ts.buildRequest(reqIngr)
	ctx, _ := ts.buildContextAndRecorder(req.GetHTTPRequest())

	err := ts.controller.PostReward(ctx)

	ts.Error(err)
	ts.True(strings.Contains(err.Error(), fmt.Sprintf("code=%d", http.StatusConflict)))
}

func (ts *ControllerTestSuite) Test_PostReward_Unprocessable() {
	reqIngr := ts.buildRequestIngredients()
	reqIngr.Checksum = reward.TestCheckSum
	ts.eventUseCase.On("GetTrackingURL", reqIngr.DeviceID, ts.buildResource(reqIngr)).Return("", nil).Once()
	ts.rewardUseCase.On("ValidateRequest", reqIngr).Return(nil).Once()
	ts.appUseCase.On("GetUnitByID", reqIngr.UnitID).Return(ts.createUnit(reqIngr.UnitID), nil).Once()
	ts.rewardUseCase.On("GiveReward", reqIngr).Return(0, reward.UnprocessableError{}).Once()
	req := ts.buildRequest(reqIngr)
	ctx, _ := ts.buildContextAndRecorder(req.GetHTTPRequest())

	err := ts.controller.PostReward(ctx)

	ts.Error(err)
	ts.True(strings.Contains(err.Error(), fmt.Sprintf("code=%d", http.StatusUnprocessableEntity)))
}

func (ts *ControllerTestSuite) buildRequestIngredients() reward.RequestIngredients {
	ingr := reward.RequestIngredients{}
	err := faker.FakeData(&ingr)
	ingr.ClickType = "l"
	ingr.AppID = int64(rand.Intn(1000) + 1)
	ingr.DeviceID = int64(rand.Intn(1000) + 1)
	ingr.UnitID = int64(rand.Intn(1000) + 1000)
	ingr.CampaignID = int64(rand.Intn(1000) + 1000000000)
	ts.NoError(err)
	return ingr
}

func (ts *ControllerTestSuite) buildResource(req reward.RequestIngredients) event.Resource {
	if dto.BuzzAdCampaignIDOffset < req.CampaignID {
		return event.Resource{
			ID:   req.CampaignID - dto.BuzzAdCampaignIDOffset,
			Type: event.ResourceTypeAd,
		}
	}
	return event.Resource{
		ID:   req.CampaignID,
		Type: event.ResourceTypeArticle,
	}
}

func (ts *ControllerTestSuite) createUnit(unitID int64) *app.Unit {
	return &app.Unit{
		ID:       unitID,
		IsActive: true,
	}
}

func (ts *ControllerTestSuite) buildRequest(reqIngr reward.RequestIngredients) *network.Request {
	req := network.Request{
		Method: http.MethodPost,
		URL:    "/api/rewards",
		Params: &url.Values{
			"app_id":                   {strconv.FormatInt(reqIngr.AppID, 10)},
			"unit_id":                  {strconv.FormatInt(reqIngr.UnitID, 10)},
			"device_id":                {strconv.FormatInt(reqIngr.DeviceID, 10)},
			"ifa":                      {reqIngr.IFA},
			"unit_device_token":        {reqIngr.UnitDeviceToken},
			"client_unit_device_token": {"__unit_device_token__"},
			"campaign_id":              {strconv.FormatInt(reqIngr.CampaignID, 10)},
			"campaign_type":            {reqIngr.CampaignType},
			"campaign_name":            {reqIngr.CampaignName},
			"campaign_owner_id":        {*reqIngr.CampaignOwnerID},
			"campaign_is_media":        {strconv.Itoa(reqIngr.CampaignIsMedia)},
			"slot":                     {strconv.Itoa(reqIngr.Slot)},
			"reward":                   {strconv.Itoa(reqIngr.Reward)},
			"base_reward":              {strconv.Itoa(reqIngr.BaseReward)},
			"check":                    {reqIngr.Checksum},
			"type":                     {"imp"},
		},
	}
	return req.Build()
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (ctx core.Context, rec *httptest.ResponseRecorder) {
	rec = httptest.NewRecorder()
	ctx = ts.engine.NewContext(httpRequest, rec)
	return
}

type ControllerTestSuite struct {
	suite.Suite
	controller    rewardsvc.Controller
	engine        *core.Engine
	rewardUseCase *mockRewardUseCase
	appUseCase    *mockAppUseCase
	eventUseCase  *mockEventUseCase
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.engine = core.NewEngine(nil)
	ts.rewardUseCase = new(mockRewardUseCase)
	ts.appUseCase = new(mockAppUseCase)
	ts.eventUseCase = new(mockEventUseCase)
	ts.controller = rewardsvc.NewController(ts.engine, ts.appUseCase, ts.eventUseCase, ts.rewardUseCase)
}

func (ts *ControllerTestSuite) AfterTest(_, _ string) {
	ts.rewardUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
	ts.eventUseCase.AssertExpectations(ts.T())
}

type mockRewardUseCase struct {
	mock.Mock
}

func (u *mockRewardUseCase) ValidateRequest(ingredients reward.RequestIngredients) error {
	ret := u.Called(ingredients)
	return ret.Error(0)
}

func (u *mockRewardUseCase) GiveReward(ingredients reward.RequestIngredients) (int, error) {
	ret := u.Called(ingredients)
	return ret.Get(0).(int), ret.Error(1)
}

func (u *mockRewardUseCase) GetReceivedStatusMap(deviceID int64, periodForCampaign reward.PeriodForCampaign) map[int64]reward.ReceivedStatus {
	ret := u.Called(deviceID, periodForCampaign)
	return ret.Get(0).(map[int64]reward.ReceivedStatus)
}

type mockAppUseCase struct {
	mock.Mock
}

func (ds *mockAppUseCase) GetAppByID(ctx context.Context, appID int64) (*app.App, error) {
	ret := ds.Called(appID)
	return ret.Get(0).(*app.App), ret.Error(1)
}

func (u *mockAppUseCase) GetRewardableWelcomeRewardConfig(ctx context.Context, unitID int64, country string, unitRegisterSeconds int64) (*app.WelcomeRewardConfig, error) {
	ret := u.Called(unitID, country, unitRegisterSeconds)
	return ret.Get(0).(*app.WelcomeRewardConfig), ret.Error(1)
}

func (u *mockAppUseCase) GetActiveWelcomeRewardConfigs(ctx context.Context, unitID int64) (app.WelcomeRewardConfigs, error) {
	ret := u.Called(unitID)
	return ret.Get(0).(app.WelcomeRewardConfigs), ret.Error(1)
}

func (u *mockAppUseCase) GetReferralRewardConfig(ctx context.Context, appID int64) (*app.ReferralRewardConfig, error) {
	ret := u.Called(appID)
	return ret.Get(0).(*app.ReferralRewardConfig), ret.Error(1)
}

func (u *mockAppUseCase) GetUnitByID(ctx context.Context, unitID int64) (*app.Unit, error) {
	ret := u.Called(unitID)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}

func (u *mockAppUseCase) GetUnitByAppID(ctx context.Context, appID int64) (*app.Unit, error) {
	ret := u.Called(appID)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}

func (u *mockAppUseCase) GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType app.UnitType) (*app.Unit, error) {
	ret := u.Called(appID, unitType)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}

type mockEventUseCase struct {
	mock.Mock
}

func (u *mockEventUseCase) TrackEvent(handler event.MessageHandler, ingredients event.Token) error {
	ret := u.Called(handler, ingredients)
	return ret.Error(0)
}

func (u *mockEventUseCase) GetRewardStatus(token event.Token, auth header.Auth) (string, error) {
	ret := u.Called(token, auth)
	return ret.Get(0).(string), ret.Error(1)
}

func (u *mockEventUseCase) GetEventsMap(resources []event.Resource, unitID int64, auth header.Auth) (map[int64]event.Events, error) {
	ret := u.Called(resources, unitID, auth)
	return ret.Get(0).(map[int64]event.Events), ret.Error(1)
}

func (u *mockEventUseCase) GetTokenEncrypter() event.TokenEncrypter {
	ret := u.Called()
	return ret.Get(0).(event.TokenEncrypter)
}

func (u *mockEventUseCase) SaveTrackingURL(deviceID int64, resource event.Resource, trackingURL string) {
	u.Called(deviceID, resource, trackingURL)
}

func (u *mockEventUseCase) GetTrackingURL(deviceID int64, resource event.Resource) (string, error) {
	ret := u.Called(deviceID, resource)
	return ret.Get(0).(string), ret.Error(1)
}
