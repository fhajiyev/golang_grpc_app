package eventsvc_test

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzlib-go/network"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/eventsvc"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/eventsvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/eventsvc/publisher"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/ad"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/auth"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/contentcampaign"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/device"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/bxcodec/faker"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (ts *ControllerTestSuite) Test_GetRewardStatus() {
	resource := ts.createResource("ad")
	token := ts.createToken(*resource)
	tokenStr := "TEST_TOKEN_STRING"
	a := ts.createAuth()
	req := &dto.RewardStatusRequest{TokenStr: tokenStr}
	status := "pending"

	ts.eventUseCase.On("GetRewardStatus", *token, *a).Return(status, nil).Once()
	ts.eventUseCase.On("GetTokenEncrypter").Return(ts.tokenEncrypter).Once()
	ts.tokenEncrypter.On("Parse", tokenStr).Return(token, nil).Once()
	networkReq := (&network.Request{
		Header: ts.createHeader(a),
		Method: http.MethodGet,
		URL:    "/api/reward-status",
		Params: &url.Values{"token": {req.TokenStr}},
	}).Build()

	ctx, rec := ts.buildContextAndRecorder(networkReq.GetHTTPRequest())
	err := ts.controller.GetRewardStatus(ctx)
	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)

	res := make(map[string]string)
	err = json.Unmarshal(rec.Body.Bytes(), &res)
	ts.NoError(err)
	ts.Equal(res["reward_status"], status)
}

func (ts *ControllerTestSuite) Test_TrackEvent_Article() {
	tokenStr := "TEST_TOKEN_STRING"
	resource := ts.createResource(event.ResourceTypeArticle)
	article := ts.createContentCampaign(resource.ID)
	token := ts.createToken(*resource)
	a := ts.createAuth()

	ts.contentCampaignUseCase.On("GetContentCampaignByID", resource.ID).Return(article, nil).Once()
	ts.eventUseCase.On("TrackEvent", mock.AnythingOfType("*publisher.handler"), *token).Return(nil).Once()
	ts.eventUseCase.On("GetTokenEncrypter").Return(ts.tokenEncrypter).Once()
	ts.tokenEncrypter.On("Parse", tokenStr).Return(token, nil).Once()

	httpReq := ts.buildTrackEventHTTPRequest(tokenStr, a)
	ctx, rec := ts.buildContextAndRecorder(httpReq)
	err := ts.controller.TrackEvent(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
}

func (ts *ControllerTestSuite) Test_TrackEvent_Ad() {
	tokenStr := "TEST_TOKEN_STRING"
	resource := ts.createResource(event.ResourceTypeAd)
	detail := ts.createAdDetail(resource.ID)
	token := ts.createToken(*resource)
	a := ts.createAuth()

	ts.adUseCase.On("GetAdDetail", token.Resource.ID).Return(detail, nil).Once()
	ts.eventUseCase.On("TrackEvent", mock.AnythingOfType("*publisher.handler"), *token).Return(nil).Once()
	ts.eventUseCase.On("GetTokenEncrypter").Return(ts.tokenEncrypter).Once()
	ts.tokenEncrypter.On("Parse", tokenStr).Return(token, nil).Once()

	httpReq := ts.buildTrackEventHTTPRequest(tokenStr, a)
	ctx, rec := ts.buildContextAndRecorder(httpReq)
	err := ts.controller.TrackEvent(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)
}

func (ts *ControllerTestSuite) buildTrackEventHTTPRequest(token string, a *header.Auth) *http.Request {
	req := &dto.TrackEventRequest{TokenStr: token}
	networkReq := (&network.Request{
		Header: ts.createHeader(a),
		Method: http.MethodGet,
		URL:    "/api/track-event",
		Params: &url.Values{"token": {req.TokenStr}},
	}).Build()
	return networkReq.GetHTTPRequest()
}

func (ts *ControllerTestSuite) createContentCampaign(id int64) *contentcampaign.ContentCampaign {
	return &contentcampaign.ContentCampaign{
		ID:             id,
		Name:           "TEST_ARTICLE_NAME",
		OrganizationID: rand.Int63n(1000000) + 1,
		ExtraData: &map[string]interface{}{
			"unit": map[string]interface{}{
				"id": rand.Int63n(1000000) + 1,
			},
		},
	}
}

func (ts *ControllerTestSuite) createAdDetail(id int64) *ad.Detail {
	return &ad.Detail{
		ID:             id,
		Name:           "TEST_AD_NAME",
		RevenueType:    "cpc",
		OrganizationID: rand.Int63n(100000) + 1,
		Extra: map[string]interface{}{
			"extra_data": map[string]interface{}{
				"id:": rand.Int63n(100000) + 1,
			},
		},
	}
}

func (ts *ControllerTestSuite) createToken(resource event.Resource) *event.Token {
	return &event.Token{
		TransactionID: uuid.NewV4().String(),
		Resource:      resource,
		EventType:     "landed",
		UnitID:        rand.Int63n(100000) + 1,
	}
}

func (ts *ControllerTestSuite) createResourceDataByAd(resource event.Resource, detail ad.Detail) publisher.ResourceData {
	resourceData := publisher.ResourceData{
		ID:             resource.ID,
		Name:           detail.Name,
		OrganizationID: detail.OrganizationID,
		RevenueType:    detail.RevenueType,
	}
	unitExtra, ok := detail.Extra["extra_data"].(map[string]interface{})
	if ok {
		resourceData.Extra = unitExtra
	}
	return resourceData
}

func (ts *ControllerTestSuite) createResourceDataByArticle(resource event.Resource, campaign contentcampaign.ContentCampaign) publisher.ResourceData {
	resourceData := publisher.ResourceData{
		ID:             resource.ID,
		Name:           campaign.Name,
		OrganizationID: campaign.OrganizationID,
		RevenueType:    "",
	}
	unitExtra, ok := (*campaign.ExtraData)["unit"].(map[string]interface{})
	if ok {
		resourceData.Extra = unitExtra
	}
	return resourceData
}

func (ts *ControllerTestSuite) createResource(resourceType string) *event.Resource {
	name := "TEST_RESOURCE_NAME"
	return &event.Resource{
		ID:   rand.Int63n(100000) + 1,
		Type: resourceType,
		Name: &name,
	}
}

func (ts *ControllerTestSuite) createHeader(a *header.Auth) *http.Header {
	return &http.Header{
		"Buzz-App-Id":            []string{strconv.FormatInt(a.AppID, 10)},
		"Buzz-Account-Id":        []string{strconv.FormatInt(a.AccountID, 10)},
		"Buzz-Publisher-User-Id": []string{a.PublisherUserID},
		"Buzz-Ifa":               []string{a.IFA},
	}
}

func (ts *ControllerTestSuite) createAuth() *header.Auth {
	a := &header.Auth{}
	ts.NoError(faker.FakeData(&a))
	return a
}

func (ts *ControllerTestSuite) Test_PostEventV1() {
	mockUnit := &app.Unit{
		ID:       1,
		AppID:    1234,
		UnitType: app.UnitTypeLockscreen,
	}
	ts.appUseCase.On("GetUnitByAppIDAndType", mockUnit.AppID, mockUnit.UnitType).Return(mockUnit, nil).Once()

	req := (&network.Request{
		Method: http.MethodGet,
		URL:    "/api/event/",
		Params: &url.Values{
			"app_id":           {"1234"},
			"carrier":          {"SKT"},
			"device_name":      {"device"},
			"device_os":        {"21"},
			"device_timestamp": {"123412134"},
			"event_name":       {"test_event"},
			"ifa":              {"ifaifaifaifa"},
			"package":          {"com.honeyscreen"},
			"resolution":       {"1280x720"},
			"sdk_version":      {"10101"},
		},
	}).Build()
	ctx, rec := ts.buildContextAndRecorder(req.GetHTTPRequest())
	err := ts.controller.PostEventV1(ctx)

	ts.NoError(err)
	ts.Equal(http.StatusOK, rec.Code)

	var res map[string]interface{}

	err = json.Unmarshal([]byte(rec.Body.String()), &res)
	ts.NoError(err)
	ts.Equal(res["gudid"].(string) != "", true)
	ts.Equal(res["period"].(float64) > 0, true)

	ts.appUseCase.AssertExpectations(ts.T())
}

type ControllerTestSuite struct {
	suite.Suite
	controller             eventsvc.Controller
	engine                 *core.Engine
	appUseCase             *mockAppUseCase
	eventUseCase           *mockEventUseCase
	tokenEncrypter         *mockTokenEncrypter
	contentCampaignUseCase *mockContentCampaignUseCase
	adUseCase              *mockAdUsecase
	deviceUseCase          *mockDeviceUseCase
	publisher              *mockMQPublisher
}

func (ts *ControllerTestSuite) SetupTest() {
	ts.engine = core.NewEngine(nil)
	ts.appUseCase = new(mockAppUseCase)
	ts.eventUseCase = new(mockEventUseCase)
	ts.tokenEncrypter = new(mockTokenEncrypter)
	ts.contentCampaignUseCase = new(mockContentCampaignUseCase)
	ts.adUseCase = new(mockAdUsecase)
	ts.publisher = new(mockMQPublisher)
	ts.deviceUseCase = new(mockDeviceUseCase)
	ts.controller = eventsvc.NewController(ts.engine, ts.appUseCase, new(mockAuthUseCase), ts.deviceUseCase, ts.eventUseCase, ts.contentCampaignUseCase, ts.adUseCase, ts.publisher)
}

func (ts *ControllerTestSuite) AfterTest(_, _ string) {
	ts.appUseCase.AssertExpectations(ts.T())
	ts.eventUseCase.AssertExpectations(ts.T())
	ts.contentCampaignUseCase.AssertExpectations(ts.T())
	ts.adUseCase.AssertExpectations(ts.T())
	ts.publisher.AssertExpectations(ts.T())
}

func (ts *ControllerTestSuite) buildContextAndRecorder(httpRequest *http.Request) (ctx core.Context, rec *httptest.ResponseRecorder) {
	rec = httptest.NewRecorder()
	ctx = ts.engine.NewContext(httpRequest, rec)
	return
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

type mockAuthUseCase struct {
	mock.Mock
}

func (u *mockAuthUseCase) CreateAuth(identifier auth.Identifier) (string, error) {
	args := u.Called(identifier)
	return args.Get(0).(string), args.Error(1)
}

func (u *mockAuthUseCase) GetAuth(token string) (*auth.Auth, error) {
	args := u.Called(token)
	return args.Get(0).(*auth.Auth), args.Error(1)
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

type mockTokenEncrypter struct {
	mock.Mock
}

func (m *mockTokenEncrypter) Build(token event.Token) (string, error) {
	args := m.Called(token)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockTokenEncrypter) Parse(tokenStr string) (*event.Token, error) {
	args := m.Called(tokenStr)
	return args.Get(0).(*event.Token), args.Error(1)
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

type mockAdUsecase struct {
	mock.Mock
}

func (u *mockAdUsecase) GetAdDetail(id int64) (*ad.Detail, error) {
	ret := u.Called(id)
	return ret.Get(0).(*ad.Detail), ret.Error(1)
}

func (r *mockAdUsecase) GetAdsV1(v1Req ad.V1AdsRequest) (*ad.V1AdsResponse, error) {
	ret := r.Called(v1Req)
	return ret.Get(0).(*ad.V1AdsResponse), ret.Error(1)
}

func (r *mockAdUsecase) GetAdsV2(v2Req ad.V2AdsRequest) (*ad.V2AdsResponse, error) {
	ret := r.Called(v2Req)
	return ret.Get(0).(*ad.V2AdsResponse), ret.Error(1)
}

func (r *mockAdUsecase) LogAdAllocationRequestV1(appID int64, v1Req ad.V1AdsRequest) {
	r.Called(appID, v1Req)
}

func (r *mockAdUsecase) LogAdAllocationRequestV2(appID int64, v2Req ad.V2AdsRequest) {
	r.Called(appID, v2Req)
}

type mockMQPublisher struct {
	mock.Mock
}

func (p *mockMQPublisher) Push(publishing amqp.Publishing, routingKey string) error {
	ret := p.Called(publishing, routingKey)
	return ret.Error(0)
}

func (p *mockMQPublisher) Close() error {
	ret := p.Called()
	return ret.Error(0)
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
