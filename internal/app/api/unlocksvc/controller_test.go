package unlocksvc_test

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/unlocksvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/unlocksvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (ts *ControllerTestSuite) Test_Unlock() {
	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()
	ts.rewardUseCase.On("GiveReward", mock.AnythingOfType("reward.RequestIngredients")).Return(structReq.Reward, nil).Once()

	unit := ts.createUnit(*structReq.UnitID)
	ts.appUseCase.On("GetUnitByID", *structReq.UnitID).Return(unit, nil)

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", mock.AnythingOfType("string")).Return(payloadStruct, nil).Once()
	ts.payloadUseCase.On("IsPayloadExpired", mock.AnythingOfType("*payload.Payload")).Return(false).Once()

	err = ts.controller.Unlock(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)

	result := map[string]interface{}{}
	json.Unmarshal(rec.Body.Bytes(), &result)
	ts.Equal(structReq.Reward, int(result["reward_received"].(float64)))
}

func (ts *ControllerTestSuite) Test_Unlock_NoPayload() {
	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.PayloadStr = ""
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.payloadUseCase.On("ParsePayload", structReq.PayloadStr).Return((*payload.Payload)(nil), errors.New("invalid payload")).Once()

	err := ts.controller.Unlock(ctx)

	ts.Error(err)
	ts.Equal(http.StatusOK, rec.Code)
}

func (ts *ControllerTestSuite) Test_Unlock_InvalidClickType() {
	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.ClickType = "i"
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	payloadStruct := &payload.Payload{}
	ts.NoError(faker.FakeData(&payloadStruct))
	ts.payloadUseCase.On("ParsePayload", structReq.PayloadStr).Return(payloadStruct, nil).Once()

	err := ts.controller.Unlock(ctx)

	ts.Error(err)
}

func (ts *ControllerTestSuite) Test_Unlock_DeactivatedUnit() {
	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()

	unit := ts.createUnit(*structReq.UnitID)
	unit.IsActive = false
	ts.appUseCase.On("GetUnitByID", *structReq.UnitID).Return(unit, nil)

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", mock.AnythingOfType("string")).Return(payloadStruct, nil).Once()
	ts.payloadUseCase.On("IsPayloadExpired", mock.AnythingOfType("*payload.Payload")).Return(false).Once()

	err = ts.controller.Unlock(ctx)

	ts.Error(err)
}

func (ts *ControllerTestSuite) Test_Unlock_InvalidChecksum() {
	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	payloadStruct := &payload.Payload{}
	ts.NoError(faker.FakeData(&payloadStruct))
	ts.payloadUseCase.On("ParsePayload", structReq.PayloadStr).Return(payloadStruct, nil).Once()
	ts.payloadUseCase.On("IsPayloadExpired", payloadStruct).Return(false).Once()
	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(errors.New("checksum is invalid")).Twice()

	err := ts.controller.Unlock(ctx)

	ts.Error(err)
}

func (ts *ControllerTestSuite) Test_Unlock_Expired() {
	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", mock.AnythingOfType("string")).Return(payloadStruct, nil).Once()
	ts.payloadUseCase.On("IsPayloadExpired", mock.AnythingOfType("*payload.Payload")).Return(true).Once()

	err = ts.controller.Unlock(ctx)
	ts.NoError(err)
}

func (ts *ControllerTestSuite) Test_Unlock_NoReward() {
	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.Reward = 0
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	payloadStruct := &payload.Payload{}
	ts.NoError(faker.FakeData(&payloadStruct))
	ts.payloadUseCase.On("ParsePayload", structReq.PayloadStr).Return(payloadStruct, nil).Once()

	err := ts.controller.Unlock(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
	// TODO json result
}

func (ts *ControllerTestSuite) createUnit(unitID int64) *app.Unit {
	return &app.Unit{
		ID:       unitID,
		Country:  "GB",
		IsActive: true,
	}
}

func (ts *ControllerTestSuite) buildNetworkRequest(req *dto.PostUnlockRequest) *network.Request {
	params := &url.Values{
		"unit_id":           {strconv.FormatInt(*req.UnitID, 10)},
		"device_id":         {strconv.FormatInt(req.DeviceID, 10)},
		"ifa":               {req.IFA},
		"unit_device_token": {req.UnitDeviceToken},

		"reward":      {strconv.Itoa(req.Reward)},
		"base_reward": {strconv.Itoa(req.BaseReward)},

		"click_campaign_id":       {strconv.FormatInt(req.CampaignID, 10)},
		"click_campaign_type":     {req.CampaignType},
		"click_campaign_name":     {req.CampaignName},
		"click_campaign_is_media": {strconv.Itoa(req.CampaignIsMedia)},
		"click_type":              {req.ClickType},

		"slot":                   {strconv.Itoa(req.Slot)},
		"click_campaign_payload": {req.PayloadStr},
		"check":                  {req.Checksum},
	}

	if req.AppID != nil {
		(*params)["app_id"] = []string{strconv.FormatInt(*req.AppID, 10)}
	}

	if req.CampaignOwnerID != nil {
		(*params)["campaign_owner_id"] = []string{*req.CampaignOwnerID}
	}

	if req.ExternalCampaignID != nil {
		(*params)["external_click_campaign_id"] = []string{*req.ExternalCampaignID}
	}

	if req.ExternalCampaignUnlocks != nil {
		(*params)["external_click_campaign_imps"] = []string{*req.ExternalCampaignUnlocks}
	}

	networkReq := &network.Request{
		Method: http.MethodPost,
		URL:    "/api/impression/",
		Params: params,
		Header: &http.Header{
			"User-Agent": {"Mozilla/5.0 (Linux; Android 4.2.1; en-us; Nexus 5 Build/JOP40D) AppleWebKit/535.19 (KHTML, like Gecko; googleweblight) Chrome/38.0.1025.166 Mobile Safari/535.19"},
		},
	}

	return networkReq.Build()
}

func (ts *ControllerTestSuite) buildBaseRequest() *dto.PostUnlockRequest {
	unlockRequest := &dto.PostUnlockRequest{}

	err := faker.FakeData(&unlockRequest.AppID)
	ts.NoError(err)
	err = faker.FakeData(&unlockRequest.IFA)
	ts.NoError(err)

	unlockRequest.DeviceID = int64(rand.Intn(1000) + 1)
	unitID := int64(rand.Intn(1000) + 1000)
	unlockRequest.UnitID = &unitID
	unlockRequest.CampaignID = int64(rand.Intn(1000) + 1000000)
	unlockRequest.CampaignName = "CampaignName"

	unlockRequest.Reward = 1
	unlockRequest.ClickType = "u"
	unlockRequest.UnitDeviceToken = "__unit_device_token__"
	ts.NoError(err)
	return unlockRequest
}

type ControllerTestSuite struct {
	suite.Suite
	controller     unlocksvc.Controller
	engine         *core.Engine
	rewardUseCase  *mockRewardUseCase
	appUseCase     *mockAppUseCase
	payloadUseCase *mockPayloadUseCase
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (ctx core.Context, rec *httptest.ResponseRecorder) {
	rec = httptest.NewRecorder()
	ctx = ts.engine.NewContext(httpRequest, rec)
	return
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.engine = core.NewEngine(nil)
	ts.rewardUseCase = new(mockRewardUseCase)
	ts.appUseCase = new(mockAppUseCase)
	ts.payloadUseCase = new(mockPayloadUseCase)

	ts.controller = unlocksvc.NewController(ts.engine, ts.rewardUseCase, ts.appUseCase, ts.payloadUseCase)
}

func (ts *ControllerTestSuite) AfterTest(_, _ string) {
	ts.appUseCase.AssertExpectations(ts.T())
	ts.payloadUseCase.AssertExpectations(ts.T())
	ts.rewardUseCase.AssertExpectations(ts.T())
}

var _ reward.UseCase = &mockRewardUseCase{}

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

type mockPayloadUseCase struct {
	mock.Mock
}

func (u *mockPayloadUseCase) ParsePayload(payloadString string) (*payload.Payload, error) {
	ret := u.Called(payloadString)
	return ret.Get(0).(*payload.Payload), ret.Error(1)
}

func (u *mockPayloadUseCase) BuildPayloadString(payload *payload.Payload) string {
	ret := u.Called(payload)
	return ret.Get(0).(string)
}

func (u *mockPayloadUseCase) IsPayloadExpired(payload *payload.Payload) bool {
	ret := u.Called(payload)
	return ret.Get(0).(bool)
}
