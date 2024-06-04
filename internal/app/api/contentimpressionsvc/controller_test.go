package contentimpressionsvc_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/contentimpressionsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/contentimpressionsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/impressiondata"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
	"github.com/bxcodec/faker"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (ts *ControllerTestSuite) Test_ContentImpression() {
	core.Loggers["impression"] = logrus.New()

	structReq, trackingData, impressionData := ts.buildBaseRequest()
	/* overwrite request parameters */
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.trackingDataUseCase.On("ParseTrackingData", *structReq.TrackingDataStr).Return(trackingData, nil).Once()
	ts.impressionDataUseCase.On("ParseImpressionData", structReq.ImpressionDataStr).Return(impressionData, nil).Once()
	ts.contentCampaignUseCase.On("IncreaseImpression", impressionData.CampaignID, impressionData.UnitID).Return(nil).Once()
	ts.deviceUseCase.On("SaveActivity", impressionData.DeviceID, impressionData.CampaignID, device.ActivityImpression).Return(nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", impressionData.UnitDeviceToken).Return(true, nil).Once()

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)

	err = ts.controller.ContentImpression(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
}

func (ts *ControllerTestSuite) Test_ContentImpression_NoTrackingData() {
	core.Loggers["impression"] = logrus.New()

	structReq, _, impressionData := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.TrackingDataStr = nil
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.impressionDataUseCase.On("ParseImpressionData", structReq.ImpressionDataStr).Return(impressionData, nil).Once()
	ts.contentCampaignUseCase.On("IncreaseImpression", impressionData.CampaignID, impressionData.UnitID).Return(nil).Once()
	ts.deviceUseCase.On("SaveActivity", impressionData.DeviceID, impressionData.CampaignID, device.ActivityImpression).Return(nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", impressionData.UnitDeviceToken).Return(true, nil).Once()

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)

	err = ts.controller.ContentImpression(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
}

func (ts *ControllerTestSuite) buildNetworkRequest(req *dto.GetContentImpressionRequest) *network.Request {
	params := &url.Values{"data": {req.ImpressionDataStr}}
	if req.TrackingDataStr != nil {
		(*params)["tracking_data"] = []string{*req.TrackingDataStr}
	}
	if req.Place != nil {
		(*params)["place"] = []string{*req.Place}
	}
	if req.Position != nil {
		(*params)["position"] = []string{*req.Position}
	}
	if req.SessionID != nil {
		(*params)["session_id"] = []string{*req.SessionID}
	}

	networkReq := &network.Request{
		Method: http.MethodGet,
		URL:    "/api/content_impression/",
		Params: params,
		Header: &http.Header{
			"User-Agent": {"Mozilla/5.0 (Linux; Android 4.2.1; en-us; Nexus 5 Build/JOP40D) AppleWebKit/535.19 (KHTML, like Gecko; googleweblight) Chrome/38.0.1025.166 Mobile Safari/535.19"},
		},
	}

	return networkReq.Build()
}

func (ts *ControllerTestSuite) buildBaseRequest() (*dto.GetContentImpressionRequest, *trackingdata.TrackingData, *impressiondata.ImpressionData) {
	contentImpressionRequest := &dto.GetContentImpressionRequest{}

	faker.FakeData(&contentImpressionRequest.Place)
	faker.FakeData(&contentImpressionRequest.Position)
	faker.FakeData(&contentImpressionRequest.SessionID)
	faker.FakeData(&contentImpressionRequest.TrackingDataStr)
	faker.FakeData(&contentImpressionRequest.ImpressionDataStr)

	trackingData := &trackingdata.TrackingData{ModelArtifact: "v1"}
	impressionData := &impressiondata.ImpressionData{}
	faker.FakeData(&impressionData)

	place := "__place__"
	position := "__position__"
	sessionID := "_session_id__"
	contentImpressionRequest.Place = &place
	contentImpressionRequest.Position = &position
	contentImpressionRequest.SessionID = &sessionID

	return contentImpressionRequest, trackingData, impressionData
}

type ControllerTestSuite struct {
	suite.Suite
	controller             contentimpressionsvc.Controller
	engine                 *core.Engine
	impressionDataUseCase  *mockImpressionDataUseCase
	contentCampaignUseCase *mockContentCampaignUseCase
	trackingDataUseCase    *mockTrackingDataUseCase
	deviceUseCase          *mockDeviceUseCase
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (ctx core.Context, rec *httptest.ResponseRecorder) {
	rec = httptest.NewRecorder()
	ctx = ts.engine.NewContext(httpRequest, rec)
	return
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.engine = core.NewEngine(nil)
	ts.contentCampaignUseCase = new(mockContentCampaignUseCase)
	ts.trackingDataUseCase = new(mockTrackingDataUseCase)
	ts.deviceUseCase = new(mockDeviceUseCase)
	ts.impressionDataUseCase = new(mockImpressionDataUseCase)

	ts.controller = contentimpressionsvc.NewController(ts.engine, ts.trackingDataUseCase, ts.impressionDataUseCase, ts.contentCampaignUseCase, ts.deviceUseCase)
}

func (ts *ControllerTestSuite) AfterTest(_, _ string) {
	ts.impressionDataUseCase.AssertExpectations(ts.T())
	ts.contentCampaignUseCase.AssertExpectations(ts.T())
	ts.trackingDataUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
}

type mockImpressionDataUseCase struct {
	mock.Mock
}

func (u *mockImpressionDataUseCase) ParseImpressionData(impressionDataString string) (*impressiondata.ImpressionData, error) {
	ret := u.Called(impressionDataString)
	return ret.Get(0).(*impressiondata.ImpressionData), ret.Error(1)
}

func (u *mockImpressionDataUseCase) BuildImpressionDataString(impressionData impressiondata.ImpressionData) string {
	ret := u.Called(impressionData)
	return ret.Get(0).(string)
}

type mockContentCampaignUseCase struct {
	mock.Mock
}

func (u *mockContentCampaignUseCase) GetContentCampaignByID(campaignID int64) (*contentcampaign.ContentCampaign, error) {
	ret := u.Called(campaignID)
	return ret.Get(0).(*contentcampaign.ContentCampaign), ret.Error(1)
}

func (u *mockContentCampaignUseCase) IncreaseClick(campaignID int64, unitID int64) error {
	ret := u.Called(campaignID, unitID)
	return ret.Error(0)
}
func (u *mockContentCampaignUseCase) IncreaseImpression(campaignID int64, unitID int64) error {
	ret := u.Called(campaignID, unitID)
	return ret.Error(0)
}

func (u *mockContentCampaignUseCase) IsContentCampaignExpired(contentCampaign *contentcampaign.ContentCampaign) bool {
	ret := u.Called(contentCampaign)
	return ret.Get(0).(bool)
}

type mockTrackingDataUseCase struct {
	mock.Mock
}

func (u *mockTrackingDataUseCase) ParseTrackingData(trackingDataString string) (*trackingdata.TrackingData, error) {
	ret := u.Called(trackingDataString)
	return ret.Get(0).(*trackingdata.TrackingData), ret.Error(1)
}

func (u *mockTrackingDataUseCase) BuildTrackingDataString(trackingData *trackingdata.TrackingData) string {
	ret := u.Called(trackingData)
	return ret.Get(0).(string)
}

type mockDeviceUseCase struct {
	mock.Mock
}

func (u *mockDeviceUseCase) GetProfile(deviceID int64) (*device.Profile, error) {
	ret := u.Called(deviceID)
	return ret.Get(0).(*device.Profile), ret.Error(1)
}

func (u *mockDeviceUseCase) GetActivity(deviceID int64) (*device.Activity, error) {
	ret := u.Called(deviceID)
	return ret.Get(0).(*device.Activity), ret.Error(1)
}

func (u *mockDeviceUseCase) SaveActivity(deviceID int64, campaignID int64, activityType device.ActivityType) error {
	ret := u.Called(deviceID, campaignID, activityType)
	return ret.Error(0)
}

func (u *mockDeviceUseCase) SaveProfile(dp device.Profile) error {
	ret := u.Called(device.Profile{})
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
	ret := u.Called(device.Profile{})
	return ret.Error(0)
}

func (u *mockDeviceUseCase) GetByID(deviceID int64) (*device.Device, error) {
	ret := u.Called(deviceID)
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (u *mockDeviceUseCase) GetByParams(params device.Params) (*device.Device, error) {
	ret := u.Called(params)
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (u *mockDeviceUseCase) UpsertDevice(d device.Device) (*device.Device, error) {
	ret := u.Called(d)
	return ret.Get(0).(*device.Device), ret.Error(1)
}

func (u *mockDeviceUseCase) ValidateUnitDeviceToken(unitDeviceToken string) (bool, error) {
	ret := u.Called(unitDeviceToken)
	return ret.Get(0).(bool), ret.Error(1)
}
