package custompreviewsvc_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/custompreviewsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/custompreviewsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	dateFormat = time.RFC3339
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

type ControllerTestSuite struct {
	suite.Suite
	controller custompreviewsvc.Controller
	engine     *core.Engine
	useCase    *mockUseCase
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.engine = core.NewEngine(nil)
	ts.useCase = new(mockUseCase)
	ts.controller = custompreviewsvc.NewController(ts.engine, ts.useCase, nil)
}

func (ts *ControllerTestSuite) AfterTest() {
	ts.useCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_GetConfig() {
	timezone := "Asia/Seoul"
	dtoconfig, config := ts.getTestConfig()
	req := ts.buildGetConfigReq(dtoconfig.UnitID, timezone)

	ts.Run("success", func() {
		ctx, rec := ts.buildContextAndRecorder(req.Build().GetHTTPRequest())
		ts.useCase.On("GetClock").Return()
		ts.useCase.On("GetConfigByUnitID", config.UnitID, timezone, mock.Anything).Return(config, nil).Once()

		err := ts.controller.GetConfig(ctx)
		ts.NoError(err)

		result := dto.GetConfigRes{}
		json.Unmarshal(rec.Body.Bytes(), &result)

		ts.Equal(http.StatusOK, rec.Code)
		ts.equalConfig(dtoconfig, result.Config)
	})

	ts.Run("config not found", func() {
		ctx, rec := ts.buildContextAndRecorder(req.Build().GetHTTPRequest())
		ts.useCase.On("GetConfigByUnitID", config.UnitID, timezone, mock.Anything).Return(nil, nil).Once()

		err := ts.controller.GetConfig(ctx)
		ts.NoError(err)

		ts.Equal(http.StatusBadRequest, rec.Code)
	})

	ts.Run("get config error", func() {
		ctx, rec := ts.buildContextAndRecorder(req.Build().GetHTTPRequest())
		ts.useCase.On("GetConfigByUnitID", config.UnitID, timezone, mock.Anything).Return(nil, fmt.Errorf("useCase error")).Once()

		err := ts.controller.GetConfig(ctx)
		ts.NoError(err)

		ts.Equal(http.StatusInternalServerError, rec.Code)
	})
}

type mockUseCase struct {
	mock.Mock
}

func (u *mockUseCase) GetConfigByUnitID(unitID int64, timezone string, now time.Time) (*custompreview.Config, error) {
	ret := u.Called(unitID, timezone, now)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*custompreview.Config), ret.Error(1)
}

func (ts *ControllerTestSuite) buildGetConfigReq(unitID int64, timezone string) network.Request {
	req := network.Request{
		Header: &http.Header{"Time-Zone": []string{timezone}},
		Method: http.MethodGet,
		URL:    "/api/v3/custom-preview-message/config",
		Params: &url.Values{
			"unit_id": {strconv.FormatInt(unitID, 10)},
		},
	}
	return req
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (core.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx := ts.engine.NewContext(httpRequest, rec)
	return ctx, rec
}

func (ts *ControllerTestSuite) getTestConfig() (*dto.Config, *custompreview.Config) {
	var dtoconfig dto.Config
	ts.NoError(faker.FakeData(&dtoconfig))

	startDate, err := time.Parse(dateFormat, dtoconfig.StartDate.Format(dateFormat))
	ts.NoError(err)
	dtoconfig.StartDate = startDate

	endDate, err := time.Parse(dateFormat, dtoconfig.EndDate.Format(dateFormat))
	ts.NoError(err)
	dtoconfig.EndDate = endDate

	dtoconfig.ID++
	dtoconfig.UnitID++

	dtoconfig.DIPU++
	dtoconfig.TIPU++
	dtoconfig.DCPU++
	dtoconfig.TCPU++

	return &dtoconfig, ts.dtoConfigToConfig(dtoconfig)
}

func (ts *ControllerTestSuite) equalConfig(config1 *dto.Config, config2 dto.Config) {
	ts.Equal(config1.UnitID, config2.UnitID)
	ts.Equal(config1.Message, config2.Message)

	ts.Equal(config1.Period.StartDate, config2.Period.StartDate)
	ts.Equal(config1.Period.EndDate, config2.Period.EndDate)
	ts.Equal(config1.Period.StartHourMinute, config2.Period.StartHourMinute)
	ts.Equal(config1.Period.EndHourMinute, config2.Period.EndHourMinute)

	ts.Equal(config1.FrequencyLimit.DIPU, config2.FrequencyLimit.DIPU)
	ts.Equal(config1.FrequencyLimit.TIPU, config2.FrequencyLimit.TIPU)
	ts.Equal(config1.FrequencyLimit.DCPU, config2.FrequencyLimit.DCPU)
	ts.Equal(config1.FrequencyLimit.TCPU, config2.FrequencyLimit.TCPU)

	ts.Equal(*config1.Icon, *config2.Icon)
}

func (ts *ControllerTestSuite) dtoConfigToConfig(dtoconfig dto.Config) *custompreview.Config {
	return &custompreview.Config{
		UnitID:     dtoconfig.UnitID,
		Message:    dtoconfig.Message,
		LandingURL: dtoconfig.LandingURL,
		Period: custompreview.Period{
			StartDate:       dtoconfig.Period.StartDate,
			EndDate:         dtoconfig.Period.EndDate,
			StartHourMinute: dtoconfig.Period.StartHourMinute,
			EndHourMinute:   dtoconfig.Period.EndHourMinute,
		},
		FrequencyLimit: custompreview.FrequencyLimit{
			DIPU: &dtoconfig.FrequencyLimit.DIPU,
			TIPU: &dtoconfig.FrequencyLimit.TIPU,
			DCPU: &dtoconfig.FrequencyLimit.DCPU,
			TCPU: &dtoconfig.FrequencyLimit.TCPU,
		},
		Icon: dtoconfig.Icon,
	}
}
