package repo

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	rewardsvc "github.com/Buzzvil/buzzapis/go/reward"
	"github.com/Buzzvil/buzzlib-go/header"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/event"
	"github.com/bxcodec/faker"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

func TestRepoSuite(t *testing.T) {
	suite.Run(t, new(RepoTestSuite))
}

func (ts *RepoTestSuite) TestSaveAndGetTrackingURL() {
	deviceID := rand.Int63n(1000000) + 1
	resource := event.Resource{
		ID:   rand.Int63n(1000000) + 1,
		Type: event.ResourceTypeAd,
	}
	trackURL := "https://tracking.url"
	cacheExpiration := time.Minute * 5

	cacheKey := fmt.Sprintf("CACHE_GO_TRACKINGURL-%v-%v-%v", deviceID, resource.ID, resource.Type)
	validateCacheKeyAndGetObject := func(key string, obj interface{}) error {
		ts.Equal(cacheKey, key)
		s := obj.(*string)
		*s = trackURL
		return nil
	}

	ts.cache.On("SetCacheAsync", cacheKey, trackURL, cacheExpiration).Once()
	ts.cache.On("GetCache", cacheKey, mock.AnythingOfType("*string")).Return(validateCacheKeyAndGetObject).Once()

	ts.repo.SaveTrackingURL(deviceID, resource, trackURL)
	result, err := ts.repo.GetTrackingURL(deviceID, resource)
	ts.NoError(err)
	ts.Equal(trackURL, result)
}

func (ts *RepoTestSuite) TestSaveAndDelteTrackingURL() {
	deviceID := rand.Int63n(1000000) + 1
	resource := event.Resource{
		ID:   rand.Int63n(1000000) + 1,
		Type: event.ResourceTypeAd,
	}
	trackURL := "https://tracking.url"
	cacheExpiration := time.Minute * 5
	cacheKey := fmt.Sprintf("CACHE_GO_TRACKINGURL-%v-%v-%v", deviceID, resource.ID, resource.Type)

	ts.cache.On("SetCacheAsync", cacheKey, trackURL, cacheExpiration).Once()
	ts.cache.On("DeleteCache", cacheKey).Return(nil).Once()

	ts.repo.SaveTrackingURL(deviceID, resource, trackURL)
	err := ts.repo.DeleteTrackingURL(deviceID, resource)
	ts.NoError(err)
}

func (ts *RepoTestSuite) TestGetRewardStatus() {
	resource := ts.createResource(rand.Intn(100000) + 1)
	unitID := rand.Int63n(1000000) + 1
	transactionID := ts.createTransactionID()
	token := ts.createToken(resource, unitID, transactionID)

	a := ts.createAuth()
	status := "pending"

	checkRewardStatusReq := &rewardsvc.CheckRewardStatusRequest{
		Resource:      ts.mapToProtoResource(resource),
		EventType:     token.EventType,
		TransactionId: transactionID,
	}
	checkRewardStatusRes := &rewardsvc.CheckRewardStatusResponse{
		Status: rewardsvc.RewardStatus(rewardsvc.RewardStatus_value[strings.ToUpper(status)]),
	}

	ts.client.On("CheckRewardStatus", mock.AnythingOfType("*context.valueCtx"), checkRewardStatusReq, mock.AnythingOfType("[]grpc.CallOption")).Return(checkRewardStatusRes, nil).Once()
	resStatus, err := ts.repo.GetRewardStatus(token, a)
	ts.NoError(err)
	ts.Equal(status, resStatus)
}

func (ts *RepoTestSuite) TestGetEventsMap() {
	totalResources := 10

	// create GetEventsMap request & response
	a := ts.createAuth()
	unitID := rand.Int63n(100000) + 1
	resources := make([]event.Resource, 0)
	eventsMap := make(map[int64]event.Events)
	for i := 0; i < totalResources; i++ {
		resource := ts.createResource(i)
		resources = append(resources, resource)
		eventsMap[resource.ID] = ts.createEvents(resource, ts.createReward(resource))
	}
	// create GetEventsMap request & response end

	// create request & response in mock function
	issueRewardsRequest := &rewardsvc.IssueRewardsRequest{}
	issueRewardsResponse := &rewardsvc.IssueRewardsResponse{
		RewardsForResourceMap: make(map[int64]*rewardsvc.IssueRewardsResponse_RewardsForResourceTypeMap),
	}
	for _, resource := range resources {
		pResourceType := rewardsvc.ResourceType(rewardsvc.ResourceType_value[strings.ToUpper(resource.Type)])

		issueRewardsRequest.Resources = append(issueRewardsRequest.Resources, &rewardsvc.Resource{
			Id:   resource.ID,
			Type: pResourceType,
		})

		_, ok := issueRewardsResponse.RewardsForResourceMap[resource.ID]
		if !ok {
			issueRewardsResponse.RewardsForResourceMap[resource.ID] = &rewardsvc.IssueRewardsResponse_RewardsForResourceTypeMap{
				RewardsForResourceTypeMap: make(map[int32]*rewardsvc.Rewards),
			}
		}

		rewards := ts.createProtoRewards(resource, eventsMap[resource.ID])
		if rewards != nil {
			issueRewardsResponse.RewardsForResourceMap[resource.ID].RewardsForResourceTypeMap[int32(pResourceType)] = rewards
		}
	}
	tokenStr := "TEST_TOKEN_STRING"
	// create request & response in mock function end

	matchToken := mock.MatchedBy(func(token event.Token) bool {
		for _, resource := range resources {
			if resource == token.Resource {
				return true
			}
		}
		return false
	})
	ts.tokenEncrypter.On("Build", matchToken).Return(tokenStr, nil)
	ts.client.On("IssueRewards", mock.AnythingOfType("*context.valueCtx"), issueRewardsRequest, mock.AnythingOfType("[]grpc.CallOption")).Return(issueRewardsResponse, nil).Once()

	resultEventsMap, err := ts.repo.GetEventsMap(resources, unitID, a, ts.tokenEncrypter)
	ts.NoError(err)
	ts.equalEventsMap(eventsMap, resultEventsMap)
}

func (ts *RepoTestSuite) TestGetEventsMapEmptyReward() {
	totalResources := 10

	// create GetEventsMap request & response
	a := ts.createAuth()
	unitID := rand.Int63n(100000) + 1
	resources := make([]event.Resource, 0)
	eventsMap := make(map[int64]event.Events)
	for i := 0; i < totalResources; i++ {
		resource := ts.createResource(i)
		resources = append(resources, resource)
	}
	// create GetEventsMap request & response end

	// create request & response in mock function
	issueRewardsRequest := &rewardsvc.IssueRewardsRequest{}
	issueRewardsResponse := &rewardsvc.IssueRewardsResponse{
		RewardsForResourceMap: nil,
	}
	for _, resource := range resources {
		issueRewardsRequest.Resources = append(issueRewardsRequest.Resources, &rewardsvc.Resource{
			Id:   resource.ID,
			Type: rewardsvc.ResourceType(rewardsvc.ResourceType_value[strings.ToUpper(resource.Type)]),
		})
	}
	// create request & response in mock function end

	ts.client.On("IssueRewards", mock.AnythingOfType("*context.valueCtx"), issueRewardsRequest, mock.AnythingOfType("[]grpc.CallOption")).Return(issueRewardsResponse, nil).Once()

	resultEventsMap, err := ts.repo.GetEventsMap(resources, unitID, a, ts.tokenEncrypter)
	ts.NoError(err)
	ts.equalEventsMap(eventsMap, resultEventsMap)
}

func (ts *RepoTestSuite) equalIngredients(expected event.Token, actual event.Token) {
	ts.Equal(expected.UnitID, actual.UnitID)
	ts.Equal(expected.EventType, actual.EventType)
	ts.Equal(expected.TransactionID, actual.TransactionID)
	ts.equalResource(expected.Resource, actual.Resource)
}

func (ts *RepoTestSuite) equalResource(expected event.Resource, actual event.Resource) {
	ts.Equal(expected.ID, actual.ID)
	ts.Equal(expected.Type, actual.Type)
}

func (ts *RepoTestSuite) equalEventsMap(expected map[int64]event.Events, actual map[int64]event.Events) {
	ts.NotNil(actual)
	ts.Equal(len(expected), len(actual))
	for k, v := range expected {
		e, ok := actual[k]
		ts.True(ok)
		ts.equalEvents(v, e)
	}
}

func (ts *RepoTestSuite) equalEvents(expectedEvents event.Events, actualEvents event.Events) {
	ts.NotNil(actualEvents)
	ts.Equal(len(expectedEvents), len(actualEvents))
	for _, expectedEvent := range expectedEvents {
		for _, actualEvent := range actualEvents {
			if expectedEvent.Type == actualEvent.Type {
				ts.equalEvent(expectedEvent, actualEvent)
			}
		}
	}
}

func (ts *RepoTestSuite) equalEvent(expected event.Event, actual event.Event) {
	ts.Equal(expected.Type, actual.Type)
	ts.Equal(len(expected.TrackingURLs), len(actual.TrackingURLs))
	for _, actualTrackingURL := range actual.TrackingURLs {
		if strings.HasPrefix(actualTrackingURL, "https://ad.buzzvil.com/") {
			ts.assertContainTrackingURL(expected.TrackingURLs, actualTrackingURL)
		}
	}

	if expected.Reward != nil {
		ts.NotNil(actual.Reward)
		ts.equalReward(*(expected.Reward), *(actual.Reward))
	} else {
		ts.Nil(actual.Reward)
	}
}

func (ts *RepoTestSuite) assertContainTrackingURL(trackingURLs []string, needleTrackingURL string) {
	for _, trackingURL := range trackingURLs {
		if trackingURL == needleTrackingURL {
			return
		}
	}
}

func (ts *RepoTestSuite) equalReward(expected event.Reward, actual event.Reward) {
	ts.Equal(expected.Amount, actual.Amount)
	ts.Equal(expected.Status, actual.Status)
	ts.Equal(expected.IssueMethod, actual.IssueMethod)
	ts.Equal(expected.TTL, actual.TTL)
}

func (ts *RepoTestSuite) mapToProtoResource(resource event.Resource) *rewardsvc.Resource {
	return &rewardsvc.Resource{
		Id:   resource.ID,
		Type: rewardsvc.ResourceType(rewardsvc.ResourceType_value[strings.ToUpper(resource.Type)]),
	}
}

func (ts *RepoTestSuite) createToken(resource event.Resource, unitID int64, transactionID string) event.Token {
	return event.Token{
		Resource:      resource,
		EventType:     "landed",
		UnitID:        unitID,
		TransactionID: transactionID,
	}
}

func (ts *RepoTestSuite) createTransactionID() string {
	return uuid.NewV4().String()
}

func (ts *RepoTestSuite) createProtoRewards(resource event.Resource, events event.Events) *rewardsvc.Rewards {
	resRewards := &rewardsvc.Rewards{}

	for _, e := range events {
		if e.Reward == nil {
			continue
		}
		resReward := &rewardsvc.Reward{
			Resource: &rewardsvc.Resource{
				Id:   resource.ID,
				Type: rewardsvc.ResourceType(rewardsvc.ResourceType_value[strings.ToUpper(resource.Type)]),
			},
			EventType:     e.Type,
			Amount:        e.Reward.Amount,
			IssueMethod:   e.Reward.IssueMethod,
			Ttl:           e.Reward.TTL,
			Status:        rewardsvc.RewardStatus(rewardsvc.RewardStatus_value[strings.ToUpper(e.Reward.Status)]),
			TransactionId: ts.createTransactionID(),
		}

		resRewards.Rewards = append(resRewards.Rewards, resReward)
	}

	return resRewards
}

func (ts *RepoTestSuite) createReward(resource event.Resource) *event.Reward {
	return &event.Reward{
		Amount:      rand.Int63n(10) + 1,
		Status:      "receivable",
		IssueMethod: "static",
		TTL:         rand.Int63n(100000) + 1,
	}
}

func (ts *RepoTestSuite) createEvents(resource event.Resource, r *event.Reward) event.Events {
	var events event.Events
	eventTypes := []string{event.TypeClicked, event.TypeImpressed}

	for _, t := range eventTypes {
		e := event.Event{
			Type:         t,
			Reward:       r,
			TrackingURLs: []string{"http://buzzscreen-test.buzzvil.com/track_event"},
		}

		events = append(events, e)
	}

	return events
}

func (ts *RepoTestSuite) createResource(i int) event.Resource {
	resource := event.Resource{
		ID:   rand.Int63n(100) + int64(i*100),
		Type: event.ResourceTypeAd,
	}

	return resource
}

func (ts *RepoTestSuite) createAuth() header.Auth {
	a := header.Auth{}
	ts.NoError(faker.FakeData(&a))
	return a
}

type RepoTestSuite struct {
	suite.Suite
	client         *mockRewardServiceClient
	tokenEncrypter *mockTokenEncrypter
	cache          *mockCache
	repo           event.Repository
}

func (ts *RepoTestSuite) SetupTest() {
	ts.client = &mockRewardServiceClient{}
	ts.tokenEncrypter = &mockTokenEncrypter{}
	ts.cache = &mockCache{}
	ts.repo = New(ts.client, "localhost", ts.cache)
}

func (ts *RepoTestSuite) AfterTest(_, _ string) {
	ts.client.AssertExpectations(ts.T())
	ts.tokenEncrypter.AssertExpectations(ts.T())
	ts.cache.AssertExpectations(ts.T())
}

type mockRewardServiceClient struct {
	mock.Mock
}

func (c *mockRewardServiceClient) IssueRewards(ctx context.Context, in *rewardsvc.IssueRewardsRequest, opts ...grpc.CallOption) (*rewardsvc.IssueRewardsResponse, error) {
	args := c.Called(ctx, in, opts)
	return args.Get(0).(*rewardsvc.IssueRewardsResponse), args.Error(1)
}

func (c *mockRewardServiceClient) CheckRewardStatus(ctx context.Context, in *rewardsvc.CheckRewardStatusRequest, opts ...grpc.CallOption) (*rewardsvc.CheckRewardStatusResponse, error) {
	args := c.Called(ctx, in, opts)
	return args.Get(0).(*rewardsvc.CheckRewardStatusResponse), args.Error(1)
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

type mockCache struct {
	mock.Mock
}

func (m *mockCache) SetCacheAsync(key string, obj interface{}, expiration time.Duration) {
	m.Called(key, obj, expiration)
}

func (m *mockCache) SetCache(key string, obj interface{}, expiration time.Duration) error {
	ret := m.Called(key, obj, expiration)
	return ret.Error(0)
}

func (m *mockCache) GetCache(key string, obj interface{}) error {
	ret := m.Called(key, obj)
	f, ok := ret.Get(0).(func(string, interface{}) error)
	if ok {
		return f(key, obj)
	}
	return ret.Error(0)
}

func (m *mockCache) DeleteCache(key string) error {
	ret := m.Called(key)
	return ret.Error(0)
}
