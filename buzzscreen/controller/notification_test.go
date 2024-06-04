package controller_test

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/datasource/dbapp"

	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/dto"
	"github.com/Buzzvil/buzzscreen-api/buzzscreen/model"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type (
	// TestNotificationsResponse type definition
	TestNotificationsResponse struct {
		Notifications []dto.Notification `json:"notifications"`
	}

	// NotificationTestSuite type definition
	NotificationTestSuite struct {
		suite.Suite
		adTestCase       AdTestCase
		buzzAds          []ad.AdV2
		receivedStatuses []reward.ReceivedStatus
		unit             dbapp.Unit
	}
)

func (suite *NotificationTestSuite) BeforeTest(suiteName, testName string) {
	suite.adTestCase = AdTestCase{AdTestConfig: V3AdTestConfig{}}
	suite.adTestCase.setUp()

	suite.unit = dbapp.Unit{ID: int64(10101)}
	assert.Nil(suite.T(), buzzscreen.Service.DB.Save(&suite.unit).Error)
}

func (suite *NotificationTestSuite) AfterTest(suiteName, testName string) {
	suite.buzzAds = []ad.AdV2{}
	suite.receivedStatuses = []reward.ReceivedStatus{}
	suite.adTestCase.tearDown()
	buzzscreen.Service.DB.Delete(model.NotificationSchedule{})
	buzzscreen.Service.DB.Delete(suite.unit)
}

func TestNotificationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationTestSuite))
}

func (suite *NotificationTestSuite) registerAd(ad ad.AdV2, receivedStatus reward.ReceivedStatus) {
	suite.buzzAds = append(suite.buzzAds, ad)
	suite.receivedStatuses = append(suite.receivedStatuses, receivedStatus)
}

func (suite *NotificationTestSuite) registerSchedule(fn func(schedule *model.NotificationSchedule)) model.NotificationSchedule {
	schedule := model.NotificationSchedule{
		UnitID:           suite.unit.ID,
		Title:            "Example Title",
		IconURL:          "http://test.com/icon.url",
		Description:      "You can earn {total_reward} Point. {dummy_key}",
		Schedule:         "* * * * * * *",
		NotificationType: model.NotificationTypeFeed,
		Importance:       model.NotificationImportanceLow,
	}
	fn(&schedule)
	assert.Nil(suite.T(), buzzscreen.Service.DB.Save(&schedule).Error)
	return schedule
}

func (suite *NotificationTestSuite) mockResponses() {
	suite.adTestCase.MockBuzzAds(suite.buzzAds)
}

func (suite *NotificationTestSuite) requestNotifications(requestParams *url.Values) TestNotificationsResponse {
	var response TestNotificationsResponse
	statusCode, err := (&network.Request{
		Method: "GET",
		Params: requestParams,
		URL:    ts.URL + "/api/v3/notifications",
	}).GetResponse(&response)

	assert.Equal(suite.T(), 200, statusCode)
	assert.Nil(suite.T(), err)

	return response
}

func (suite *NotificationTestSuite) TestGetNotificationsAnonymousUser() {
	requestParams, _ := suite.adTestCase.getRequestParams(suite.T())
	requestParams.Del("session_key")

	var response TestNotificationsResponse
	statusCode, _ := (&network.Request{
		Method: "GET",
		Params: requestParams,
		URL:    ts.URL + "/api/v3/notifications",
	}).GetResponse(&response)
	assert.Equal(suite.T(), http.StatusUnauthorized, statusCode)
}

func (suite *NotificationTestSuite) TestGetNotificationsTotalPoint() {
	assert := assert.New(suite.T())

	rewards := []int{100, 200, 300}
	suite.registerAd(*buildBuzzAdV2From(func(ad *ad.AdV2) {
		ad.ID = 100001
		ad.LandingReward = rewards[0]
	}), reward.StatusUnknown)
	suite.registerAd(*buildBuzzAdV2From(func(ad *ad.AdV2) {
		ad.ID = 100002
		ad.ActionReward = rewards[1]
	}), reward.StatusUnknown)
	suite.registerAd(*buildBuzzAdV2From(func(ad *ad.AdV2) {
		ad.ID = 100003
		ad.LandingReward = rewards[2]
	}), reward.StatusReceived)

	requestParams, device := suite.adTestCase.getRequestParams(suite.T())
	requestParams.Set("unit_id", strconv.FormatInt(suite.unit.ID, 10))

	adID := requestParams.Get("ad_id")
	suite.adTestCase.createPoint(suite.T(), 100003+dto.BuzzAdCampaignIDOffset, adID, int64(device.Result["device_id"].(float64)))
	suite.mockResponses()

	schedule := suite.registerSchedule(func(schedule *model.NotificationSchedule) {})
	response := suite.requestNotifications(requestParams)
	assert.Equal(1, len(response.Notifications))

	notification := response.Notifications[0]
	assert.Equal(schedule.Importance.String(), notification.Importance)
	assert.Equal(schedule.Link(), notification.Link)
	assert.Equal(schedule.IconURL, notification.IconURL)
	assert.True(strings.Contains(notification.Description, strconv.Itoa(rewards[0]+rewards[1])))

	totalReward, err := strconv.Atoi(notification.Payload["total_reward"])
	assert.Nil(err)
	assert.Equal(rewards[0]+rewards[1], totalReward)
}

func (suite *NotificationTestSuite) TestGetNotificationsWithoutTotalReward() {
	suite.registerAd(*buildBuzzAdV2From(func(ad *ad.AdV2) {
		ad.ID = 100001
		ad.LandingReward = 100
	}), reward.StatusReceived)
	requestParams, device := suite.adTestCase.getRequestParams(suite.T())
	requestParams.Set("unit_id", strconv.FormatInt(suite.unit.ID, 10))
	adID := requestParams.Get("ad_id")

	suite.adTestCase.createPoint(suite.T(), 100001+dto.BuzzAdCampaignIDOffset, adID, int64(device.Result["device_id"].(float64)))

	suite.registerSchedule(func(schedule *model.NotificationSchedule) {})
	suite.mockResponses()
	response := suite.requestNotifications(requestParams)
	assert.Equal(suite.T(), 0, len(response.Notifications))
}

func (suite *NotificationTestSuite) TestGetNotificationsNotInSchedule() {
	suite.registerAd(*buildBuzzAdV2From(func(ad *ad.AdV2) {
		ad.ID = 100001
		ad.LandingReward = 100
	}), reward.StatusUnknown)
	suite.registerSchedule(func(schedule *model.NotificationSchedule) {
		schedule.Schedule = "* " + strconv.Itoa(time.Now().Add(-time.Minute).Minute()) + " * * * * *"
	})
	suite.mockResponses()
	requestParams, _ := suite.adTestCase.getRequestParams(suite.T())
	requestParams.Set("unit_id", strconv.FormatInt(suite.unit.ID, 10))
	response := suite.requestNotifications(requestParams)
	assert.Equal(suite.T(), 0, len(response.Notifications))
}

func (suite *NotificationTestSuite) TestGetNotificationsFiltersNotEligibleSchedules() {
	assert := assert.New(suite.T())
	suite.registerAd(*buildBuzzAdV2From(func(ad *ad.AdV2) {
		ad.ID = 100001
		ad.LandingReward = 100
	}), reward.StatusUnknown)
	suite.mockResponses()

	suite.T().Run("EligibleVersionCode", func(t *testing.T) {
		schedule := suite.registerSchedule(func(schedule *model.NotificationSchedule) {
			schedule.MinVersionCode = new(int)
			*schedule.MinVersionCode = 10
			schedule.MaxVersionCode = new(int)
			*schedule.MaxVersionCode = 999999999
		})
		defer buzzscreen.Service.DB.Delete(&schedule)
		requestParams, _ := suite.adTestCase.getRequestParams(suite.T())
		requestParams.Set("unit_id", strconv.FormatInt(suite.unit.ID, 10))
		response := suite.requestNotifications(requestParams)
		assert.Equal(1, len(response.Notifications))
	})

	suite.T().Run("IneligibleMinVersionCode", func(t *testing.T) {
		schedule := suite.registerSchedule(func(schedule *model.NotificationSchedule) {
			schedule.MinVersionCode = new(int)
			*schedule.MinVersionCode = 999999999
		})
		defer buzzscreen.Service.DB.Delete(&schedule)
		requestParams, _ := suite.adTestCase.getRequestParams(suite.T())
		requestParams.Set("unit_id", strconv.FormatInt(suite.unit.ID, 10))
		response := suite.requestNotifications(requestParams)
		assert.Equal(0, len(response.Notifications))
	})

	suite.T().Run("IneligibleMaxVersionCode", func(t *testing.T) {
		schedule := suite.registerSchedule(func(schedule *model.NotificationSchedule) {
			schedule.MaxVersionCode = new(int)
			*schedule.MaxVersionCode = 10
		})
		defer buzzscreen.Service.DB.Delete(&schedule)
		requestParams, _ := suite.adTestCase.getRequestParams(suite.T())
		requestParams.Set("unit_id", strconv.FormatInt(suite.unit.ID, 10))
		response := suite.requestNotifications(requestParams)
		assert.Equal(0, len(response.Notifications))
	})
}
