package custompreviewsvc_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/custompreviewsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview"
	"github.com/Buzzvil/buzzscreen-api/tests"
	"github.com/jinzhu/gorm"

	customPreviewRepo "github.com/Buzzvil/buzzscreen-api/internal/pkg/custompreview/repo"
	"github.com/stretchr/testify/suite"
)

const (
	timezoneAsiaSeoulStr      = "Asia/Seoul"
	timezoneAmericaNewyorkStr = "America/New_York"
	unitID                    = int64(1024)
)

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

type IntegrationTestSuite struct {
	suite.Suite
	server     *httptest.Server
	controller custompreviewsvc.Controller
	gdb        *gorm.DB
	engine     *core.Engine
	uc         custompreview.UseCase
}

func (ts *IntegrationTestSuite) SetupTest() {
	ts.server = tests.GetTestServer(nil)
	ts.engine = core.NewEngine(nil)
	gdb, err := env.GetDatabase()
	ts.NotNil(gdb)
	ts.NoError(err)
	ts.gdb = gdb
	ts.gdb.AutoMigrate(&customPreviewRepo.DBConfig{})
	r := customPreviewRepo.New(gdb)
	ts.uc = custompreview.NewUseCase(r)
}

func (ts *IntegrationTestSuite) TearDownTest() {
	ts.gdb.DropTable(&customPreviewRepo.DBConfig{})
	ts.server.Close()
}

func (ts *IntegrationTestSuite) Test_Main() {
	cases := ts.getTestCases()

	for _, tc := range cases {
		ts.Run(tc.description, func() {
			// Apply Precondition
			ts.controller = custompreviewsvc.NewController(ts.engine, ts.uc, &tc.preCondition.currentTime)
			for _, config := range tc.preCondition.dbStatus {
				ts.gdb.Create(&config)
			}
			defer ts.gdb.Exec("ALTER TABLE custom_preview_config AUTO_INCREMENT=1")
			defer ts.gdb.Exec("DELETE FROM custom_preview_config")

			// Call API
			req := ts.buildGetConfigReq(tc.request.unitID, tc.request.timezone)
			ctx, rec := ts.buildContextAndRecorder(req.Build().GetHTTPRequest())
			err := ts.controller.GetConfig(ctx)
			ts.NoError(err)

			// Check Return Value
			result := apiResponse{}
			json.Unmarshal(rec.Body.Bytes(), &result)
			ts.Equal(tc.expectedResponse, result)
		})
	}
}

type testCase struct {
	description      string
	preCondition     preCondition
	request          apiRequest
	expectedResponse apiResponse
}

type preCondition struct {
	dbStatus    []customPreviewRepo.DBConfig
	currentTime time.Time
}

type apiRequest struct {
	unitID   int64
	timezone string
}

type apiResponse struct {
	UnitID          int64  `json:"unit_id"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	StartHourMinute string `json:"start_hour_minute"`
	EndHourMinute   string `json:"end_hour_minute"`
}

func (ts *IntegrationTestSuite) getTestCases() []testCase {
	config1 := customPreviewRepo.DBConfig{
		UnitID:          unitID,
		StartDate:       ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
		EndDate:         ts.getDate(2020, 3, 2, 23, 59, 59, 0, time.UTC),
		StartHourMinute: "02:00",
		EndHourMinute:   "03:00",
		IsActive:        true,
	}

	config2 := customPreviewRepo.DBConfig{
		UnitID:          unitID,
		StartDate:       ts.getDate(2020, 3, 3, 0, 0, 0, 0, time.UTC),
		EndDate:         ts.getDate(2020, 3, 3, 23, 59, 59, 0, time.UTC),
		StartHourMinute: "05:00",
		EndHourMinute:   "07:00",
		IsActive:        true,
	}

	// Asia/Seoul : +09:00
	timezoneAsiaSeoul, err := time.LoadLocation(timezoneAsiaSeoulStr)
	ts.NoError(err)
	// America/New_York : -05:00 (Time without dailight saving time)
	timezoneAmericaNewyork, err := time.LoadLocation(timezoneAmericaNewyorkStr)
	ts.NoError(err)

	cases := []testCase{
		testCase{
			description: "Case 1 - Asia/Seoul - Day 1 - 00:00:00",
			preCondition: preCondition{
				dbStatus: []customPreviewRepo.DBConfig{
					config1,
					config2,
				},
				currentTime: ts.getDate(2020, 3, 2, 0, 0, 0, 0, timezoneAsiaSeoul),
			},
			request: apiRequest{
				unitID:   unitID,
				timezone: timezoneAsiaSeoulStr,
			},
			expectedResponse: apiResponse{
				UnitID:          unitID,
				StartDate:       "2020-03-01T15:00:00Z",
				EndDate:         "2020-03-02T14:59:59Z",
				StartHourMinute: "17:00",
				EndHourMinute:   "18:00",
			},
		},
		testCase{
			description: "Case 2 - Asia/Seoul - Day 1 - 23:59:59",
			preCondition: preCondition{
				dbStatus: []customPreviewRepo.DBConfig{
					config1,
					config2,
				},
				currentTime: ts.getDate(2020, 3, 2, 23, 59, 59, 0, timezoneAsiaSeoul),
			},
			request: apiRequest{
				unitID:   unitID,
				timezone: timezoneAsiaSeoulStr,
			},
			expectedResponse: apiResponse{
				UnitID:          unitID,
				StartDate:       "2020-03-01T15:00:00Z",
				EndDate:         "2020-03-02T14:59:59Z",
				StartHourMinute: "17:00",
				EndHourMinute:   "18:00",
			},
		},
		testCase{
			description: "Case 3 - Asia/Seoul - Day 2 - 00:00:00",
			preCondition: preCondition{
				dbStatus: []customPreviewRepo.DBConfig{
					config1,
					config2,
				},
				currentTime: ts.getDate(2020, 3, 3, 0, 0, 0, 0, timezoneAsiaSeoul),
			},
			request: apiRequest{
				unitID:   unitID,
				timezone: timezoneAsiaSeoulStr,
			},
			expectedResponse: apiResponse{
				UnitID:          unitID,
				StartDate:       "2020-03-02T15:00:00Z",
				EndDate:         "2020-03-03T14:59:59Z",
				StartHourMinute: "20:00",
				EndHourMinute:   "22:00",
			},
		},
		testCase{
			description: "Case 4 - America/New_York - Day 1 - 00:00:00",
			preCondition: preCondition{
				dbStatus: []customPreviewRepo.DBConfig{
					config1,
					config2,
				},
				currentTime: ts.getDate(2020, 3, 2, 0, 0, 0, 0, timezoneAmericaNewyork),
			},
			request: apiRequest{
				unitID:   unitID,
				timezone: timezoneAmericaNewyorkStr,
			},
			expectedResponse: apiResponse{
				UnitID:          unitID,
				StartDate:       "2020-03-02T05:00:00Z",
				EndDate:         "2020-03-03T04:59:59Z",
				StartHourMinute: "07:00",
				EndHourMinute:   "08:00",
			},
		},
		testCase{
			description: "Case 5 - America/New_York - Day 1 - 23:59:59",
			preCondition: preCondition{
				dbStatus: []customPreviewRepo.DBConfig{
					config1,
					config2,
				},
				currentTime: ts.getDate(2020, 3, 2, 23, 59, 59, 0, timezoneAmericaNewyork),
			},
			request: apiRequest{
				unitID:   unitID,
				timezone: timezoneAmericaNewyorkStr,
			},
			expectedResponse: apiResponse{
				UnitID:          unitID,
				StartDate:       "2020-03-02T05:00:00Z",
				EndDate:         "2020-03-03T04:59:59Z",
				StartHourMinute: "07:00",
				EndHourMinute:   "08:00",
			},
		},
		testCase{
			description: "Case 6 - America/New_York - Day 2 - 00:00:00",
			preCondition: preCondition{
				dbStatus: []customPreviewRepo.DBConfig{
					config1,
					config2,
				},
				currentTime: ts.getDate(2020, 3, 3, 0, 0, 0, 0, timezoneAmericaNewyork),
			},
			request: apiRequest{
				unitID:   unitID,
				timezone: timezoneAmericaNewyorkStr,
			},
			expectedResponse: apiResponse{
				UnitID:          unitID,
				StartDate:       "2020-03-03T05:00:00Z",
				EndDate:         "2020-03-04T04:59:59Z",
				StartHourMinute: "10:00",
				EndHourMinute:   "12:00",
			},
		},
	}
	return cases
}

func (ts *IntegrationTestSuite) buildGetConfigReq(unitID int64, timezone string) network.Request {
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

func (ts *IntegrationTestSuite) buildContextAndRecorder(httprequest *http.Request) (core.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx := ts.engine.NewContext(httprequest, rec)
	return ctx, rec
}

func (ts *IntegrationTestSuite) getDate(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) time.Time {
	Date, err := time.Parse(dateFormat, time.Date(year, month, day, hour, min, sec, nsec, loc).Format(dateFormat))
	ts.NoError(err)
	return Date
}
