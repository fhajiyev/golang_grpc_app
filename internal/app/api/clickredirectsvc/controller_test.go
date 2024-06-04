package clickredirectsvc_test

import (
	"context"
	"errors"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/clickredirectsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/clickredirectsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/payload"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/profilerequest"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/reward"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/trackingdata"
	"github.com/Buzzvil/buzzscreen-api/tests"
	gotestmock "github.com/Buzzvil/go-test/mock"
	"github.com/bxcodec/faker"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (ts *ControllerTestSuite) Test_ClickRedirect() {
	clientPatcher := ts.getBuzzAdMock(nil)
	defer clientPatcher.RemovePatch()
	core.Loggers["click"] = logrus.New()

	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()
	ts.rewardUseCase.On("GiveReward", mock.AnythingOfType("reward.RequestIngredients")).Return(structReq.Reward, nil).Once()

	unit := ts.createUnit(structReq.UnitID)
	ts.appUseCase.On("GetUnitByID", structReq.UnitID).Return(unit, nil)

	ts.contentCampaignUseCase.On("IncreaseClick", structReq.CampaignID, structReq.UnitID).Return(nil).Once()
	ts.deviceUseCase.On("SaveActivity", structReq.DeviceID, structReq.CampaignID, device.ActivityClick).Return(nil).Once()

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", mock.AnythingOfType("string")).Return(payloadStruct, nil).Once()
	ts.payloadUseCase.On("IsPayloadExpired", mock.AnythingOfType("*payload.Payload")).Return(false).Once()
	ts.trackingDataUseCase.On("ParseTrackingData", structReq.TrackingDataStr).Return(&trackingdata.TrackingData{ModelArtifact: "v1"}, nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", structReq.GetUDT()).Return(true, nil).Once()
	ts.profileRequestUseCase.On("PopulateProfile", mock.AnythingOfType("profilerequest.Account")).Return(nil).Once()

	err = ts.controller.ClickRedirect(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusFound, rec.Code)
}

func (ts *ControllerTestSuite) Test_ClickRedirect_NoPayload() {
	clientPatcher := ts.getBuzzAdMock(nil)
	defer clientPatcher.RemovePatch()
	core.Loggers["click"] = logrus.New()

	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.PayloadStr = ""
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()
	ts.rewardUseCase.On("GiveReward", mock.AnythingOfType("reward.RequestIngredients")).Return(structReq.Reward, nil).Once()

	unit := ts.createUnit(structReq.UnitID)
	ts.appUseCase.On("GetUnitByID", structReq.UnitID).Return(unit, nil)

	contentCampaign := &contentcampaign.ContentCampaign{ID: structReq.CampaignID}
	ts.contentCampaignUseCase.On("GetContentCampaignByID", structReq.CampaignID).Return(contentCampaign, nil).Once()
	ts.contentCampaignUseCase.On("IsContentCampaignExpired", contentCampaign).Return(false).Once()
	ts.contentCampaignUseCase.On("IncreaseClick", structReq.CampaignID, structReq.UnitID).Return(nil).Once()
	ts.deviceUseCase.On("SaveActivity", structReq.DeviceID, structReq.CampaignID, device.ActivityClick).Return(nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", structReq.GetUDT()).Return(true, nil).Once()
	ts.profileRequestUseCase.On("PopulateProfile", mock.AnythingOfType("profilerequest.Account")).Return(nil).Once()

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", structReq.PayloadStr).Return((*payload.Payload)(nil), errors.New("invalid payload")).Once()
	ts.trackingDataUseCase.On("ParseTrackingData", structReq.TrackingDataStr).Return(&trackingdata.TrackingData{ModelArtifact: "v1"}, nil).Once()

	err = ts.controller.ClickRedirect(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusFound, rec.Code)
}

func (ts *ControllerTestSuite) Test_ClickRedirect_Expired() {
	clientPatcher := ts.getBuzzAdMock(nil)
	defer clientPatcher.RemovePatch()
	core.Loggers["click"] = logrus.New()

	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", structReq.GetUDT()).Return(true, nil).Once()

	unit := ts.createUnit(structReq.UnitID)
	ts.appUseCase.On("GetUnitByID", structReq.UnitID).Return(unit, nil)

	ts.contentCampaignUseCase.On("IncreaseClick", structReq.CampaignID, structReq.UnitID).Return(nil).Once()
	ts.deviceUseCase.On("SaveActivity", structReq.DeviceID, structReq.CampaignID, device.ActivityClick).Return(nil).Once()

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", mock.AnythingOfType("string")).Return(payloadStruct, nil).Once()
	ts.payloadUseCase.On("IsPayloadExpired", mock.AnythingOfType("*payload.Payload")).Return(true).Once()
	ts.trackingDataUseCase.On("ParseTrackingData", structReq.TrackingDataStr).Return(&trackingdata.TrackingData{ModelArtifact: "v1"}, nil).Once()
	ts.profileRequestUseCase.On("PopulateProfile", mock.AnythingOfType("profilerequest.Account")).Return(nil).Once()

	err = ts.controller.ClickRedirect(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusFound, rec.Code)
}

func (ts *ControllerTestSuite) Test_ClickRedirect_NoReward() {
	clientPatcher := ts.getBuzzAdMock(nil)
	defer clientPatcher.RemovePatch()
	core.Loggers["click"] = logrus.New()

	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.Reward = 0
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", structReq.GetUDT()).Return(true, nil).Once()

	unit := ts.createUnit(structReq.UnitID)
	ts.appUseCase.On("GetUnitByID", structReq.UnitID).Return(unit, nil)

	ts.contentCampaignUseCase.On("IncreaseClick", structReq.CampaignID, structReq.UnitID).Return(nil).Once()
	ts.deviceUseCase.On("SaveActivity", structReq.DeviceID, structReq.CampaignID, device.ActivityClick).Return(nil).Once()

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", mock.AnythingOfType("string")).Return(payloadStruct, nil).Once()
	ts.payloadUseCase.On("IsPayloadExpired", mock.AnythingOfType("*payload.Payload")).Return(false).Once()
	ts.trackingDataUseCase.On("ParseTrackingData", structReq.TrackingDataStr).Return(&trackingdata.TrackingData{ModelArtifact: "v1"}, nil).Once()
	ts.profileRequestUseCase.On("PopulateProfile", mock.AnythingOfType("profilerequest.Account")).Return(nil).Once()

	err = ts.controller.ClickRedirect(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusFound, rec.Code)
}

func (ts *ControllerTestSuite) Test_ClickRedirect_DeactivatedUnit() {
	clientPatcher := ts.getBuzzAdMock(nil)
	defer clientPatcher.RemovePatch()
	core.Loggers["click"] = logrus.New()

	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.Reward = 0
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, _ := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", structReq.GetUDT()).Return(true, nil).Once()

	unit := ts.createUnit(structReq.UnitID)
	unit.IsActive = false
	ts.appUseCase.On("GetUnitByID", structReq.UnitID).Return(unit, nil)

	err := ts.controller.ClickRedirect(ctx)
	ts.Error(err)
}

func (ts *ControllerTestSuite) Test_ClickRedirect_Tracker() {
	trackEventCalled := false
	clientPatcher := ts.getBuzzAdMock(&trackEventCalled)
	defer clientPatcher.RemovePatch()
	core.Loggers["click"] = logrus.New()

	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.PayloadStr = ""
	structReq.BaseReward = 0
	trackingURL := "https://ad.buzzvil.com/api/track_event?token=DC-2AG_LBRrgc_ttbRtyZJtL1ZkIR1EhQ1lGT815HVgOinjI_prlunpLEI-IW0xNW7bkdeKaCJrYrmqmbcpr3Bph01Y89cwt-YNa-znz-5K2Uu1BXwXCepF-e9dgf-0QSU3ICpRR-1JOUGb1KwkI-05BBX8a4MNz2HNqgE5ccgz4awhJNbJRsbf0dJxD_SwSlRBIpA9B7TvrT_Eq9kdjn7-qlWZtarlNI_HU5BvEpC-OpsWKNdvKIXjKii8RfQ9o8aFelPnRnshP1T3hsbYw1gussGw4t0m9ViuQfK8wCT2SjcriBtaP4mYHvCJuZWyi"
	structReq.TrackingURL = &trackingURL
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()

	unit := ts.createUnit(structReq.UnitID)
	ts.appUseCase.On("GetUnitByID", structReq.UnitID).Return(unit, nil)

	contentCampaign := &contentcampaign.ContentCampaign{ID: structReq.CampaignID}
	ts.contentCampaignUseCase.On("GetContentCampaignByID", structReq.CampaignID).Return(contentCampaign, nil).Once()
	ts.contentCampaignUseCase.On("IsContentCampaignExpired", contentCampaign).Return(false).Once()
	ts.contentCampaignUseCase.On("IncreaseClick", structReq.CampaignID, structReq.UnitID).Return(nil).Once()
	ts.deviceUseCase.On("SaveActivity", structReq.DeviceID, structReq.CampaignID, device.ActivityClick).Return(nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", structReq.GetUDT()).Return(true, nil).Once()
	ts.profileRequestUseCase.On("PopulateProfile", mock.AnythingOfType("profilerequest.Account")).Return(nil).Once()

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", structReq.PayloadStr).Return((*payload.Payload)(nil), errors.New("invalid payload")).Once()
	ts.trackingDataUseCase.On("ParseTrackingData", structReq.TrackingDataStr).Return(&trackingdata.TrackingData{ModelArtifact: "v1"}, nil).Once()

	err = ts.controller.ClickRedirect(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusFound, rec.Code)

	/* case specific flag */
	ts.True(trackEventCalled)
	/* case specific flag end */
}

func (ts *ControllerTestSuite) Test_ClickRedirect_SaveTrackURL() {
	trackEventCalled := false
	clientPatcher := ts.getBuzzAdMock(&trackEventCalled)
	defer clientPatcher.RemovePatch()
	core.Loggers["click"] = logrus.New()

	structReq := ts.buildBaseRequest()
	/* overwrite request parameters */
	structReq.PayloadStr = ""
	structReq.BaseReward = 0
	trackingURL := "https://ad.buzzvil.com/api/track_event?token=DC-2AG_LBRrgc_ttbRtyZJtL1ZkIR1EhQ1lGT815HVgOinjI_prlunpLEI-IW0xNW7bkdeKaCJrYrmqmbcpr3Bph01Y89cwt-YNa-znz-5K2Uu1BXwXCepF-e9dgf-0QSU3ICpRR-1JOUGb1KwkI-05BBX8a4MNz2HNqgE5ccgz4awhJNbJRsbf0dJxD_SwSlRBIpA9B7TvrT_Eq9kdjn7-qlWZtarlNI_HU5BvEpC-OpsWKNdvKIXjKii8RfQ9o8aFelPnRnshP1T3hsbYw1gussGw4t0m9ViuQfK8wCT2SjcriBtaP4mYHvCJuZWyi"
	structReq.TrackingURL = &trackingURL
	structReq.UseRewardAPI = true
	/* overwrite request parameters end */
	networkReq := ts.buildNetworkRequest(structReq)
	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())

	ts.rewardUseCase.On("ValidateRequest", mock.AnythingOfType("reward.RequestIngredients")).Return(nil).Once()

	unit := ts.createUnit(structReq.UnitID)
	ts.appUseCase.On("GetUnitByID", structReq.UnitID).Return(unit, nil)

	contentCampaign := &contentcampaign.ContentCampaign{ID: structReq.CampaignID}
	ts.contentCampaignUseCase.On("GetContentCampaignByID", structReq.CampaignID).Return(contentCampaign, nil).Once()
	ts.contentCampaignUseCase.On("IsContentCampaignExpired", contentCampaign).Return(false).Once()
	ts.contentCampaignUseCase.On("IncreaseClick", structReq.CampaignID, structReq.UnitID).Return(nil).Once()
	ts.deviceUseCase.On("SaveActivity", structReq.DeviceID, structReq.CampaignID, device.ActivityClick).Return(nil).Once()
	ts.deviceUseCase.On("ValidateUnitDeviceToken", structReq.GetUDT()).Return(true, nil).Once()
	ts.eventUseCase.On("SaveTrackingURL", structReq.DeviceID, ts.buildResource(structReq.CampaignID), ts.replaceInternalBAURL(trackingURL)).Once()

	payloadStruct := &payload.Payload{}
	err := faker.FakeData(&payloadStruct)
	ts.NoError(err)
	ts.payloadUseCase.On("ParsePayload", structReq.PayloadStr).Return((*payload.Payload)(nil), errors.New("invalid payload")).Once()
	ts.trackingDataUseCase.On("ParseTrackingData", structReq.TrackingDataStr).Return(&trackingdata.TrackingData{ModelArtifact: "v1"}, nil).Once()
	ts.profileRequestUseCase.On("PopulateProfile", mock.AnythingOfType("profilerequest.Account")).Return(nil).Once()

	err = ts.controller.ClickRedirect(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusFound, rec.Code)

	/* case specific flag */
	/* case specific flag end */
}

func (ts *ControllerTestSuite) buildResource(campaignID int64) event.Resource {
	if dto.BuzzAdCampaignIDOffset < campaignID {
		return event.Resource{
			ID:   campaignID - dto.BuzzAdCampaignIDOffset,
			Type: event.ResourceTypeAd,
		}
	}
	return event.Resource{
		ID:   campaignID,
		Type: event.ResourceTypeArticle,
	}
}

func (ts *ControllerTestSuite) replaceInternalBAURL(url string) string {
	// BA API 호출 시, Public Domain보다 Internal Domain이 비용을 아낄 수 있으니 Replace 해준다.

	url = strings.Replace(url, "https://api.buzzad.io/", ts.buzzAdURL, -1)
	return strings.Replace(url, "https://ad.buzzvil.com", ts.buzzAdURL, -1)
}

func (ts *ControllerTestSuite) createUnit(unitID int64) *app.Unit {
	return &app.Unit{
		ID:       unitID,
		Country:  "GB",
		IsActive: true,
	}
}

func (ts *ControllerTestSuite) getBuzzAdMock(trackEventCalled *bool) *gotestmock.ClientPatcher {
	httpClient := network.DefaultHTTPClient
	buzzAdServer := gotestmock.NewTargetServer(network.GetHost(ts.buzzAdURL)).AddResponseHandler(&gotestmock.ResponseHandler{
		WriteToBody: func() []byte {
			return []byte(`{"status": "ok"}`)
		},
		StatusCode: http.StatusMovedPermanently,
		Path:       "/api/v1/click",
		Method:     http.MethodGet,
	}).AddResponseHandler(&gotestmock.ResponseHandler{
		WriteToBody: func() []byte {
			// TODO SARI API가 호출되었다는 것을 좀 더 나이스하게 체크하는 방법 적용 필요
			if trackEventCalled != nil {
				*trackEventCalled = true
			}
			return []byte(`{"status": "ok"}`)
		},
		StatusCode: http.StatusOK,
		Path:       "/api/track_event",
		Method:     http.MethodGet,
	})

	return gotestmock.PatchClient(httpClient, buzzAdServer)
}

func (ts *ControllerTestSuite) buildNetworkRequest(req *dto.GetClickRedirectRequest) *network.Request {
	params := &url.Values{
		"unit_id":           {strconv.FormatInt(req.UnitID, 10)},
		"device_id":         {strconv.FormatInt(req.DeviceID, 10)},
		"ifa":               {req.IFA},
		"unit_device_token": {req.UnitDeviceToken},

		"reward":      {strconv.Itoa(req.Reward)},
		"base_reward": {strconv.Itoa(req.BaseReward)},

		"campaign_id":       {strconv.FormatInt(req.CampaignID, 10)},
		"campaign_type":     {req.CampaignType},
		"campaign_owner_id": {req.CampaignOwnerID},
		"campaign_is_media": {strconv.Itoa(req.CampaignIsMedia)},

		"redirect_url":     {req.RedirectURL},
		"slot":             {strconv.Itoa(req.Slot)},
		"position":         {req.Position},
		"session_id":       {req.SessionID},
		"campaign_name":    {req.CampaignName},
		"campaign_payload": {req.PayloadStr},
		"tracking_data":    {req.TrackingDataStr},
		"check":            {req.Checksum},
	}

	if req.UnitDeviceTokenClient != nil {
		(*params)["unit_device_token_client"] = []string{*req.UnitDeviceTokenClient}
	}

	if req.AppID != nil {
		(*params)["app_id"] = []string{strconv.FormatInt(*req.AppID, 10)}
	}

	if req.ExternalCampaignID != nil {
		(*params)["external_campaign_id"] = []string{*req.ExternalCampaignID}
	}
	if req.RedirectURLClean != nil {
		(*params)["redirect_url_clean"] = []string{*req.RedirectURLClean}
	}

	if req.TrackingURL != nil {
		(*params)["tracking_url"] = []string{*req.TrackingURL}
		(*params)["use_reward_api"] = []string{strconv.FormatBool(req.UseRewardAPI)}
	}

	networkReq := &network.Request{
		Method: http.MethodGet,
		URL:    "/api/click_redirect/",
		Params: params,
		Header: &http.Header{
			"User-Agent": {"Mozilla/5.0 (Linux; Android 4.2.1; en-us; Nexus 5 Build/JOP40D) AppleWebKit/535.19 (KHTML, like Gecko; googleweblight) Chrome/38.0.1025.166 Mobile Safari/535.19"},
		},
	}

	return networkReq.Build()
}

func (ts *ControllerTestSuite) buildBaseRequest() *dto.GetClickRedirectRequest {
	clickRedirectRequest := &dto.GetClickRedirectRequest{}

	err := faker.FakeData(&clickRedirectRequest.AppID)
	ts.NoError(err)
	err = faker.FakeData(&clickRedirectRequest.IFA)
	ts.NoError(err)

	clickRedirectRequest.DeviceID = int64(rand.Intn(1000) + 1)
	clickRedirectRequest.UnitID = int64(rand.Intn(1000) + 1000)
	clickRedirectRequest.CampaignID = int64(rand.Intn(1000) + 1000000)
	clickRedirectRequest.CampaignName = "CampaignName"

	clickRedirectRequest.Position = "__position__"
	clickRedirectRequest.SessionID = "__session_id__"
	clickRedirectRequest.Reward = 1

	clickRedirectRequest.UnitDeviceToken = "TEST_UNIT_DEVICE_TOKEN"
	ts.NoError(err)
	return clickRedirectRequest
}

type ControllerTestSuite struct {
	suite.Suite
	buzzAdURL              string
	controller             clickredirectsvc.Controller
	engine                 *core.Engine
	rewardUseCase          *mockRewardUseCase
	appUseCase             *mockAppUseCase
	contentCampaignUseCase *mockContentCampaignUseCase
	payloadUseCase         *mockPayloadUseCase
	trackingDataUseCase    *mockTrackingDataUseCase
	deviceUseCase          *mockDeviceUseCase
	eventUseCase           *mockEventUseCase
	profileRequestUseCase  *mockProfileRequestUseCase
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (ctx core.Context, rec *httptest.ResponseRecorder) {
	rec = httptest.NewRecorder()
	ctx = ts.engine.NewContext(httpRequest, rec)
	return
}

func (ts *ControllerTestSuite) SetupSuite() {
	tests.GetTestServer(nil)
	ts.engine = core.NewEngine(nil)
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.buzzAdURL = os.Getenv("BUZZAD_URL")
	ts.rewardUseCase = new(mockRewardUseCase)
	ts.appUseCase = new(mockAppUseCase)
	ts.contentCampaignUseCase = new(mockContentCampaignUseCase)
	ts.payloadUseCase = new(mockPayloadUseCase)
	ts.trackingDataUseCase = new(mockTrackingDataUseCase)
	ts.deviceUseCase = new(mockDeviceUseCase)
	ts.eventUseCase = new(mockEventUseCase)
	ts.profileRequestUseCase = new(mockProfileRequestUseCase)

	ts.controller = clickredirectsvc.NewController(
		ts.engine,
		ts.rewardUseCase,
		ts.appUseCase,
		ts.contentCampaignUseCase,
		ts.payloadUseCase,
		ts.trackingDataUseCase,
		ts.deviceUseCase,
		ts.eventUseCase,
		ts.profileRequestUseCase,
		ts.buzzAdURL,
	)
}

func (ts *ControllerTestSuite) AfterTest(_, _ string) {
	ts.rewardUseCase.AssertExpectations(ts.T())
	ts.appUseCase.AssertExpectations(ts.T())
	ts.contentCampaignUseCase.AssertExpectations(ts.T())
	ts.payloadUseCase.AssertExpectations(ts.T())
	ts.trackingDataUseCase.AssertExpectations(ts.T())
	ts.deviceUseCase.AssertExpectations(ts.T())
	ts.profileRequestUseCase.AssertExpectations(ts.T())
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

type mockEventUseCase struct {
	mock.Mock
}

func (u *mockEventUseCase) TrackEvent(handler event.MessageHandler, ingredients event.Token) error {
	ret := u.Called(handler, ingredients)
	return ret.Error(0)
}

func (u *mockEventUseCase) GetRewardStatus(token event.Token, auth header.Auth) (string, error) {
	ret := u.Called(token, auth)
	return ret.Get(0).(string), ret.Error(1)
}

func (u *mockEventUseCase) GetEventsMap(resources []event.Resource, unitID int64, auth header.Auth) (map[int64]event.Events, error) {
	ret := u.Called(resources, unitID, auth)
	return ret.Get(0).(map[int64]event.Events), ret.Error(1)
}

func (u *mockEventUseCase) GetTokenEncrypter() event.TokenEncrypter {
	ret := u.Called()
	return ret.Get(0).(event.TokenEncrypter)
}

func (u *mockEventUseCase) SaveTrackingURL(deviceID int64, resource event.Resource, trackingURL string) {
	u.Called(deviceID, resource, trackingURL)
}

func (u *mockEventUseCase) GetTrackingURL(deviceID int64, resource event.Resource) (string, error) {
	ret := u.Called(deviceID, resource)
	return ret.Get(0).(string), ret.Error(1)
}

type mockProfileRequestUseCase struct {
	mock.Mock
}

func (u *mockProfileRequestUseCase) PopulateProfile(account profilerequest.Account) error {
	ret := u.Called(account)
	return ret.Error(0)
}
