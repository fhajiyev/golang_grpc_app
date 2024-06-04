package configsvc_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/configsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/configsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/config"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (ts *ControllerTestSuite) TestGetConfigurationsPatchUnitID() {
	r := dto.GetConfigsRequest{}
	ts.NoError(faker.FakeData(&r))
	r.UnitID = 0
	r.Package = "com.buzzvil.honeyscreen.jp"
	r.SDKVersion = 3500

	configReq := config.RequestIngredients{
		UnitID:       100000045,
		Manufacturer: r.Manufacturer,
	}

	networkReq := &network.Request{
		Method: http.MethodGet,
		URL:    "/api/v3/configurations",
		Params: ts.buildGetConfigQueryParams(r),
	}
	networkReq.Build()
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	var configs []config.Config // TODO return adequate config struct
	ts.useCase.On("GetConfigs", configReq).Return(&configs)

	err := ts.controller.GetConfigurations(ctx)
	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
}

func (ts *ControllerTestSuite) buildGetConfigQueryParams(r dto.GetConfigsRequest) *url.Values {
	v := &url.Values{}
	v.Add("unit_id", strconv.FormatInt(r.UnitID, 10))
	v.Add("manufacturer", r.Manufacturer)
	v.Add("package", r.Package)
	v.Add("sdk_version", strconv.Itoa(r.SDKVersion))
	return v
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (core.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx := ts.engine.NewContext(httpRequest, rec)
	return ctx, rec
}

type ControllerTestSuite struct {
	suite.Suite
	useCase    *mockUseCase
	engine     *core.Engine
	controller configsvc.Controller
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.useCase = new(mockUseCase)
	ts.engine = core.NewEngine(nil)
	ts.controller = configsvc.NewController(ts.engine, ts.useCase)
}

type mockUseCase struct {
	mock.Mock
}

func (u *mockUseCase) GetConfigs(configReq config.RequestIngredients) *[]config.Config {
	ret := u.Called(configReq)
	return ret.Get(0).(*[]config.Config)
}
