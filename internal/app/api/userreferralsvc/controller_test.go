package userreferralsvc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/userreferralsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/common/jwt"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/userreferral"
	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

type ControllerTestSuite struct {
	suite.Suite
	controller          userreferralsvc.Controller
	engine              *core.Engine
	userReferralUseCase *mockUserReferralUseCase
	deviceUseCase       *mockDeviceUseCase
	appUseCase          *mockAppUseCase
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.engine = core.NewEngine(nil)
	ts.userReferralUseCase = new(mockUserReferralUseCase)
	ts.deviceUseCase = new(mockDeviceUseCase)
	ts.appUseCase = new(mockAppUseCase)
	ts.controller = userreferralsvc.NewController(ts.engine, ts.userReferralUseCase, ts.deviceUseCase, ts.appUseCase)
}

func (ts *ControllerTestSuite) Test_GetUser_Success() {
	device, user, config := ts.getGetUserTestValues()

	req := ts.buildGetUserRequest(device.ID)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("device")
	ctx.SetParamValues(strconv.FormatInt(device.ID, 10))

	ts.deviceUseCase.On("GetByID", device.ID).Return(device, nil).Once()
	ts.appUseCase.On("GetReferralRewardConfig", device.AppID).Return(config, nil).Once()
	ts.userReferralUseCase.On("GetOrCreateUserByDevice", device.ID, device.AppID, device.UnitDeviceToken, config.VerifyURL).Return(user, nil).Once()

	err := ts.controller.GetUser(ctx)
	ts.NoError(err)

	result := map[string]map[string]interface{}{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	ts.Equal(http.StatusOK, rec.Code)
	ts.equalUser(user, result["user"])
	ts.equalConfig(config, result["referral_config"])

	ts.userReferralUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_GetUser_DisableReferral_OldDevice_Success() {
	device, user, config := ts.getGetUserTestValues()

	user.ReferrerID = 0
	ts.makeRefereeDeviceInvalid(device, config)

	req := ts.buildGetUserRequest(device.ID)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("device")
	ctx.SetParamValues(strconv.FormatInt(device.ID, 10))

	ts.deviceUseCase.On("GetByID", device.ID).Return(device, nil).Once()
	ts.appUseCase.On("GetReferralRewardConfig", device.AppID).Return(config, nil).Once()
	ts.userReferralUseCase.On("GetOrCreateUserByDevice", device.ID, device.AppID, device.UnitDeviceToken, config.VerifyURL).Return(user, nil).Once()

	err := ts.controller.GetUser(ctx)
	ts.NoError(err)

	result := map[string]map[string]interface{}{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	ts.Equal(http.StatusOK, rec.Code)
	user.ReferrerID = 1
	ts.equalUser(user, result["user"])
	ts.equalConfig(config, result["referral_config"])

	ts.userReferralUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_GetUser_DisableReferral_EndedCampaign_Success() {
	device, user, config := ts.getGetUserTestValues()

	ts.makeCampaignEnded(config)

	req := ts.buildGetUserRequest(device.ID)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("device")
	ctx.SetParamValues(strconv.FormatInt(device.ID, 10))

	ts.deviceUseCase.On("GetByID", device.ID).Return(device, nil).Once()
	ts.appUseCase.On("GetReferralRewardConfig", device.AppID).Return(config, nil).Once()
	ts.userReferralUseCase.On("GetOrCreateUserByDevice", device.ID, device.AppID, device.UnitDeviceToken, config.VerifyURL).Return(user, nil).Once()

	err := ts.controller.GetUser(ctx)
	ts.NoError(err)

	result := map[string]map[string]interface{}{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	ts.Equal(http.StatusOK, rec.Code)
	ts.equalUser(user, result["user"])
	ts.equalConfig(config, result["referral_config"])
	ts.True(result["referral_config"]["ended"].(bool))

	ts.userReferralUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_GetUser_NotFoundError() {
	device, _, config := ts.getGetUserTestValues()
	req := ts.buildGetUserRequest(device.ID)

	ts.Run("device not found error", func() {
		ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
		ctx.SetParamNames("device")
		ctx.SetParamValues(strconv.FormatInt(device.ID, 10))

		ts.deviceUseCase.On("GetByID", device.ID).Return(nil, nil).Once()

		err := ts.controller.GetUser(ctx)
		ts.NoError(err)
		ts.Equal(http.StatusBadRequest, rec.Code)

		ts.deviceUseCase.AssertExpectations(ts.T())
	})

	ts.Run("config not found error", func() {
		ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
		ctx.SetParamNames("device")
		ctx.SetParamValues(strconv.FormatInt(device.ID, 10))

		ts.deviceUseCase.On("GetByID", device.ID).Return(device, nil).Once()
		ts.appUseCase.On("GetReferralRewardConfig", device.AppID).Return(nil, nil).Once()

		err := ts.controller.GetUser(ctx)
		ts.NoError(err)
		ts.Equal(http.StatusInternalServerError, rec.Code)

		ts.deviceUseCase.AssertExpectations(ts.T())
		ts.appUseCase.AssertExpectations(ts.T())
	})

	ts.Run("user not found error", func() {
		ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
		ctx.SetParamNames("device")
		ctx.SetParamValues(strconv.FormatInt(device.ID, 10))

		ts.deviceUseCase.On("GetByID", device.ID).Return(device, nil).Once()
		ts.appUseCase.On("GetReferralRewardConfig", device.AppID).Return(config, nil).Once()
		ts.userReferralUseCase.On("GetOrCreateUserByDevice", device.ID, device.AppID, device.UnitDeviceToken, config.VerifyURL).Return(nil, userreferral.NotFoundError{}).Once()

		err := ts.controller.GetUser(ctx)
		ts.NoError(err)
		ts.Equal(http.StatusInternalServerError, rec.Code)

		ts.userReferralUseCase.AssertExpectations(ts.T())
		ts.deviceUseCase.AssertExpectations(ts.T())
		ts.appUseCase.AssertExpectations(ts.T())
	})
}

func (ts *ControllerTestSuite) Test_PostUserReferral_Success() {
	refereeDevice, referee, referrer, referrerDevice, config, createIngr := ts.getPostUserReferralTestValues()

	ts.makeRefereeDeviceValid(refereeDevice, config)
	ts.makeReferrerDeviceValid(referrerDevice, config)

	req := ts.buildPostUserReferralRequest(referrer.Code, referee.DeviceID)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("device")
	ctx.SetParamValues(strconv.FormatInt(refereeDevice.ID, 10))

	ts.deviceUseCase.On("GetByID", refereeDevice.ID).Return(refereeDevice, nil).Once()
	ts.userReferralUseCase.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
	ts.deviceUseCase.On("GetByID", referrerDevice.ID).Return(referrerDevice, nil).Once()
	ts.appUseCase.On("GetReferralRewardConfig", refereeDevice.AppID).Return(config, nil).Once()
	ts.userReferralUseCase.On("CreateReferral", *createIngr).Return(true, nil).Once()

	err := ts.controller.PostUserReferral(ctx)
	ts.NoError(err)

	result := map[string]interface{}{}
	json.Unmarshal(rec.Body.Bytes(), &result)

	ts.Equal(http.StatusOK, rec.Code)
	ts.True(result["success"].(bool))

	ts.userReferralUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_PostUserReferral_RefereeValidationError() {
	refereeDevice, referee, referrer, referrerDevice, config, _ := ts.getPostUserReferralTestValues()

	ts.makeRefereeDeviceInvalid(refereeDevice, config)

	req := ts.buildPostUserReferralRequest(referrer.Code, referee.DeviceID)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("device")
	ctx.SetParamValues(strconv.FormatInt(refereeDevice.ID, 10))

	ts.deviceUseCase.On("GetByID", refereeDevice.ID).Return(refereeDevice, nil).Once()
	ts.userReferralUseCase.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
	ts.deviceUseCase.On("GetByID", referrerDevice.ID).Return(referrerDevice, nil).Once()
	ts.appUseCase.On("GetReferralRewardConfig", refereeDevice.AppID).Return(config, nil).Once()

	err := ts.controller.PostUserReferral(ctx)
	ts.NoError(err)
	ts.Equal(http.StatusBadRequest, rec.Code)

	ts.userReferralUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_PostUserReferral_ReferrerValidationError() {
	refereeDevice, referee, referrer, referrerDevice, config, _ := ts.getPostUserReferralTestValues()

	ts.makeRefereeDeviceValid(refereeDevice, config)
	ts.makeReferrerDeviceInvalid(referrerDevice, config)

	req := ts.buildPostUserReferralRequest(referrer.Code, referee.DeviceID)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("device")
	ctx.SetParamValues(strconv.FormatInt(refereeDevice.ID, 10))

	ts.deviceUseCase.On("GetByID", refereeDevice.ID).Return(refereeDevice, nil).Once()
	ts.userReferralUseCase.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
	ts.deviceUseCase.On("GetByID", referrerDevice.ID).Return(referrerDevice, nil).Once()
	ts.appUseCase.On("GetReferralRewardConfig", refereeDevice.AppID).Return(config, nil).Once()

	err := ts.controller.PostUserReferral(ctx)
	ts.NoError(err)
	ts.Equal(http.StatusBadRequest, rec.Code)

	ts.userReferralUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_PostUserReferral_CreateReferralError() {
	refereeDevice, referee, referrer, referrerDevice, config, createIngr := ts.getPostUserReferralTestValues()

	ts.makeRefereeDeviceValid(refereeDevice, config)
	ts.makeReferrerDeviceValid(referrerDevice, config)

	req := ts.buildPostUserReferralRequest(referrer.Code, referee.DeviceID)
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	ctx.SetParamNames("device")
	ctx.SetParamValues(strconv.FormatInt(refereeDevice.ID, 10))

	ts.deviceUseCase.On("GetByID", refereeDevice.ID).Return(refereeDevice, nil).Once()
	ts.userReferralUseCase.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
	ts.deviceUseCase.On("GetByID", referrerDevice.ID).Return(referrerDevice, nil).Once()
	ts.appUseCase.On("GetReferralRewardConfig", refereeDevice.AppID).Return(config, nil).Once()
	ts.userReferralUseCase.On("CreateReferral", *createIngr).Return(false, userreferral.InvalidArgumentError{ArgName: "Code", ArgValue: referrer.Code}).Once()

	err := ts.controller.PostUserReferral(ctx)
	ts.NoError(err)
	ts.Equal(http.StatusInternalServerError, rec.Code)

	ts.userReferralUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) Test_PostUserReferral_NotFoundError() {
	refereeDevice, referee, referrer, referrerDevice, _, _ := ts.getPostUserReferralTestValues()
	req := ts.buildPostUserReferralRequest(referrer.Code, referee.DeviceID)

	ts.Run("refereeDevice not found error", func() {
		ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
		ctx.SetParamNames("device")
		ctx.SetParamValues(strconv.FormatInt(refereeDevice.ID, 10))

		ts.deviceUseCase.On("GetByID", refereeDevice.ID).Return(nil, nil).Once()

		err := ts.controller.PostUserReferral(ctx)
		ts.NoError(err)
		ts.Equal(http.StatusBadRequest, rec.Code)

		ts.deviceUseCase.AssertExpectations(ts.T())
	})

	ts.Run("referrer not found error", func() {
		ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
		ctx.SetParamNames("device")
		ctx.SetParamValues(strconv.FormatInt(refereeDevice.ID, 10))

		ts.deviceUseCase.On("GetByID", refereeDevice.ID).Return(refereeDevice, nil).Once()
		ts.userReferralUseCase.On("GetUserByCode", referrer.Code).Return(nil, userreferral.NotFoundError{}).Once()

		err := ts.controller.PostUserReferral(ctx)
		ts.NoError(err)
		ts.Equal(http.StatusBadRequest, rec.Code)

		ts.deviceUseCase.AssertExpectations(ts.T())
		ts.userReferralUseCase.AssertExpectations(ts.T())
	})

	ts.Run("referrerDevice not found error", func() {
		ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
		ctx.SetParamNames("device")
		ctx.SetParamValues(strconv.FormatInt(refereeDevice.ID, 10))

		ts.deviceUseCase.On("GetByID", refereeDevice.ID).Return(refereeDevice, nil).Once()
		ts.userReferralUseCase.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
		ts.deviceUseCase.On("GetByID", referrerDevice.ID).Return(nil, nil).Once()

		err := ts.controller.PostUserReferral(ctx)
		ts.NoError(err)
		ts.Equal(http.StatusInternalServerError, rec.Code)

		ts.deviceUseCase.AssertExpectations(ts.T())
		ts.userReferralUseCase.AssertExpectations(ts.T())
	})

	ts.Run("config not found error", func() {
		ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
		ctx.SetParamNames("device")
		ctx.SetParamValues(strconv.FormatInt(refereeDevice.ID, 10))

		ts.deviceUseCase.On("GetByID", refereeDevice.ID).Return(refereeDevice, nil).Once()
		ts.userReferralUseCase.On("GetUserByCode", referrer.Code).Return(referrer, nil).Once()
		ts.deviceUseCase.On("GetByID", referrerDevice.ID).Return(referrerDevice, nil).Once()
		ts.appUseCase.On("GetReferralRewardConfig", refereeDevice.AppID).Return(nil, nil).Once()

		err := ts.controller.PostUserReferral(ctx)
		ts.NoError(err)
		ts.Equal(http.StatusInternalServerError, rec.Code)

		ts.userReferralUseCase.AssertExpectations(ts.T())
		ts.deviceUseCase.AssertExpectations(ts.T())
		ts.appUseCase.AssertExpectations(ts.T())
	})
}

func (ts *ControllerTestSuite) makeRefereeDeviceValid(refereeDevice *device.Device, config *app.ReferralRewardConfig) {
	config.Enabled = true
	refereeDevice.CreatedAt = config.StartDate.Add(time.Hour)
	config.ExpireHours = 0 // TODO add cases for ExpireHours != 0
	config.MinSdkVersion = 10
	sdkVersion := 15
	refereeDevice.SDKVersion = &sdkVersion
}

func (ts *ControllerTestSuite) makeRefereeDeviceInvalid(refereeDevice *device.Device, config *app.ReferralRewardConfig) {
	// Make it old device
	config.Enabled = true
	refereeDevice.CreatedAt = config.StartDate.Add(-time.Hour)
}

func (ts *ControllerTestSuite) makeReferrerDeviceValid(referrerDevice *device.Device, config *app.ReferralRewardConfig) {
	referrerDevice.AppID = config.AppID
	referrerDevice.CreatedAt = config.EndDate.Add(-time.Hour)
}

func (ts *ControllerTestSuite) makeReferrerDeviceInvalid(referrerDevice *device.Device, config *app.ReferralRewardConfig) {
	referrerDevice.AppID = config.AppID
	referrerDevice.CreatedAt = config.EndDate.Add(time.Hour)
}

func (ts *ControllerTestSuite) makeCampaignEnded(config *app.ReferralRewardConfig) {
	config.Enabled = true
	endedTime := time.Now().Add(-time.Hour)
	config.EndDate = &endedTime
}

func (ts *ControllerTestSuite) buildGetUserRequest(deviceID int64) *network.Request {
	req := network.Request{
		Method: http.MethodGet,
		URL:    "/api/users/:device",
		Params: &url.Values{},
	}
	return req.Build()
}

func (ts *ControllerTestSuite) buildPostUserReferralRequest(code string, deviceID int64) *network.Request {
	req := network.Request{
		Method: http.MethodPost,
		URL:    "/api/users/:device/referral",
		Params: &url.Values{
			"code": {code},
		},
	}
	return req.Build()
}

func (ts *ControllerTestSuite) equalUser(user1 *userreferral.DeviceUser, jsonUser2 map[string]interface{}) {
	ts.True(user1.ID == int64(jsonUser2["id"].(float64)))
	ts.True(user1.DeviceID == int64(jsonUser2["device_id"].(float64)))
	ts.True(user1.Code == jsonUser2["code"].(string))
	ts.True(user1.ReferrerID == int64(jsonUser2["referrer_id"].(float64)))
	ts.True(user1.IsVerified == jsonUser2["is_verified"].(bool))
}

func (ts *ControllerTestSuite) equalConfig(config1 *app.ReferralRewardConfig, jsonConfig2 map[string]interface{}) {
	ts.True(config1.AppID == int64(jsonConfig2["app_id"].(float64)))
	ts.True(config1.Enabled == jsonConfig2["enabled"].(bool))
	ts.True(config1.Amount == int(jsonConfig2["amount"].(float64)))
}

func (ts *ControllerTestSuite) getPostUserReferralTestValues() (refereeDevice *device.Device, referee *userreferral.DeviceUser, referrer *userreferral.DeviceUser, referrerDevice *device.Device, config *app.ReferralRewardConfig, createIngr *userreferral.CreateReferralIngredients) {
	refereeDevice = ts.getTestDevice()

	referee = ts.getTestUser()
	referee.DeviceID = refereeDevice.ID

	referrer = ts.getTestUser()

	referrerDevice = ts.getTestDevice()
	referrer.DeviceID = referrerDevice.ID

	config = ts.getTestConfig()
	config.AppID = refereeDevice.AppID

	createIngr = &userreferral.CreateReferralIngredients{
		DeviceID:        refereeDevice.ID,
		AppID:           refereeDevice.AppID,
		UnitDeviceToken: refereeDevice.UnitDeviceToken,
		Code:            referrer.Code,
		JWT:             jwt.GetServiceToken(),
		VerifyURL:       config.VerifyURL,
		RewardAmount:    config.Amount,
		MaxReferral:     config.MaxReferral,
		TitleForReferral: userreferral.TitleForReferral{
			TitleForReferee:     config.TitleForReferee,
			TitleForReferrer:    config.TitleForReferrer,
			TitleForMaxReferrer: config.TitleForMaxReferrer,
		},
	}

	return
}

func (ts *ControllerTestSuite) getGetUserTestValues() (*device.Device, *userreferral.DeviceUser, *app.ReferralRewardConfig) {
	device := ts.getTestDevice()

	user := ts.getTestUser()
	user.DeviceID = device.ID

	config := ts.getTestConfig()
	config.AppID = device.AppID

	return device, user, config
}

func (ts *ControllerTestSuite) getTestDevice() *device.Device {
	var device *device.Device
	ts.NoError(faker.FakeData(&device))
	device.ID++
	return device
}

func (ts *ControllerTestSuite) getTestUser() *userreferral.DeviceUser {
	var user *userreferral.DeviceUser
	ts.NoError(faker.FakeData(&user))
	user.ID++
	user.DeviceID++
	return user
}

func (ts *ControllerTestSuite) getTestConfig() *app.ReferralRewardConfig {
	var config *app.ReferralRewardConfig
	ts.NoError(faker.FakeData(&config))
	config.AppID++
	return config
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (core.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx := ts.engine.NewContext(httpRequest, rec)
	return ctx, rec
}

type mockUserReferralUseCase struct {
	mock.Mock
}

func (u *mockUserReferralUseCase) GetUserByCode(code string) (*userreferral.DeviceUser, error) {
	ret := u.Called(code)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*userreferral.DeviceUser), ret.Error(1)
}

func (u *mockUserReferralUseCase) GetOrCreateUserByDevice(deviceID int64, appID int64, udt string, verifyURL string) (*userreferral.DeviceUser, error) {
	ret := u.Called(deviceID, appID, udt, verifyURL)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*userreferral.DeviceUser), ret.Error(1)
}

func (u *mockUserReferralUseCase) CreateReferral(ingr userreferral.CreateReferralIngredients) (bool, error) {
	ret := u.Called(ingr)
	return ret.Bool(0), ret.Error(1)
}

type mockDeviceUseCase struct {
	mock.Mock
}

func (u *mockDeviceUseCase) GetProfile(deviceID int64) (*device.Profile, error) {
	ret := u.Called(deviceID)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*device.Profile), ret.Error(1)
}

func (u *mockDeviceUseCase) GetActivity(deviceID int64) (*device.Activity, error) {
	ret := u.Called(deviceID)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*device.Activity), ret.Error(1)
}

func (u *mockDeviceUseCase) SaveActivity(deviceID int64, campaignID int64, activityType device.ActivityType) error {
	ret := u.Called(deviceID, campaignID)
	return ret.Error(0)
}

func (u *mockDeviceUseCase) SaveProfile(dp device.Profile) error {
	ret := u.Called(dp)
	return ret.Error(0)
}

func (u *mockDeviceUseCase) SaveProfilePackage(dp device.Profile) error {
	ret := u.Called(device.Profile{})
	return ret.Error(0)
}

func (u *mockDeviceUseCase) SaveProfileUnitRegisteredSeconds(dp device.Profile) error {
	ret := u.Called(device.Profile{})
	return ret.Error(0)
}

func (u *mockDeviceUseCase) DeleteProfile(dp device.Profile) error {
	ret := u.Called(dp)
	return ret.Error(0)
}

func (u *mockDeviceUseCase) GetByID(deviceID int64) (*device.Device, error) {
	ret := u.Called(deviceID)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (u *mockDeviceUseCase) GetByParams(params device.Params) (*device.Device, error) {
	ret := u.Called(params)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (u *mockDeviceUseCase) UpsertDevice(deviceArg device.Device) (*device.Device, error) {
	ret := u.Called(deviceArg)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (u *mockDeviceUseCase) ValidateUnitDeviceToken(unitDeviceToken string) (bool, error) {
	ret := u.Called(unitDeviceToken)
	return ret.Get(0).(bool), ret.Error(1)
}

type mockAppUseCase struct {
	mock.Mock
}

func (u *mockAppUseCase) GetAppByID(ctx context.Context, appID int64) (*app.App, error) {
	ret := u.Called(appID)
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
	if ret.Get(0) == nil {
		return nil, nil
	}
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
