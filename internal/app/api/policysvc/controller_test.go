package policysvc_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/policysvc"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func (ts *ControllerTestSuite) Test_GetPrivacyPolicy() {
	mockUnit := &app.Unit{
		ID:      1,
		Country: "GB",
	}
	ts.useCase.On("GetUnitByAppID", mock.Anything).Return(mockUnit, nil).Once()

	req := (&network.Request{
		Method: http.MethodGet,
		URL:    "/api/policies/privacy",
		Params: &url.Values{"appId": {"1"}, "country": {"GB"}},
	}).Build()
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	err := ts.controller.GetPrivacyPolicy(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
	ts.useCase.AssertExpectations(ts.T())
}

type ControllerTestSuite struct {
	suite.Suite
	controller policysvc.Controller
	engine     *core.Engine
	useCase    *mockUseCase
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.engine = core.NewEngine(nil)
	ts.useCase = new(mockUseCase)
	ts.controller = policysvc.NewController(ts.engine, ts.useCase, ts.useCase)
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (ctx core.Context, rec *httptest.ResponseRecorder) {
	rec = httptest.NewRecorder()
	ctx = ts.engine.NewContext(httpRequest, rec)
	return
}

type mockUseCase struct {
	mock.Mock
}

// Write test definition with TestSuite.
func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (ds *mockUseCase) GetAppByID(ctx context.Context, appID int64) (*app.App, error) {
	ret := ds.Called(appID)
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

func (r *mockUseCase) GetUnitByID(ctx context.Context, unitId int64) (*app.Unit, error) {
	ret := r.Called(unitId)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}

func (r *mockUseCase) GetUnitByAppID(ctx context.Context, appID int64) (*app.Unit, error) {
	ret := r.Called(appID)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}

func (r *mockUseCase) GetUnitByAppIDAndType(ctx context.Context, appID int64, unitType app.UnitType) (*app.Unit, error) {
	ret := r.Called(appID, unitType)
	return ret.Get(0).(*app.Unit), ret.Error(1)
}

func (r *mockUseCase) GetClientLocation(httpRequest *http.Request, countryFromLocale string) *location.Location {
	ret := r.Called(httpRequest, countryFromLocale)
	return ret.Get(0).(*location.Location)
}

func (r *mockUseCase) GetCountryFromIP(ip net.IP) (string, error) {
	ret := r.Called(ip)
	return ret.Get(0).(string), ret.Error(1)
}
