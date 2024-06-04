package custompreview

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	timezoneAsiaSeoulStr      = "Asia/Seoul"
	timezoneAmericaNewyorkStr = "America/New_York"
	dateFormat                = time.RFC3339
)

func TestUseCaseSuite(t *testing.T) {
	suite.Run(t, new(UseCaseTestSuite))
}

type UseCaseTestSuite struct {
	suite.Suite
	repo    *mockRepo
	useCase useCase
}

func (ts *UseCaseTestSuite) SetupTest() {
	ts.repo = new(mockRepo)
	ts.useCase = useCase{ts.repo}
}

func (ts *UseCaseTestSuite) AfterTest() {
	ts.repo.AssertExpectations(ts.T())
}

func (ts *UseCaseTestSuite) Test_GetConfigByUnitID() {
	ts.Run("Asia Seoul", func() {
		// In DB, date is stored without timezone
		repoConfig := &Config{}
		repoConfig.StartDate = time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)
		repoConfig.EndDate = time.Date(2020, 3, 2, 23, 59, 59, 0, time.UTC)

		// Target time is UTC 1st day , Asia/Seoul 2nd day. So, above config works target time for Asia/Seoul
		targetTime := time.Date(2020, 3, 1, 18, 0, 0, 0, time.UTC)
		// To search, it should be 2nd day
		addedTargetTime := time.Date(2020, 3, 2, 3, 0, 0, 0, time.UTC)

		repoConfig.StartHourMinute = "02:00"
		repoConfig.EndHourMinute = "03:00"

		expectedConfig := &Config{}
		expectedConfig.UnitID = repoConfig.UnitID
		expectedConfig.StartDate = time.Date(2020, 3, 1, 15, 0, 0, 0, time.UTC)
		expectedConfig.EndDate = time.Date(2020, 3, 2, 14, 59, 59, 0, time.UTC)

		expectedConfig.StartHourMinute = "17:00"
		expectedConfig.EndHourMinute = "18:00"

		ts.repo.On("GetConfigByUnitID", repoConfig.UnitID, true, addedTargetTime).Return(repoConfig, nil).Once()
		res, err := ts.useCase.GetConfigByUnitID(repoConfig.UnitID, timezoneAsiaSeoulStr, targetTime)
		ts.NoError(err)
		ts.equalConfig(expectedConfig, res)
	})
}

func (ts *UseCaseTestSuite) Test_replaceTimezone() {
	// Asia/Seoul : +09:00
	timezoneAsiaSeoul, err := time.LoadLocation(timezoneAsiaSeoulStr)
	ts.NoError(err)
	// America/New_York : -05:00 (Time without dailight saving time)
	timezoneAmericaNewyork, err := time.LoadLocation(timezoneAmericaNewyorkStr)
	ts.NoError(err)

	cases := []struct {
		description        string
		inputTime          time.Time
		inputTimezone      *time.Location
		expectedOutputTime time.Time
	}{
		{
			description:        "Case 1 - Asia/Seoul : Day 2, hour 0 -> Day 1, hour 15",
			inputTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			inputTimezone:      timezoneAsiaSeoul,
			expectedOutputTime: ts.getDate(2020, 3, 1, 15, 0, 0, 0, time.UTC),
		},
		{
			description:        "Case 2 - Asia/Seoul : Day 2, hour 9 -> Day 1, hour 0",
			inputTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			inputTimezone:      timezoneAsiaSeoul,
			expectedOutputTime: ts.getDate(2020, 3, 1, 15, 0, 0, 0, time.UTC),
		},
		{
			description:        "Case 3 - Asia/Seoul : Day 2, hour 13 -> Day 2, hour 4",
			inputTime:          ts.getDate(2020, 3, 2, 13, 0, 0, 0, time.UTC),
			inputTimezone:      timezoneAsiaSeoul,
			expectedOutputTime: ts.getDate(2020, 3, 2, 4, 0, 0, 0, time.UTC),
		},
		{
			description:        "Case 4 - America/New_York : Day 2, hour 0 -> Day 2, hour 5",
			inputTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			inputTimezone:      timezoneAmericaNewyork,
			expectedOutputTime: ts.getDate(2020, 3, 2, 5, 0, 0, 0, time.UTC),
		},
		{
			description:        "Case 5 - America/New_York : Day 2, hour 9 -> Day 2, hour 14",
			inputTime:          ts.getDate(2020, 3, 2, 9, 0, 0, 0, time.UTC),
			inputTimezone:      timezoneAmericaNewyork,
			expectedOutputTime: ts.getDate(2020, 3, 2, 14, 0, 0, 0, time.UTC),
		},
		{
			description:        "Case 6 - America/New_York : Day 2, hour 13 -> Day 2, hour 18",
			inputTime:          ts.getDate(2020, 3, 2, 13, 0, 0, 0, time.UTC),
			inputTimezone:      timezoneAmericaNewyork,
			expectedOutputTime: ts.getDate(2020, 3, 2, 18, 0, 0, 0, time.UTC),
		},
		{
			description:        "Case 7 - UTC : Day 2, hour 0 -> Day 2, hour 0",
			inputTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			inputTimezone:      time.UTC,
			expectedOutputTime: ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			description:        "Case 8 - UTC : Day 2, hour 9 -> Day 2, hour 9",
			inputTime:          ts.getDate(2020, 3, 2, 9, 0, 0, 0, time.UTC),
			inputTimezone:      time.UTC,
			expectedOutputTime: ts.getDate(2020, 3, 2, 9, 0, 0, 0, time.UTC),
		},
		{
			description:        "Case 9 - UTC : Day 2, hour 13 -> Day 2, hour 13",
			inputTime:          ts.getDate(2020, 3, 2, 13, 0, 0, 0, time.UTC),
			inputTimezone:      time.UTC,
			expectedOutputTime: ts.getDate(2020, 3, 2, 13, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range cases {
		ts.Run(tc.description, func() {
			res := ts.useCase.replaceTimezone(tc.inputTime, tc.inputTimezone)
			ts.Equal(tc.expectedOutputTime, res)
		})
	}
}

func (ts *UseCaseTestSuite) Test_removeTimezoneToHourMinute() {
	// Asia/Seoul : +09:00
	timezoneAsiaSeoul, err := time.LoadLocation(timezoneAsiaSeoulStr)
	ts.NoError(err)
	// America/New_York : -05:00 (Time without dailight saving time)
	timezoneAmericaNewyork, err := time.LoadLocation(timezoneAmericaNewyorkStr)
	ts.NoError(err)

	cases := []struct {
		description              string
		inputHourMinute          string
		inputTimezone            *time.Location
		inputTargetTime          time.Time
		expectedOutputHourMinute string
	}{
		{
			description:              "Case 1 - Asia/Seoul : 00:00 -> 15:00",
			inputHourMinute:          "00:00",
			inputTimezone:            timezoneAsiaSeoul,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "15:00",
		},
		{
			description:              "Case 2 - Asia/Seoul : 08:11 -> 23:11",
			inputHourMinute:          "08:11",
			inputTimezone:            timezoneAsiaSeoul,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "23:11",
		},
		{
			description:              "Case 3 - Asia/Seoul : 09:59 -> 00:59",
			inputHourMinute:          "09:59",
			inputTimezone:            timezoneAsiaSeoul,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "00:59",
		},
		{
			description:              "Case 4 - Asia/Seoul : 23:43 -> 14:43",
			inputHourMinute:          "23:43",
			inputTimezone:            timezoneAsiaSeoul,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "14:43",
		},
		{
			description:              "Case 5 - America/New_York : 00:03 -> 05:03",
			inputHourMinute:          "00:03",
			inputTimezone:            timezoneAmericaNewyork,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "05:03",
		},
		{
			description:              "Case 6 - America/New_York : 08:17 -> 13:17",
			inputHourMinute:          "08:17",
			inputTimezone:            timezoneAmericaNewyork,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "13:17",
		},
		{
			description:              "Case 7 - America/New_York : 09:49 -> 14:49",
			inputHourMinute:          "09:49",
			inputTimezone:            timezoneAmericaNewyork,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "14:49",
		},
		{
			description:              "Case 8 - America/New_York : 23:33 -> 04:33",
			inputHourMinute:          "23:33",
			inputTimezone:            timezoneAmericaNewyork,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "04:33",
		},
		{
			description:              "Case 9 - UTC : 00:42 -> 00:42",
			inputHourMinute:          "00:42",
			inputTimezone:            time.UTC,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "00:42",
		},
		{
			description:              "Case 10 - UTC : 08:11 -> 08:11",
			inputHourMinute:          "08:11",
			inputTimezone:            time.UTC,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "08:11",
		},
		{
			description:              "Case 11 - UTC : 09:44 -> 09:44",
			inputHourMinute:          "09:44",
			inputTimezone:            time.UTC,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "09:44",
		},
		{
			description:              "Case 12 - UTC : 23:33 -> 23:33",
			inputHourMinute:          "23:33",
			inputTimezone:            time.UTC,
			inputTargetTime:          ts.getDate(2020, 3, 2, 0, 0, 0, 0, time.UTC),
			expectedOutputHourMinute: "23:33",
		},
	}

	for _, tc := range cases {
		ts.Run(tc.description, func() {
			res, err := ts.useCase.removeTimezoneToHourMinute(tc.inputHourMinute, tc.inputTimezone, tc.inputTargetTime)
			ts.NoError(err)
			ts.Equal(tc.expectedOutputHourMinute, res)
		})
	}
}

func (ts *UseCaseTestSuite) equalConfig(config1 *Config, config2 *Config) {
	ts.NotNil(config1)
	ts.NotNil(config2)

	ts.Equal(config1.UnitID, config2.UnitID)
	ts.Equal(config1.Period.StartDate, config2.Period.StartDate)
	ts.Equal(config1.Period.EndDate, config2.Period.EndDate)
	ts.Equal(config1.Period.StartHourMinute, config2.Period.StartHourMinute)
	ts.Equal(config1.Period.EndHourMinute, config2.Period.EndHourMinute)
}

func (ts *UseCaseTestSuite) getDate(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) time.Time {
	Date, err := time.Parse(dateFormat, time.Date(year, month, day, hour, min, sec, nsec, loc).Format(dateFormat))
	ts.NoError(err)
	return Date
}

type mockRepo struct {
	mock.Mock
}

func (r *mockRepo) GetConfigByUnitID(unitID int64, isActive bool, targetTime time.Time) (*Config, error) {
	ret := r.Called(unitID, isActive, targetTime)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*Config), ret.Error(1)
}
