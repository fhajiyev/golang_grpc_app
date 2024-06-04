package appsvc_test

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/bxcodec/faker"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/appsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (ts *ControllerTestSuite) Test_GetApp() {
	var mockApp app.App
	ts.NoError(faker.FakeData(&mockApp))
	ts.useCase.On("GetAppByID", mockApp.ID).Return(&mockApp, nil).Once()
	req := (&network.Request{
		Method: http.MethodGet,
		URL:    "/api/apps/:appID",
		Params: &url.Values{},
	}).Build()
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("appID")
	ctx.SetParamValues(strconv.FormatInt(mockApp.ID, 10))

	err := ts.controller.GetApp(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal([]byte(rec.Body.String()), &res)
	ts.Equal(int64(res["id"].(float64)), mockApp.ID)

	ts.useCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_GetUnit() {
	var mockUnit app.Unit
	ts.NoError(faker.FakeData(&mockUnit))
	ts.useCase.On("GetUnitByID", mockUnit.ID).Return(&mockUnit, nil).Once()
	req := (&network.Request{
		Method: http.MethodGet,
		URL:    "/api/units/:unitID",
		Params: &url.Values{},
	}).Build()
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("unitID")
	ctx.SetParamValues(strconv.FormatInt(mockUnit.ID, 10))

	err := ts.controller.GetUnit(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal([]byte(rec.Body.String()), &res)
	ts.Equal(int64(res["id"].(float64)), mockUnit.ID)
	settings := res["settings"].(map[string]interface{})
	ts.Equal(int(settings["base_reward"].(float64)), mockUnit.BaseReward)
	ts.Equal(int(settings["page_limit"].(float64)), mockUnit.PageLimit)

	ts.useCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_GetUnitConfigs() {
	var mockUnit app.Unit
	ts.NoError(faker.FakeData(&mockUnit))

	ts.useCase.On("GetUnitByID", mockUnit.ID).Return(&mockUnit, nil).Once()

	req := (&network.Request{
		Header: &http.Header{"Authorization": []string{os.Getenv("BASIC_AUTHORIZATION_VALUE")}},
		Method: http.MethodGet,
		URL:    "/api/units/:unitID/configs",
		Params: &url.Values{},
	}).Build()

	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("unitID")
	ctx.SetParamValues(strconv.FormatInt(mockUnit.ID, 10))

	err := ts.controller.GetUnitConfigs(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal([]byte(rec.Body.String()), &res)
	ts.NoError(err)
	log.Printf("%+v", res)
	postback, ok := res["postback"].(map[string]interface{})
	ts.True(ok)
	ts.Equal(int64(res["organization_id"].(float64)), mockUnit.OrganizationID)
	ts.Equal(postback["url"].(string), mockUnit.PostbackURL)
	ts.Equal(postback["aes_iv"].(string), mockUnit.PostbackAESIv)
	ts.Equal(postback["aes_key"].(string), mockUnit.PostbackAESKey)
	ts.Equal(postback["headers"].(string), mockUnit.PostbackHeaders)
	ts.Equal(postback["hmac_key"].(string), mockUnit.PostbackHMACKey)
	ts.Equal(postback["params"].(string), mockUnit.PostbackParams)
	ts.Equal(postback["class"].(string), mockUnit.PostbackClass)
	ts.Equal(postback["config"].(string), mockUnit.PostbackConfig)
	ts.useCase.AssertExpectations(ts.T())
}
func (ts *ControllerTestSuite) Test_GetPromotions() {
	mockWRC := &app.WelcomeRewardConfig{
		ID:     1,
		UnitID: 1234,
		Amount: 500,
	}
	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now().Add(time.Hour * 23)
	mockWRC.StartTime = &startTime
	mockWRC.EndTime = &endTime

	var mockWRCs app.WelcomeRewardConfigs
	mockWRCs = append(mockWRCs, *mockWRC)
	ts.useCase.On("GetActiveWelcomeRewardConfigs", mockWRC.UnitID).Return(mockWRCs, nil).Once()
	req := (&network.Request{
		Method: http.MethodGet,
		URL:    "/api/apps/:appID/promotions",
		Params: &url.Values{},
	}).Build()
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("appID")
	ctx.SetParamValues(strconv.FormatInt(mockWRC.UnitID, 10))

	err := ts.controller.GetPromotions(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)

	var res map[string]interface{}
	err = json.Unmarshal([]byte(rec.Body.String()), &res)
	welcomePromotion := res["promotions"].([]interface{})[0].(map[string]interface{})
	ts.Equal(int(welcomePromotion["amount"].(float64)), mockWRC.Amount)
	ts.Equal(welcomePromotion["type"].(string), "welcome")

	ts.useCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (ctx core.Context, rec *httptest.ResponseRecorder) {
	rec = httptest.NewRecorder()
	ctx = ts.engine.NewContext(httpRequest, rec)
	return
}

type ControllerTestSuite struct {
	suite.Suite
	controller appsvc.Controller
	engine     *core.Engine
	useCase    *mockUseCase
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.engine = echo.New()
	ts.useCase = new(mockUseCase)
	ts.controller = appsvc.NewController(ts.engine, ts.useCase)
}

var _ app.UseCase = &mockUseCase{}

type mockUseCase struct {
	mock.Mock
}

func (u *mockUseCase) GetAppByID(ctx context.Context, appID int64) (*app.App, error) {
	ret := u.Called(appID)
	return ret.Get(0).(*app.App), ret.Error(1)
}

func (u *mockUseCase) GetRewardableWelcomeRewardConfig(ctx context.Context, unitID int64, country string, unitRegisterSeconds int64) (*app.WelcomeRewardConfig, error) {
	ret := u.Called(unitID, country, unitRegisterSeconds)
	return ret.Get(0).(*app.WelcomeRewardConfig), ret.Error(1)
}

func (u *mockUseCase) GetActiveWelcomeRewardConfigs(ctx context.Context, unitID int64) (app.WelcomeRewardConfigs, error) {
	ret := u.Called(unitID)
	return ret.Get(0).(app.WelcomeRewardConfigs), ret.Error(1)
}

func (u *mockUseCase) GetReferralRewardConfig(ctx context.Context, appID int64) (*app.ReferralRewardConfig, error) {
	ret := u.Called(appID)
	return ret.Get(0).(*app.ReferralRewardConfig), ret.Error(1)
}

func (u *mockUseCase) GetUnitByID(ctx context.Context, unitID int64) (*app.Unit, error) {
	ret := u.Called(unitID)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}

func (u *mockUseCase) GetUnitByAppID(ctx context.Context, appID int64) (*app.Unit, error) {
	ret := u.Called(appID)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}

func (u *mockUseCase) GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType app.UnitType) (*app.Unit, error) {
	ret := u.Called(appID, unitType)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}
